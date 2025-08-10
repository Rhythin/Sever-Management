package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// InitLogger sets zap as the global logger and returns the logger
func InitLogger() *zap.Logger {
	zapConfig := zap.NewDevelopmentConfig()
	zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logger, err := zapConfig.Build()
	if err != nil {
		logger, _ = zap.NewDevelopment()
	}

	zap.ReplaceGlobals(logger)
	return logger
}
