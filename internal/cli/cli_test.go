package cli_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Lillevang/agent-init/internal/cli"
	_ "github.com/Lillevang/agent-init/internal/testflags"
)

func TestListFlavors(t *testing.T) {
	t.Parallel()
	var out bytes.Buffer
	app := cli.New(&out, &bytes.Buffer{}, cli.Version{Commit: "test", BuildDate: "today"})

	if err := app.Run(context.Background(), []string{"list-flavors"}); err != nil {
		t.Fatalf("Run(list-flavors) error = %v", err)
	}
	if !bytes.Contains(out.Bytes(), []byte("fullstack")) {
		t.Fatalf("list-flavors output missing fullstack:\n%s", out.String())
	}
}

func TestInitLegacyTargetArgument(t *testing.T) {
	t.Parallel()
	target := filepath.Join(t.TempDir(), "app")
	var out bytes.Buffer
	app := cli.New(&out, &bytes.Buffer{}, cli.Version{})

	err := app.Run(context.Background(), []string{"--no-git", target})
	if err != nil {
		t.Fatalf("Run(init legacy target) error = %v", err)
	}
	content, err := os.ReadFile(filepath.Join(target, ".agent", "AGENTS.md"))
	if err != nil {
		t.Fatalf("read scaffolded AGENTS.md: %v", err)
	}
	if !bytes.Contains(content, []byte("Agent Instructions for app")) {
		t.Fatalf("project name substitution failed:\n%s", string(content))
	}
}

func TestInitLegacyTargetArgumentWithoutFlag(t *testing.T) {
	t.Parallel()
	target := filepath.Join(t.TempDir(), "app")
	app := cli.New(&bytes.Buffer{}, &bytes.Buffer{}, cli.Version{})

	err := app.Run(context.Background(), []string{target})
	if err != nil {
		t.Fatalf("Run(init legacy target without flag) error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, ".agent", "AGENTS.md")); err != nil {
		t.Fatalf("stat scaffolded AGENTS.md: %v", err)
	}
}

func TestInitRejectsUnknownFlavor(t *testing.T) {
	t.Parallel()
	app := cli.New(&bytes.Buffer{}, &bytes.Buffer{}, cli.Version{})

	err := app.Run(context.Background(), []string{"init", "missing-flavor", t.TempDir()})
	if err == nil {
		t.Fatal("Run(init missing-flavor) error = nil")
	}
	if !strings.Contains(err.Error(), "fullstack") {
		t.Fatalf("error = %v; want to mention known flavor 'fullstack'", err)
	}
}

func TestInitRejectsBareIdentifierAsSingleArg(t *testing.T) {
	t.Parallel()
	app := cli.New(&bytes.Buffer{}, &bytes.Buffer{}, cli.Version{})

	err := app.Run(context.Background(), []string{"init", "fulstack"})
	if err == nil {
		t.Fatal("Run(init fulstack) error = nil; bare identifier should not be silently treated as a target dir")
	}
	if !strings.Contains(err.Error(), "unknown flavor") {
		t.Fatalf("error = %v; want unknown flavor", err)
	}
	if !strings.Contains(err.Error(), "fullstack") {
		t.Fatalf("error = %v; want to suggest known flavor 'fullstack'", err)
	}
}

func TestInitHelpFlagDoesNotError(t *testing.T) {
	t.Parallel()
	var out, errOut bytes.Buffer
	app := cli.New(&out, &errOut, cli.Version{})

	if err := app.Run(context.Background(), []string{"init", "--help"}); err != nil {
		t.Fatalf("Run(init --help) error = %v; flag.ErrHelp should not surface as an error", err)
	}
	if !bytes.Contains(errOut.Bytes(), []byte("-force")) {
		t.Fatalf("init --help did not print flag usage:\n%s", errOut.String())
	}
}

func TestInitAgentsOnlyDropsFreshOnlyFiles(t *testing.T) {
	t.Parallel()
	target := filepath.Join(t.TempDir(), "existing")
	app := cli.New(&bytes.Buffer{}, &bytes.Buffer{}, cli.Version{})

	err := app.Run(context.Background(), []string{"init", "--no-git", "--agents-only", "go-cli", target})
	if err != nil {
		t.Fatalf("Run(init --agents-only go-cli) error = %v", err)
	}
	for _, p := range []string{"cmd", "go.mod", filepath.Join("internal", "version", "version.go")} {
		if _, err := os.Stat(filepath.Join(target, p)); !os.IsNotExist(err) {
			t.Errorf("--agents-only shipped %s, stat err = %v", p, err)
		}
	}
	for _, p := range []string{filepath.Join(".agent", "AGENTS.md"), "Justfile", ".pre-commit-config.yaml"} {
		if _, err := os.Stat(filepath.Join(target, p)); err != nil {
			t.Errorf("--agents-only missing %s: %v", p, err)
		}
	}
	justfile, err := os.ReadFile(filepath.Join(target, "Justfile"))
	if err != nil {
		t.Fatalf("read Justfile: %v", err)
	}
	if bytes.Contains(justfile, []byte("./cmd/")) {
		t.Errorf("--agents-only Justfile still references ./cmd/, want layout-agnostic recipes:\n%s", string(justfile))
	}
}

func TestInitAgentsOnlyRejectsUnsupportedFlavor(t *testing.T) {
	t.Parallel()
	app := cli.New(&bytes.Buffer{}, &bytes.Buffer{}, cli.Version{})

	// project-management is a doc-collab flavor; it has no "fresh project"
	// vs "existing project" distinction so --agents-only is rejected.
	err := app.Run(context.Background(), []string{"init", "--no-git", "--agents-only", "project-management", t.TempDir()})
	if err == nil {
		t.Fatal("Run(init --agents-only project-management) error = nil; want rejection")
	}
	if !strings.Contains(err.Error(), "agents-only") {
		t.Fatalf("error = %v; want to mention --agents-only", err)
	}
}

func TestRejectsUnknownCommandTypo(t *testing.T) {
	t.Parallel()
	app := cli.New(&bytes.Buffer{}, &bytes.Buffer{}, cli.Version{})

	err := app.Run(context.Background(), []string{"versoin"})
	if err == nil {
		t.Fatal("Run(unknown command) error = nil")
	}
}

func TestListTrackers(t *testing.T) {
	t.Parallel()
	var out bytes.Buffer
	app := cli.New(&out, &bytes.Buffer{}, cli.Version{})

	if err := app.Run(context.Background(), []string{"list-trackers"}); err != nil {
		t.Fatalf("Run(list-trackers) error = %v", err)
	}
	got := out.String()
	for _, name := range []string{"gh", "jira", "ado"} {
		if !strings.Contains(got, name) {
			t.Errorf("list-trackers output missing %q:\n%s", name, got)
		}
	}
}

func TestAddTrackerWritesFilesAndMergesMCP(t *testing.T) {
	t.Parallel()
	target := filepath.Join(t.TempDir(), "pm")
	app := cli.New(&bytes.Buffer{}, &bytes.Buffer{}, cli.Version{})

	// First scaffold the base project-management workspace.
	if err := app.Run(context.Background(), []string{"init", "--no-git", "project-management", target}); err != nil {
		t.Fatalf("init: %v", err)
	}

	// Add the GitHub tracker.
	if err := app.Run(context.Background(), []string{"add-tracker", "gh", target}); err != nil {
		t.Fatalf("add-tracker gh: %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, "integrations", "github", "README.md")); err != nil {
		t.Fatalf("integrations/github/README.md should exist: %v", err)
	}
	mcp, err := os.ReadFile(filepath.Join(target, ".mcp.json"))
	if err != nil {
		t.Fatalf("read .mcp.json: %v", err)
	}
	if !strings.Contains(string(mcp), `"github"`) {
		t.Fatalf(".mcp.json missing 'github' entry:\n%s", string(mcp))
	}
}

func TestAddTrackerIsIdempotent(t *testing.T) {
	t.Parallel()
	target := filepath.Join(t.TempDir(), "pm")
	var out bytes.Buffer
	app := cli.New(&out, &bytes.Buffer{}, cli.Version{})

	if err := app.Run(context.Background(), []string{"init", "--no-git", "project-management", target}); err != nil {
		t.Fatalf("init: %v", err)
	}
	if err := app.Run(context.Background(), []string{"add-tracker", "gh", target}); err != nil {
		t.Fatalf("add-tracker gh (first): %v", err)
	}
	out.Reset()
	if err := app.Run(context.Background(), []string{"add-tracker", "gh", target}); err != nil {
		t.Fatalf("add-tracker gh (second): %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "already present") && !strings.Contains(got, "skip") {
		t.Fatalf("second add-tracker run should report no changes; got:\n%s", got)
	}
}

func TestAddTrackerMultipleCoexist(t *testing.T) {
	t.Parallel()
	target := filepath.Join(t.TempDir(), "pm")
	app := cli.New(&bytes.Buffer{}, &bytes.Buffer{}, cli.Version{})

	if err := app.Run(context.Background(), []string{"init", "--no-git", "project-management", target}); err != nil {
		t.Fatalf("init: %v", err)
	}
	for _, name := range []string{"gh", "jira", "ado"} {
		if err := app.Run(context.Background(), []string{"add-tracker", name, target}); err != nil {
			t.Fatalf("add-tracker %s: %v", name, err)
		}
	}
	mcp, err := os.ReadFile(filepath.Join(target, ".mcp.json"))
	if err != nil {
		t.Fatalf("read .mcp.json: %v", err)
	}
	for _, key := range []string{`"github"`, `"atlassian"`, `"azure-devops"`} {
		if !strings.Contains(string(mcp), key) {
			t.Errorf(".mcp.json missing %s after multi-tracker add:\n%s", key, string(mcp))
		}
	}
}

func TestAddTrackerRejectsMissingScaffold(t *testing.T) {
	t.Parallel()
	target := t.TempDir() // empty dir, no .mcp.json
	app := cli.New(&bytes.Buffer{}, &bytes.Buffer{}, cli.Version{})

	err := app.Run(context.Background(), []string{"add-tracker", "gh", target})
	if err == nil {
		t.Fatal("expected error when target has no .mcp.json")
	}
	if !strings.Contains(err.Error(), "project-management scaffold") {
		t.Fatalf("error should mention the missing scaffold; got: %v", err)
	}
}

func TestAddTrackerRejectsUnknownTracker(t *testing.T) {
	t.Parallel()
	target := filepath.Join(t.TempDir(), "pm")
	app := cli.New(&bytes.Buffer{}, &bytes.Buffer{}, cli.Version{})
	if err := app.Run(context.Background(), []string{"init", "--no-git", "project-management", target}); err != nil {
		t.Fatalf("init: %v", err)
	}
	err := app.Run(context.Background(), []string{"add-tracker", "github-but-misspelled", target})
	if err == nil {
		t.Fatal("expected error for unknown tracker")
	}
}

func TestVersion(t *testing.T) {
	t.Parallel()
	var out bytes.Buffer
	app := cli.New(&out, &bytes.Buffer{}, cli.Version{Commit: "abc123", BuildDate: "2026-05-12"})

	if err := app.Run(context.Background(), []string{"version"}); err != nil {
		t.Fatalf("Run(version) error = %v", err)
	}
	if got := out.String(); got != "agent-init commit=abc123 buildDate=2026-05-12\n" {
		t.Fatalf("version output = %q", got)
	}
}
