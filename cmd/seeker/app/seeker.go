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
)

type Seeker struct {
	cfg      Config
	plugins  []plugin.Plugin
	handlers *ahandlers.Handlers
}

type Config struct{}

func New(shutdownCtx context.Context, cfg Config, handlers ahandlers.Handlers, plugins []plugin.Plugin) (*Seeker, error) {
	if len(plugins) == 0 {
		return nil, errors.New("there should be at least one plugin provided")
	}

	return &Seeker{
		cfg,
		plugins,
		&handlers,
	}, nil
}

var i int = 0

func (s *Seeker) ReadTasks(pluginName string) []queue.Task {
	if i == 0 {
		i += 1
		return []queue.Task{
			[]byte(`{"source": "Haiti", "dest": "Hawaii"}`),
		}
	} else if i == 1 {
		i += 1
		return []queue.Task{
			[]byte(`{"source": "Hawaii", "dest": "Madagascar"}`),
		}
	} else {
		i += 1
		return []queue.Task{}
	}
}

func (s *Seeker) startQueues() {
	for _, _p := range s.plugins {
		_q := queue.New(_p.GetQueueConfig())

		consumeTaskCh := _q.StartConsuming(context.Background())

		go func(p plugin.Plugin, q *queue.Queue) {
			for {
				for _, t := range s.ReadTasks(p.GetName()) {
					_q.Publish(t)
				}
			}
		}(_p, &_q)

		go func(p plugin.Plugin, consumeTaskCh <-chan queue.Task) {
			for task := range consumeTaskCh {
				var request plugin.Request

				err := json.Unmarshal(task, &request)
				if err != nil {
					// Add error handling
					continue
				}

				res, err := p.DoRequest(request)
				if err != nil {
					// Add error handling
					continue
				}

				// add proper res processing
				fmt.Println("Res: ", res)
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
