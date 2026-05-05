package repository

import (
	"errors"

	"github.com/flavio10araujo/GoSimpleRestApi/internal/model"
)

var ErrTaskNotFound = errors.New("task not found")

type TaskRepository interface {
	AddTask(title string, done bool) (model.Task, error)
	ListTasks() ([]model.Task, error)
	UpdateTask(id int, title string, done bool) (model.Task, error)
	DeleteTask(id int) error
}
