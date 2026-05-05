package service

import (
	"fmt"

	"github.com/flavio10araujo/GoSimpleRestApi/internal/model"
	"github.com/flavio10araujo/GoSimpleRestApi/internal/repository"
)

type TaskService struct {
	repository repository.TaskRepository
}

func NewTaskService(taskRepository repository.TaskRepository) *TaskService {
	return &TaskService{repository: taskRepository}
}

func (s *TaskService) AddTask(title string, done bool) (model.Task, error) {
	task, err := s.repository.AddTask(title, done)
	if err != nil {
		return model.Task{}, fmt.Errorf("add task: %w", err)
	}

	return task, nil
}

func (s *TaskService) ListTasks() ([]model.Task, error) {
	tasks, err := s.repository.ListTasks()
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}

	return tasks, nil
}

func (s *TaskService) UpdateTask(id int, title string, done bool) (model.Task, error) {
	task, err := s.repository.UpdateTask(id, title, done)
	if err != nil {
		return model.Task{}, fmt.Errorf("update task: %w", err)
	}

	return task, nil
}

func (s *TaskService) DeleteTask(id int) error {
	if err := s.repository.DeleteTask(id); err != nil {
		return fmt.Errorf("delete task: %w", err)
	}

	return nil
}
