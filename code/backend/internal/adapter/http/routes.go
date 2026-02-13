package http

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"logistics/internal/adapter/http/handlers"
)

// Router строит HTTP-маршруты.
type Router struct {
	health *handlers.HealthHandler
	cities *handlers.CityHandler
}

// NewRouter создаёт роутер с зависимостями хендлеров.
func NewRouter(health *handlers.HealthHandler, cities *handlers.CityHandler) *Router {
	return &Router{health: health, cities: cities}
}

// Handler возвращает готовый http.Handler.
func (r *Router) Handler() http.Handler {
	chiRouter := chi.NewRouter()

	chiRouter.Use(middleware.RequestID)
	chiRouter.Use(middleware.RealIP)
	chiRouter.Use(middleware.Recoverer)
	chiRouter.Use(middleware.Timeout(15 * time.Second))

	chiRouter.Get("/health", r.health.Handle)
	chiRouter.Get("/v1/cities", r.cities.HandleList)

	return chiRouter
}
