package seeker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/malcolmmadsheep/handshakes-seeker/pkg/plugin"
	"github.com/malcolmmadsheep/handshakes-seeker/pkg/queue"
)

type Seeker struct {
	cfg     Config
	plugins []plugin.Plugin
}

type Config struct{}

func New(cfg Config, plugins []plugin.Plugin) (*Seeker, error) {
	if len(plugins) == 0 {
		return nil, errors.New("there should be at least one plugin provided")
	}

	return &Seeker{
		cfg,
		plugins,
	}, nil
}

func (s *Seeker) ReadTasks(pluginName string) []queue.Task {
	return make([]queue.Task, 0)
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

func createHTTPServer() *http.Server {
	apiRouter := mux.NewRouter().StrictSlash(false).PathPrefix("/api/v1").Subrouter()

	apiRouter.HandleFunc("/task", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "POST task\n")
	}).Methods(http.MethodPost)

	taskSubrouter := apiRouter.PathPrefix("/task/{taskId}").Subrouter()

	taskSubrouter.HandleFunc("", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "GET task")
	}).Methods(http.MethodGet)

	taskSubrouter.HandleFunc("", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "DELETE task")
	}).Methods(http.MethodDelete)

	return &http.Server{
		Handler:      apiRouter,
		Addr:         ":8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
}

func (s *Seeker) Run() error {
	s.startQueues()

	srv := createHTTPServer()

	return srv.ListenAndServe()
}
