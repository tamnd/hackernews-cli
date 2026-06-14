package hackernews

import (
	"testing"
	"time"
)

func TestIsoDate(t *testing.T) {
	// 1175714200 = 2007-04-04T19:16:40Z (the first HN item's timestamp era).
	if got := isoDate(1175714200); got != "2007-04-04T19:16:40Z" {
		t.Errorf("isoDate = %q", got)
	}
}

func TestHNURL(t *testing.T) {
	if got := hnURL(42); got != "https://news.ycombinator.com/item?id=42" {
		t.Errorf("hnURL = %q", got)
	}
}

func TestStripTags(t *testing.T) {
	cases := map[string]string{
		"<p>hello</p>":                "hello",
		"a &amp; b":                   "a & b",
		"&lt;tag&gt;":                 "<tag>",
		"say &quot;hi&quot;":          `say "hi"`,
		"it&#39;s &apos;quoted&apos;": "it's 'quoted'",
		"<a href=\"x\">link</a>":      "link",
		"https:&#x2F;&#x2F;a.example": "https://a.example",
		"&#x27;hex&#x27;":             "'hex'",
		"  spaced  ":                  "spaced",
		"":                            "",
	}
	for in, want := range cases {
		if got := stripTags(in); got != want {
			t.Errorf("stripTags(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestParseDuration(t *testing.T) {
	cases := []struct {
		in      string
		want    time.Duration
		wantErr bool
	}{
		{"24h", 24 * time.Hour, false},
		{"7d", 7 * 24 * time.Hour, false},
		{"1d", 24 * time.Hour, false},
		{"90m", 90 * time.Minute, false},
		{"bad", 0, true},
		{"xd", 0, true},
		{"", 0, true},
	}
	for _, c := range cases {
		got, err := parseDuration(c.in)
		if c.wantErr {
			if err == nil {
				t.Errorf("parseDuration(%q) expected error", c.in)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseDuration(%q) unexpected error: %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("parseDuration(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestItemToStoryLinkAndSelf(t *testing.T) {
	link := itemToStory(&hnItem{ID: 1, Type: "story", Title: "T", URL: "https://x.example", By: "a", Score: 5, Descendants: 9, Time: 1175714200, Text: "<i>body</i>"}, 3)
	if link.URL != "https://x.example" {
		t.Errorf("link story URL = %q", link.URL)
	}
	if link.Comments != 9 {
		t.Errorf("Comments should map from descendants, got %d", link.Comments)
	}
	if link.Text != "body" {
		t.Errorf("Text should be stripped, got %q", link.Text)
	}
	if link.HNURL != "https://news.ycombinator.com/item?id=1" {
		t.Errorf("HNURL = %q", link.HNURL)
	}
	if link.Rank != 3 {
		t.Errorf("Rank = %d", link.Rank)
	}

	self := itemToStory(&hnItem{ID: 7, Type: "story", Title: "Ask"}, 1)
	if self.URL != "https://news.ycombinator.com/item?id=7" {
		t.Errorf("self-post URL should fall back to HN link, got %q", self.URL)
	}
}

func TestItemToComment(t *testing.T) {
	c := itemToComment(&hnItem{ID: 11, By: "u", Parent: 10, Time: 1175714200, Text: "<p>hi &amp; bye</p>"}, 2)
	if c.Depth != 2 {
		t.Errorf("Depth = %d", c.Depth)
	}
	if c.Parent != 10 {
		t.Errorf("Parent = %d", c.Parent)
	}
	if c.Text != "hi & bye" {
		t.Errorf("Text = %q", c.Text)
	}
}

func TestWireUserToUser(t *testing.T) {
	u := wireUserToUser(&hnUser{ID: "pg", Karma: 100, Created: 1175714200, About: "<b>founder</b>", Submitted: []int{1, 2, 3}})
	if u.Username != "pg" {
		t.Errorf("Username = %q", u.Username)
	}
	if u.Submitted != 3 {
		t.Errorf("Submitted should be count, got %d", u.Submitted)
	}
	if u.About != "founder" {
		t.Errorf("About = %q", u.About)
	}
	if u.URL != "https://news.ycombinator.com/user?id=pg" {
		t.Errorf("URL = %q", u.URL)
	}
}

func TestAlgoliaHitToRecordStoryVsComment(t *testing.T) {
	sid := 500
	story := algoliaHitToRecord(algoliaHit{
		ObjectID: "501", Title: "A story", URL: "https://s.example", Author: "a",
		Points: 10, NumComments: 2, CreatedAtI: 1175714200,
		StoryID: &sid, StoryTitle: "echoed", Tags: []string{"story", "author_a"},
	}, 1)
	if story.Type != "story" {
		t.Errorf("type = %q", story.Type)
	}
	if story.StoryID != 0 || story.StoryTitle != "" {
		t.Errorf("story hit should drop story_id/story_title, got %d %q", story.StoryID, story.StoryTitle)
	}

	comment := algoliaHitToRecord(algoliaHit{
		ObjectID: "502", Author: "b", CommentText: "<p>nice &amp; short</p>",
		CreatedAtI: 1175714200, StoryID: &sid, StoryTitle: "parent story",
		Tags: []string{"comment", "author_b"},
	}, 2)
	if comment.Type != "comment" {
		t.Errorf("type = %q", comment.Type)
	}
	if comment.StoryID != 500 {
		t.Errorf("comment should carry story_id, got %d", comment.StoryID)
	}
	if comment.StoryTitle != "parent story" {
		t.Errorf("story_title = %q", comment.StoryTitle)
	}
	if comment.Text != "nice & short" {
		t.Errorf("comment text from comment_text, stripped: %q", comment.Text)
	}
}

func TestUpdatesChanges(t *testing.T) {
	u := Updates{Items: []int{10, 11}, Profiles: []string{"pg", "dang"}}
	changes := u.Changes()
	if len(changes) != 4 {
		t.Fatalf("want 4 changes, got %d", len(changes))
	}
	if changes[0].Kind != "item" || changes[0].Value != "10" {
		t.Errorf("change0 = %+v", changes[0])
	}
	if changes[0].URL != "https://news.ycombinator.com/item?id=10" {
		t.Errorf("item url = %q", changes[0].URL)
	}
	if changes[2].Kind != "profile" || changes[2].Value != "pg" {
		t.Errorf("change2 = %+v", changes[2])
	}
	if changes[2].URL != "https://news.ycombinator.com/user?id=pg" {
		t.Errorf("profile url = %q", changes[2].URL)
	}
}
