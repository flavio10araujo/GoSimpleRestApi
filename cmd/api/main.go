package main

import (
	"log"
	"net/http"

	"github.com/flavio10araujo/GoSimpleRestApi/internal/config"
	"github.com/flavio10araujo/GoSimpleRestApi/internal/db"
	"github.com/flavio10araujo/GoSimpleRestApi/internal/handler"
	"github.com/flavio10araujo/GoSimpleRestApi/internal/repository"
	"github.com/flavio10araujo/GoSimpleRestApi/internal/service"
)

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

	mux := http.NewServeMux()
	taskRepository := repository.NewSQLiteTaskRepository(sqliteDB)
	taskService := service.NewTaskService(taskRepository)
	taskHandler := handler.NewTaskHandler(taskService)

	mux.HandleFunc("GET /tasks", taskHandler.GetTasks)
	mux.HandleFunc("POST /tasks", taskHandler.CreateTask)
	mux.HandleFunc("PUT /tasks/{id}", taskHandler.UpdateTask)
	mux.HandleFunc("DELETE /tasks/{id}", taskHandler.DeleteTask)

	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
