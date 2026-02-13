package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"logistics/internal/adapter/http/dto"
	"logistics/internal/domain"
)

// CityService описывает операции сервиса городов, нужные хендлеру.
type CityService interface {
	ListCities(ctx context.Context) ([]domain.City, error)
}

// CityHandler обрабатывает запросы, связанные с городами.
type CityHandler struct {
	service CityService
	log     *zap.Logger
}

// NewCityHandler создаёт новый хендлер городов.
func NewCityHandler(service CityService, log *zap.Logger) *CityHandler {
	return &CityHandler{service: service, log: log}
}

// HandleList возвращает список городов.
func (h *CityHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cities, err := h.service.ListCities(ctx)
	if err != nil {
		h.log.Error("ошибка получения городов", zap.Error(err))
		http.Error(w, "внутренняя ошибка", http.StatusInternalServerError)
		return
	}

	resp := make([]dto.CityResponse, 0, len(cities))
	for _, c := range cities {
		resp = append(resp, dto.FromDomain(c))
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.log.Error("не удалось сериализовать ответ", zap.Error(err))
	}
}
