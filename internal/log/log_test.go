package log_test

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/brpaz/github-notifications-cleaner/internal/log"
)

func TestLvlFromEnv(t *testing.T) {
	testCases := []struct {
		name     string
		envValue string
		expected slog.Level
	}{
		{"Debug", "debug", slog.LevelDebug},
		{"Info", "info", slog.LevelInfo},
		{"Warn", "warn", slog.LevelWarn},
		{"Error", "error", slog.LevelError},
		{"Empty", "", slog.LevelInfo},
		{"Invalid", "invalid", slog.LevelInfo},
		{"Mixed Case", "Debug", slog.LevelDebug},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("LOG_LEVEL", tc.envValue)
			level := log.LvlFromEnv()
			assert.Equal(t, tc.expected, level)
		})
	}
}
