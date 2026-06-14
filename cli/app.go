// Package cli builds the hackernews command tree on top of the hackernews
// library and the any-cli/kit framework. Every command is a kit operation:
// declared once and exposed as a CLI subcommand, an HTTP route, and an MCP
// tool, with --limit, the --db store tee, and the output formats handled by
// the framework.
package cli

import (
	"time"

	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/hackernews-cli/hackernews"
)

// Build metadata, injected via -ldflags at release time.
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

// NewApp assembles the kit application: identity, defaults, client factory,
// and the hackernews operations.
func NewApp() *kit.App {
	app := kit.New(kit.Identity{
		Binary:  "hackernews",
		Version: Version,
		Short:   "Read public Hacker News data",
		Long: `hackernews turns news.ycombinator.com into a fast, scriptable command line.

Fetch the front page, new/best/ask/show/job stories, individual items, and user
profiles — all from the open Firebase API, no key required.

Quick start:
  hackernews top                   top 30 stories
  hackernews new --limit 10        10 newest stories
  hackernews best -o jsonl         best stories as newline-delimited JSON
  hackernews item 1                the very first HN item
  hackernews user pg               pg's profile`,
		Site: "news.ycombinator.com",
		Repo: "https://github.com/tamnd/hackernews-cli",
	}, kit.WithDefaults(func(c *kit.Config) {
		c.Rate = 100 * time.Millisecond
		c.Retries = 3
		c.Timeout = 30 * time.Second
		c.UserAgent = hackernews.DefaultUserAgent
	}))

	// Register the hackernews domain's operations and client factory onto the app.
	hackernews.Domain{}.Register(app)

	return app
}
