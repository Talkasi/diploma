package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"logistics/internal/infrastructure/config"
)

// NewPool создаёт пул соединений к PostgreSQL и проверяет подключение.
func NewPool(ctx context.Context, cfg *config.Config, log *zap.Logger) (*pgxpool.Pool, error) {
	pgxCfg, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("разбор DSN: %w", err)
	}
	pgxCfg.MaxConns = int32(cfg.DBMaxConns)

	pool, err := pgxpool.NewWithConfig(ctx, pgxCfg)
	if err != nil {
		return nil, fmt.Errorf("создание пула: %w", err)
	}

	ctxPing, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(ctxPing); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping базы: %w", err)
	}

	log.Info("подключение к БД установлено", zap.String("host", cfg.DBHost), zap.String("db", cfg.DBName))
	return pool, nil
}
