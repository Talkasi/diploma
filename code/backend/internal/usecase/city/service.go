package city

import (
	"context"

	"logistics/internal/domain"
)

// Service реализует бизнес-логику для работы с городами.
type Service struct {
	repo Repository
}

// NewService создаёт сервис городов.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// ListCities возвращает список городов.
func (s *Service) ListCities(ctx context.Context) ([]domain.City, error) {
	cities, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	return cities, nil
}
