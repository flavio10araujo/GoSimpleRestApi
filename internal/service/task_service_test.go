package service

import (
	"errors"
	"strings"
	"testing"

	"github.com/flavio10araujo/GoSimpleRestApi/internal/model"
	"github.com/flavio10araujo/GoSimpleRestApi/internal/repository"
)

type fakeTaskRepository struct {
	addTaskFn    func(title string, done bool) (model.Task, error)
	listTasksFn  func() ([]model.Task, error)
	updateTaskFn func(id int, title string, done bool) (model.Task, error)
	deleteTaskFn func(id int) error
}

func (f *fakeTaskRepository) AddTask(title string, done bool) (model.Task, error) {
	return f.addTaskFn(title, done)
}

func (f *fakeTaskRepository) ListTasks() ([]model.Task, error) {
	return f.listTasksFn()
}

func (f *fakeTaskRepository) UpdateTask(id int, title string, done bool) (model.Task, error) {
	return f.updateTaskFn(id, title, done)
}

func (f *fakeTaskRepository) DeleteTask(id int) error {
	return f.deleteTaskFn(id)
}

func TestTaskServiceAddTaskSuccess(t *testing.T) {
	var gotTitle string
	var gotDone bool

	expected := model.Task{ID: 1, Title: "Study Go", Done: false}
	repo := &fakeTaskRepository{
		addTaskFn: func(title string, done bool) (model.Task, error) {
			gotTitle = title
			gotDone = done
			return expected, nil
		},
	}

	svc := NewTaskService(repo)
	got, err := svc.AddTask("Study Go", false)
	if err != nil {
		t.Fatalf("AddTask returned unexpected error: %v", err)
	}
	if got != expected {
		t.Fatalf("AddTask returned %+v, expected %+v", got, expected)
	}
	if gotTitle != "Study Go" || gotDone != false {
		t.Fatalf("AddTask forwarded wrong args: title=%q done=%t", gotTitle, gotDone)
	}
}

func TestTaskServiceAddTaskError(t *testing.T) {
	repoErr := errors.New("insert failed")
	repo := &fakeTaskRepository{
		addTaskFn: func(title string, done bool) (model.Task, error) {
			return model.Task{}, repoErr
		},
	}

	svc := NewTaskService(repo)
	got, err := svc.AddTask("X", true)
	if err == nil {
		t.Fatal("AddTask expected error, got nil")
	}
	if !errors.Is(err, repoErr) {
		t.Fatalf("AddTask error should wrap repo error")
	}
	if !strings.Contains(err.Error(), "add task:") {
		t.Fatalf("AddTask error missing context, got: %v", err)
	}
	if got != (model.Task{}) {
		t.Fatalf("AddTask should return zero-value task on error")
	}
}

func TestTaskServiceListTasksSuccess(t *testing.T) {
	expected := []model.Task{{ID: 1, Title: "A", Done: false}, {ID: 2, Title: "B", Done: true}}
	repo := &fakeTaskRepository{
		listTasksFn: func() ([]model.Task, error) {
			return expected, nil
		},
	}

	svc := NewTaskService(repo)
	got, err := svc.ListTasks()
	if err != nil {
		t.Fatalf("ListTasks returned unexpected error: %v", err)
	}
	if len(got) != len(expected) {
		t.Fatalf("ListTasks returned %d tasks, expected %d", len(got), len(expected))
	}
	for i := range expected {
		if got[i] != expected[i] {
			t.Fatalf("ListTasks task[%d] = %+v, expected %+v", i, got[i], expected[i])
		}
	}
}

func TestTaskServiceListTasksError(t *testing.T) {
	repoErr := errors.New("query failed")
	repo := &fakeTaskRepository{
		listTasksFn: func() ([]model.Task, error) {
			return nil, repoErr
		},
	}

	svc := NewTaskService(repo)
	got, err := svc.ListTasks()
	if err == nil {
		t.Fatal("ListTasks expected error, got nil")
	}
	if !errors.Is(err, repoErr) {
		t.Fatalf("ListTasks error should wrap repo error")
	}
	if !strings.Contains(err.Error(), "list tasks:") {
		t.Fatalf("ListTasks error missing context, got: %v", err)
	}
	if got != nil {
		t.Fatalf("ListTasks should return nil slice on error")
	}
}

func TestTaskServiceUpdateTaskSuccess(t *testing.T) {
	var gotID int
	var gotTitle string
	var gotDone bool

	expected := model.Task{ID: 7, Title: "Updated", Done: true}
	repo := &fakeTaskRepository{
		updateTaskFn: func(id int, title string, done bool) (model.Task, error) {
			gotID = id
			gotTitle = title
			gotDone = done
			return expected, nil
		},
	}

	svc := NewTaskService(repo)
	got, err := svc.UpdateTask(7, "Updated", true)
	if err != nil {
		t.Fatalf("UpdateTask returned unexpected error: %v", err)
	}
	if got != expected {
		t.Fatalf("UpdateTask returned %+v, expected %+v", got, expected)
	}
	if gotID != 7 || gotTitle != "Updated" || gotDone != true {
		t.Fatalf("UpdateTask forwarded wrong args: id=%d title=%q done=%t", gotID, gotTitle, gotDone)
	}
}

func TestTaskServiceUpdateTaskError(t *testing.T) {
	repo := &fakeTaskRepository{
		updateTaskFn: func(id int, title string, done bool) (model.Task, error) {
			return model.Task{}, repository.ErrTaskNotFound
		},
	}

	svc := NewTaskService(repo)
	got, err := svc.UpdateTask(99, "Missing", false)
	if err == nil {
		t.Fatal("UpdateTask expected error, got nil")
	}
	if !errors.Is(err, repository.ErrTaskNotFound) {
		t.Fatalf("UpdateTask error should wrap ErrTaskNotFound")
	}
	if !strings.Contains(err.Error(), "update task:") {
		t.Fatalf("UpdateTask error missing context, got: %v", err)
	}
	if got != (model.Task{}) {
		t.Fatalf("UpdateTask should return zero-value task on error")
	}
}

func TestTaskServiceDeleteTaskSuccess(t *testing.T) {
	var gotID int
	repo := &fakeTaskRepository{
		deleteTaskFn: func(id int) error {
			gotID = id
			return nil
		},
	}

	svc := NewTaskService(repo)
	if err := svc.DeleteTask(3); err != nil {
		t.Fatalf("DeleteTask returned unexpected error: %v", err)
	}
	if gotID != 3 {
		t.Fatalf("DeleteTask forwarded wrong id: %d", gotID)
	}
}

func TestTaskServiceDeleteTaskError(t *testing.T) {
	repo := &fakeTaskRepository{
		deleteTaskFn: func(id int) error {
			return repository.ErrTaskNotFound
		},
	}

	svc := NewTaskService(repo)
	err := svc.DeleteTask(77)
	if err == nil {
		t.Fatal("DeleteTask expected error, got nil")
	}
	if !errors.Is(err, repository.ErrTaskNotFound) {
		t.Fatalf("DeleteTask error should wrap ErrTaskNotFound")
	}
	if !strings.Contains(err.Error(), "delete task:") {
		t.Fatalf("DeleteTask error missing context, got: %v", err)
	}
}
