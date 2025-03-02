package cleaner

import (
	"fmt"
	"net/url"
	"path"
	"strconv"
	"strings"
)

// parseNotificationURL extracts owner, repo, and number from the GitHub API URL.
// Example URL: https://api.github.com/repos/owner/repo/pulls/123
func parseNotificationURL(rawURL string) (owner string, repo string, number int, err error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", "", 0, err
	}
	segments := strings.Split(u.Path, "/")
	// Expected path is: /repos/{owner}/{repo}/{resource}/{number}
	if len(segments) < 5 {
		return "", "", 0, fmt.Errorf("invalid URL format")
	}
	owner = segments[2]
	repo = segments[3]
	// The resource segment (e.g., "pulls" or "issues") is segments[4]
	numberStr := path.Base(u.Path)
	number, err = strconv.Atoi(numberStr)
	if err != nil {
		return "", "", 0, err
	}
	return owner, repo, number, nil
}
