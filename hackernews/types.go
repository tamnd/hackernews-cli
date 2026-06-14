package hackernews

import (
	"fmt"
	"html"
	"strconv"
	"strings"
	"time"
)

// Story is the record emitted for stories, Ask HN, Show HN, and job posts.
type Story struct {
	Rank     int    `json:"rank"`
	ID       int    `json:"id"`
	Type     string `json:"type"`
	Title    string `json:"title"`
	URL      string `json:"url"`
	By       string `json:"by"`
	Score    int    `json:"score"`
	Comments int    `json:"comments"`
	Time     int64  `json:"time"`
	Date     string `json:"date"`
	Text     string `json:"text"`
	HNURL    string `json:"hn_url"`
}

// Comment is the record emitted for comment items.
type Comment struct {
	ID     int    `json:"id"`
	By     string `json:"by"`
	Parent int    `json:"parent"`
	Time   int64  `json:"time"`
	Date   string `json:"date"`
	Depth  int    `json:"depth"`
	Text   string `json:"text"`
	HNURL  string `json:"hn_url"`
}

// User is the record emitted for HN user profiles.
type User struct {
	Username    string `json:"username"`
	Karma       int    `json:"karma"`
	Created     int64  `json:"created"`
	CreatedDate string `json:"created_date"`
	About       string `json:"about"`
	Submitted   int    `json:"submitted"`
	URL         string `json:"url"`
}

// SearchHit is the record for an Algolia search result (story or comment).
type SearchHit struct {
	Rank       int    `json:"rank"`
	ID         int    `json:"id"`
	Type       string `json:"type"`
	Title      string `json:"title"`
	URL        string `json:"url"`
	By         string `json:"by"`
	Score      int    `json:"score"`
	Comments   int    `json:"comments"`
	Time       int64  `json:"time"`
	Date       string `json:"date"`
	Text       string `json:"text"`
	StoryID    int    `json:"story_id,omitempty"`
	StoryTitle string `json:"story_title,omitempty"`
	HNURL      string `json:"hn_url"`
}

// Updates is returned by the /updates endpoint.
type Updates struct {
	Items    []int    `json:"items"`
	Profiles []string `json:"profiles"`
}

// Change is one entry from the updates firehose: either a changed item or a
// changed profile. It is the flat record the updates command renders, so the
// firehose prints in any output format like every other surface.
type Change struct {
	Kind  string `json:"kind"`  // "item" or "profile"
	Value string `json:"value"` // item id (decimal) or username
	URL   string `json:"url"`   // canonical HN link
}

// Changes flattens the updates payload into a single ordered record slice,
// items first then profiles, matching the order HN returns them.
func (u Updates) Changes() []Change {
	out := make([]Change, 0, len(u.Items)+len(u.Profiles))
	for _, id := range u.Items {
		out = append(out, Change{
			Kind:  "item",
			Value: strconv.Itoa(id),
			URL:   hnURL(id),
		})
	}
	for _, name := range u.Profiles {
		out = append(out, Change{
			Kind:  "profile",
			Value: name,
			URL:   fmt.Sprintf("https://news.ycombinator.com/user?id=%s", name),
		})
	}
	return out
}

// ─── wire types from Firebase ────────────────────────────────────────────────

type hnItem struct {
	ID          int    `json:"id"`
	Type        string `json:"type"`
	By          string `json:"by"`
	Time        int64  `json:"time"`
	Text        string `json:"text"`
	URL         string `json:"url"`
	Title       string `json:"title"`
	Score       int    `json:"score"`
	Descendants int    `json:"descendants"`
	Kids        []int  `json:"kids"`
	Parent      int    `json:"parent"`
	Dead        bool   `json:"dead"`
	Deleted     bool   `json:"deleted"`
}

type hnUser struct {
	ID        string `json:"id"`
	Created   int64  `json:"created"`
	Karma     int    `json:"karma"`
	About     string `json:"about"`
	Submitted []int  `json:"submitted"`
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func hnURL(id int) string {
	return fmt.Sprintf("https://news.ycombinator.com/item?id=%d", id)
}

func isoDate(unix int64) string {
	return time.Unix(unix, 0).UTC().Format(time.RFC3339)
}

// stripTags turns HN's HTML comment/story bodies into plain text: it drops the
// small tag vocabulary HN emits (<p>, <a>, <i>, <pre><code>) and then decodes
// every HTML entity, including numeric and hex forms like &#x2F; that HN uses
// heavily in URLs. <p> becomes a blank line so paragraph breaks survive.
func stripTags(s string) string {
	s = strings.ReplaceAll(s, "<p>", "\n\n")
	var b strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			b.WriteRune(r)
		}
	}
	return strings.TrimSpace(html.UnescapeString(b.String()))
}

func itemToStory(it *hnItem, rank int) Story {
	u := it.URL
	if u == "" {
		u = hnURL(it.ID)
	}
	return Story{
		Rank:     rank,
		ID:       it.ID,
		Type:     it.Type,
		Title:    it.Title,
		URL:      u,
		By:       it.By,
		Score:    it.Score,
		Comments: it.Descendants,
		Time:     it.Time,
		Date:     isoDate(it.Time),
		Text:     stripTags(it.Text),
		HNURL:    hnURL(it.ID),
	}
}

func itemToComment(it *hnItem, depth int) Comment {
	return Comment{
		ID:     it.ID,
		By:     it.By,
		Parent: it.Parent,
		Time:   it.Time,
		Date:   isoDate(it.Time),
		Depth:  depth,
		Text:   stripTags(it.Text),
		HNURL:  hnURL(it.ID),
	}
}

func wireUserToUser(u *hnUser) User {
	return User{
		Username:    u.ID,
		Karma:       u.Karma,
		Created:     u.Created,
		CreatedDate: isoDate(u.Created),
		About:       stripTags(u.About),
		Submitted:   len(u.Submitted),
		URL:         fmt.Sprintf("https://news.ycombinator.com/user?id=%s", u.ID),
	}
}
