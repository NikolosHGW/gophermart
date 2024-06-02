package logger

import (
	"fmt"

	"go.uber.org/zap"
)

func NewLogger(level string) (*zap.Logger, error) {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, fmt.Errorf("не удалось разобрать уровень логирования: %w", err)
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = lvl

	zl, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("не удалось построить конфигурацию логгера: %w", err)
	}

	return zl, nil
}
