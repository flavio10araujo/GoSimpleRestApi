# GoSimpleRestApi

Simple REST API in Go.

## Current status

- `GET /tasks`, `POST /tasks`, `PUT /tasks/{id}` and `DELETE /tasks/{id}` are implemented.
- Task persistence uses SQLite through repository layer (`internal/repository/sqlite_task_repository.go`).
- The `tasks` table is created automatically at startup.

## Endpoints

- `GET /tasks`
- `POST /tasks`
- `PUT /tasks/{id}`
- `DELETE /tasks/{id}`

## SQLite configuration

The app reads `DB_PATH` from environment.

- Default: `./data/tasks.db`
- If the `data` directory does not exist, it is created automatically.

## Run

```powershell
cd C:\Dev\Projects\flavio\fma-go\GoSimpleRestApi
go run ./...
```

## Run with custom SQLite path

```powershell
cd C:\Dev\Projects\flavio\fma-go\GoSimpleRestApi
$env:DB_PATH = "./data/dev-tasks.db"
go run ./...
```