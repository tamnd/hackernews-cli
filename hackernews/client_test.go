package hackernews

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// newTestClient wires a Client to a mock server, with pacing off so tests are fast.
func newTestClient(base string) *Client {
	c := NewClient(Config{UserAgent: "test", Rate: 0, Retries: 1, Workers: 4, Timeout: 5 * time.Second})
	c.fbBase = base + "/v0"
	c.algBase = base + "/api/v1"
	return c
}

// mockHN is a minimal Firebase + Algolia server for the client tests.
func mockHN(t *testing.T, items map[string]string, lists map[string]string) (*httptest.Server, *recordedQuery) {
	t.Helper()
	rec := &recordedQuery{}
	algoliaBody := lists["__algolia__"]
	mux := http.NewServeMux()
	for path, body := range lists {
		if path == "__algolia__" {
			continue
		}
		b := body
		mux.HandleFunc(path, func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(b))
		})
	}
	// Firebase returns the literal null (HTTP 200) for any missing id or user.
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("null"))
	})
	mux.HandleFunc("/v0/item/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/v0/item/"), ".json")
		body, ok := items[id]
		if !ok {
			_, _ = w.Write([]byte("null"))
			return
		}
		_, _ = w.Write([]byte(body))
	})
	mux.HandleFunc("/api/v1/search", func(w http.ResponseWriter, r *http.Request) {
		rec.capture(r)
		_, _ = w.Write([]byte(algoliaBody))
	})
	mux.HandleFunc("/api/v1/search_by_date", func(w http.ResponseWriter, r *http.Request) {
		rec.capture(r)
		_, _ = w.Write([]byte(algoliaBody))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv, rec
}

type recordedQuery struct {
	path   string
	values map[string]string
}

func (rq *recordedQuery) capture(r *http.Request) {
	rq.path = r.URL.Path
	rq.values = map[string]string{}
	for k, v := range r.URL.Query() {
		rq.values[k] = v[0]
	}
}

func TestStoryListResolvesInRank(t *testing.T) {
	srv, _ := mockHN(t, map[string]string{
		"1": `{"id":1,"type":"story","title":"One","by":"a","score":5,"descendants":2,"url":"https://one.example","time":1175714200}`,
		"2": `{"id":2,"type":"story","title":"Two","by":"b","score":3,"time":1175714201}`,
		"3": `{"id":3,"type":"story","title":"Gone","deleted":true}`,
	}, map[string]string{
		"/v0/topstories.json": `[1,2,3]`,
	})
	c := newTestClient(srv.URL)
	stories, err := c.StoryList(context.Background(), "topstories", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(stories) != 2 {
		t.Fatalf("deleted item should be dropped, got %d stories", len(stories))
	}
	if stories[0].Rank != 1 || stories[0].ID != 1 {
		t.Errorf("rank1 = %+v", stories[0])
	}
	if stories[1].Rank != 2 || stories[1].ID != 2 {
		t.Errorf("rank2 = %+v", stories[1])
	}
	if stories[1].URL != "https://news.ycombinator.com/item?id=2" {
		t.Errorf("self-post URL fallback = %q", stories[1].URL)
	}
}

func TestStoryListLimitTruncates(t *testing.T) {
	srv, _ := mockHN(t, map[string]string{
		"1": `{"id":1,"type":"story","title":"One"}`,
		"2": `{"id":2,"type":"story","title":"Two"}`,
	}, map[string]string{
		"/v0/topstories.json": `[1,2,3,4,5]`,
	})
	c := newTestClient(srv.URL)
	stories, err := c.StoryList(context.Background(), "topstories", 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(stories) != 2 {
		t.Fatalf("limit should cap fetched ids, got %d", len(stories))
	}
}

func TestItemTreeFlattenAndDepth(t *testing.T) {
	// 10 -> [11,12]; 11 -> [13]; tree flattens depth-first preserving sibling order.
	srv, _ := mockHN(t, map[string]string{
		"10": `{"id":10,"type":"story","title":"Root","kids":[11,12]}`,
		"11": `{"id":11,"type":"comment","by":"a","parent":10,"text":"first","kids":[13]}`,
		"12": `{"id":12,"type":"comment","by":"b","parent":10,"text":"second"}`,
		"13": `{"id":13,"type":"comment","by":"c","parent":11,"text":"nested"}`,
	}, nil)
	c := newTestClient(srv.URL)

	story, comments, err := c.Item(context.Background(), 10, -1)
	if err != nil {
		t.Fatal(err)
	}
	if story.ID != 10 {
		t.Errorf("story id = %d", story.ID)
	}
	gotIDs := []int{}
	for _, cm := range comments {
		gotIDs = append(gotIDs, cm.ID)
	}
	want := []int{11, 13, 12}
	if len(gotIDs) != len(want) {
		t.Fatalf("flatten order = %v, want %v", gotIDs, want)
	}
	for i := range want {
		if gotIDs[i] != want[i] {
			t.Fatalf("flatten order = %v, want %v", gotIDs, want)
		}
	}
	// depth of 11 is 1, depth of 13 is 2.
	if comments[0].Depth != 1 || comments[1].Depth != 2 {
		t.Errorf("depths = %d,%d", comments[0].Depth, comments[1].Depth)
	}

	// depth=1 returns only the direct children.
	_, shallow, err := c.Item(context.Background(), 10, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(shallow) != 2 {
		t.Errorf("depth=1 should return 2 comments, got %d", len(shallow))
	}

	// depth=0 returns no comments.
	_, none, err := c.Item(context.Background(), 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(none) != 0 {
		t.Errorf("depth=0 should return no comments, got %d", len(none))
	}
}

func TestItemNotFound(t *testing.T) {
	srv, _ := mockHN(t, map[string]string{}, nil)
	c := newTestClient(srv.URL)
	_, _, err := c.Item(context.Background(), 999, 0)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}

func TestUser(t *testing.T) {
	srv, _ := mockHN(t, nil, map[string]string{
		"/v0/user/pg.json": `{"id":"pg","created":1175714200,"karma":155000,"about":"<b>hi</b>","submitted":[1,2,3,4]}`,
	})
	c := newTestClient(srv.URL)
	u, submitted, err := c.User(context.Background(), "pg")
	if err != nil {
		t.Fatal(err)
	}
	if u.Username != "pg" || u.Karma != 155000 {
		t.Errorf("user = %+v", u)
	}
	if u.Submitted != 4 {
		t.Errorf("submitted count = %d", u.Submitted)
	}
	if len(submitted) != 4 {
		t.Errorf("raw submitted ids = %v", submitted)
	}
}

func TestUserNotFound(t *testing.T) {
	srv, _ := mockHN(t, nil, nil)
	c := newTestClient(srv.URL)
	_, _, err := c.User(context.Background(), "ghost")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}

func TestMaxItemAndUpdates(t *testing.T) {
	srv, _ := mockHN(t, nil, map[string]string{
		"/v0/maxitem.json": `40000000`,
		"/v0/updates.json": `{"items":[1,2],"profiles":["pg"]}`,
	})
	c := newTestClient(srv.URL)
	max, err := c.MaxItem(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if max != 40000000 {
		t.Errorf("maxitem = %d", max)
	}
	u, err := c.Updates(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(u.Items) != 2 || len(u.Profiles) != 1 {
		t.Errorf("updates = %+v", u)
	}
}

func TestSearchRelevanceParamsAndHits(t *testing.T) {
	srv, rec := mockHN(t, nil, map[string]string{
		"__algolia__": `{"hits":[{"objectID":"1","title":"Rust","url":"https://r.example","author":"a","points":100,"num_comments":10,"created_at_i":1175714200,"_tags":["story"]}],"nbPages":1}`,
	})
	c := newTestClient(srv.URL)
	hits, err := c.Search(context.Background(), SearchOptions{Query: "rust", Tags: "story", MinPoints: 50, MinComments: 5, Limit: 20})
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 1 || hits[0].Title != "Rust" {
		t.Fatalf("hits = %+v", hits)
	}
	if rec.path != "/api/v1/search" {
		t.Errorf("relevance should hit /search, got %q", rec.path)
	}
	if rec.values["query"] != "rust" || rec.values["tags"] != "story" {
		t.Errorf("query params = %v", rec.values)
	}
	nf := rec.values["numericFilters"]
	if !strings.Contains(nf, "points>=50") || !strings.Contains(nf, "num_comments>=5") {
		t.Errorf("numericFilters = %q", nf)
	}
}

func TestSearchByDateEndpoint(t *testing.T) {
	srv, rec := mockHN(t, nil, map[string]string{
		"__algolia__": `{"hits":[],"nbPages":1}`,
	})
	c := newTestClient(srv.URL)
	if _, err := c.Search(context.Background(), SearchOptions{Query: "go", Sort: "date"}); err != nil {
		t.Fatal(err)
	}
	if rec.path != "/api/v1/search_by_date" {
		t.Errorf("sort=date should hit /search_by_date, got %q", rec.path)
	}
}

func TestSearchSinceBuildsCutoff(t *testing.T) {
	srv, rec := mockHN(t, nil, map[string]string{
		"__algolia__": `{"hits":[],"nbPages":1}`,
	})
	c := newTestClient(srv.URL)
	if _, err := c.Search(context.Background(), SearchOptions{Query: "x", Since: "24h"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(rec.values["numericFilters"], "created_at_i>") {
		t.Errorf("since should add created_at_i filter, got %q", rec.values["numericFilters"])
	}
}

func TestSearchInvalidSince(t *testing.T) {
	srv, _ := mockHN(t, nil, map[string]string{"__algolia__": `{"hits":[],"nbPages":1}`})
	c := newTestClient(srv.URL)
	_, err := c.Search(context.Background(), SearchOptions{Query: "x", Since: "bogus"})
	if err == nil {
		t.Error("invalid --since should error")
	}
}

func TestGetNoRetryOn404(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)
	c := newTestClient(srv.URL)
	if _, err := c.MaxItem(context.Background()); err == nil {
		t.Error("expected error on 404")
	}
	if calls != 1 {
		t.Errorf("404 should not retry, got %d calls", calls)
	}
}
