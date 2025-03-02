// Package cmd contains the command definitions for the application.
package cmd

import (
	"github.com/spf13/cobra"

	"github.com/brpaz/github-notifications-cleaner/cmd/clean"
	"github.com/brpaz/github-notifications-cleaner/cmd/version"
)

// NewRootCmd returns a new instance of the root command for the application
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "github-notifications-cleaner",
		Short: "A CLI tool to clean up GitHub notifications.",
	}

	// Reggister subcommands
	rootCmd.AddCommand(version.NewCmd())
	rootCmd.AddCommand(clean.NewCleanCmd())

	return rootCmd
}
