package main

import (
	"log/slog"
	"os"
)

var logger *slog.Logger

func InitLogger(lvl string) {

	var log_level slog.Level

	switch lvl {
	case "DEBUG":
		log_level = slog.LevelDebug
	case "WARN":
		log_level = slog.LevelWarn
	case "ERROR":
		log_level = slog.LevelError
	default:
		log_level = slog.LevelInfo

	}

	logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: log_level}))

}
