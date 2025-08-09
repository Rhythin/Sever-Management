package logging

import (
	"go.uber.org/zap"
)

// InitLogger sets zap as the global logger and returns the logger
func InitLogger() *zap.Logger {
	logger, err := zap.NewProduction()
	if err != nil {
		logger, _ = zap.NewDevelopment()
	}
	zap.ReplaceGlobals(logger)
	return logger
}
