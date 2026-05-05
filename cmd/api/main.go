package main

import (
	"log"
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger/v2"

	_ "github.com/flavio10araujo/GoSimpleRestApi/docs"
	"github.com/flavio10araujo/GoSimpleRestApi/internal/config"
	"github.com/flavio10araujo/GoSimpleRestApi/internal/db"
	"github.com/flavio10araujo/GoSimpleRestApi/internal/handler"
	"github.com/flavio10araujo/GoSimpleRestApi/internal/repository"
	"github.com/flavio10araujo/GoSimpleRestApi/internal/service"
)

// @title           GoSimpleRestApi
// @version         1.0.0
// @description     Simple REST API for managing tasks with pagination
// @termsOfService  http://example.com/terms/

// @contact.name   API Support
// @contact.url    http://example.com/support
// @contact.email  support@example.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /

func main() {
	dbPath := config.GetEnv("DB_PATH", "./data/tasks.db")
	sqliteDB, err := db.OpenSQLite(dbPath)
	if err != nil {
		log.Fatalf("failed to initialize sqlite: %v", err)
	}
	if err := db.EnsureTasksSchema(sqliteDB); err != nil {
		_ = sqliteDB.Close()
		log.Fatalf("failed to ensure sqlite schema: %v", err)
	}
	defer func() {
		if closeErr := sqliteDB.Close(); closeErr != nil {
			log.Printf("failed to close sqlite connection: %v", closeErr)
		}
	}()

	log.Printf("SQLite configured at %s", dbPath)

	paginationConfig := config.LoadPaginationConfig()
	log.Printf("Pagination: default_limit=%d, max_limit=%d", paginationConfig.DefaultLimit, paginationConfig.MaxLimit)

	mux := http.NewServeMux()
	taskRepository := repository.NewSQLiteTaskRepository(sqliteDB)
	taskService := service.NewTaskService(taskRepository)
	taskHandler := handler.NewTaskHandler(taskService, paginationConfig)

	// Swagger UI
	mux.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"),
	))

	mux.HandleFunc("GET /tasks", taskHandler.GetTasks)
	mux.HandleFunc("POST /tasks", taskHandler.CreateTask)
	mux.HandleFunc("PUT /tasks/{id}", taskHandler.UpdateTask)
	mux.HandleFunc("DELETE /tasks/{id}", taskHandler.DeleteTask)

	log.Println("Server running on :8080")
	log.Println("Swagger UI available at http://localhost:8080/swagger/index.html")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
