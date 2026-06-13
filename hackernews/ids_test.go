package hackernews

import "testing"

func TestParseItemIDBareInt(t *testing.T) {
	id, err := ParseItemID("1")
	if err != nil || id != 1 {
		t.Fatalf("got (%d, %v)", id, err)
	}
}

func TestParseItemIDURL(t *testing.T) {
	id, err := ParseItemID("https://news.ycombinator.com/item?id=12345")
	if err != nil || id != 12345 {
		t.Fatalf("got (%d, %v)", id, err)
	}
}

func TestParseItemIDInvalid(t *testing.T) {
	_, err := ParseItemID("notanid")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseItemIDNegative(t *testing.T) {
	_, err := ParseItemID("-1")
	if err == nil {
		t.Fatal("expected error for negative")
	}
}

func TestParseUsernameBareName(t *testing.T) {
	name, err := ParseUsername("pg")
	if err != nil || name != "pg" {
		t.Fatalf("got (%q, %v)", name, err)
	}
}

func TestParseUsernameURL(t *testing.T) {
	name, err := ParseUsername("https://news.ycombinator.com/user?id=pg")
	if err != nil || name != "pg" {
		t.Fatalf("got (%q, %v)", name, err)
	}
}

func TestStripTagsBasic(t *testing.T) {
	cases := []struct{ in, want string }{
		{"<p>Hello <b>world</b></p>", "Hello world"},
		{"no tags", "no tags"},
		{"&amp;&lt;&gt;&quot;&#39;", "&<>\"'"},
		{"", ""},
	}
	for _, tc := range cases {
		got := stripTags(tc.in)
		if got != tc.want {
			t.Errorf("stripTags(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
