# FirstApp

Simple REST API in Go.

## Current status

- `GET /tasks` and `POST /tasks` are implemented.
- Task persistence is still in-memory (`internal/service/task_service.go`).
- SQLite is already configured at startup, but not used by the service yet.

## SQLite configuration

The app reads `DB_PATH` from environment.

- Default: `./data/tasks.db`
- If the `data` directory does not exist, it is created automatically.

## Run

```powershell
cd C:\Dev\Projects\flavio\fma-go\FirstApp
go run ./...
```

## Run with custom SQLite path

```powershell
cd C:\Dev\Projects\flavio\fma-go\FirstApp
$env:DB_PATH = "./data/dev-tasks.db"
go run ./...
```