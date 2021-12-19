package dbhandlers

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/jackc/pgx/v4"
	"github.com/malcolmmadsheep/handshakes-seeker/pkg/services"
)

type Handlers struct {
	conn        *pgx.Conn
	taskService services.TaskService
}

func New(conn *pgx.Conn, taskService services.TaskService) *Handlers {
	return &Handlers{
		conn,
		taskService,
	}
}

type CreateTaskReq struct {
	SourceUrl string `json:"source_url"`
	DestUrl   string `json:"dest_url"`
}

func (h *Handlers) CreateTask(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if task, err := h.taskService.GetTaskByBody(body); err == nil {
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `{"taskId": "%d"}`, task.Id)
		return
	} else if err != pgx.ErrNoRows {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"error": "%s"}`, err)
		return
	}

	if task, err := h.taskService.CreateNewTask(body); err == nil {
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `{"taskId": "%d"}`, task.Id)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"error": "%s"}`, err)
	}
}

func (h *Handlers) DeleteTask(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	fmt.Fprint(w, "DELETE OK")
}

func (h *Handlers) GetPath(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	fmt.Fprint(w, "GET OK")
}
