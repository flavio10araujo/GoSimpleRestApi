package repository

import (
	"context"
	"errors"

	"github.com/flavio10araujo/GoSimpleRestApi/internal/model"
)

var ErrTaskNotFound = errors.New("task not found")

type TaskRepository interface {
	AddTask(ctx context.Context, title string, done bool) (model.Task, error)
	GetTask(ctx context.Context, id int) (model.Task, error)
	ListTasks(ctx context.Context, offset, limit int) ([]model.Task, int, error)
	UpdateTask(ctx context.Context, id int, title string, done bool) (model.Task, error)
	DeleteTask(ctx context.Context, id int) error
}
