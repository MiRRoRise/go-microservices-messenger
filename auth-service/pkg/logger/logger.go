package logger

import (
	"log/slog"
	"os"
)

type Logger struct {
	Logger *slog.Logger
}

func New() *Logger {
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	handler := slog.NewJSONHandler(os.Stdout, opts)
	return &Logger{
		Logger: slog.New(handler),
	}
}

func (l *Logger) Info(msg string, data ...interface{}) {
	l.Logger.Info(msg, data...)
}

func (l *Logger) Error(msg string, data interface{}) {
	l.Logger.Error(msg, data)
}

func (l *Logger) Debug(msg string, data interface{}) {
	l.Logger.Debug(msg, data)
}

func (l *Logger) Fatal(msg string, data interface{}) {
	l.Logger.Error(msg, data)
	os.Exit(1)
}
