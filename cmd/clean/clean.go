package clean

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/google/go-github/v69/github"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"

	"github.com/brpaz/github-notifications-cleaner/internal/cleaner"
)

const (
	flagToken  = "token"
	flagDays   = "days-threshold"
	flagDryRun = "dry-run"
)

// Cleaner defines the interface for the service that cleans up notifications.
type Cleaner interface {
	Clean(ctx context.Context) error
}

// NewCleanCmd creates a new instance of the clean command.
func NewCleanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "clean",
		Short:   "Cleans up GitHub notifications.",
		Example: `github-notifications-cleaner clean --token <GITHUB_TOKEN> --days-threshold 15`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Flags().Changed(flagToken) {
				return nil
			}
			_ = cmd.Flags().Set(flagToken, os.Getenv("GITHUB_TOKEN"))
			return nil
		},
		RunE: run,
	}

	cmd.Flags().StringP(flagToken, "t", "", "GitHub Personal Access Token with notifications access")
	cmd.Flags().IntP(flagDays, "d", cleaner.DefaultDaysThreshold, "Mark notifications older than this number of days as done.")
	cmd.Flags().BoolP(flagDryRun, "n", false, "Dry run mode")

	_ = cmd.MarkFlagRequired(flagToken)
	return cmd
}

// initCleaner initializes the Cleaner using the GitHub token flag.
func initCleaner(cmd *cobra.Command) (Cleaner, context.Context, error) {
	githubToken, err := cmd.Flags().GetString(flagToken)
	if err != nil {
		return nil, nil, err
	}
	if githubToken == "" {
		return nil, nil, fmt.Errorf("GitHub token is required")
	}

	daysThreshold, err := cmd.Flags().GetInt(flagDays)
	if err != nil {
		return nil, nil, err
	}

	dryRun, err := cmd.Flags().GetBool(flagDryRun)
	if err != nil {
		return nil, nil, err
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: githubToken})
	tc := oauth2.NewClient(ctx, ts)

	ghClient := github.NewClient(tc)
	nc := cleaner.NewNotificationsCleaner(
		cleaner.WithGitHubClient(ghClient),
		cleaner.WithOlderThanDays(daysThreshold),
		cleaner.WithDryRun(dryRun),
	)
	return nc, ctx, nil
}

func run(cmd *cobra.Command, args []string) error {
	cleanerInstance, ctx, err := initCleaner(cmd)
	if err != nil {
		return err
	}

	if err := cleanerInstance.Clean(ctx); err != nil {
		return fmt.Errorf("error cleaning notifications: %w", err)
	}

	slog.Info("Notifications cleaned successfully.")
	return nil
}
