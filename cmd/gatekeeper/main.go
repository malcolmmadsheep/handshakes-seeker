package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func main() {
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

	srv := &http.Server{
		Handler:      apiRouter,
		Addr:         ":8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	fmt.Println("Running server...")
	log.Fatal(srv.ListenAndServe())
}
