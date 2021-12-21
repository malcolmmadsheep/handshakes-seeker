package seeker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4"
	ahandlers "github.com/malcolmmadsheep/handshakes-seeker/pkg/handlers"
	aplugin "github.com/malcolmmadsheep/handshakes-seeker/pkg/plugin"
	aqueue "github.com/malcolmmadsheep/handshakes-seeker/pkg/queue"
	"github.com/malcolmmadsheep/handshakes-seeker/pkg/services"
)

type Seeker struct {
	cfg         Config
	plugins     []aplugin.Plugin
	handlers    *ahandlers.Handlers
	taskService services.TaskService
	pathService services.PathService
	errorLogger *log.Logger
}

type Config struct{}

func New(
	shutdownCtx context.Context,
	cfg Config,
	handlers ahandlers.Handlers,
	taskService services.TaskService,
	pathService services.PathService,
	plugins []aplugin.Plugin,
) (*Seeker, error) {
	if len(plugins) == 0 {
		return nil, errors.New("there should be at least one plugin provided")
	}

	errorLogger := log.New(os.Stdout, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	return &Seeker{
		cfg,
		plugins,
		&handlers,
		taskService,
		pathService,
		errorLogger,
	}, nil
}

func taskToQueueTask(task *services.Task) (aqueue.Task, error) {
	queueTask, err := json.Marshal(task)
	if err != nil {
		return nil, err
	}

	return queueTask, nil
}

func queueTaskToTask(queueTask aqueue.Task) (*services.Task, error) {
	var task services.Task

	err := json.Unmarshal(queueTask, &task)
	if err != nil {
		return nil, err
	}

	return &task, nil
}

func (s *Seeker) GetTasks(pluginName string, n uint) ([]*services.Task, error) {
	return s.taskService.GetNEarliestTasks(n)
}

func (s *Seeker) shouldSkipTask(task *services.Task) bool {
	if s.taskService.ShouldSkipTask(task) {
		return true
	}

	path, err := s.pathService.GetPathByTaskId(task.OriginTaskId)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false
		}
		s.errorLogger.Println("seeker:s.pathService.GetPathByTaskId:", err)
		return false
	}

	return path.Status == services.PathStatusFound.String() || path.Status == services.PathStatusNotFound.String()
}

func (s *Seeker) startQueues() {
	for _, plugin := range s.plugins {
		queue := aqueue.New(plugin.GetQueueConfig())

		consumeTaskCh := queue.StartConsuming(context.Background())
		go func(p aplugin.Plugin, q *aqueue.Queue) {
			for {
				tasks, err := s.GetTasks(p.GetName(), p.GetQueueConfig().QueueSize)
				if err != nil {
					s.errorLogger.Printf("GetTasks: %s\n", err)
					continue
				}

				if len(tasks) == 0 {
					time.Sleep(5 * time.Second)
					continue
				}

				for _, task := range tasks {
					queueTask, err := taskToQueueTask(task)
					if err != nil {
						s.errorLogger.Printf("taskToQueueTask: %s\n", err)
						continue
					}
					queue.Publish(queueTask)

					err = s.taskService.DeleteTaskByIds(task.Id, task.OriginTaskId)
					if err != nil {
						s.errorLogger.Printf("DeleteTaskById: %s; %s\n", task.Id, err)
					}
				}
			}
		}(plugin, &queue)

		go func(p aplugin.Plugin, consumeTaskCh <-chan aqueue.Task) {
			for queueTask := range consumeTaskCh {
				task, err := queueTaskToTask(queueTask)
				if err != nil {
					s.errorLogger.Printf("queueTaskToTask: %s\n", err)
					continue
				}

				if s.shouldSkipTask(task) {
					continue
				}

				request := aplugin.Request{
					SourceUrl: task.SourceUrl,
					DestUrl:   task.DestUrl,
					Cursor:    task.Cursor,
				}

				response, err := p.DoRequest(request)
				if err != nil {
					s.errorLogger.Printf("DoRequest. Plugin: %s; Error: %s\n", p.GetName(), err)
					continue
				}

				var foundConnection *aplugin.Connection

				for _, connection := range response.Connections {
					_, err := s.taskService.CreateNewTask(
						s.taskService.GenerateId(connection.SourceUrl, connection.DestUrl),
						task.OriginTaskId,
						connection.SourceUrl,
						connection.DestUrl,
						connection.Cursor,
						task.RequestsCount,
					)
					if err != nil {
						s.errorLogger.Printf("s.taskService.CreateNewTask. Plugin: %s; Error: %s\n", p.GetName(), err)
						continue
					}

					_, err = s.pathService.CreateFoundPath(
						s.taskService.GenerateId(task.SourceUrl, connection.SourceUrl),
						task.SourceUrl,
						connection.SourceUrl,
						fmt.Sprintf("%s,%s", task.SourceUrl, connection.SourceUrl),
					)
					if err != nil {
						s.errorLogger.Printf("s.pathService.CreateFoundPath. Plugin: %s; Error: %s\n", p.GetName(), err)
						continue
					}

					if connection.SourceUrl == connection.DestUrl {
						foundConnection = &connection
						break
					}
				}

				if foundConnection != nil {
					fmt.Println("Success found path:", task.Id, foundConnection)
					err = s.taskService.DeleteAllTasksWithOrigin(task.OriginTaskId)
					if err != nil {
						s.errorLogger.Printf("taskService.DeleteAllTasksWithOrigin. Plugin: %s; Error: %s\n", p.GetName(), err)
					}

					err = s.pathService.UpdatePathStatusByTaskId(task.OriginTaskId, services.PathStatusFound)
					if err != nil {
						s.errorLogger.Printf("pathService.UpdatePathStatusByTaskId. Plugin: %s; Error: %s\n", p.GetName(), err)
					}
					continue
				}
			}
		}(plugin, consumeTaskCh)
	}
}

func createHTTPServer(handlers *ahandlers.Handlers) *http.Server {
	apiRouter := mux.NewRouter().StrictSlash(false).PathPrefix("/api/v1").Subrouter()

	apiRouter.HandleFunc("/task", (*handlers).CreateTask).Methods(http.MethodPost)

	taskSubrouter := apiRouter.PathPrefix("/task/{taskId}").Subrouter()

	taskSubrouter.HandleFunc("", (*handlers).GetPath).Methods(http.MethodGet)
	taskSubrouter.HandleFunc("", (*handlers).DeleteTask).Methods(http.MethodDelete)

	return &http.Server{
		Handler:      apiRouter,
		Addr:         ":8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
}

func (s *Seeker) Run() error {
	s.startQueues()

	server := createHTTPServer(s.handlers)

	return server.ListenAndServe()
}
