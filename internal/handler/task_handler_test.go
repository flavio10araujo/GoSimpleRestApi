package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flavio10araujo/GoSimpleRestApi/internal/config"
	"github.com/flavio10araujo/GoSimpleRestApi/internal/model"
	"github.com/flavio10araujo/GoSimpleRestApi/internal/repository"
	"github.com/flavio10araujo/GoSimpleRestApi/internal/service"
)

type fakeHandlerRepository struct {
	addTaskFn    func(ctx context.Context, title string, done bool) (model.Task, error)
	getTaskFn    func(ctx context.Context, id int) (model.Task, error)
	listTasksFn  func(ctx context.Context, offset, limit int) ([]model.Task, int, error)
	updateTaskFn func(ctx context.Context, id int, title string, done bool) (model.Task, error)
	deleteTaskFn func(ctx context.Context, id int) error
}

func (f *fakeHandlerRepository) AddTask(ctx context.Context, title string, done bool) (model.Task, error) {
	return f.addTaskFn(ctx, title, done)
}

func (f *fakeHandlerRepository) GetTask(ctx context.Context, id int) (model.Task, error) {
	return f.getTaskFn(ctx, id)
}

func (f *fakeHandlerRepository) ListTasks(ctx context.Context, offset, limit int) ([]model.Task, int, error) {
	return f.listTasksFn(ctx, offset, limit)
}

func (f *fakeHandlerRepository) UpdateTask(ctx context.Context, id int, title string, done bool) (model.Task, error) {
	return f.updateTaskFn(ctx, id, title, done)
}

func (f *fakeHandlerRepository) DeleteTask(ctx context.Context, id int) error {
	return f.deleteTaskFn(ctx, id)
}

func newTestService(addFn func(context.Context, string, bool) (model.Task, error),
	getFn func(context.Context, int) (model.Task, error),
	listFn func(context.Context, int, int) ([]model.Task, int, error),
	updateFn func(context.Context, int, string, bool) (model.Task, error),
	deleteFn func(context.Context, int) error) *service.TaskService {
	fakeRepo := &fakeHandlerRepository{
		addTaskFn:    addFn,
		getTaskFn:    getFn,
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

func assertJSONError(t *testing.T, w *httptest.ResponseRecorder, expected string) {
	t.Helper()

	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("Error Content-Type is %q, expected application/json", ct)
	}

	var got ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("Failed to unmarshal error response: %v", err)
	}
	if got.Error != expected {
		t.Fatalf("Error message is %q, expected %q", got.Error, expected)
	}
}

func TestGetTasksSuccess(t *testing.T) {
	expected := []model.Task{{ID: 1, Title: "Task A", Done: false}, {ID: 2, Title: "Task B", Done: true}}
	svc := newTestService(
		nil,
		nil,
		func(_ context.Context, offset, limit int) ([]model.Task, int, error) {
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
		nil,
		func(_ context.Context, offset, limit int) ([]model.Task, int, error) {
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
	svc := newTestService(nil, nil, nil, nil, nil)
	handler := newTestHandler(svc)
	req := httptest.NewRequest("GET", "/tasks?limit=101", nil)
	w := httptest.NewRecorder()
	handler.GetTasks(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("GetTasks with limit>max returned status %d, expected %d", w.Code, http.StatusBadRequest)
	}
	assertJSONError(t, w, "limit exceeds maximum allowed")
}

func TestGetTasksInvalidOffset(t *testing.T) {
	svc := newTestService(nil, nil, nil, nil, nil)
	handler := newTestHandler(svc)
	req := httptest.NewRequest("GET", "/tasks?offset=-1", nil)
	w := httptest.NewRecorder()
	handler.GetTasks(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("GetTasks with negative offset returned status %d, expected %d", w.Code, http.StatusBadRequest)
	}
	assertJSONError(t, w, "invalid offset")
}

func TestGetTasksError(t *testing.T) {
	svc := newTestService(
		nil,
		nil,
		func(_ context.Context, offset, limit int) ([]model.Task, int, error) {
			return nil, 0, errors.New("db error")
		},
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
	assertJSONError(t, w, "failed to list tasks")
}

func TestCreateTaskSuccess(t *testing.T) {
	expected := model.Task{ID: 1, Title: "New Task", Done: false}
	svc := newTestService(
		func(_ context.Context, title string, done bool) (model.Task, error) {
			return expected, nil
		},
		nil,
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
	svc := newTestService(nil, nil, nil, nil, nil)
	handler := newTestHandler(svc)
	req := httptest.NewRequest("POST", "/tasks", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()
	handler.CreateTask(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("CreateTask with invalid body returned status %d, expected %d", w.Code, http.StatusBadRequest)
	}
	assertJSONError(t, w, "invalid request body")
}

func TestCreateTaskEmptyTitle(t *testing.T) {
	svc := newTestService(nil, nil, nil, nil, nil)
	handler := newTestHandler(svc)
	body := []byte(`{"title":"","done":false}`)
	req := httptest.NewRequest("POST", "/tasks", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.CreateTask(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("CreateTask with empty title returned status %d, expected %d", w.Code, http.StatusBadRequest)
	}
	assertJSONError(t, w, "title is required")
}

func TestCreateTaskServiceError(t *testing.T) {
	svc := newTestService(
		func(_ context.Context, title string, done bool) (model.Task, error) {
			return model.Task{}, errors.New("db error")
		},
		nil,
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
	assertJSONError(t, w, "failed to create task")
}

func TestReplaceTaskSuccess(t *testing.T) {
	expected := model.Task{ID: 5, Title: "Updated", Done: true}
	svc := newTestService(
		nil,
		nil,
		nil,
		func(_ context.Context, id int, title string, done bool) (model.Task, error) {
			return expected, nil
		},
		nil,
	)
	handler := newTestHandler(svc)
	body := []byte(`{"title":"Updated","done":true}`)
	req := httptest.NewRequest("PUT", "/tasks/5", bytes.NewReader(body))
	req.SetPathValue("id", "5")
	w := httptest.NewRecorder()
	handler.ReplaceTask(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("ReplaceTask returned status %d, expected %d", w.Code, http.StatusOK)
	}
	var got model.Task
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if got != expected {
		t.Fatalf("ReplaceTask returned %+v, expected %+v", got, expected)
	}
}

func TestReplaceTaskInvalidID(t *testing.T) {
	svc := newTestService(nil, nil, nil, nil, nil)
	handler := newTestHandler(svc)
	body := []byte(`{"title":"X","done":false}`)
	req := httptest.NewRequest("PUT", "/tasks/invalid", bytes.NewReader(body))
	req.SetPathValue("id", "invalid")
	w := httptest.NewRecorder()
	handler.ReplaceTask(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("ReplaceTask with invalid ID returned status %d, expected %d", w.Code, http.StatusBadRequest)
	}
	assertJSONError(t, w, "invalid task id")
}

func TestReplaceTaskInvalidBody(t *testing.T) {
	svc := newTestService(nil, nil, nil, nil, nil)
	handler := newTestHandler(svc)
	req := httptest.NewRequest("PUT", "/tasks/1", bytes.NewReader([]byte("invalid")))
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()
	handler.ReplaceTask(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("ReplaceTask with invalid body returned status %d, expected %d", w.Code, http.StatusBadRequest)
	}
	assertJSONError(t, w, "invalid request body")
}

func TestReplaceTaskEmptyTitle(t *testing.T) {
	svc := newTestService(nil, nil, nil, nil, nil)
	handler := newTestHandler(svc)
	body := []byte(`{"title":"","done":false}`)
	req := httptest.NewRequest("PUT", "/tasks/1", bytes.NewReader(body))
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()
	handler.ReplaceTask(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("ReplaceTask with empty title returned status %d, expected %d", w.Code, http.StatusBadRequest)
	}
	assertJSONError(t, w, "title is required")
}

func TestReplaceTaskNotFound(t *testing.T) {
	svc := newTestService(
		nil,
		nil,
		nil,
		func(_ context.Context, id int, title string, done bool) (model.Task, error) {
			return model.Task{}, repository.ErrTaskNotFound
		},
		nil,
	)
	handler := newTestHandler(svc)
	body := []byte(`{"title":"X","done":false}`)
	req := httptest.NewRequest("PUT", "/tasks/99", bytes.NewReader(body))
	req.SetPathValue("id", "99")
	w := httptest.NewRecorder()
	handler.ReplaceTask(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("ReplaceTask with not found error returned status %d, expected %d", w.Code, http.StatusNotFound)
	}
	assertJSONError(t, w, "task not found")
}

func TestReplaceTaskServiceError(t *testing.T) {
	svc := newTestService(
		nil,
		nil,
		nil,
		func(_ context.Context, id int, title string, done bool) (model.Task, error) {
			return model.Task{}, errors.New("db error")
		},
		nil,
	)
	handler := newTestHandler(svc)
	body := []byte(`{"title":"X","done":false}`)
	req := httptest.NewRequest("PUT", "/tasks/1", bytes.NewReader(body))
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()
	handler.ReplaceTask(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("ReplaceTask with service error returned status %d, expected %d", w.Code, http.StatusInternalServerError)
	}
	assertJSONError(t, w, "failed to update task")
}

func TestUpdateTaskPatchDoneOnlySuccess(t *testing.T) {
	var updatedTitle string
	var updatedDone bool

	expected := model.Task{ID: 8, Title: "Existing", Done: true}
	svc := newTestService(
		nil,
		func(_ context.Context, id int) (model.Task, error) {
			return model.Task{ID: id, Title: "Existing", Done: false}, nil
		},
		nil,
		func(_ context.Context, id int, title string, done bool) (model.Task, error) {
			updatedTitle = title
			updatedDone = done
			return expected, nil
		},
		nil,
	)
	handler := newTestHandler(svc)
	body := []byte(`{"done":true}`)
	req := httptest.NewRequest("PATCH", "/tasks/8", bytes.NewReader(body))
	req.SetPathValue("id", "8")
	w := httptest.NewRecorder()
	handler.UpdateTask(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("UpdateTask PATCH returned status %d, expected %d", w.Code, http.StatusOK)
	}
	if updatedTitle != "Existing" || !updatedDone {
		t.Fatalf("PATCH should preserve title and update done, got title=%q done=%t", updatedTitle, updatedDone)
	}

	var got model.Task
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if got != expected {
		t.Fatalf("UpdateTask PATCH returned %+v, expected %+v", got, expected)
	}
}

func TestUpdateTaskPatchWithoutFields(t *testing.T) {
	svc := newTestService(nil, nil, nil, nil, nil)
	handler := newTestHandler(svc)
	req := httptest.NewRequest("PATCH", "/tasks/1", bytes.NewReader([]byte(`{}`)))
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()
	handler.UpdateTask(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("UpdateTask PATCH without fields returned status %d, expected %d", w.Code, http.StatusBadRequest)
	}
	assertJSONError(t, w, "at least one field is required")
}

func TestDeleteTaskSuccess(t *testing.T) {
	svc := newTestService(
		nil,
		nil,
		nil,
		nil,
		func(_ context.Context, id int) error {
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
	svc := newTestService(nil, nil, nil, nil, nil)
	handler := newTestHandler(svc)
	req := httptest.NewRequest("DELETE", "/tasks/invalid", nil)
	req.SetPathValue("id", "invalid")
	w := httptest.NewRecorder()
	handler.DeleteTask(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("DeleteTask with invalid ID returned status %d, expected %d", w.Code, http.StatusBadRequest)
	}
	assertJSONError(t, w, "invalid task id")
}

func TestDeleteTaskNotFound(t *testing.T) {
	svc := newTestService(
		nil,
		nil,
		nil,
		nil,
		func(_ context.Context, id int) error {
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
	assertJSONError(t, w, "task not found")
}

func TestDeleteTaskServiceError(t *testing.T) {
	svc := newTestService(
		nil,
		nil,
		nil,
		nil,
		func(_ context.Context, id int) error {
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
	assertJSONError(t, w, "failed to delete task")
}
