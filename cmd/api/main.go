package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	serverConfig, err := config.LoadServerConfig()
	if err != nil {
		log.Fatalf("failed to load server config: %v", err)
	}

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
	taskRepository := repository.NewSQLiteTaskRepository(sqliteDB, serverConfig.QueryTimeout)
	taskService := service.NewTaskService(taskRepository)
	taskHandler := handler.NewTaskHandler(taskService, paginationConfig)

	// Swagger UI
	mux.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	mux.HandleFunc("POST /tasks", taskHandler.CreateTask)
	mux.HandleFunc("GET /tasks", taskHandler.GetTasks)
	mux.HandleFunc("GET /tasks/{id}", taskHandler.GetTask)
	mux.HandleFunc("PUT /tasks/{id}", taskHandler.ReplaceTask)
	mux.HandleFunc("PATCH /tasks/{id}", taskHandler.UpdateTask)
	mux.HandleFunc("DELETE /tasks/{id}", taskHandler.DeleteTask)

	server := &http.Server{
		Addr:              serverConfig.Address(),
		Handler:           mux,
		ReadHeaderTimeout: serverConfig.ReadHeaderTimeout,
		ReadTimeout:       serverConfig.ReadTimeout,
		WriteTimeout:      serverConfig.WriteTimeout,
		IdleTimeout:       serverConfig.IdleTimeout,
	}

	log.Printf("Server running on %s", serverConfig.Address())
	log.Printf("Swagger UI available at http://localhost:%s/swagger/index.html", serverConfig.Port)

	// Channel to capture OS signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a separate goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	sig := <-sigChan
	log.Printf("received signal: %v, initiating graceful shutdown", sig)

	// Graceful shutdown with 30-second timeout for in-flight requests
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("shutdown error: %v", err)
	}

	log.Println("server stopped")
}
