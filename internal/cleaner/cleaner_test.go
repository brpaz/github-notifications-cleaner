package cleaner_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-github/v69/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"

	"github.com/brpaz/github-notifications-cleaner/internal/cleaner"
)

func setupMockClient(t *testing.T) *github.Client {
	httpClient := &http.Client{}
	gock.InterceptClient(httpClient)
	return github.NewClient(httpClient)
}

func TestNewNotificationsCleaner(t *testing.T) {
	t.Run("default initialization sets expected values", func(t *testing.T) {
		nc := cleaner.NewNotificationsCleaner()
		assert.NotNil(t, nc.GitHubClient, "expected default GitHubClient to be non-nil")
		assert.Equal(t, cleaner.DefaultDaysThereshold, nc.OlderThanDays, "expected default OlderThanDays value")
		assert.False(t, nc.DryRun, "expected default DryRun to be false")
	})

	t.Run("WithGitHubClient option sets the client", func(t *testing.T) {
		customHTTPClient := &http.Client{}
		customClient := github.NewClient(customHTTPClient)

		nc := cleaner.NewNotificationsCleaner(cleaner.WithGitHubClient(customClient))
		assert.Equal(t, customClient, nc.GitHubClient, "expected GitHubClient to be the custom one")
	})

	t.Run("WithOlderThanDays option sets the threshold", func(t *testing.T) {
		customDays := 30
		nc := cleaner.NewNotificationsCleaner(cleaner.WithOlderThanDays(customDays))
		assert.Equal(t, customDays, nc.OlderThanDays, "expected OlderThanDays to be set to custom value")
	})

	t.Run("WithDryRun option enables dry-run mode", func(t *testing.T) {
		nc := cleaner.NewNotificationsCleaner(cleaner.WithDryRun(true))
		assert.True(t, nc.DryRun, "expected DryRun to be enabled")
	})
}

func TestClean(t *testing.T) {
	t.Run("notification filtering", func(t *testing.T) {
		t.Run("marks old notifications as done", func(t *testing.T) {
			defer gock.Off()

			// Setup fixtures
			oldDate := time.Now().AddDate(0, 0, -20)
			recentDate := time.Now()

			// Mock API responses
			gock.New("https://api.github.com").
				Get("/notifications").
				Reply(200).
				JSON([]*github.Notification{
					{
						ID:        github.Ptr("old-1"),
						UpdatedAt: &github.Timestamp{Time: oldDate},
						Subject: &github.NotificationSubject{
							Title: github.Ptr("Old Notification"),
							Type:  github.Ptr(cleaner.TypePullRequest),
						},
					},
					{
						ID:        github.Ptr("recent-1"),
						UpdatedAt: &github.Timestamp{Time: recentDate},
						Subject: &github.NotificationSubject{
							Title: github.Ptr("Recent Notification"),
							Type:  github.Ptr(cleaner.TypePullRequest),
						},
					},
				})

			gock.New("https://api.github.com").
				Patch("/notifications/threads/old-1").
				Reply(205)

			// Test execution
			githubClient := setupMockClient(t)
			nc := cleaner.NewNotificationsCleaner(
				cleaner.WithGitHubClient(githubClient),
				cleaner.WithOlderThanDays(15),
			)

			err := nc.Clean(context.Background())

			// Assertions
			require.NoError(t, err)
			assert.True(t, gock.IsDone(), "expected all API mocks to be called")
		})

		t.Run("marks closed PR notifications as done", func(t *testing.T) {
			defer gock.Off()

			gock.New("https://api.github.com").
				Get("/notifications").
				Reply(200).
				JSON([]*github.Notification{
					{
						ID:        github.Ptr("pr-closed"),
						UpdatedAt: &github.Timestamp{Time: time.Now()},
						Subject: &github.NotificationSubject{
							Title: github.Ptr("Closed Pull Request"),
							Type:  github.Ptr(cleaner.TypePullRequest),
							URL:   github.Ptr("https://api.github.com/repos/owner/repo/pulls/123"),
						},
					},
				})

			gock.New("https://api.github.com").
				Get("/repos/owner/repo/pulls/123").
				Reply(200).
				JSON(map[string]any{
					"state":  "closed",
					"title":  "Closed Pull Request",
					"number": 123,
				})

			gock.New("https://api.github.com").
				Patch("/notifications/threads/pr-closed").
				Reply(205)

			githubClient := setupMockClient(t)
			nc := cleaner.NewNotificationsCleaner(
				cleaner.WithGitHubClient(githubClient),
				cleaner.WithOlderThanDays(15),
			)

			err := nc.Clean(context.Background())
			require.NoError(t, err)
			assert.True(t, gock.IsDone())
		})

		t.Run("marks closed issue notifications as done", func(t *testing.T) {
			defer gock.Off()

			gock.New("https://api.github.com").
				Get("/notifications").
				Reply(200).
				JSON([]*github.Notification{
					{
						ID:        github.Ptr("issue-closed"),
						UpdatedAt: &github.Timestamp{Time: time.Now()},
						Subject: &github.NotificationSubject{
							Title: github.Ptr("Closed Issue"),
							Type:  github.Ptr(cleaner.TypeIssue),
							URL:   github.Ptr("https://api.github.com/repos/owner/repo/issues/456"),
						},
					},
				})

			gock.New("https://api.github.com").
				Get("/repos/owner/repo/issues/456").
				Reply(200).
				JSON(map[string]any{
					"state":  "closed",
					"title":  "Closed Issue",
					"number": 456,
				})

			gock.New("https://api.github.com").
				Patch("/notifications/threads/issue-closed").
				Reply(205)

			githubClient := setupMockClient(t)
			nc := cleaner.NewNotificationsCleaner(
				cleaner.WithGitHubClient(githubClient),
				cleaner.WithOlderThanDays(15),
			)

			err := nc.Clean(context.Background())
			require.NoError(t, err)
			assert.True(t, gock.IsDone())
		})

		t.Run("does not mark open issue notifications as done", func(t *testing.T) {
			defer gock.Off()

			gock.New("https://api.github.com").
				Get("/notifications").
				Reply(200).
				JSON([]*github.Notification{
					{
						ID:        github.Ptr("issue-open"),
						UpdatedAt: &github.Timestamp{Time: time.Now()},
						Subject: &github.NotificationSubject{
							Title: github.Ptr("Open Issue"),
							Type:  github.Ptr(cleaner.TypeIssue),
							URL:   github.Ptr("https://api.github.com/repos/owner/repo/issues/789"),
						},
					},
				})

			gock.New("https://api.github.com").
				Get("/repos/owner/repo/issues/789").
				Reply(200).
				JSON(map[string]any{
					"state":  "open",
					"title":  "Open Issue",
					"number": 789,
				})

			// No MarkThreadRead call expected

			githubClient := setupMockClient(t)
			nc := cleaner.NewNotificationsCleaner(
				cleaner.WithGitHubClient(githubClient),
				cleaner.WithOlderThanDays(15),
			)

			err := nc.Clean(context.Background())
			require.NoError(t, err)
			assert.True(t, gock.IsDone())
		})
	})

	t.Run("settings behavior", func(t *testing.T) {
		t.Run("respects dry-run mode", func(t *testing.T) {
			defer gock.Off()

			gock.New("https://api.github.com").
				Get("/notifications").
				Reply(200).
				JSON([]*github.Notification{
					{
						ID:        github.Ptr("dryrun-test"),
						UpdatedAt: &github.Timestamp{Time: time.Now().AddDate(0, 0, -20)},
						Subject: &github.NotificationSubject{
							Title: github.Ptr("Old PR"),
							Type:  github.Ptr(cleaner.TypePullRequest),
						},
					},
				})

			// No MarkThreadRead call expected in dry-run mode

			githubClient := setupMockClient(t)
			nc := cleaner.NewNotificationsCleaner(
				cleaner.WithGitHubClient(githubClient),
				cleaner.WithOlderThanDays(15),
				cleaner.WithDryRun(true),
			)

			err := nc.Clean(context.Background())
			require.NoError(t, err)
			assert.True(t, gock.IsDone())
		})
	})

	t.Run("error handling", func(t *testing.T) {
		t.Run("handles API errors with listing notifications", func(t *testing.T) {
			defer gock.Off()

			gock.New("https://api.github.com").
				Get("/notifications").
				Reply(500).
				JSON(map[string]string{
					"message": "Internal Server Error",
				})

			githubClient := setupMockClient(t)
			nc := cleaner.NewNotificationsCleaner(
				cleaner.WithGitHubClient(githubClient),
			)

			err := nc.Clean(context.Background())
			require.Error(t, err)
			assert.Contains(t, err.Error(), "error listing notifications")
		})

		t.Run("handles API errors with fetching PR details", func(t *testing.T) {
			defer gock.Off()

			gock.New("https://api.github.com").
				Get("/notifications").
				Reply(200).
				JSON([]*github.Notification{
					{
						ID:        github.Ptr("pr-error"),
						UpdatedAt: &github.Timestamp{Time: time.Now()},
						Subject: &github.NotificationSubject{
							Title: github.Ptr("PR with Error"),
							Type:  github.Ptr(cleaner.TypePullRequest),
							URL:   github.Ptr("https://api.github.com/repos/owner/repo/pulls/999"),
						},
					},
				})

			gock.New("https://api.github.com").
				Get("/repos/owner/repo/pulls/999").
				Reply(404).
				JSON(map[string]string{
					"message": "Not Found",
				})

			githubClient := setupMockClient(t)
			nc := cleaner.NewNotificationsCleaner(
				cleaner.WithGitHubClient(githubClient),
			)

			// We expect the operation to complete but log the error
			err := nc.Clean(context.Background())
			require.NoError(t, err)
			assert.True(t, gock.IsDone())
		})
	})
}
