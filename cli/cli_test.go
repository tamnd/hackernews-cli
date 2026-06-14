package cli

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/tamnd/hackernews-cli/hackernews"
)

func TestVersionCommand(t *testing.T) {
	root := Root()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"version", "--short"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out.String()) != Version {
		t.Errorf("version --short = %q, want %q", out.String(), Version)
	}
}

func TestBadFormatThroughExecute(t *testing.T) {
	root := Root()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	// version needs no network, but PersistentPreRunE still validates --output.
	root.SetArgs([]string{"-o", "nope", "version"})
	err := root.Execute()
	var ee *ExitError
	if !errors.As(err, &ee) || ee.Code != exitUsage {
		t.Errorf("bad format should surface exitUsage, got %v", err)
	}
}

func TestSearchRequiresQuery(t *testing.T) {
	root := Root()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"search"})
	if err := root.Execute(); err == nil {
		t.Error("search with no query should error on arg validation")
	}
}

func TestRootHasAllSubcommands(t *testing.T) {
	root := Root()
	want := []string{"top", "best", "new", "ask", "show", "jobs", "item", "user", "search", "updates", "maxitem", "version"}
	have := map[string]bool{}
	for _, c := range root.Commands() {
		have[c.Name()] = true
	}
	for _, w := range want {
		if !have[w] {
			t.Errorf("root is missing subcommand %q", w)
		}
	}
}

func TestRootPersistentFlags(t *testing.T) {
	root := Root()
	pf := root.PersistentFlags()
	for _, name := range []string{"output", "fields", "no-header", "template", "limit", "quiet", "workers", "delay", "timeout", "retries", "user-agent"} {
		if pf.Lookup(name) == nil {
			t.Errorf("missing persistent flag --%s", name)
		}
	}
}

func TestSetupRejectsBadFormat(t *testing.T) {
	a := &App{cfg: hackernews.DefaultConfig(), output: "nonsense"}
	err := a.setup()
	if err == nil {
		t.Fatal("expected error for bad format")
	}
	var ee *ExitError
	if !errors.As(err, &ee) || ee.Code != exitUsage {
		t.Errorf("want exitUsage ExitError, got %v", err)
	}
}

func TestSetupValidFormatBuildsClient(t *testing.T) {
	a := &App{cfg: hackernews.DefaultConfig(), output: "json"}
	if err := a.setup(); err != nil {
		t.Fatal(err)
	}
	if a.client == nil {
		t.Error("setup should build the client")
	}
}

func TestEffectiveLimit(t *testing.T) {
	a := &App{}
	if got := a.effectiveLimit(30); got != 30 {
		t.Errorf("default = %d", got)
	}
	a.limit = 5
	if got := a.effectiveLimit(30); got != 5 {
		t.Errorf("override = %d", got)
	}
}

func TestMapFetchErr(t *testing.T) {
	if mapFetchErr(nil) != nil {
		t.Error("nil should map to nil")
	}

	nf := mapFetchErr(fmt.Errorf("wrap: %w", hackernews.ErrNotFound))
	var ee *ExitError
	if !errors.As(nf, &ee) || ee.Code != exitNoData {
		t.Errorf("not-found should map to exitNoData, got %v", nf)
	}

	other := mapFetchErr(errors.New("boom"))
	if !errors.As(other, &ee) || ee.Code != exitError {
		t.Errorf("generic error should map to exitError, got %v", other)
	}
}

func TestExitError(t *testing.T) {
	base := errors.New("root cause")
	e := &ExitError{Code: 7, Err: base}
	if e.Error() != "root cause" {
		t.Errorf("Error() = %q", e.Error())
	}
	if !errors.Is(e, base) {
		t.Error("Unwrap should expose the wrapped error")
	}

	noErr := &ExitError{Code: 3}
	if noErr.Error() != "exit 3" {
		t.Errorf("codeless Error() = %q", noErr.Error())
	}
}

func TestIsNotFound(t *testing.T) {
	if !isNotFound(fmt.Errorf("x: %w", hackernews.ErrNotFound)) {
		t.Error("should detect wrapped ErrNotFound")
	}
	if isNotFound(errors.New("other")) {
		t.Error("should not match unrelated error")
	}
}
