package city

import (
	"context"

	"logistics/internal/domain"
)

// Repository описывает порт доступа к хранилищу городов.
type Repository interface {
	List(ctx context.Context) ([]domain.City, error)
}
