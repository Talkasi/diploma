package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"logistics/internal/infrastructure/config"
	"logistics/internal/infrastructure/db"
	"logistics/internal/infrastructure/logger"
)

func main() {
	if err := loadEnv(); err != nil {
		log.Printf("предупреждение: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("ошибка загрузки конфига: %v", err)
	}

	logr, err := logger.New(cfg.AppEnv)
	if err != nil {
		log.Fatalf("ошибка инициализации логгера: %v", err)
	}
	defer logr.Sync()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := db.NewPool(ctx, cfg, logr)
	if err != nil {
		logr.Fatal("не удалось подключиться к БД", zap.Error(err))
	}
	defer pool.Close()

	var version string
	if err := pool.QueryRow(ctx, "SELECT PostGIS_Version() AS version").Scan(&version); err != nil {
		logr.Fatal("не удалось узнать версию PostGIS", zap.Error(err))
	}

	logr.Info("версия PostGIS", zap.String("version", version))
	logr.Info("импортер готов")
}

func loadEnv() error {
	primary := "../../infra/.env"
	if _, err := os.Stat(primary); err == nil {
		return godotenv.Load(primary)
	}
	return godotenv.Load()
}
