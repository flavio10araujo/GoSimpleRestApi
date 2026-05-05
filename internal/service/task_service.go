package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/flavio10araujo/GoSimpleRestApi/internal/model"
)

type TaskService struct {
	db *sql.DB
}

func NewTaskService(db *sql.DB) *TaskService {
	return &TaskService{db: db}
}

func (s *TaskService) AddTask(title string, done bool) (model.Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	result, err := s.db.ExecContext(ctx, "INSERT INTO tasks (title, done) VALUES (?, ?)", title, boolToSQLite(done))
	if err != nil {
		return model.Task{}, fmt.Errorf("insert task: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return model.Task{}, fmt.Errorf("get inserted task id: %w", err)
	}

	return model.Task{
		ID:    int(id),
		Title: title,
		Done:  done,
	}, nil
}

func (s *TaskService) ListTasks() ([]model.Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, "SELECT id, title, done FROM tasks ORDER BY id")
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	tasks := make([]model.Task, 0)
	for rows.Next() {
		var task model.Task
		var done int
		if err := rows.Scan(&task.ID, &task.Title, &done); err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		task.Done = done == 1
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tasks: %w", err)
	}

	return tasks, nil
}

func boolToSQLite(done bool) int {
	if done {
		return 1
	}

	return 0
}
