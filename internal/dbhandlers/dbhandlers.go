package dbhandlers

import (
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v4"
)

type Handlers struct {
	conn *pgx.Conn
}

func New(conn *pgx.Conn) *Handlers {
	return &Handlers{
		conn,
	}
}

func (h *Handlers) CreateTask(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)

	fmt.Fprint(w, "POST OK")
}

func (h *Handlers) DeleteTask(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	fmt.Fprint(w, "DELETE OK")
}

func (h *Handlers) GetPath(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	fmt.Fprint(w, "GET OK")
}
