package logger

import (
	"log/slog"
	"os"
	"strings"
)

func NewLogger(lvl string) *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     selectLogLevel(lvl),
		AddSource: true,
	}))
}

func selectLogLevel(lvl string) slog.Level {
	logLevel := strings.ToUpper(lvl)

	switch logLevel {
	case "INFO":
		return slog.LevelInfo
	case "ERROR":
		return slog.LevelError
	case "WARN":
		return slog.LevelWarn
	case "DEBUG":
		return slog.LevelDebug
	}

	return slog.LevelInfo
}
