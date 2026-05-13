# GoSimpleRestApi

Simple REST API in Go.

## Current status

- `GET /tasks`, `GET /tasks/{id}`, `POST /tasks`, `PUT /tasks/{id}`, `PATCH /tasks/{id}` and `DELETE /tasks/{id}` are implemented.
- Task persistence uses SQLite through repository layer (`internal/repository/sqlite_task_repository.go`).
- The `tasks` table is created automatically at startup.
- Swagger / OpenAPI is available through `swaggo/swag`.

## Endpoints

- `GET /tasks`
- `GET /tasks/{id}`
- `POST /tasks`
- `PUT /tasks/{id}`
- `PATCH /tasks/{id}`
- `DELETE /tasks/{id}`

### Pagination

#### List with defaults (offset=0, limit=20)
GET /tasks

#### List page 2 (11-20)
GET /tasks?offset=10&limit=10

#### List with custom limit
GET /tasks?limit=50

#### Error: limit above maximum (100)
GET /tasks?limit=101  # → 400 Bad Request

## Swagger / OpenAPI

The project exposes interactive API documentation through Swagger UI.

### URLs

- Swagger UI: `http://localhost:8080/swagger/index.html`
- OpenAPI JSON: `http://localhost:8080/swagger/doc.json`

### How to use

1. Start the application.
2. Open the Swagger UI URL in your browser.
3. Use **Try it out** to test the endpoints directly from the interface.

### Generate / update the documentation

Whenever you change Swagger annotations in files such as `cmd/api/main.go` or `internal/handler/task_handler.go`, regenerate the files in the `docs/` folder:

```powershell
go run github.com/swaggo/swag/cmd/swag init -g cmd/api/main.go --output docs
```

Generated files:

- `docs/docs.go`
- `docs/swagger.json`
- `docs/swagger.yaml`

### Notes

Since `main.go` imports the `docs` package, it is recommended to version the `docs/` folder in the repository to avoid build issues in other environments.

## SQLite configuration

The app reads `DB_PATH` from environment.

- Default: `./data/tasks.db`
- If the `data` directory does not exist, it is created automatically.

## Run

```powershell
go run ./...
```

After starting the application, open:

```text
http://localhost:8080/swagger/index.html
```

## Run with custom SQLite path

```powershell
$env:DB_PATH = "./data/dev-tasks.db"
go run ./...
```