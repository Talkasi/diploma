package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"

	httpAdapter "logistics/internal/adapter/http"
	"logistics/internal/adapter/http/handlers"
	"logistics/internal/infrastructure/config"
	"logistics/internal/infrastructure/db"
	"logistics/internal/infrastructure/logger"
	"logistics/internal/infrastructure/repository"
	"logistics/internal/usecase/city"
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

	cityRepo := repository.NewCityPostgres(pool)
	citySvc := city.NewService(cityRepo)

	healthHandler := handlers.NewHealthHandler()
	cityHandler := handlers.NewCityHandler(citySvc, logr)
	router := httpAdapter.NewRouter(healthHandler, cityHandler)

	if err := runServer(ctx, cfg.AppHTTPAddr, router.Handler(), logr); err != nil {
		logr.Fatal("сервер завершился с ошибкой", zap.Error(err))
	}
}

func loadEnv() error {
	primary := "../../infra/.env"
	if _, err := os.Stat(primary); err == nil {
		return godotenv.Load(primary)
	}
	return godotenv.Load()
}

func runServer(ctx context.Context, addr string, handler http.Handler, log *zap.Logger) error {
	srv := &http.Server{Addr: addr, Handler: handler}

	errCh := make(chan error, 1)
	go func() {
		log.Info("HTTP сервер запущен", zap.String("addr", addr))
		errCh <- srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}
