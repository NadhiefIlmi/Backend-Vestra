package handler

import (
	"encoding/json"
	"net/http"
)

type PaginationMeta struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

type SuccessEnvelope struct {
	Data       any             `json:"data"`
	Pagination *PaginationMeta `json:"pagination,omitempty"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

type ErrorEnvelope struct {
	Error ErrorDetail `json:"error"`
}

func WriteSuccess(w http.ResponseWriter, statusCode int, data any, pagination *PaginationMeta) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	resp := SuccessEnvelope{
		Data:       data,
		Pagination: pagination,
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func WriteError(w http.ResponseWriter, statusCode int, code string, message string, details any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	resp := ErrorEnvelope{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
	_ = json.NewEncoder(w).Encode(resp)
}

