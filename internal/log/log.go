// package log provides utilities for working with logs.
package log

import (
	"log/slog"
	"os"
	"strings"
)

// LvlFromEnv returns the log level from the LOG_LEVEL environment variable.
func LvlFromEnv() slog.Level {
	logLevel := os.Getenv("LOG_LEVEL")

	switch strings.ToLower(logLevel) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
