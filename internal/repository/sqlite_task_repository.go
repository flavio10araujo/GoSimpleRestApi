package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/flavio10araujo/GoSimpleRestApi/internal/model"
)

type SQLiteTaskRepository struct {
	db *sql.DB
}

func NewSQLiteTaskRepository(db *sql.DB) *SQLiteTaskRepository {
	return &SQLiteTaskRepository{db: db}
}

func (r *SQLiteTaskRepository) AddTask(title string, done bool) (model.Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	result, err := r.db.ExecContext(ctx, "INSERT INTO tasks (title, done) VALUES (?, ?)", title, boolToSQLite(done))
	if err != nil {
		return model.Task{}, fmt.Errorf("insert task: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return model.Task{}, fmt.Errorf("get inserted task id: %w", err)
	}

	return model.Task{ID: int(id), Title: title, Done: done}, nil
}

func (r *SQLiteTaskRepository) GetTask(id int) (model.Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var task model.Task
	var done int
	err := r.db.QueryRowContext(ctx, "SELECT id, title, done FROM tasks WHERE id = ?", id).Scan(&task.ID, &task.Title, &done)
	if err == sql.ErrNoRows {
		return model.Task{}, ErrTaskNotFound
	}
	if err != nil {
		return model.Task{}, fmt.Errorf("get task: %w", err)
	}
	task.Done = done == 1

	return task, nil
}

func (r *SQLiteTaskRepository) ListTasks(offset, limit int) ([]model.Task, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var total int
	countErr := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM tasks").Scan(&total)
	if countErr != nil {
		return nil, 0, fmt.Errorf("count tasks: %w", countErr)
	}

	rows, err := r.db.QueryContext(ctx, "SELECT id, title, done FROM tasks ORDER BY id LIMIT ? OFFSET ?", limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("query tasks: %w", err)
	}
	defer rows.Close()

	tasks := make([]model.Task, 0)
	for rows.Next() {
		var task model.Task
		var done int
		if err := rows.Scan(&task.ID, &task.Title, &done); err != nil {
			return nil, 0, fmt.Errorf("scan task: %w", err)
		}
		task.Done = done == 1
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate tasks: %w", err)
	}

	return tasks, total, nil
}

func (r *SQLiteTaskRepository) UpdateTask(id int, title string, done bool) (model.Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	result, err := r.db.ExecContext(ctx, "UPDATE tasks SET title = ?, done = ? WHERE id = ?", title, boolToSQLite(done), id)
	if err != nil {
		return model.Task{}, fmt.Errorf("update task: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return model.Task{}, fmt.Errorf("get updated rows: %w", err)
	}
	if rowsAffected == 0 {
		return model.Task{}, ErrTaskNotFound
	}

	return model.Task{ID: id, Title: title, Done: done}, nil
}

func (r *SQLiteTaskRepository) DeleteTask(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	result, err := r.db.ExecContext(ctx, "DELETE FROM tasks WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get deleted rows: %w", err)
	}
	if rowsAffected == 0 {
		return ErrTaskNotFound
	}

	return nil
}

func boolToSQLite(done bool) int {
	if done {
		return 1
	}

	return 0
}
