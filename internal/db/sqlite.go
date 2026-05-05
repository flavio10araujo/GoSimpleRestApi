package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

func OpenSQLite(path string) (*sql.DB, error) {
	if path == "" {
		return nil, errors.New("sqlite path is required")
	}

	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("create sqlite directory: %w", err)
		}
	}

	sqliteDB, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := sqliteDB.PingContext(ctx); err != nil {
		_ = sqliteDB.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	return sqliteDB, nil
}

func EnsureTasksSchema(sqliteDB *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	const schema = `
CREATE TABLE IF NOT EXISTS tasks (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	title TEXT NOT NULL,
	done INTEGER NOT NULL DEFAULT 0
);`

	if _, err := sqliteDB.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("ensure tasks schema: %w", err)
	}

	return nil
}
