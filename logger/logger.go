package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func GetProductionLogger() (*zap.Logger, error) {
	logger, err := zap.NewProduction()
	return logger, err
}

func GetDevelopmentLogger() (*zap.Logger, error) {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logger, err := config.Build()
	return logger, err
}
