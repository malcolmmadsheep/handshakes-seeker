package dbhandlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
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
	var createTaskReq CreateTaskReq
	err := json.NewDecoder(r.Body).Decode(&createTaskReq)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	taskId := h.taskService.GenerateId(createTaskReq.SourceUrl, createTaskReq.DestUrl)

	if task, err := h.taskService.GetTaskById(taskId); err == nil {
		count, err := h.taskService.UpdateTaskRequestsCount(task.OriginTaskId, 1)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"error": "%s"}`, err)
			return
		}
		fmt.Println("for id", taskId, "count", count)
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `{"taskId": "%s"}`, task.Id)
		return
	} else if err != pgx.ErrNoRows {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"error": "%s"}`, err)
		return
	}

	if task, err := h.taskService.CreateNewTask(taskId, taskId, createTaskReq.SourceUrl, createTaskReq.DestUrl, "", 1); err == nil {
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `{"taskId": "%s"}`, task.Id)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"error": "%s"}`, err)
	}
}

func (h *Handlers) DeleteTask(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	if taskId, contains := params["taskId"]; contains {
		count, err := h.taskService.UpdateTaskRequestsCount(taskId, -1)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if count > 0 {
			w.WriteHeader(http.StatusOK)
			return
		}

		h.taskService.DeleteAllTasksWithOrigin(taskId)

		w.WriteHeader(http.StatusOK)
		return
	}

	w.WriteHeader(http.StatusBadRequest)
}

func (h *Handlers) GetPath(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	fmt.Fprint(w, "GET OK")
}
