package seeker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
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
}

type Config struct{}

func New(shutdownCtx context.Context, cfg Config, handlers ahandlers.Handlers, taskService services.TaskService, plugins []plugin.Plugin) (*Seeker, error) {
	if len(plugins) == 0 {
		return nil, errors.New("there should be at least one plugin provided")
	}

	return &Seeker{
		cfg,
		plugins,
		&handlers,
		taskService,
	}, nil
}

func (s *Seeker) ReadTasks(pluginName string, n uint) ([]queue.Task, error) {
	queueTasks := make([]queue.Task, 0, n)
	tasks, err := s.taskService.GetNEarliestTasks(n)
	if err != nil {
		return nil, err
	}

	for _, task := range tasks {
		queueTasks = append(queueTasks, queue.Task{
			Id:   task.Id,
			Body: task.Body,
		})
	}

	return queueTasks, nil
}

func (s *Seeker) startQueues() {
	for _, _p := range s.plugins {
		fmt.Println("run plugin")
		_q := queue.New(_p.GetQueueConfig())

		consumeTaskCh := _q.StartConsuming(context.Background())

		go func(p plugin.Plugin, q *queue.Queue) {
			for {
				tasks, err := s.ReadTasks(p.GetName(), p.GetQueueConfig().QueueSize)
				if err != nil {
					// add logging?
					continue
				}

				if len(tasks) == 0 {
					time.Sleep(5 * time.Second)
					continue
				}

				for _, t := range tasks {
					_q.Publish(t)

					err := s.taskService.DeleteTaskById(t.Id)
					if err != nil {
						// add logging
					}
				}
			}
		}(_p, &_q)

		go func(p plugin.Plugin, consumeTaskCh <-chan queue.Task) {
			for task := range consumeTaskCh {
				var request plugin.Request

				err := json.Unmarshal(task.Body, &request)
				if err != nil {
					// Add error handling
					continue
				}

				_, err = p.DoRequest(request)
				if err != nil {
					// Add error handling
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

	srv := createHTTPServer(s.handlers)

	return srv.ListenAndServe()
}
