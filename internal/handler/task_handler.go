package handler

import (
	"encoding/json"
	"net/http"

	"github.com/flavio10araujo/GoSimpleRestApi/internal/service"
)

type TaskHandler struct {
	taskService *service.TaskService
}

type createTaskRequest struct {
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

func NewTaskHandler(taskService *service.TaskService) *TaskHandler {
	return &TaskHandler{taskService: taskService}
}

func (h *TaskHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	_ = r

	tasks := h.taskService.ListTasks()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(tasks); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var req createTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}

	createdTask := h.taskService.AddTask(req.Title, req.Done)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(createdTask); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
