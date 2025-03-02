package cleaner

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/go-github/v69/github"
)

const (
	TypeIssue             = "Issue"
	TypePullRequest       = "PullRequest"
	DefaultDaysThereshold = 15
)

// Cleaner defines the interface for cleaning notifications.
type Cleaner interface {
	Clean(ctx context.Context) error
}

// NotificationsCleaner defines the cleaner struct.
type NotificationsCleaner struct {
	GitHubClient  *github.Client
	OlderThanDays int
	DryRun        bool
}

// Option defines a functional option for NotificationsCleaner.
type Option func(*NotificationsCleaner)

// NewNotificationsCleaner creates a new NotificationsCleaner instance
// with the provided options. It initializes with a default GitHubClient.
func NewNotificationsCleaner(opts ...Option) *NotificationsCleaner {
	nc := &NotificationsCleaner{
		GitHubClient:  github.NewClient(nil),
		OlderThanDays: DefaultDaysThereshold,
		DryRun:        false,
	}

	for _, opt := range opts {
		opt(nc)
	}

	return nc
}

// WithGitHubClient is an option to set a custom GitHub client.
func WithGitHubClient(client *github.Client) Option {
	return func(nc *NotificationsCleaner) {
		nc.GitHubClient = client
	}
}

// WithOlderThanDays is an option to set the age threshold (in days)
// for cleaning notifications.
func WithOlderThanDays(days int) Option {
	return func(nc *NotificationsCleaner) {
		nc.OlderThanDays = days
	}
}

// WithDryRun is an option to enable dry-run mode.
func WithDryRun(dryRun bool) Option {
	return func(nc *NotificationsCleaner) {
		nc.DryRun = dryRun
	}
}

// Clean performs cleaning notifications.
// It marks notifications as done if they are related to closed pull requests/issues
// or if they are older than the configured number of days.
func (nc *NotificationsCleaner) Clean(ctx context.Context) error {
	notifications, _, err := nc.GitHubClient.Activity.ListNotifications(ctx, &github.NotificationListOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	})
	if err != nil {
		return fmt.Errorf("error listing notifications: %w", err)
	}

	threshold := time.Now().AddDate(0, 0, -nc.OlderThanDays)

	// TODO Process in parallel with a worker pool
	for _, n := range notifications {
		markDone, err := nc.canBeMarkedAsDone(ctx, n, threshold)
		if err != nil {
			slog.Error("error checking notification",
				slog.String("notification_id", n.GetID()),
				slog.String("error", err.Error()),
			)
			continue
		}

		if markDone {
			slog.Info("found notification to mark as done",
				slog.String("notification_id", n.GetID()),
				slog.String("subject", n.GetSubject().GetTitle()),
				slog.String("type", n.GetSubject().GetType()),
			)

			if nc.DryRun {
				slog.Info("dry-run mode enabled. Skipping marking notification as done.")
				continue
			}

			_, err := nc.GitHubClient.Activity.MarkThreadRead(ctx, n.GetID())
			if err != nil {
				// Log error but continue processing other notifications.
				slog.Error("error marking notification as done",
					slog.String("notification_id", n.GetID()),
					slog.String("error", err.Error()),
				)
			}
		}
	}

	return nil
}

// canBeMarkedAsDone checks if a notification should be marked as done
// nolint: gocyclo
func (nc *NotificationsCleaner) canBeMarkedAsDone(ctx context.Context, n *github.Notification, threshold time.Time) (bool, error) {
	// Rule 1: Check notiications older than the threshold
	if n.UpdatedAt != nil && n.UpdatedAt.Time.Before(threshold) {
		fmt.Printf("Notification %s is older than the threshold\n", n.GetID())
		return true, nil
	}

	// Rule 2: Check if the notification is related to a closed issue or pull request
	subjectType := n.GetSubject().GetType()
	if subjectType == TypeIssue || subjectType == TypePullRequest {
		owner, repo, number, err := parseNotificationURL(n.GetSubject().GetURL())
		if err != nil {
			return false, fmt.Errorf("error parsing notification URL for notification %s: %w", n.GetID(), err)
		}

		switch subjectType {
		case TypePullRequest:
			pr, _, err := nc.GitHubClient.PullRequests.Get(ctx, owner, repo, number)
			if err != nil {
				return false, fmt.Errorf("error fetching pull request %s/%s#%d: %w", owner, repo, number, err)
			}

			if pr.GetState() == "closed" {
				return true, nil
			}
		case TypeIssue:
			issue, _, err := nc.GitHubClient.Issues.Get(ctx, owner, repo, number)
			if err != nil {
				return false, fmt.Errorf("error fetching issue %s/%s#%d: %w", owner, repo, number, err)
			}

			if issue.GetState() == "closed" {
				return true, nil
			}
		}
	}

	return false, nil
}
