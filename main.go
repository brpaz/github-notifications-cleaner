// Package main provides the entry point for the application.
// It is responsible for initializing the global dependencies and executing the root command.
package main

import (
	"log/slog"
	"os"

	"github.com/brpaz/github-notifications-cleaner/cmd"
	"github.com/brpaz/github-notifications-cleaner/internal/log"
)

func main() {
	// Initialize the logger.
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: log.LvlFromEnv(),
	}))
	slog.SetDefault(logger)

	// Execute the root command.
	rootCmd := cmd.NewRootCmd()
	if err := rootCmd.Execute(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
