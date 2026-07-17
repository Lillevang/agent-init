package cli_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Lillevang/agent-init/internal/cli"
	"github.com/Lillevang/agent-init/internal/gitignore"
)

// statusSnapshot records the on-disk state of every file `status` could
// possibly examine, so tests can prove the subcommand is read-only by diffing
// the snapshot before and after the run. A missing file is represented by
// (exists: false), distinct from an empty file (exists: true, content: "").
type statusSnapshot struct {
	localExists, hiddenExists, globalExists bool
	local, hidden, global                   string
}

func snapshotStatusFiles(t *testing.T, target, global string) statusSnapshot {
	t.Helper()
	read := func(path string) (bool, string) {
		b, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				return false, ""
			}
			t.Fatalf("snapshot read %s: %v", path, err)
		}
		return true, string(b)
	}
	localPath, _ := gitignore.LocalPath(target)
	hiddenPath, _ := gitignore.HiddenPath(target)
	var s statusSnapshot
	s.localExists, s.local = read(localPath)
	s.hiddenExists, s.hidden = read(hiddenPath)
	s.globalExists, s.global = read(global)
	return s
}

func assertStatusReadOnly(t *testing.T, before, after statusSnapshot) {
	t.Helper()
	if before != after {
		t.Errorf("status mutated files (must be read-only)\nbefore: %+v\nafter:  %+v", before, after)
	}
}

func TestStatusSharedWhenNoBlockAnywhere(t *testing.T) {
	t.Parallel()
	target := t.TempDir()
	var out, errOut bytes.Buffer
	app := cli.New(&out, &errOut, cli.Version{})

	if err := app.Run(context.Background(), []string{"status", target}); err != nil {
		t.Fatalf("Run(status) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{"mode:   shared", "target: " + target, "no agent-init ignore block found"} {
		if !strings.Contains(got, want) {
			t.Errorf("status output missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "ignore:") {
		t.Errorf("shared mode should not print an ignore: line:\n%s", got)
	}
	if errOut.Len() != 0 {
		t.Errorf("status wrote to stderr; want stdout only:\n%s", errOut.String())
	}
}

func TestStatusLocalWhenBlockInGitignore(t *testing.T) {
	t.Parallel()
	target := t.TempDir()
	gitignorePath := filepath.Join(target, ".gitignore")
	// Seed an existing .gitignore that already has the managed block, so the
	// scenario looks exactly like an `init --visibility=local` run.
	if err := os.WriteFile(gitignorePath, []byte("node_modules/\n\n"+gitignore.Block()), 0o644); err != nil {
		t.Fatalf("seed .gitignore: %v", err)
	}

	before := snapshotStatusFiles(t, target, "")

	var out bytes.Buffer
	app := cli.New(&out, &bytes.Buffer{}, cli.Version{})
	if err := app.Run(context.Background(), []string{"status", target}); err != nil {
		t.Fatalf("Run(status) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"mode:   local",
		"ignore: " + gitignorePath,
		"undo:",
		gitignore.MarkerStart,
		gitignore.MarkerEnd,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("status output missing %q:\n%s", want, got)
		}
	}

	after := snapshotStatusFiles(t, target, "")
	assertStatusReadOnly(t, before, after)
}

func TestStatusHiddenWhenBlockInGitInfoExclude(t *testing.T) {
	t.Parallel()
	target := t.TempDir()
	hiddenPath := filepath.Join(target, ".git", "info", "exclude")
	if err := os.MkdirAll(filepath.Dir(hiddenPath), 0o755); err != nil {
		t.Fatalf("mkdir .git/info: %v", err)
	}
	if err := os.WriteFile(hiddenPath, []byte(gitignore.Block()), 0o644); err != nil {
		t.Fatalf("seed exclude: %v", err)
	}

	before := snapshotStatusFiles(t, target, "")

	var out bytes.Buffer
	app := cli.New(&out, &bytes.Buffer{}, cli.Version{})
	if err := app.Run(context.Background(), []string{"status", target}); err != nil {
		t.Fatalf("Run(status) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"mode:   hidden",
		"ignore: " + hiddenPath,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("status output missing %q:\n%s", want, got)
		}
	}

	after := snapshotStatusFiles(t, target, "")
	assertStatusReadOnly(t, before, after)
}

// TestStatusShadowedByGlobalWhenOnlyGlobalCarriesBlock seeds a managed block in
// the machine-wide excludes file and verifies status reports shadowed-by-global
// plus the force-add hint. The fake HOME pattern (HOME + GIT_CONFIG_GLOBAL +
// XDG_CONFIG_HOME) mirrors isolateGlobalGitConfig so the developer's real
// global config is never read or written.
func TestStatusShadowedByGlobalWhenOnlyGlobalCarriesBlock(t *testing.T) {
	home := isolateGlobalGitConfig(t)
	globalPath := filepath.Join(home, ".config", "git", "ignore")
	if err := os.MkdirAll(filepath.Dir(globalPath), 0o755); err != nil {
		t.Fatalf("mkdir global excludes dir: %v", err)
	}
	if err := os.WriteFile(globalPath, []byte(gitignore.Block()), 0o644); err != nil {
		t.Fatalf("seed global excludes: %v", err)
	}

	target := filepath.Join(t.TempDir(), "proj")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatalf("mkdir target: %v", err)
	}

	before := snapshotStatusFiles(t, target, globalPath)

	var out bytes.Buffer
	app := cli.New(&out, &bytes.Buffer{}, cli.Version{})
	if err := app.Run(context.Background(), []string{"status", target}); err != nil {
		t.Fatalf("Run(status) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"mode:   shadowed-by-global",
		"ignore: " + globalPath,
		"(machine-wide)",
		"note:",
		"git add -f",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("status output missing %q:\n%s", want, got)
		}
	}

	after := snapshotStatusFiles(t, target, globalPath)
	assertStatusReadOnly(t, before, after)
}

// TestStatusPrefersLocalOverHiddenOverGlobal locks down the precedence rule:
// when blocks exist in multiple carriers simultaneously, status reports the
// most-local (highest-precedence) one. Git's precedence is .gitignore >
// .git/info/exclude > core.excludesfile, and the reported mode follows that.
func TestStatusPrefersLocalOverHiddenOverGlobal(t *testing.T) {
	home := isolateGlobalGitConfig(t)
	globalPath := filepath.Join(home, ".config", "git", "ignore")
	if err := os.MkdirAll(filepath.Dir(globalPath), 0o755); err != nil {
		t.Fatalf("mkdir global excludes dir: %v", err)
	}
	if err := os.WriteFile(globalPath, []byte(gitignore.Block()), 0o644); err != nil {
		t.Fatalf("seed global excludes: %v", err)
	}

	target := filepath.Join(t.TempDir(), "proj")
	hiddenPath := filepath.Join(target, ".git", "info", "exclude")
	if err := os.MkdirAll(filepath.Dir(hiddenPath), 0o755); err != nil {
		t.Fatalf("mkdir .git/info: %v", err)
	}
	if err := os.WriteFile(hiddenPath, []byte(gitignore.Block()), 0o644); err != nil {
		t.Fatalf("seed exclude: %v", err)
	}
	localPath := filepath.Join(target, ".gitignore")
	if err := os.WriteFile(localPath, []byte(gitignore.Block()), 0o644); err != nil {
		t.Fatalf("seed .gitignore: %v", err)
	}

	var out bytes.Buffer
	app := cli.New(&out, &bytes.Buffer{}, cli.Version{})
	if err := app.Run(context.Background(), []string{"status", target}); err != nil {
		t.Fatalf("Run(status) error = %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "mode:   local") {
		t.Errorf("status should pick most-local carrier (local), got:\n%s", got)
	}
	if strings.Contains(got, "shadowed-by-global") || strings.Contains(got, "mode:   hidden") {
		t.Errorf("status leaked lower-precedence mode in output:\n%s", got)
	}
}

func TestStatusDefaultTargetIsCwd(t *testing.T) {
	// t.Chdir (Go 1.24+) restores cwd on cleanup and refuses t.Parallel, which
	// is exactly what we want here: many sibling tests resolve filepath.Abs(".")
	// and would race if this one ran in parallel.
	target := t.TempDir()
	t.Chdir(target)

	var out bytes.Buffer
	app := cli.New(&out, &bytes.Buffer{}, cli.Version{})
	if err := app.Run(context.Background(), []string{"status"}); err != nil {
		t.Fatalf("Run(status) error = %v", err)
	}
	// The target line must contain the resolved cwd. Some platforms canonicalize
	// the path differently (macOS /var/folders -> /private/var/folders), so we
	// check via the same resolution status uses internally: filepath.Abs(".").
	absCwd, err := filepath.Abs(".")
	if err != nil {
		t.Fatalf("abs(.): %v", err)
	}
	if !strings.Contains(out.String(), "target: "+absCwd) {
		t.Errorf("default target should be cwd %q:\n%s", absCwd, out.String())
	}
}

func TestStatusExplicitTargetIsResolvedToAbsolutePath(t *testing.T) {
	t.Parallel()
	target := t.TempDir()
	// Pass a non-clean relative-looking form (the trailing slash and parent
	// hop) to confirm status canonicalizes before printing.
	noisy := target + string(filepath.Separator) + "."

	var out bytes.Buffer
	app := cli.New(&out, &bytes.Buffer{}, cli.Version{})
	if err := app.Run(context.Background(), []string{"status", noisy}); err != nil {
		t.Fatalf("Run(status) error = %v", err)
	}
	if !strings.Contains(out.String(), "target: "+target) {
		t.Errorf("explicit target not canonicalized to %q:\n%s", target, out.String())
	}
}

func TestStatusRejectsExtraPositionalArgs(t *testing.T) {
	t.Parallel()
	app := cli.New(&bytes.Buffer{}, &bytes.Buffer{}, cli.Version{})
	err := app.Run(context.Background(), []string{"status", t.TempDir(), "extra"})
	if err == nil {
		t.Fatal("Run(status target extra) = nil, want usage error")
	}
	if !strings.Contains(err.Error(), "--help") {
		t.Errorf("usage error should reference --help; got: %v", err)
	}
}

func TestStatusHelp(t *testing.T) {
	t.Parallel()
	for _, args := range [][]string{{"status", "--help"}, {"status", "-h"}, {"help", "status"}} {
		args := args
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			t.Parallel()
			var out bytes.Buffer
			app := cli.New(&out, &bytes.Buffer{}, cli.Version{})
			if err := app.Run(context.Background(), args); err != nil {
				t.Fatalf("Run(%v) error = %v", args, err)
			}
			for _, want := range []string{"status", "Usage", "read-only"} {
				if !strings.Contains(out.String(), want) {
					t.Errorf("help (%v) missing %q:\n%s", args, want, out.String())
				}
			}
		})
	}
}

// TestStatusIgnoresGlobalLookupErrors verifies status does not crash when the
// global-excludes resolver can't locate a HOME — the user still gets the
// repo-local detection result, which is the part they care about.
func TestStatusIgnoresGlobalLookupErrors(t *testing.T) {
	// Force os.UserHomeDir to fail by clearing the env vars it reads. On Linux
	// HOME is consulted; on darwin HOME or USER. Both cleared, plus XDG_CONFIG_HOME
	// to make sure the XDG fallback path is also unreachable.
	t.Setenv("HOME", "")
	t.Setenv("USER", "")
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(t.TempDir(), "no-such-gitconfig"))

	target := t.TempDir()
	var out bytes.Buffer
	app := cli.New(&out, &bytes.Buffer{}, cli.Version{})
	if err := app.Run(context.Background(), []string{"status", target}); err != nil {
		t.Fatalf("Run(status) with broken HOME error = %v; status should degrade gracefully", err)
	}
	if !strings.Contains(out.String(), "mode:   shared") {
		t.Errorf("status fell off the rails when HOME was unreadable:\n%s", out.String())
	}
}
