package hackernews

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// ParseItemID extracts a numeric HN item id from a bare integer string, a
// news.ycombinator.com/item?id=N URL, or a full https URL.
func ParseItemID(s string) (int, error) {
	s = strings.TrimSpace(s)
	// bare integer
	if id, err := strconv.Atoi(s); err == nil {
		if id <= 0 {
			return 0, fmt.Errorf("item id must be positive, got %d", id)
		}
		return id, nil
	}
	// URL
	u, err := url.Parse(s)
	if err != nil {
		return 0, fmt.Errorf("not a valid item id or URL: %q", s)
	}
	idStr := u.Query().Get("id")
	if idStr == "" {
		// try path segment last part
		parts := strings.Split(strings.Trim(u.Path, "/"), "/")
		idStr = parts[len(parts)-1]
	}
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("cannot extract item id from %q", s)
	}
	return id, nil
}

// ParseUsername extracts a HN username from a bare name or a
// news.ycombinator.com/user?id=name URL.
func ParseUsername(s string) (string, error) {
	s = strings.TrimSpace(s)
	if !strings.Contains(s, "/") {
		if s == "" {
			return "", fmt.Errorf("username must not be empty")
		}
		return s, nil
	}
	u, err := url.Parse(s)
	if err != nil {
		return "", fmt.Errorf("not a valid username or URL: %q", s)
	}
	name := u.Query().Get("id")
	if name == "" {
		return "", fmt.Errorf("cannot extract username from %q", s)
	}
	return name, nil
}
