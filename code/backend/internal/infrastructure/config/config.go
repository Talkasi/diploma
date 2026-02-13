package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

// Config хранит параметры приложения, читаемые из окружения.
type Config struct {
	AppEnv      string `envconfig:"APP_ENV" default:"dev"`
	AppHTTPAddr string `envconfig:"APP_HTTP_ADDR" default:":8080"`

	DBHost     string `envconfig:"DB_HOST" default:"localhost"`
	DBPort     int    `envconfig:"DB_PORT" default:"5433"`
	DBName     string `envconfig:"DB_NAME" default:"logistics"`
	DBUser     string `envconfig:"DB_USER" required:"true"`
	DBPassword string `envconfig:"DB_PASSWORD" required:"true"`
	DBSSLMode  string `envconfig:"DB_SSLMODE" default:"disable"`
	DBMaxConns int    `envconfig:"DB_MAX_CONNS" default:"10"`
}

// Load читает переменные окружения и возвращает конфиг.
func Load() (*Config, error) {
	cfg := &Config{}
	if err := envconfig.Process("", cfg); err != nil {
		return nil, fmt.Errorf("не удалось разобрать окружение: %w", err)
	}
	if cfg.DBUser == "" {
		return nil, fmt.Errorf("DB_USER не задан")
	}
	if cfg.DBPassword == "" {
		return nil, fmt.Errorf("DB_PASSWORD не задан")
	}
	return cfg, nil
}

// DSN формирует строку подключения для pgxpool.
func (c *Config) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s", c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode)
}
