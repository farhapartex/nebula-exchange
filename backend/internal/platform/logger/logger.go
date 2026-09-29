package logger

import (
	"io"
	"log/slog"
	"os"
)

func New(level slog.Level, isProduction bool) *slog.Logger {
	return NewWithWriter(os.Stdout, level, isProduction)
}

func NewWithWriter(writer io.Writer, level slog.Level, isProduction bool) *slog.Logger {
	handlerOptions := &slog.HandlerOptions{Level: level}
	if isProduction {
		return slog.New(slog.NewJSONHandler(writer, handlerOptions))
	}
	return slog.New(slog.NewTextHandler(writer, handlerOptions))
}
