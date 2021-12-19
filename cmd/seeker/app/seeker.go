package seeker

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	ahandlers "github.com/malcolmmadsheep/handshakes-seeker/pkg/handlers"
	"github.com/malcolmmadsheep/handshakes-seeker/pkg/plugin"
	"github.com/malcolmmadsheep/handshakes-seeker/pkg/queue"
	"github.com/malcolmmadsheep/handshakes-seeker/pkg/services"
)

type Seeker struct {
	cfg         Config
	plugins     []plugin.Plugin
	handlers    *ahandlers.Handlers
	taskService services.TaskService
	errorLogger *log.Logger
}

type Config struct{}

func New(shutdownCtx context.Context, cfg Config, handlers ahandlers.Handlers, taskService services.TaskService, plugins []plugin.Plugin) (*Seeker, error) {
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

func taskToQueueTask(task *services.Task) (queue.Task, error) {
	queueTask, err := json.Marshal(task)
	if err != nil {
		return nil, err
	}

	return queueTask, nil
}

func queueTaskToTask(queueTask queue.Task) (*services.Task, error) {
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
	for _, _p := range s.plugins {
		_q := queue.New(_p.GetQueueConfig())

		consumeTaskCh := _q.StartConsuming(context.Background())

		go func(p plugin.Plugin, q *queue.Queue) {
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
					_q.Publish(queueTask)

					err = s.taskService.DeleteTaskById(task.Id)
					if err != nil {
						s.errorLogger.Printf("DeleteTaskById: %s; %s\n", task.Id, err)
					}
				}
			}
		}(_p, &_q)

		go func(p plugin.Plugin, consumeTaskCh <-chan queue.Task) {
			for queueTask := range consumeTaskCh {
				task, err := queueTaskToTask(queueTask)
				if err != nil {
					s.errorLogger.Printf("queueTaskToTask: %s\n", err)
					continue
				}
				request := plugin.Request{
					SourceUrl: task.SourceUrl,
					DestUrl:   task.DestUrl,
					Cursor:    task.Cursor,
				}

				_, err = p.DoRequest(request)
				if err != nil {
					s.errorLogger.Printf("DoRequest. Plugin: %s; Error: %s\n", p.GetName(), err)
					continue
				}

				// add proper res processing
			}
		}(_p, consumeTaskCh)
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
