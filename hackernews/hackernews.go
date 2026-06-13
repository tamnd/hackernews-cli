// Package hackernews is the library behind the hn command: the HTTP client,
// request shaping, and the typed data models for Hacker News.
//
// Two APIs: the official Firebase endpoint at hacker-news.firebaseio.com for
// live data, and the Algolia endpoint at hn.algolia.com for full-text search.
// Both are open, no key required.
package hackernews

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	firebaseBase = "https://hacker-news.firebaseio.com/v0"
	algoliaBase  = "https://hn.algolia.com/api/v1"
)

// DefaultUserAgent identifies the client to both APIs.
const DefaultUserAgent = "hn/dev (+https://github.com/tamnd/hackernews-cli)"

// ErrNotFound is returned when the Firebase API returns null for an id or user.
var ErrNotFound = errors.New("not found")

// Client talks to the HN Firebase and Algolia APIs.
type Client struct {
	httpClient *http.Client
	userAgent  string
	rate       time.Duration
	retries    int
	workers    int
	mu         sync.Mutex
	last       time.Time
}

// Config holds constructor parameters.
type Config struct {
	UserAgent string
	Rate      time.Duration
	Retries   int
	Workers   int
	Timeout   time.Duration
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		UserAgent: DefaultUserAgent,
		Rate:      50 * time.Millisecond,
		Retries:   5,
		Workers:   16,
		Timeout:   30 * time.Second,
	}
}

// NewClient returns a Client with the given config.
func NewClient(cfg Config) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: cfg.Timeout},
		userAgent:  cfg.UserAgent,
		rate:       cfg.Rate,
		retries:    cfg.Retries,
		workers:    cfg.Workers,
	}
}

// get fetches a URL with pacing and retries.
func (c *Client) get(ctx context.Context, rawURL string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff(attempt)):
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
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
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

func (c *Client) pace() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.rate <= 0 {
		return
	}
	if wait := c.rate - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
}

func backoff(attempt int) time.Duration {
	d := time.Duration(attempt) * 500 * time.Millisecond
	if d > 5*time.Second {
		d = 5 * time.Second
	}
	return d
}

// getJSON fetches and JSON-decodes into v. Returns ErrNotFound when the body is null.
func (c *Client) getJSON(ctx context.Context, rawURL string, v any) error {
	body, err := c.get(ctx, rawURL)
	if err != nil {
		return err
	}
	trimmed := strings.TrimSpace(string(body))
	if trimmed == "null" {
		return ErrNotFound
	}
	if err := json.Unmarshal(body, v); err != nil {
		return fmt.Errorf("decode %s: %w", rawURL, err)
	}
	return nil
}

// ─── Firebase ────────────────────────────────────────────────────────────────

// StoryList fetches a named story list and resolves the top limit ids to stories.
// endpoint is one of: topstories, beststories, newstories, askstories,
// showstories, jobstories.
func (c *Client) StoryList(ctx context.Context, endpoint string, limit int) ([]Story, error) {
	var ids []int
	if err := c.getJSON(ctx, firebaseBase+"/"+endpoint+".json", &ids); err != nil {
		return nil, err
	}
	if limit > 0 && limit < len(ids) {
		ids = ids[:limit]
	}
	return c.resolveIDs(ctx, ids)
}

// resolveIDs fetches each id concurrently and returns stories in rank order.
func (c *Client) resolveIDs(ctx context.Context, ids []int) ([]Story, error) {
	results := make([]*hnItem, len(ids))
	sem := make(chan struct{}, c.workers)
	var wg sync.WaitGroup
	for i, id := range ids {
		wg.Add(1)
		sem <- struct{}{}
		go func(idx, itemID int) {
			defer wg.Done()
			defer func() { <-sem }()
			var it hnItem
			u := fmt.Sprintf("%s/item/%d.json", firebaseBase, itemID)
			if err := c.getJSON(ctx, u, &it); err == nil && !it.Deleted && !it.Dead {
				results[idx] = &it
			}
		}(i, id)
	}
	wg.Wait()

	out := make([]Story, 0, len(ids))
	rank := 1
	for _, it := range results {
		if it != nil {
			out = append(out, itemToStory(it, rank))
			rank++
		}
	}
	return out, nil
}

// Item fetches a single item and its comment tree to depth levels.
// depth=0 returns the item only; depth=-1 returns the full tree.
func (c *Client) Item(ctx context.Context, id int, depth int) (Story, []Comment, error) {
	var it hnItem
	u := fmt.Sprintf("%s/item/%d.json", firebaseBase, id)
	if err := c.getJSON(ctx, u, &it); err != nil {
		return Story{}, nil, fmt.Errorf("item %d: %w", id, err)
	}
	story := itemToStory(&it, 0)
	var comments []Comment
	if depth != 0 && len(it.Kids) > 0 {
		comments = c.fetchTree(ctx, it.Kids, 1, depth)
	}
	return story, comments, nil
}

func (c *Client) fetchTree(ctx context.Context, kids []int, currentDepth, maxDepth int) []Comment {
	if maxDepth > 0 && currentDepth > maxDepth {
		return nil
	}
	items := make([]*hnItem, len(kids))
	sem := make(chan struct{}, c.workers)
	var wg sync.WaitGroup
	for i, id := range kids {
		wg.Add(1)
		sem <- struct{}{}
		go func(idx, itemID int) {
			defer wg.Done()
			defer func() { <-sem }()
			var it hnItem
			u := fmt.Sprintf("%s/item/%d.json", firebaseBase, itemID)
			if err := c.getJSON(ctx, u, &it); err == nil && !it.Deleted && !it.Dead {
				items[idx] = &it
			}
		}(i, id)
	}
	wg.Wait()

	var out []Comment
	for _, it := range items {
		if it == nil {
			continue
		}
		out = append(out, itemToComment(it, currentDepth))
		if len(it.Kids) > 0 {
			out = append(out, c.fetchTree(ctx, it.Kids, currentDepth+1, maxDepth)...)
		}
	}
	return out
}

// User fetches a user profile. Returns the profile, the raw submitted ids, and any error.
func (c *Client) User(ctx context.Context, username string) (User, []int, error) {
	var u hnUser
	rawURL := fmt.Sprintf("%s/user/%s.json", firebaseBase, url.PathEscape(username))
	if err := c.getJSON(ctx, rawURL, &u); err != nil {
		return User{}, nil, fmt.Errorf("user %q: %w", username, err)
	}
	return wireUserToUser(&u), u.Submitted, nil
}

// MaxItem returns the current maximum item id.
func (c *Client) MaxItem(ctx context.Context) (int, error) {
	var id int
	if err := c.getJSON(ctx, firebaseBase+"/maxitem.json", &id); err != nil {
		return 0, err
	}
	return id, nil
}

// Updates returns recently changed item ids and profiles.
func (c *Client) Updates(ctx context.Context) (Updates, error) {
	var u Updates
	if err := c.getJSON(ctx, firebaseBase+"/updates.json", &u); err != nil {
		return Updates{}, err
	}
	return u, nil
}

// UserSubmissions resolves a user's submitted ids to Story records.
func (c *Client) UserSubmissions(ctx context.Context, ids []int, limit int) ([]Story, error) {
	if limit > 0 && limit < len(ids) {
		ids = ids[:limit]
	}
	return c.resolveIDs(ctx, ids)
}

// ─── Algolia ─────────────────────────────────────────────────────────────────

// SearchOptions controls the Algolia query.
type SearchOptions struct {
	Query       string
	Tags        string
	Sort        string
	Since       string
	MinPoints   int
	MinComments int
	Limit       int
}

type algoliaResp struct {
	Hits    []algoliaHit `json:"hits"`
	NbPages int          `json:"nbPages"`
}

type algoliaHit struct {
	ObjectID    string   `json:"objectID"`
	Title       string   `json:"title"`
	URL         string   `json:"url"`
	Author      string   `json:"author"`
	Points      int      `json:"points"`
	NumComments int      `json:"num_comments"`
	CreatedAtI  int64    `json:"created_at_i"`
	StoryText   string   `json:"story_text"`
	CommentText string   `json:"comment_text"`
	StoryID     *int     `json:"story_id"`
	StoryTitle  string   `json:"story_title"`
	Tags        []string `json:"_tags"`
}

// Search searches HN via Algolia.
func (c *Client) Search(ctx context.Context, opts SearchOptions) ([]SearchHit, error) {
	endpoint := "/search"
	if opts.Sort == "date" {
		endpoint = "/search_by_date"
	}

	tags := opts.Tags
	if tags == "" {
		tags = "story"
	}

	params := url.Values{}
	params.Set("query", opts.Query)
	params.Set("tags", tags)

	var numFilters []string
	if opts.Since != "" {
		since, err := parseDuration(opts.Since)
		if err != nil {
			return nil, fmt.Errorf("--since: %w", err)
		}
		cutoff := time.Now().Add(-since).Unix()
		numFilters = append(numFilters, fmt.Sprintf("created_at_i>%d", cutoff))
	}
	if opts.MinPoints > 0 {
		numFilters = append(numFilters, fmt.Sprintf("points>=%d", opts.MinPoints))
	}
	if opts.MinComments > 0 {
		numFilters = append(numFilters, fmt.Sprintf("num_comments>=%d", opts.MinComments))
	}
	if len(numFilters) > 0 {
		params.Set("numericFilters", strings.Join(numFilters, ","))
	}

	limit := opts.Limit
	if limit <= 0 {
		limit = 20
	}
	pageSize := limit
	if pageSize > 50 {
		pageSize = 50
	}
	params.Set("hitsPerPage", strconv.Itoa(pageSize))

	var out []SearchHit
	page := 0
	for {
		params.Set("page", strconv.Itoa(page))
		rawURL := algoliaBase + endpoint + "?" + params.Encode()
		var resp algoliaResp
		if err := c.getJSON(ctx, rawURL, &resp); err != nil {
			return out, err
		}
		for _, h := range resp.Hits {
			out = append(out, algoliaHitToRecord(h, len(out)+1))
			if len(out) >= limit {
				return out, nil
			}
		}
		page++
		if page >= resp.NbPages || len(resp.Hits) == 0 {
			break
		}
	}
	return out, nil
}

func algoliaHitToRecord(h algoliaHit, rank int) SearchHit {
	id, _ := strconv.Atoi(h.ObjectID)
	typ := "story"
	for _, t := range h.Tags {
		if t == "comment" || t == "story" || t == "job" || t == "poll" {
			typ = t
			break
		}
	}
	text := stripTags(h.StoryText)
	if h.CommentText != "" {
		text = stripTags(h.CommentText)
	}
	sh := SearchHit{
		Rank:     rank,
		ID:       id,
		Type:     typ,
		Title:    h.Title,
		URL:      h.URL,
		By:       h.Author,
		Score:    h.Points,
		Comments: h.NumComments,
		Time:     h.CreatedAtI,
		Date:     isoDate(h.CreatedAtI),
		Text:     text,
		HNURL:    hnURL(id),
	}
	if h.StoryID != nil {
		sh.StoryID = *h.StoryID
		sh.StoryTitle = h.StoryTitle
	}
	return sh
}

func parseDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if strings.HasSuffix(s, "d") {
		days, err := strconv.Atoi(s[:len(s)-1])
		if err != nil {
			return 0, fmt.Errorf("invalid duration %q", s)
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}
	return time.ParseDuration(s)
}
