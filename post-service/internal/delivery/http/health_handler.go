// Package http provides handlers and so on
package http

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// HealthResponse is a struct with minimum info to recognize health of http-connection to app
type HealthResponse struct {
	Status string `json:"status"`
}

// HealthHandler
type HealthHandler struct {
	log *slog.Logger
}

// NewHealthHandler
func NewHealthHandler(log *slog.Logger) *HealthHandler {
	return &HealthHandler{log: log}
}

// HealthCheck
func (h *HealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {

	response := HealthResponse{Status: "ok"}
	data, _ := json.Marshal(response)

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(data)
}
