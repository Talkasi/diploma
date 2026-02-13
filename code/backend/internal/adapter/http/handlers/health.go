package handlers

import (
	"encoding/json"
	"net/http"
)

// HealthHandler отвечает за эндпоинт проверки работоспособности.
type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Handle отдает статус сервиса.
func (h *HealthHandler) Handle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
