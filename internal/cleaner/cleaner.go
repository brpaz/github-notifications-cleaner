package cleaner

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/google/go-github/v69/github"
)

const (
	TypeIssue            = "Issue"
	TypePullRequest      = "PullRequest"
	DefaultDaysThreshold = 30
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
		OlderThanDays: DefaultDaysThreshold,
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
	threshold := time.Now().AddDate(0, 0, -nc.OlderThanDays)

	opts := &github.NotificationListOptions{
		All: true,
		ListOptions: github.ListOptions{
			PerPage: 100,
			Page:    1,
		},
	}

	allNotications := make([]*github.Notification, 0)
	for {
		slog.Info("fetching notifications",
			slog.Int("page", opts.Page),
		)
		notifications, resp, err := nc.GitHubClient.Activity.ListNotifications(ctx, opts)
		if err != nil {
			return fmt.Errorf("error listing notifications: %w", err)
		}

		allNotications = append(allNotications, notifications...)

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	for _, n := range allNotications {
		nc.processNotification(ctx, n, threshold)
	}

	return nil
}

// processNotification processes a single notification.
func (nc *NotificationsCleaner) processNotification(ctx context.Context, n *github.Notification, threshold time.Time) {
	markDone, err := nc.canBeMarkedAsDone(ctx, n, threshold)
	if err != nil {
		slog.Error("error checking notification",
			slog.String("notification_id", n.GetID()),
			slog.String("error", err.Error()),
		)
		return
	}

	if !markDone {
		return
	}

	slog.Info("marking notification as done",
		slog.String("id", n.GetID()),
		slog.String(("repository"), n.GetRepository().GetFullName()),
		slog.String("subject", n.GetSubject().GetTitle()),
	)

	if nc.DryRun {
		slog.Debug("dry-run mode enabled. Skipping marking notification as done.")
		return
	}

	nID, err := strconv.Atoi(n.GetID())
	if err != nil {
		slog.Error("error converting notification ID to int",
			slog.String("notification_id", n.GetID()),
			slog.String("error", err.Error()),
		)
		return
	}

	_, err = nc.GitHubClient.Activity.MarkThreadDone(ctx, int64(nID))
	if err != nil {
		// Log error but continue processing other notifications.
		slog.Error("error marking notification as done",
			slog.String("notification_id", n.GetID()),
			slog.String("error", err.Error()),
		)
	}
}

// canBeMarkedAsDone checks if a notification should be marked as done
// nolint: gocyclo
func (nc *NotificationsCleaner) canBeMarkedAsDone(ctx context.Context, n *github.Notification, threshold time.Time) (bool, error) {
	// Rule 1: Check notiications older than the threshold
	if n.UpdatedAt != nil && n.UpdatedAt.Time.Before(threshold) {
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
