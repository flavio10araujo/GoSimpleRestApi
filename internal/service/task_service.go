package service

import (
	"sync"

	"github.com/flavio10araujo/GoSimpleRestApi/internal/model"
)

type TaskService struct {
	mu     sync.RWMutex
	tasks  []model.Task
	nextID int
}

func NewTaskService() *TaskService {
	return &TaskService{
		tasks:  make([]model.Task, 0),
		nextID: 1,
	}
}

func (s *TaskService) AddTask(title string, done bool) model.Task {
	s.mu.Lock()
	defer s.mu.Unlock()

	task := model.Task{
		ID:    s.nextID,
		Title: title,
		Done:  done,
	}

	s.nextID++
	s.tasks = append(s.tasks, task)

	return task
}

func (s *TaskService) ListTasks() []model.Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]model.Task, len(s.tasks))
	copy(tasks, s.tasks)

	return tasks
}
