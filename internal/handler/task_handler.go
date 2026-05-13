package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/flavio10araujo/GoSimpleRestApi/internal/config"
	"github.com/flavio10araujo/GoSimpleRestApi/internal/model"
	"github.com/flavio10araujo/GoSimpleRestApi/internal/repository"
	"github.com/flavio10araujo/GoSimpleRestApi/internal/service"
)

type TaskHandler struct {
	taskService      *service.TaskService
	paginationConfig *config.PaginationConfig
}

type createTaskRequest struct {
	Title string `json:"title" example:"Buy groceries"`
	Done  bool   `json:"done" example:"false"`
}

type updateTaskRequest struct {
	Title string `json:"title" example:"Buy milk"`
	Done  bool   `json:"done" example:"true"`
}

type patchTaskRequest struct {
	Title *string `json:"title" example:"Buy milk"`
	Done  *bool   `json:"done" example:"true"`
}

func NewTaskHandler(taskService *service.TaskService, paginationConfig *config.PaginationConfig) *TaskHandler {
	return &TaskHandler{taskService: taskService, paginationConfig: paginationConfig}
}

// CreateTask godoc
// @Summary      Create a new task
// @Description  Add a new task to the database
// @Tags         tasks
// @Accept       json
// @Param        body  body  createTaskRequest  true  "Task data"
// @Produce      json
// @Success      201  {object}  model.Task
// @Failure      400  {object}  ErrorResponse  "Missing title or invalid JSON"
// @Failure      500  {object}  ErrorResponse  "Database error"
// @Router       /tasks [post]
func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var req createTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Title == "" {
		writeErrorJSON(w, http.StatusBadRequest, "title is required")
		return
	}

	createdTask, err := h.taskService.AddTask(req.Title, req.Done)
	if err != nil {
		writeErrorJSON(w, http.StatusInternalServerError, "failed to create task")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(createdTask); err != nil {
		writeErrorJSON(w, http.StatusInternalServerError, "failed to encode response")
	}
}

// GetTasks godoc
// @Summary      List all tasks with pagination
// @Description  Retrieve paginated list of tasks from the database
// @Tags         tasks
// @Param        offset  query  int  false  "Offset for pagination (default: 0)"  default(0)
// @Param        limit   query  int  false  "Limit per page (default: 20, max: 100)"  default(20)
// @Produce      json
// @Success      200  {object}  PaginatedResponse
// @Failure      400  {object}  ErrorResponse  "Invalid pagination parameters"
// @Failure      500  {object}  ErrorResponse  "Database error"
// @Router       /tasks [get]
func (h *TaskHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	offset, limit, err := parsePaginationParams(r, h.paginationConfig)
	if err != nil {
		writeErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	tasks, total, err := h.taskService.ListTasks(offset, limit)
	if err != nil {
		writeErrorJSON(w, http.StatusInternalServerError, "failed to list tasks")
		return
	}

	if tasks == nil {
		tasks = make([]model.Task, 0)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := PaginatedResponse{
		Data:   tasks,
		Total:  total,
		Offset: offset,
		Limit:  limit,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		writeErrorJSON(w, http.StatusInternalServerError, "failed to encode response")
	}
}

// GetTask godoc
// @Summary      Get a task by ID
// @Description  Retrieve a single task by its ID
// @Tags         tasks
// @Param        id  path  int  true  "Task ID"
// @Produce      json
// @Success      200  {object}  model.Task
// @Failure      400  {object}  ErrorResponse  "Invalid ID"
// @Failure      404  {object}  ErrorResponse  "Task not found"
// @Failure      500  {object}  ErrorResponse  "Database error"
// @Router       /tasks/{id} [get]
func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	id, err := parseTaskID(r)
	if err != nil {
		writeErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	task, err := h.taskService.GetTask(id)
	if err != nil {
		if errors.Is(err, repository.ErrTaskNotFound) {
			writeErrorJSON(w, http.StatusNotFound, "task not found")
			return
		}

		writeErrorJSON(w, http.StatusInternalServerError, "failed to get task")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(task); err != nil {
		writeErrorJSON(w, http.StatusInternalServerError, "failed to encode response")
	}
}

// ReplaceTask godoc
// @Summary      Replace an existing task
// @Description  Replace all task fields by ID
// @Tags         tasks
// @Accept       json
// @Param        id    path  int                    true  "Task ID"
// @Param        body  body  updateTaskRequest  true  "Updated task data"
// @Produce      json
// @Success      200  {object}  model.Task
// @Failure      400  {object}  ErrorResponse  "Invalid ID or request"
// @Failure      404  {object}  ErrorResponse  "Task not found"
// @Failure      500  {object}  ErrorResponse  "Database error"
// @Router       /tasks/{id} [put]
func (h *TaskHandler) ReplaceTask(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	id, err := parseTaskID(r)
	if err != nil {
		writeErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	var req updateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Title == "" {
		writeErrorJSON(w, http.StatusBadRequest, "title is required")
		return
	}

	updatedTask, err := h.taskService.UpdateTask(id, req.Title, req.Done)
	if err != nil {
		if errors.Is(err, repository.ErrTaskNotFound) {
			writeErrorJSON(w, http.StatusNotFound, "task not found")
			return
		}

		writeErrorJSON(w, http.StatusInternalServerError, "failed to update task")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(updatedTask); err != nil {
		writeErrorJSON(w, http.StatusInternalServerError, "failed to encode response")
	}
}

// UpdateTask godoc
// @Summary      Partially update an existing task
// @Description  Modify one or more task fields by ID
// @Tags         tasks
// @Accept       json
// @Param        id    path  int               true  "Task ID"
// @Param        body  body  patchTaskRequest  true  "Fields to update"
// @Produce      json
// @Success      200  {object}  model.Task
// @Failure      400  {object}  ErrorResponse  "Invalid ID or request"
// @Failure      404  {object}  ErrorResponse  "Task not found"
// @Failure      500  {object}  ErrorResponse  "Database error"
// @Router       /tasks/{id} [patch]
func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	id, err := parseTaskID(r)
	if err != nil {
		writeErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	var req patchTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Title == nil && req.Done == nil {
		writeErrorJSON(w, http.StatusBadRequest, "at least one field is required")
		return
	}

	currentTask, err := h.taskService.GetTask(id)
	if err != nil {
		if errors.Is(err, repository.ErrTaskNotFound) {
			writeErrorJSON(w, http.StatusNotFound, "task not found")
			return
		}

		writeErrorJSON(w, http.StatusInternalServerError, "failed to get task")
		return
	}

	if req.Title != nil {
		if *req.Title == "" {
			writeErrorJSON(w, http.StatusBadRequest, "title cannot be empty")
			return
		}
		currentTask.Title = *req.Title
	}
	if req.Done != nil {
		currentTask.Done = *req.Done
	}

	updatedTask, err := h.taskService.UpdateTask(id, currentTask.Title, currentTask.Done)
	if err != nil {
		if errors.Is(err, repository.ErrTaskNotFound) {
			writeErrorJSON(w, http.StatusNotFound, "task not found")
			return
		}

		writeErrorJSON(w, http.StatusInternalServerError, "failed to update task")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(updatedTask); err != nil {
		writeErrorJSON(w, http.StatusInternalServerError, "failed to encode response")
	}
}

// DeleteTask godoc
// @Summary      Delete a task
// @Description  Remove a task by ID
// @Tags         tasks
// @Param        id  path  int  true  "Task ID"
// @Produce      json
// @Success      204
// @Failure      400  {object}  ErrorResponse  "Invalid ID"
// @Failure      404  {object}  ErrorResponse  "Task not found"
// @Failure      500  {object}  ErrorResponse  "Database error"
// @Router       /tasks/{id} [delete]
func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	id, err := parseTaskID(r)
	if err != nil {
		writeErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.taskService.DeleteTask(id); err != nil {
		if errors.Is(err, repository.ErrTaskNotFound) {
			writeErrorJSON(w, http.StatusNotFound, "task not found")
			return
		}

		writeErrorJSON(w, http.StatusInternalServerError, "failed to delete task")
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

func parsePaginationParams(r *http.Request, cfg *config.PaginationConfig) (int, int, error) {
	offsetStr := r.URL.Query().Get("offset")
	limitStr := r.URL.Query().Get("limit")

	offset := 0
	if offsetStr != "" {
		o, err := strconv.Atoi(offsetStr)
		if err != nil || o < 0 {
			return 0, 0, errors.New("invalid offset")
		}
		offset = o
	}

	limit := cfg.DefaultLimit
	if limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err != nil || l <= 0 {
			return 0, 0, errors.New("invalid limit")
		}
		if l > cfg.MaxLimit {
			return 0, 0, errors.New("limit exceeds maximum allowed")
		}
		limit = l
	}

	return offset, limit, nil
}
