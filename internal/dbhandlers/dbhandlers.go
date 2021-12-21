package dbhandlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/malcolmmadsheep/handshakes-seeker/pkg/services"
)

type Handlers struct {
	conn        *pgxpool.Pool
	taskService services.TaskService
	pathService services.PathService
}

func New(
	conn *pgxpool.Pool,
	taskService services.TaskService,
	pathService services.PathService,
) *Handlers {
	return &Handlers{
		conn,
		taskService,
		pathService,
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

	sourceUrlTitle := h.taskService.CutUrlTitle(createTaskReq.SourceUrl)
	destUrlTitle := h.taskService.CutUrlTitle(createTaskReq.DestUrl)
	taskId := h.taskService.GenerateId(sourceUrlTitle, destUrlTitle)

	if task, err := h.taskService.GetTaskById(taskId); err == nil {
		_, err := h.taskService.UpdateTaskRequestsCount(task.OriginTaskId, 1)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"error": "%s"}`, err)
			return
		}
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `{"taskId": "%s"}`, task.Id)
		return
	} else if err != pgx.ErrNoRows {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"error": "%s"}`, err)
		return
	}

	task, err := h.taskService.CreateNewTask(taskId, taskId, sourceUrlTitle, destUrlTitle, "", 1)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"error": "%s"}`, err)
		return
	}

	_, err = h.pathService.GetPathByTaskId(taskId)
	if err != nil {
		if err != pgx.ErrNoRows {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"error": "%s"}`, err)
			return
		}
	}

	_, err = h.pathService.CreateNewPath(task)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"error": "%s"}`, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, `{"taskId": "%s"}`, task.Id)
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
	taskId := mux.Vars(r)["taskId"]

	path, err := h.pathService.GetPathByTaskId(taskId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"error": "%s"}`, err)
		return
	}

	if path.Status == services.PathStatusFound.String() && path.Trace == "" {
		path, err = h.pathService.BuildFullTraceAndUpdate(path)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"error": "%s"}`, err)
			return
		}
	}

	pathStr, err := json.Marshal(path)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"error": "%s"}`, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"path": %s}`, pathStr)
}
