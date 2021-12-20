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
	errorLogger *log.Logger
}

type Config struct{}

func New(shutdownCtx context.Context, cfg Config, handlers ahandlers.Handlers, taskService services.TaskService, plugins []aplugin.Plugin) (*Seeker, error) {
	if len(plugins) == 0 {
		return nil, errors.New("there should be at least one plugin provided")
	}

	errorLogger := log.New(os.Stdout, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	return &Seeker{
		cfg,
		plugins,
		&handlers,
		taskService,
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

				if s.taskService.ShouldSkipTask(task) {
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
					if connection.SourceUrl == connection.DestUrl {
						foundConnection = &connection
						break
					}
				}

				if foundConnection != nil {
					fmt.Println("Success found path:", task.Id, foundConnection)
					err = s.taskService.DeleteAllTasksWithOrigin(task.OriginTaskId)
					if err != nil {
						s.errorLogger.Printf("DeleteAllTasksWithOrigin. Plugin: %s; Error: %s\n", p.GetName(), err)
					}
					continue
				}

				for _, connection := range response.Connections {
					s.taskService.CreateNewTask(
						s.taskService.GenerateId(connection.SourceUrl, connection.DestUrl),
						task.OriginTaskId,
						connection.SourceUrl,
						connection.DestUrl,
						connection.Cursor,
						task.RequestsCount,
					)
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
