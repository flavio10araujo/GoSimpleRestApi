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

### Pagination

#### Listar com defaults (offset=0, limit=20)
GET /tasks

#### Listar página 2 (11-20)
GET /tasks?offset=10&limit=10

#### Listar com limite personalizado
GET /tasks?limit=50

#### Erro: limit acima do máximo (100)
GET /tasks?limit=101  # → 400 Bad Request

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