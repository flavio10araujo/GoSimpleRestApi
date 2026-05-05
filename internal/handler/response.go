package handler

import "github.com/flavio10araujo/GoSimpleRestApi/internal/model"

type PaginatedResponse struct {
	Data   []model.Task `json:"data"`
	Total  int          `json:"total"`
	Offset int          `json:"offset"`
	Limit  int          `json:"limit"`
}
