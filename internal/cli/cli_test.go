package cli_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mikeschinkel/agent-init/internal/cli"
	_ "github.com/mikeschinkel/agent-init/internal/testflags"
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
}

func TestRejectsUnknownCommandTypo(t *testing.T) {
	t.Parallel()
	app := cli.New(&bytes.Buffer{}, &bytes.Buffer{}, cli.Version{})

	err := app.Run(context.Background(), []string{"versoin"})
	if err == nil {
		t.Fatal("Run(unknown command) error = nil")
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
