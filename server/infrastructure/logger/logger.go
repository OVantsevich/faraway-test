// Package logger - loggers initialization.
package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Development - case of Environment: Dev
func Development(opts ...zap.Option) (*zap.Logger, error) {
	cfg := zap.NewDevelopmentConfig()
	cfg.OutputPaths = []string{"stdout"}
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	return cfg.Build(opts...)
}

// Production - case of Environment: Prod
func Production(opts ...zap.Option) (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()
	cfg.DisableStacktrace = true
	cfg.OutputPaths = []string{"stdout"}
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	return cfg.Build(opts...)
}
