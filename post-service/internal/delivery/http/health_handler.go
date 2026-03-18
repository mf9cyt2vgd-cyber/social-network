package http

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type HealthResponse struct {
	Status string `json:"status"`
}

type HealthHandler struct {
	log *slog.Logger
}

func NewHealthHandler(log *slog.Logger) *HealthHandler {
	return &HealthHandler{log: log}
}

func (h *HealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {

	response := HealthResponse{Status: "ok"}
	data, _ := json.Marshal(response)

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(data)
}
