package hackernews

import (
	"testing"

	"github.com/tamnd/any-cli/kit"
)

func TestDomainInfo(t *testing.T) {
	info := Domain{}.Info()
	if info.Scheme != "hackernews" {
		t.Errorf("scheme = %q, want hackernews", info.Scheme)
	}
	found := false
	for _, a := range info.Aliases {
		if a == "hn" {
			found = true
		}
	}
	if !found {
		t.Errorf("aliases = %v, want to contain hn", info.Aliases)
	}
	if info.Identity.Binary != "hackernews" {
		t.Errorf("binary = %q, want hackernews", info.Identity.Binary)
	}
}

func TestClassify(t *testing.T) {
	d := Domain{}

	typ, id, err := d.Classify("48527700")
	if err != nil || typ != "item" || id != "48527700" {
		t.Errorf("Classify(numeric) = %q/%q/%v, want item/48527700/nil", typ, id, err)
	}

	typ, id, err = d.Classify("pg")
	if err != nil || typ != "user" || id != "pg" {
		t.Errorf("Classify(username) = %q/%q/%v, want user/pg/nil", typ, id, err)
	}

	_, _, err = d.Classify("")
	if err == nil {
		t.Error("Classify('') = nil error, want error")
	}

	_, _, err = d.Classify("has spaces")
	if err == nil {
		t.Error("Classify('has spaces') = nil error, want error")
	}
}

func TestLocate(t *testing.T) {
	d := Domain{}

	url, err := d.Locate("item", "1")
	if err != nil || url != "https://news.ycombinator.com/item?id=1" {
		t.Errorf("Locate(item,1) = %q/%v", url, err)
	}

	url, err = d.Locate("user", "pg")
	if err != nil || url != "https://news.ycombinator.com/user?id=pg" {
		t.Errorf("Locate(user,pg) = %q/%v", url, err)
	}

	_, err = d.Locate("unknown", "foo")
	if err == nil {
		t.Error("Locate(unknown) = nil error, want error")
	}

	_, err = d.Locate("item", "notanumber")
	if err == nil {
		t.Error("Locate(item, notanumber) = nil error, want error")
	}
}

func TestDomainRegistered(t *testing.T) {
	// init() registered the domain; kit.Open should find it.
	h, err := kit.Open()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := h.Domain("hackernews"); !ok {
		t.Fatal("hackernews domain not registered")
	}
	if _, ok := h.Domain("hn"); !ok {
		t.Fatal("hn alias not registered")
	}
}
