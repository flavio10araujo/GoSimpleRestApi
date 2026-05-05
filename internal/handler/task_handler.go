package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/flavio10araujo/GoSimpleRestApi/internal/repository"
	"github.com/flavio10araujo/GoSimpleRestApi/internal/service"
)

type TaskHandler struct {
	taskService *service.TaskService
}

type createTaskRequest struct {
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

type updateTaskRequest struct {
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

func NewTaskHandler(taskService *service.TaskService) *TaskHandler {
	return &TaskHandler{taskService: taskService}
}

func (h *TaskHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	_ = r

	tasks, err := h.taskService.ListTasks()
	if err != nil {
		http.Error(w, "failed to list tasks", http.StatusInternalServerError)
		return
	}

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

	createdTask, err := h.taskService.AddTask(req.Title, req.Done)
	if err != nil {
		http.Error(w, "failed to create task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(createdTask); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	id, err := parseTaskID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var req updateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}

	updatedTask, err := h.taskService.UpdateTask(id, req.Title, req.Done)
	if err != nil {
		if errors.Is(err, repository.ErrTaskNotFound) {
			http.Error(w, "task not found", http.StatusNotFound)
			return
		}

		http.Error(w, "failed to update task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(updatedTask); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	id, err := parseTaskID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.taskService.DeleteTask(id); err != nil {
		if errors.Is(err, repository.ErrTaskNotFound) {
			http.Error(w, "task not found", http.StatusNotFound)
			return
		}

		http.Error(w, "failed to delete task", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func parseTaskID(r *http.Request) (int, error) {
	idValue := r.PathValue("id")
	id, err := strconv.Atoi(idValue)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid task id")
	}

	return id, nil
}
