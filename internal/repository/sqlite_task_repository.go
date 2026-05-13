package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/flavio10araujo/GoSimpleRestApi/internal/model"
)

const sqliteDateTimeLayout = "2006-01-02 15:04:05"

type SQLiteTaskRepository struct {
	db           *sql.DB
	queryTimeout time.Duration
}

func NewSQLiteTaskRepository(db *sql.DB, queryTimeout time.Duration) *SQLiteTaskRepository {
	return &SQLiteTaskRepository{db: db, queryTimeout: queryTimeout}
}

func (r *SQLiteTaskRepository) AddTask(ctx context.Context, title string, done bool) (model.Task, error) {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	result, err := r.db.ExecContext(
		ctx,
		"INSERT INTO tasks (title, done, created_at, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)",
		title,
		boolToSQLite(done),
	)
	if err != nil {
		return model.Task{}, fmt.Errorf("insert task: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return model.Task{}, fmt.Errorf("get inserted task id: %w", err)
	}

	task, err := r.GetTask(ctx, int(id))
	if err != nil {
		return model.Task{}, fmt.Errorf("fetch inserted task: %w", err)
	}

	return task, nil
}

func (r *SQLiteTaskRepository) GetTask(ctx context.Context, id int) (model.Task, error) {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	var task model.Task
	var done int
	var createdAtRaw string
	var updatedAtRaw string
	err := r.db.QueryRowContext(
		ctx,
		"SELECT id, title, done, created_at, updated_at FROM tasks WHERE id = ?",
		id,
	).Scan(&task.ID, &task.Title, &done, &createdAtRaw, &updatedAtRaw)
	if err == sql.ErrNoRows {
		return model.Task{}, ErrTaskNotFound
	}
	if err != nil {
		return model.Task{}, fmt.Errorf("get task: %w", err)
	}
	task.Done = done == 1
	task.CreatedAt, err = parseSQLiteTimestamp(createdAtRaw)
	if err != nil {
		return model.Task{}, fmt.Errorf("parse created_at: %w", err)
	}
	task.UpdatedAt, err = parseSQLiteTimestamp(updatedAtRaw)
	if err != nil {
		return model.Task{}, fmt.Errorf("parse updated_at: %w", err)
	}

	return task, nil
}

func (r *SQLiteTaskRepository) ListTasks(ctx context.Context, offset, limit int) ([]model.Task, int, error) {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	var total int
	countErr := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM tasks").Scan(&total)
	if countErr != nil {
		return nil, 0, fmt.Errorf("count tasks: %w", countErr)
	}

	rows, err := r.db.QueryContext(ctx, "SELECT id, title, done, created_at, updated_at FROM tasks ORDER BY id LIMIT ? OFFSET ?", limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("query tasks: %w", err)
	}
	defer rows.Close()

	tasks := make([]model.Task, 0)
	for rows.Next() {
		var task model.Task
		var done int
		var createdAtRaw string
		var updatedAtRaw string
		if err := rows.Scan(&task.ID, &task.Title, &done, &createdAtRaw, &updatedAtRaw); err != nil {
			return nil, 0, fmt.Errorf("scan task: %w", err)
		}
		task.Done = done == 1
		task.CreatedAt, err = parseSQLiteTimestamp(createdAtRaw)
		if err != nil {
			return nil, 0, fmt.Errorf("parse created_at: %w", err)
		}
		task.UpdatedAt, err = parseSQLiteTimestamp(updatedAtRaw)
		if err != nil {
			return nil, 0, fmt.Errorf("parse updated_at: %w", err)
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate tasks: %w", err)
	}

	return tasks, total, nil
}

func (r *SQLiteTaskRepository) UpdateTask(ctx context.Context, id int, title string, done bool) (model.Task, error) {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
	defer cancel()

	result, err := r.db.ExecContext(ctx, "UPDATE tasks SET title = ?, done = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", title, boolToSQLite(done), id)
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

	task, err := r.GetTask(ctx, id)
	if err != nil {
		return model.Task{}, fmt.Errorf("fetch updated task: %w", err)
	}

	return task, nil
}

func (r *SQLiteTaskRepository) DeleteTask(ctx context.Context, id int) error {
	ctx, cancel := context.WithTimeout(ctx, r.queryTimeout)
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

func parseSQLiteTimestamp(raw string) (time.Time, error) {
	if t, err := time.Parse(sqliteDateTimeLayout, raw); err == nil {
		return t.UTC(), nil
	}
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return t.UTC(), nil
	}
	if t, err := time.Parse(time.RFC3339Nano, raw); err == nil {
		return t.UTC(), nil
	}

	return time.Time{}, fmt.Errorf("unsupported timestamp format: %q", raw)
}
