package logger

import "go.uber.org/zap"

// New создаёт zap-логгер в зависимости от окружения.
func New(env string) (*zap.Logger, error) {
	if env == "prod" {
		return zap.NewProduction()
	}
	return zap.NewDevelopment()
}
