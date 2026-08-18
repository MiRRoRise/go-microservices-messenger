package logger

import (
	"log/slog"
	"os"
)

type Logger struct {
	*slog.Logger
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

func (l *Logger) Info(msg string, args ...any) {
	l.Logger.Info(msg, args...)
}

func (l *Logger) Error(msg string, err any) {
	l.Logger.Error(msg, "err", err)
}

func (l *Logger) Debug(msg string, args ...any) {
	l.Logger.Debug(msg, args...)
}

func (l *Logger) Fatal(msg string, err any) {
	l.Logger.Error(msg, "err", err)
	os.Exit(1)
}
