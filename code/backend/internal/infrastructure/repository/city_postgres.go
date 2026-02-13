package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"logistics/internal/domain"
	"logistics/internal/usecase/city"
)

// CityPostgres реализует репозиторий городов на PostgreSQL.
type CityPostgres struct {
	pool *pgxpool.Pool
}

// NewCityPostgres создаёт репозиторий.
func NewCityPostgres(pool *pgxpool.Pool) *CityPostgres {
	return &CityPostgres{pool: pool}
}

// List возвращает все города.
func (r *CityPostgres) List(ctx context.Context) ([]domain.City, error) {
	rows, err := r.pool.Query(ctx, "SELECT id, name FROM cities ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cities := make([]domain.City, 0)
	for rows.Next() {
		var c domain.City
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			return nil, err
		}
		cities = append(cities, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return cities, nil
}

var _ city.Repository = (*CityPostgres)(nil)
