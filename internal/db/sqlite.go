package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
					done INTEGER NOT NULL DEFAULT 0,
					created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
					updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);`

	if _, err := sqliteDB.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("ensure tasks schema: %w", err)
	}

	if _, err := sqliteDB.ExecContext(ctx, "ALTER TABLE tasks ADD COLUMN created_at TEXT"); err != nil && !isDuplicateColumnErr(err) {
		return fmt.Errorf("add created_at column: %w", err)
	}
	if _, err := sqliteDB.ExecContext(ctx, "ALTER TABLE tasks ADD COLUMN updated_at TEXT"); err != nil && !isDuplicateColumnErr(err) {
		return fmt.Errorf("add updated_at column: %w", err)
	}

	if _, err := sqliteDB.ExecContext(ctx, `
UPDATE tasks
SET created_at = COALESCE(created_at, CURRENT_TIMESTAMP),
	updated_at = COALESCE(updated_at, CURRENT_TIMESTAMP)
`); err != nil {
		return fmt.Errorf("backfill task timestamps: %w", err)
	}

	return nil
}

func isDuplicateColumnErr(err error) bool {
	return err != nil && strings.Contains(err.Error(), "duplicate column name")
}
