package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger interface defines the logging methods
type Logger interface {
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
}

// logger implements the Logger interface
type logger struct {
	zapLogger *zap.Logger
}

// NewLogger creates a new logger instance
func NewLogger(mode string) Logger {
	var config zap.Config

	if mode == "production" {
		config = zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	zapLogger, err := config.Build()
	if err != nil {
		panic(err)
	}

	return &logger{
		zapLogger: zapLogger,
	}
}

// Debug logs a debug message
func (l *logger) Debug(msg string, fields ...zap.Field) {
	l.zapLogger.Debug(msg, fields...)
}

// Info logs an info message
func (l *logger) Info(msg string, fields ...zap.Field) {
	l.zapLogger.Info(msg, fields...)
}

// Warn logs a warning message
func (l *logger) Warn(msg string, fields ...zap.Field) {
	l.zapLogger.Warn(msg, fields...)
}

// Error logs an error message
func (l *logger) Error(msg string, fields ...zap.Field) {
	l.zapLogger.Error(msg, fields...)
}

// Fatal logs a fatal message and exits
func (l *logger) Fatal(msg string, fields ...zap.Field) {
	l.zapLogger.Fatal(msg, fields...)
}

// Sync flushes any buffered log entries
func (l *logger) Sync() error {
	return l.zapLogger.Sync()
}
