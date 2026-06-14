package hackernews

import (
	"context"
	"fmt"
	"strconv"
	"unicode"

	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/any-cli/kit/errs"
)

// domain.go registers the hackernews kit Domain so a blank import in a
// multi-domain host (ant) enables the driver:
//
//	import _ "github.com/tamnd/hackernews-cli/hackernews"
//
// The Domain also builds the standalone hackernews binary via NewApp.
func init() { kit.Register(Domain{}) }

// Domain is the Hacker News driver. It carries no state; the per-run client is
// built by the factory Register hands kit.
type Domain struct{}

// Info describes the scheme and the identity the single-site binary inherits.
func (Domain) Info() kit.DomainInfo {
	return kit.DomainInfo{
		Scheme:  "hackernews",
		Aliases: []string{"hn"},
		Hosts:   []string{Host, "news.ycombinator.com"},
		Identity: kit.Identity{
			Binary: "hackernews",
			Short:  "Read public Hacker News data",
			Site:   "news.ycombinator.com",
			Repo:   "https://github.com/tamnd/hackernews-cli",
		},
	}
}

// Register installs the client factory and the five HN operations onto app.
func (Domain) Register(app *kit.App) {
	app.SetClient(newClient)

	kit.Handle(app, kit.OpMeta{
		Name:    "top",
		Group:   "stories",
		Summary: "List the top stories",
		Args:    []kit.Arg{{Name: "limit", Help: "max stories", Optional: true}},
	}, topStories)

	kit.Handle(app, kit.OpMeta{
		Name:    "new",
		Group:   "stories",
		Summary: "List the newest stories",
		Args:    []kit.Arg{{Name: "limit", Help: "max stories", Optional: true}},
	}, newStories)

	kit.Handle(app, kit.OpMeta{
		Name:    "best",
		Group:   "stories",
		Summary: "List the best stories",
		Args:    []kit.Arg{{Name: "limit", Help: "max stories", Optional: true}},
	}, bestStories)

	kit.Handle(app, kit.OpMeta{
		Name:    "ask",
		Group:   "stories",
		Summary: "List Ask HN stories",
		Args:    []kit.Arg{{Name: "limit", Help: "max stories", Optional: true}},
	}, askStories)

	kit.Handle(app, kit.OpMeta{
		Name:    "show",
		Group:   "stories",
		Summary: "List Show HN stories",
		Args:    []kit.Arg{{Name: "limit", Help: "max stories", Optional: true}},
	}, showStories)

	kit.Handle(app, kit.OpMeta{
		Name:    "jobs",
		Group:   "stories",
		Summary: "List HN job posts",
		Args:    []kit.Arg{{Name: "limit", Help: "max stories", Optional: true}},
	}, jobStories)

	kit.Handle(app, kit.OpMeta{
		Name:     "item",
		Group:    "read",
		Single:   true,
		Resolver: true,
		URIType:  "item",
		Summary:  "Fetch a single item by ID",
		Args:     []kit.Arg{{Name: "id", Help: "item ID"}},
	}, getItem)

	kit.Handle(app, kit.OpMeta{
		Name:     "user",
		Group:    "read",
		Single:   true,
		Resolver: true,
		URIType:  "user",
		Summary:  "Fetch a user profile",
		Args:     []kit.Arg{{Name: "id", Help: "username"}},
	}, getUser)
}

// newClient builds a Client from the resolved kit Config.
func newClient(_ context.Context, cfg kit.Config) (any, error) {
	c := DefaultConfig()
	if cfg.Rate > 0 {
		c.Rate = cfg.Rate
	}
	if cfg.Retries > 0 {
		c.Retries = cfg.Retries
	}
	if cfg.Timeout > 0 {
		c.Timeout = cfg.Timeout
	}
	if cfg.UserAgent != "" {
		c.UserAgent = cfg.UserAgent
	}
	return NewClient(c), nil
}

// --- input structs ---

type storiesInput struct {
	Limit  int     `kit:"flag,inherit" help:"max stories" default:"30"`
	Client *Client `kit:"inject"`
}

type itemInput struct {
	ID     int     `kit:"arg" help:"item ID"`
	Client *Client `kit:"inject"`
}

type userInput struct {
	ID     string  `kit:"arg" help:"username"`
	Client *Client `kit:"inject"`
}

// --- handlers ---

func topStories(ctx context.Context, in storiesInput, emit func(Item) error) error {
	items, err := in.Client.TopStories(ctx, in.Limit)
	if err != nil {
		return err
	}
	for _, it := range items {
		if err := emit(it); err != nil {
			return err
		}
	}
	return nil
}

func newStories(ctx context.Context, in storiesInput, emit func(Item) error) error {
	items, err := in.Client.NewStories(ctx, in.Limit)
	if err != nil {
		return err
	}
	for _, it := range items {
		if err := emit(it); err != nil {
			return err
		}
	}
	return nil
}

func bestStories(ctx context.Context, in storiesInput, emit func(Item) error) error {
	items, err := in.Client.BestStories(ctx, in.Limit)
	if err != nil {
		return err
	}
	for _, it := range items {
		if err := emit(it); err != nil {
			return err
		}
	}
	return nil
}

func askStories(ctx context.Context, in storiesInput, emit func(Item) error) error {
	items, err := in.Client.AskStories(ctx, in.Limit)
	if err != nil {
		return err
	}
	for _, it := range items {
		if err := emit(it); err != nil {
			return err
		}
	}
	return nil
}

func showStories(ctx context.Context, in storiesInput, emit func(Item) error) error {
	items, err := in.Client.ShowStories(ctx, in.Limit)
	if err != nil {
		return err
	}
	for _, it := range items {
		if err := emit(it); err != nil {
			return err
		}
	}
	return nil
}

func jobStories(ctx context.Context, in storiesInput, emit func(Item) error) error {
	items, err := in.Client.JobStories(ctx, in.Limit)
	if err != nil {
		return err
	}
	for _, it := range items {
		if err := emit(it); err != nil {
			return err
		}
	}
	return nil
}

func getItem(ctx context.Context, in itemInput, emit func(*Item) error) error {
	it, err := in.Client.GetItem(ctx, in.ID)
	if err != nil {
		return mapErr(err)
	}
	return emit(it)
}

func getUser(ctx context.Context, in userInput, emit func(*User) error) error {
	u, err := in.Client.GetUser(ctx, in.ID)
	if err != nil {
		return mapErr(err)
	}
	return emit(u)
}

// --- Resolver ---

// Classify turns any accepted input into the canonical (uriType, id).
// Numeric input → ("item", id); username-like → ("user", input); otherwise error.
func (Domain) Classify(input string) (uriType, id string, err error) {
	if input == "" {
		return "", "", errs.Usage("hackernews: empty input")
	}
	// numeric id → item
	if _, err := strconv.Atoi(input); err == nil {
		return "item", input, nil
	}
	// username: letters, digits, hyphens, underscores
	if isUsername(input) {
		return "user", input, nil
	}
	return "", "", errs.Usage("hackernews: unrecognized reference: %q", input)
}

// Locate returns the canonical HN URL for a (uriType, id).
func (Domain) Locate(uriType, id string) (string, error) {
	switch uriType {
	case "item":
		n, err := strconv.Atoi(id)
		if err != nil {
			return "", errs.Usage("hackernews: item id must be numeric, got %q", id)
		}
		return fmt.Sprintf("https://news.ycombinator.com/item?id=%d", n), nil
	case "user":
		return fmt.Sprintf("https://news.ycombinator.com/user?id=%s", id), nil
	default:
		return "", errs.Usage("hackernews has no resource type %q", uriType)
	}
}

// isUsername returns true when s looks like a valid HN username: non-empty,
// and composed of letters, digits, hyphens, and underscores.
func isUsername(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '-' && r != '_' {
			return false
		}
	}
	return true
}

// mapErr translates library errors into kit error kinds with appropriate exit codes.
func mapErr(err error) error {
	if err == nil {
		return nil
	}
	if err == ErrNotFound {
		return errs.NotFound("%s", err.Error())
	}
	return err
}
