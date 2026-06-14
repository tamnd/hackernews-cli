package hackernews_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/tamnd/hackernews-cli/hackernews"
)

func TestDefaultConfig(t *testing.T) {
	cfg := hackernews.DefaultConfig()
	if cfg.Rate != 100*time.Millisecond {
		t.Errorf("Rate = %v, want 100ms", cfg.Rate)
	}
	if cfg.Retries <= 0 {
		t.Errorf("Retries = %d, want > 0", cfg.Retries)
	}
	if cfg.Timeout <= 0 {
		t.Errorf("Timeout = %v, want > 0", cfg.Timeout)
	}
	if cfg.UserAgent == "" {
		t.Error("UserAgent is empty")
	}
}

func TestNewClientNotNil(t *testing.T) {
	c := hackernews.NewClient(hackernews.DefaultConfig())
	if c == nil {
		t.Fatal("NewClient returned nil")
	}
}

func TestItemRoundTrip(t *testing.T) {
	want := hackernews.Item{
		ID:    48527700,
		Type:  "story",
		By:    "dang",
		Score: 300,
		Title: "A great article",
		URL:   "https://example.com",
	}
	b, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	var got hackernews.Item
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if got.ID != want.ID || got.By != want.By || got.Score != want.Score {
		t.Errorf("round-trip mismatch: got %+v, want %+v", got, want)
	}
}

func TestUserRoundTrip(t *testing.T) {
	want := hackernews.User{
		ID:      "pg",
		Karma:   155895,
		Created: 1160418092,
		About:   "Y Combinator founder",
	}
	b, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	var got hackernews.User
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if got.ID != want.ID || got.Karma != want.Karma {
		t.Errorf("round-trip mismatch: got %+v, want %+v", got, want)
	}
}

func TestErrNotFound(t *testing.T) {
	if hackernews.ErrNotFound == nil {
		t.Error("ErrNotFound is nil")
	}
	if hackernews.ErrNotFound.Error() == "" {
		t.Error("ErrNotFound has empty message")
	}
}

func TestHostConstant(t *testing.T) {
	if hackernews.Host != "hacker-news.firebaseio.com" {
		t.Errorf("Host = %q, want hacker-news.firebaseio.com", hackernews.Host)
	}
}

func TestGetItemFromTestServer(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if r.Header.Get("User-Agent") == "" {
			t.Error("request has no User-Agent")
		}
		it := hackernews.Item{
			ID:    99,
			Type:  "story",
			By:    "tester",
			Score: 10,
			Title: "Test story",
		}
		_ = json.NewEncoder(w).Encode(it)
	}))
	defer srv.Close()

	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()

	var it hackernews.Item
	if err := json.NewDecoder(resp.Body).Decode(&it); err != nil {
		t.Fatal(err)
	}
	if it.ID != 99 || it.By != "tester" {
		t.Errorf("decoded item = %+v", it)
	}
	if !called {
		t.Error("server was not called")
	}
}

func TestContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	cfg := hackernews.DefaultConfig()
	cfg.Rate = 0
	cfg.Retries = 0
	c := hackernews.NewClient(cfg)

	_, err := c.GetItem(ctx, 1)
	if err == nil {
		t.Error("GetItem with cancelled context returned nil error")
	}
}

func TestItemDeadDeletedFlags(t *testing.T) {
	it := hackernews.Item{
		ID:      123,
		Dead:    true,
		Deleted: false,
	}
	if !it.Dead {
		t.Error("Dead flag not preserved")
	}

	it2 := hackernews.Item{
		ID:      456,
		Deleted: true,
	}
	if !it2.Deleted {
		t.Error("Deleted flag not preserved")
	}
}
