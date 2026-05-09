package controllers

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type DetailedErrorResponse struct {
	Error   string      `json:"error"`
	Code    string      `json:"code,omitempty"`
	Details interface{} `json:"details,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, ErrorResponse{Error: msg})
}

func writeDetailedError(w http.ResponseWriter, status int, code string, msg string, details interface{}) {
	writeJSON(w, status, DetailedErrorResponse{
		Error:   msg,
		Code:    code,
		Details: details,
	})
}
