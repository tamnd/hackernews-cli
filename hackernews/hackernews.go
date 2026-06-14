// Package hackernews is the library behind the hackernews command: the HTTP
// client, pacing, and the typed data models for the Hacker News Firebase API.
//
// The official Firebase endpoint at hacker-news.firebaseio.com is open: no API
// key, no auth, no rate limits beyond basic politeness. This package wraps it
// with a sequential, rate-limited client that the kit operations consume.
package hackernews

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Host is the Firebase base hostname for the Hacker News API.
const Host = "hacker-news.firebaseio.com"

const (
	baseURL          = "https://" + Host + "/v0"
	DefaultUserAgent = "hackernews/dev (+https://github.com/tamnd/hackernews-cli)"
)

// ErrNotFound is returned when the Firebase API returns null for an id or user.
var ErrNotFound = errors.New("not found")

// Item is one record from the Firebase item endpoint. It covers stories,
// comments, Ask HN, Show HN, jobs, and polls.
type Item struct {
	ID          int    `kit:"id" json:"id"`
	Type        string `json:"type"`
	By          string `json:"by"`
	Time        int64  `json:"time"`
	Title       string `json:"title"`
	URL         string `json:"url"`
	Score       int    `json:"score"`
	Descendants int    `json:"descendants"`
	Kids        []int  `json:"kids"`
	Text        string `json:"text"`
	Dead        bool   `json:"dead"`
	Deleted     bool   `json:"deleted"`
}

// User is one record from the Firebase user endpoint.
type User struct {
	ID      string `kit:"id" json:"id"`
	About   string `json:"about"`
	Karma   int    `json:"karma"`
	Created int64  `json:"created"`
}

// Config holds constructor parameters for Client.
type Config struct {
	UserAgent string
	Rate      time.Duration
	Retries   int
	Timeout   time.Duration
}

// DefaultConfig returns sensible defaults for the Firebase API.
func DefaultConfig() Config {
	return Config{
		UserAgent: DefaultUserAgent,
		Rate:      100 * time.Millisecond,
		Retries:   3,
		Timeout:   30 * time.Second,
	}
}

// Client is a rate-limited HTTP client for the HN Firebase API.
type Client struct {
	cfg  Config
	http *http.Client
}

// NewClient returns a Client configured with cfg.
func NewClient(cfg Config) *Client {
	return &Client{
		cfg:  cfg,
		http: &http.Client{Timeout: cfg.Timeout},
	}
}

// TopStories fetches the top stories list and resolves up to limit items.
func (c *Client) TopStories(ctx context.Context, limit int) ([]Item, error) {
	return c.storyList(ctx, "topstories", limit)
}

// NewStories fetches the new stories list and resolves up to limit items.
func (c *Client) NewStories(ctx context.Context, limit int) ([]Item, error) {
	return c.storyList(ctx, "newstories", limit)
}

// BestStories fetches the best stories list and resolves up to limit items.
func (c *Client) BestStories(ctx context.Context, limit int) ([]Item, error) {
	return c.storyList(ctx, "beststories", limit)
}

// AskStories fetches the Ask HN stories list and resolves up to limit items.
func (c *Client) AskStories(ctx context.Context, limit int) ([]Item, error) {
	return c.storyList(ctx, "askstories", limit)
}

// ShowStories fetches the Show HN stories list and resolves up to limit items.
func (c *Client) ShowStories(ctx context.Context, limit int) ([]Item, error) {
	return c.storyList(ctx, "showstories", limit)
}

// JobStories fetches the job stories list and resolves up to limit items.
func (c *Client) JobStories(ctx context.Context, limit int) ([]Item, error) {
	return c.storyList(ctx, "jobstories", limit)
}

// GetItem fetches a single item by id.
func (c *Client) GetItem(ctx context.Context, id int) (*Item, error) {
	var it Item
	u := fmt.Sprintf("%s/item/%d.json", baseURL, id)
	if err := c.getJSON(ctx, u, &it); err != nil {
		return nil, fmt.Errorf("item %d: %w", id, err)
	}
	return &it, nil
}

// GetUser fetches a single user profile by username.
func (c *Client) GetUser(ctx context.Context, id string) (*User, error) {
	var u User
	url := fmt.Sprintf("%s/user/%s.json", baseURL, id)
	if err := c.getJSON(ctx, url, &u); err != nil {
		return nil, fmt.Errorf("user %q: %w", id, err)
	}
	return &u, nil
}

// storyList fetches the named list endpoint and resolves items sequentially.
func (c *Client) storyList(ctx context.Context, endpoint string, limit int) ([]Item, error) {
	var ids []int
	if err := c.getJSON(ctx, baseURL+"/"+endpoint+".json", &ids); err != nil {
		return nil, err
	}
	if limit > 0 && limit < len(ids) {
		ids = ids[:limit]
	}

	out := make([]Item, 0, len(ids))
	for i, id := range ids {
		if i > 0 && c.cfg.Rate > 0 {
			select {
			case <-ctx.Done():
				return out, ctx.Err()
			case <-time.After(c.cfg.Rate):
			}
		}
		it, err := c.GetItem(ctx, id)
		if err != nil {
			continue // skip items that fail to fetch
		}
		if it.Dead || it.Deleted {
			continue
		}
		out = append(out, *it)
	}
	return out, nil
}

// getJSON fetches a URL and JSON-decodes into v. Returns ErrNotFound when the
// body is the literal JSON null.
func (c *Client) getJSON(ctx context.Context, rawURL string, v any) error {
	body, err := c.get(ctx, rawURL)
	if err != nil {
		return err
	}
	if strings.TrimSpace(string(body)) == "null" {
		return ErrNotFound
	}
	if err := json.Unmarshal(body, v); err != nil {
		return fmt.Errorf("decode %s: %w", rawURL, err)
	}
	return nil
}

// get fetches a URL with retries on transient errors.
func (c *Client) get(ctx context.Context, rawURL string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.cfg.Retries; attempt++ {
		if attempt > 0 {
			wait := time.Duration(attempt) * 500 * time.Millisecond
			if wait > 5*time.Second {
				wait = 5 * time.Second
			}
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(wait):
			}
		}
		body, retry, err := c.do(ctx, rawURL)
		if err == nil {
			return body, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}
	return nil, fmt.Errorf("get %s: %w", rawURL, lastErr)
}

func (c *Client) do(ctx context.Context, rawURL string) ([]byte, bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", c.cfg.UserAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return nil, true, fmt.Errorf("http %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("http %d", resp.StatusCode)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, true, err
	}
	return b, false, nil
}
