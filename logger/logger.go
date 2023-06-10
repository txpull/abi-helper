package logger

import "go.uber.org/zap"

func GetProductionLogger() (*zap.Logger, error) {
	logger, err := zap.NewProduction()
	return logger, err
}

func GetDevelopmentLogger() (*zap.Logger, error) {
	logger, err := zap.NewDevelopment()
	return logger, err
}
