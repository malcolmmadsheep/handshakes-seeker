package handlers

import "net/http"

type Handlers interface {
	CreateTask(http.ResponseWriter, *http.Request)
	DeleteTask(http.ResponseWriter, *http.Request)
	GetPath(http.ResponseWriter, *http.Request)
}
