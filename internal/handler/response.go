package handler

import (
	"encoding/json"
	"net/http"

	"github.com/flavio10araujo/GoSimpleRestApi/internal/model"
)

type PaginatedResponse struct {
	Data   []TaskResponse `json:"data"`
	Total  int            `json:"total"`
	Offset int            `json:"offset"`
	Limit  int            `json:"limit"`
}

type TaskResponse struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func toTaskResponse(task model.Task) TaskResponse {
	return TaskResponse{ID: task.ID, Title: task.Title, Done: task.Done}
}

func toTaskResponses(tasks []model.Task) []TaskResponse {
	responses := make([]TaskResponse, 0, len(tasks))
	for _, task := range tasks {
		responses = append(responses, toTaskResponse(task))
	}

	return responses
}

func writeErrorJSON(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}
