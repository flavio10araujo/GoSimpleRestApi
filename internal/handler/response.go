package handler

import (
	"encoding/json"
	"net/http"

	"github.com/flavio10araujo/GoSimpleRestApi/internal/model"
)

type PaginatedResponse struct {
	Data   []model.Task `json:"data"`
	Total  int          `json:"total"`
	Offset int          `json:"offset"`
	Limit  int          `json:"limit"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func writeErrorJSON(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}
