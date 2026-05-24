package gitignore

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBlockHasMarkersAndEnvelope(t *testing.T) {
	t.Parallel()
	block := Block()
	for _, want := range []string{
		blockStart, blockEnd,
		".agent/", "/AGENTS.md", "/CLAUDE.md",
		".devcontainer/", "/Justfile", ".pre-commit-config.yaml",
	} {
		if !strings.Contains(block, want) {
			t.Errorf("Block() missing %q:\n%s", want, block)
		}
	}
	if !strings.HasSuffix(block, "\n") {
		t.Errorf("Block() must end with a newline:\n%q", block)
	}
}

func TestEnsureLocalCreatesAndAppends(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		initial string // "" means no existing .gitignore
	}{
		{name: "no existing file"},
		{name: "existing file without block", initial: "node_modules/\ndist/\n"},
		{name: "existing file without trailing newline", initial: "node_modules/"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			path := filepath.Join(dir, ".gitignore")
			if tt.initial != "" {
				if err := os.WriteFile(path, []byte(tt.initial), 0o644); err != nil {
					t.Fatalf("seed .gitignore: %v", err)
				}
			}

			got, err := EnsureLocal(dir)
			if err != nil {
				t.Fatalf("EnsureLocal() error = %v", err)
			}
			if got != path {
				t.Errorf("EnsureLocal() path = %q, want %q", got, path)
			}

			content := readFile(t, path)
			if tt.initial != "" && !strings.Contains(content, strings.TrimSpace(tt.initial)) {
				t.Errorf("EnsureLocal dropped pre-existing content:\n%s", content)
			}
			if !strings.Contains(content, Block()) {
				t.Errorf("EnsureLocal did not write the block:\n%s", content)
			}
			if strings.Count(content, blockStart) != 1 {
				t.Errorf("want exactly one block, content:\n%s", content)
			}
		})
	}
}

func TestEnsureLocalIsIdempotent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, ".gitignore")
	if err := os.WriteFile(path, []byte("node_modules/\n"), 0o644); err != nil {
		t.Fatalf("seed .gitignore: %v", err)
	}

	if _, err := EnsureLocal(dir); err != nil {
		t.Fatalf("first EnsureLocal() error = %v", err)
	}
	first := readFile(t, path)
	if _, err := EnsureLocal(dir); err != nil {
		t.Fatalf("second EnsureLocal() error = %v", err)
	}
	second := readFile(t, path)

	if first != second {
		t.Errorf("EnsureLocal not idempotent:\nfirst:\n%s\nsecond:\n%s", first, second)
	}
	if n := strings.Count(second, blockStart); n != 1 {
		t.Errorf("idempotent re-run produced %d blocks, want 1:\n%s", n, second)
	}
}

func TestEnsureLocalReplacesStaleBlockInPlace(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, ".gitignore")
	// A pre-existing block with a stale envelope entry, surrounded by user
	// content, must be replaced in place — not duplicated, and the surrounding
	// content preserved.
	stale := "node_modules/\n" + blockStart + "\n.agent/\n/OLD-ENTRY\n" + blockEnd + "\ncoverage/\n"
	if err := os.WriteFile(path, []byte(stale), 0o644); err != nil {
		t.Fatalf("seed .gitignore: %v", err)
	}

	if _, err := EnsureLocal(dir); err != nil {
		t.Fatalf("EnsureLocal() error = %v", err)
	}
	content := readFile(t, path)

	if strings.Contains(content, "/OLD-ENTRY") {
		t.Errorf("stale block entry survived:\n%s", content)
	}
	if n := strings.Count(content, blockStart); n != 1 {
		t.Errorf("got %d blocks, want 1:\n%s", n, content)
	}
	for _, surrounding := range []string{"node_modules/", "coverage/"} {
		if !strings.Contains(content, surrounding) {
			t.Errorf("surrounding content %q lost:\n%s", surrounding, content)
		}
	}
	if !strings.Contains(content, "/Justfile") {
		t.Errorf("refreshed block missing current envelope:\n%s", content)
	}
}

func TestEnsureHiddenCreatesAndAppends(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		initial string // "" means no existing exclude file
	}{
		{name: "no existing file"},
		{name: "existing file without block", initial: "# git ls-files --others\nbuild/\n"},
		{name: "existing file without trailing newline", initial: "build/"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			path := filepath.Join(dir, ".git", "info", "exclude")
			if tt.initial != "" {
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					t.Fatalf("mkdir .git/info: %v", err)
				}
				if err := os.WriteFile(path, []byte(tt.initial), 0o644); err != nil {
					t.Fatalf("seed exclude: %v", err)
				}
			}

			got, err := EnsureHidden(dir)
			if err != nil {
				t.Fatalf("EnsureHidden() error = %v", err)
			}
			if got != path {
				t.Errorf("EnsureHidden() path = %q, want %q", got, path)
			}

			content := readFile(t, path)
			if tt.initial != "" && !strings.Contains(content, strings.TrimSpace(tt.initial)) {
				t.Errorf("EnsureHidden dropped pre-existing content:\n%s", content)
			}
			if !strings.Contains(content, Block()) {
				t.Errorf("EnsureHidden did not write the block:\n%s", content)
			}
			if strings.Count(content, blockStart) != 1 {
				t.Errorf("want exactly one block, content:\n%s", content)
			}
		})
	}
}

func TestEnsureHiddenIsIdempotent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, ".git", "info", "exclude")

	if _, err := EnsureHidden(dir); err != nil {
		t.Fatalf("first EnsureHidden() error = %v", err)
	}
	first := readFile(t, path)
	if _, err := EnsureHidden(dir); err != nil {
		t.Fatalf("second EnsureHidden() error = %v", err)
	}
	second := readFile(t, path)

	if first != second {
		t.Errorf("EnsureHidden not idempotent:\nfirst:\n%s\nsecond:\n%s", first, second)
	}
	if n := strings.Count(second, blockStart); n != 1 {
		t.Errorf("idempotent re-run produced %d blocks, want 1:\n%s", n, second)
	}
}

func TestEnsureHiddenReplacesStaleBlockInPlace(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, ".git", "info", "exclude")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir .git/info: %v", err)
	}
	// A pre-existing block with a stale envelope entry, surrounded by user
	// content, must be replaced in place — not duplicated, and the surrounding
	// content preserved.
	stale := "build/\n" + blockStart + "\n.agent/\n/OLD-ENTRY\n" + blockEnd + "\nscratch/\n"
	if err := os.WriteFile(path, []byte(stale), 0o644); err != nil {
		t.Fatalf("seed exclude: %v", err)
	}

	if _, err := EnsureHidden(dir); err != nil {
		t.Fatalf("EnsureHidden() error = %v", err)
	}
	content := readFile(t, path)

	if strings.Contains(content, "/OLD-ENTRY") {
		t.Errorf("stale block entry survived:\n%s", content)
	}
	if n := strings.Count(content, blockStart); n != 1 {
		t.Errorf("got %d blocks, want 1:\n%s", n, content)
	}
	for _, surrounding := range []string{"build/", "scratch/"} {
		if !strings.Contains(content, surrounding) {
			t.Errorf("surrounding content %q lost:\n%s", surrounding, content)
		}
	}
	if !strings.Contains(content, "/Justfile") {
		t.Errorf("refreshed block missing current envelope:\n%s", content)
	}
}

// TestEnsureHiddenWritesNoGitignore guards the core distinction from local
// mode: hidden must leave the committed .gitignore untouched.
func TestEnsureHiddenWritesNoGitignore(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if _, err := EnsureHidden(dir); err != nil {
		t.Fatalf("EnsureHidden() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".gitignore")); !os.IsNotExist(err) {
		t.Errorf("EnsureHidden touched .gitignore, stat err = %v", err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(b)
}
