package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/flavio10araujo/GoSimpleRestApi/internal/config"
	"github.com/flavio10araujo/GoSimpleRestApi/internal/model"
	"github.com/flavio10araujo/GoSimpleRestApi/internal/repository"
	"github.com/flavio10araujo/GoSimpleRestApi/internal/service"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeHandlerRepository struct {
	addTaskFn    func(title string, done bool) (model.Task, error)
	listTasksFn  func(offset, limit int) ([]model.Task, int, error)
	updateTaskFn func(id int, title string, done bool) (model.Task, error)
	deleteTaskFn func(id int) error
}

func (f *fakeHandlerRepository) AddTask(title string, done bool) (model.Task, error) {
	return f.addTaskFn(title, done)
}
func (f *fakeHandlerRepository) ListTasks(offset, limit int) ([]model.Task, int, error) {
	return f.listTasksFn(offset, limit)
}
func (f *fakeHandlerRepository) UpdateTask(id int, title string, done bool) (model.Task, error) {
	return f.updateTaskFn(id, title, done)
}
func (f *fakeHandlerRepository) DeleteTask(id int) error {
	return f.deleteTaskFn(id)
}
func newTestService(addFn func(string, bool) (model.Task, error),
	listFn func(int, int) ([]model.Task, int, error),
	updateFn func(int, string, bool) (model.Task, error),
	deleteFn func(int) error) *service.TaskService {
	fakeRepo := &fakeHandlerRepository{
		addTaskFn:    addFn,
		listTasksFn:  listFn,
		updateTaskFn: updateFn,
		deleteTaskFn: deleteFn,
	}
	return service.NewTaskService(fakeRepo)
}
func newTestHandler(svc *service.TaskService) *TaskHandler {
	paginationCfg := &config.PaginationConfig{DefaultLimit: 20, MaxLimit: 100}
	return NewTaskHandler(svc, paginationCfg)
}
func TestGetTasksSuccess(t *testing.T) {
	expected := []model.Task{{ID: 1, Title: "Task A", Done: false}, {ID: 2, Title: "Task B", Done: true}}
	svc := newTestService(
		nil,
		func(offset, limit int) ([]model.Task, int, error) {
			return expected, 2, nil
		},
		nil,
		nil,
	)
	handler := newTestHandler(svc)
	req := httptest.NewRequest("GET", "/tasks", nil)
	w := httptest.NewRecorder()
	handler.GetTasks(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GetTasks returned status %d, expected %d", w.Code, http.StatusOK)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("GetTasks Content-Type is %q, expected application/json", ct)
	}
	var got PaginatedResponse
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if got.Total != 2 {
		t.Fatalf("GetTasks total is %d, expected 2", got.Total)
	}
	if got.Limit != 20 {
		t.Fatalf("GetTasks limit is %d, expected 20", got.Limit)
	}
	if len(got.Data) != len(expected) {
		t.Fatalf("GetTasks returned %d tasks, expected %d", len(got.Data), len(expected))
	}
}
func TestGetTasksWithPagination(t *testing.T) {
	svc := newTestService(
		nil,
		func(offset, limit int) ([]model.Task, int, error) {
			if offset != 10 || limit != 5 {
				t.Errorf("Expected offset=10, limit=5 but got offset=%d, limit=%d", offset, limit)
			}
			return []model.Task{{ID: 11, Title: "Task 11", Done: false}}, 50, nil
		},
		nil,
		nil,
	)
	handler := newTestHandler(svc)
	req := httptest.NewRequest("GET", "/tasks?offset=10&limit=5", nil)
	w := httptest.NewRecorder()
	handler.GetTasks(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GetTasks returned status %d, expected %d", w.Code, http.StatusOK)
	}
	var got PaginatedResponse
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if got.Offset != 10 {
		t.Fatalf("GetTasks offset is %d, expected 10", got.Offset)
	}
	if got.Limit != 5 {
		t.Fatalf("GetTasks limit is %d, expected 5", got.Limit)
	}
	if got.Total != 50 {
		t.Fatalf("GetTasks total is %d, expected 50", got.Total)
	}
}
func TestGetTasksInvalidLimit(t *testing.T) {
	svc := newTestService(nil, nil, nil, nil)
	handler := newTestHandler(svc)
	req := httptest.NewRequest("GET", "/tasks?limit=101", nil)
	w := httptest.NewRecorder()
	handler.GetTasks(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("GetTasks with limit>max returned status %d, expected %d", w.Code, http.StatusBadRequest)
	}
}
func TestGetTasksInvalidOffset(t *testing.T) {
	svc := newTestService(nil, nil, nil, nil)
	handler := newTestHandler(svc)
	req := httptest.NewRequest("GET", "/tasks?offset=-1", nil)
	w := httptest.NewRecorder()
	handler.GetTasks(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("GetTasks with negative offset returned status %d, expected %d", w.Code, http.StatusBadRequest)
	}
}
func TestGetTasksError(t *testing.T) {
	svc := newTestService(
		nil,
		func(offset, limit int) ([]model.Task, int, error) { return nil, 0, errors.New("db error") },
		nil,
		nil,
	)
	handler := newTestHandler(svc)
	req := httptest.NewRequest("GET", "/tasks", nil)
	w := httptest.NewRecorder()
	handler.GetTasks(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("GetTasks returned status %d, expected %d", w.Code, http.StatusInternalServerError)
	}
}
func TestCreateTaskSuccess(t *testing.T) {
	expected := model.Task{ID: 1, Title: "New Task", Done: false}
	svc := newTestService(
		func(title string, done bool) (model.Task, error) {
			return expected, nil
		},
		nil,
		nil,
		nil,
	)
	handler := newTestHandler(svc)
	body := []byte(`{"title":"New Task","done":false}`)
	req := httptest.NewRequest("POST", "/tasks", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.CreateTask(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("CreateTask returned status %d, expected %d", w.Code, http.StatusCreated)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("CreateTask Content-Type is %q, expected application/json", ct)
	}
	var got model.Task
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if got != expected {
		t.Fatalf("CreateTask returned %+v, expected %+v", got, expected)
	}
}
func TestCreateTaskInvalidBody(t *testing.T) {
	svc := newTestService(nil, nil, nil, nil)
	handler := newTestHandler(svc)
	req := httptest.NewRequest("POST", "/tasks", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()
	handler.CreateTask(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("CreateTask with invalid body returned status %d, expected %d", w.Code, http.StatusBadRequest)
	}
}
func TestCreateTaskEmptyTitle(t *testing.T) {
	svc := newTestService(nil, nil, nil, nil)
	handler := newTestHandler(svc)
	body := []byte(`{"title":"","done":false}`)
	req := httptest.NewRequest("POST", "/tasks", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.CreateTask(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("CreateTask with empty title returned status %d, expected %d", w.Code, http.StatusBadRequest)
	}
}
func TestCreateTaskServiceError(t *testing.T) {
	svc := newTestService(
		func(title string, done bool) (model.Task, error) {
			return model.Task{}, errors.New("db error")
		},
		nil,
		nil,
		nil,
	)
	handler := newTestHandler(svc)
	body := []byte(`{"title":"Task","done":false}`)
	req := httptest.NewRequest("POST", "/tasks", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.CreateTask(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("CreateTask with service error returned status %d, expected %d", w.Code, http.StatusInternalServerError)
	}
}
func TestUpdateTaskSuccess(t *testing.T) {
	expected := model.Task{ID: 5, Title: "Updated", Done: true}
	svc := newTestService(
		nil,
		nil,
		func(id int, title string, done bool) (model.Task, error) {
			return expected, nil
		},
		nil,
	)
	handler := newTestHandler(svc)
	body := []byte(`{"title":"Updated","done":true}`)
	req := httptest.NewRequest("PUT", "/tasks/5", bytes.NewReader(body))
	req.SetPathValue("id", "5")
	w := httptest.NewRecorder()
	handler.UpdateTask(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("UpdateTask returned status %d, expected %d", w.Code, http.StatusOK)
	}
	var got model.Task
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if got != expected {
		t.Fatalf("UpdateTask returned %+v, expected %+v", got, expected)
	}
}
func TestUpdateTaskInvalidID(t *testing.T) {
	svc := newTestService(nil, nil, nil, nil)
	handler := newTestHandler(svc)
	body := []byte(`{"title":"X","done":false}`)
	req := httptest.NewRequest("PUT", "/tasks/invalid", bytes.NewReader(body))
	req.SetPathValue("id", "invalid")
	w := httptest.NewRecorder()
	handler.UpdateTask(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("UpdateTask with invalid ID returned status %d, expected %d", w.Code, http.StatusBadRequest)
	}
}
func TestUpdateTaskInvalidBody(t *testing.T) {
	svc := newTestService(nil, nil, nil, nil)
	handler := newTestHandler(svc)
	req := httptest.NewRequest("PUT", "/tasks/1", bytes.NewReader([]byte("invalid")))
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()
	handler.UpdateTask(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("UpdateTask with invalid body returned status %d, expected %d", w.Code, http.StatusBadRequest)
	}
}
func TestUpdateTaskEmptyTitle(t *testing.T) {
	svc := newTestService(nil, nil, nil, nil)
	handler := newTestHandler(svc)
	body := []byte(`{"title":"","done":false}`)
	req := httptest.NewRequest("PUT", "/tasks/1", bytes.NewReader(body))
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()
	handler.UpdateTask(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("UpdateTask with empty title returned status %d, expected %d", w.Code, http.StatusBadRequest)
	}
}
func TestUpdateTaskNotFound(t *testing.T) {
	svc := newTestService(
		nil,
		nil,
		func(id int, title string, done bool) (model.Task, error) {
			return model.Task{}, repository.ErrTaskNotFound
		},
		nil,
	)
	handler := newTestHandler(svc)
	body := []byte(`{"title":"X","done":false}`)
	req := httptest.NewRequest("PUT", "/tasks/99", bytes.NewReader(body))
	req.SetPathValue("id", "99")
	w := httptest.NewRecorder()
	handler.UpdateTask(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("UpdateTask with not found error returned status %d, expected %d", w.Code, http.StatusNotFound)
	}
}
func TestUpdateTaskServiceError(t *testing.T) {
	svc := newTestService(
		nil,
		nil,
		func(id int, title string, done bool) (model.Task, error) {
			return model.Task{}, errors.New("db error")
		},
		nil,
	)
	handler := newTestHandler(svc)
	body := []byte(`{"title":"X","done":false}`)
	req := httptest.NewRequest("PUT", "/tasks/1", bytes.NewReader(body))
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()
	handler.UpdateTask(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("UpdateTask with service error returned status %d, expected %d", w.Code, http.StatusInternalServerError)
	}
}
func TestDeleteTaskSuccess(t *testing.T) {
	svc := newTestService(
		nil,
		nil,
		nil,
		func(id int) error {
			return nil
		},
	)
	handler := newTestHandler(svc)
	req := httptest.NewRequest("DELETE", "/tasks/3", nil)
	req.SetPathValue("id", "3")
	w := httptest.NewRecorder()
	handler.DeleteTask(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("DeleteTask returned status %d, expected %d", w.Code, http.StatusNoContent)
	}
	if w.Body.Len() > 0 {
		t.Fatalf("DeleteTask should return empty body, got %q", w.Body.String())
	}
}
func TestDeleteTaskInvalidID(t *testing.T) {
	svc := newTestService(nil, nil, nil, nil)
	handler := newTestHandler(svc)
	req := httptest.NewRequest("DELETE", "/tasks/invalid", nil)
	req.SetPathValue("id", "invalid")
	w := httptest.NewRecorder()
	handler.DeleteTask(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("DeleteTask with invalid ID returned status %d, expected %d", w.Code, http.StatusBadRequest)
	}
}
func TestDeleteTaskNotFound(t *testing.T) {
	svc := newTestService(
		nil,
		nil,
		nil,
		func(id int) error {
			return repository.ErrTaskNotFound
		},
	)
	handler := newTestHandler(svc)
	req := httptest.NewRequest("DELETE", "/tasks/99", nil)
	req.SetPathValue("id", "99")
	w := httptest.NewRecorder()
	handler.DeleteTask(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("DeleteTask with not found error returned status %d, expected %d", w.Code, http.StatusNotFound)
	}
}
func TestDeleteTaskServiceError(t *testing.T) {
	svc := newTestService(
		nil,
		nil,
		nil,
		func(id int) error {
			return errors.New("db error")
		},
	)
	handler := newTestHandler(svc)
	req := httptest.NewRequest("DELETE", "/tasks/1", nil)
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()
	handler.DeleteTask(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("DeleteTask with service error returned status %d, expected %d", w.Code, http.StatusInternalServerError)
	}
}
