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
		t.Fatalf("Run(init --help) error = %v; an explicit --help should not surface as an error", err)
	}
	if !bytes.Contains(out.Bytes(), []byte("--force")) {
		t.Fatalf("init --help did not print flag usage to stdout:\n%s", out.String())
	}
	if errOut.Len() != 0 {
		t.Fatalf("init --help wrote to stderr; explicit help belongs on stdout:\n%s", errOut.String())
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

func TestInitVisibilityLocalAppendsGitignoreBlock(t *testing.T) {
	t.Parallel()
	target := filepath.Join(t.TempDir(), "proj")
	var out bytes.Buffer
	app := cli.New(&out, &bytes.Buffer{}, cli.Version{})

	err := app.Run(context.Background(), []string{"init", "--no-git", "--visibility=local", "go-cli", target})
	if err != nil {
		t.Fatalf("Run(init --visibility=local go-cli) error = %v", err)
	}

	gitignore := filepath.Join(target, ".gitignore")
	content, err := os.ReadFile(gitignore)
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	for _, want := range []string{"agent-init", ".agent/", "/AGENTS.md", ".devcontainer/", "/Justfile"} {
		if !strings.Contains(string(content), want) {
			t.Errorf(".gitignore missing %q:\n%s", want, content)
		}
	}
	// The side effect must be announced with the absolute path.
	if !strings.Contains(out.String(), gitignore) {
		t.Errorf("init output did not announce the .gitignore path %q:\n%s", gitignore, out.String())
	}
}

func TestInitVisibilityLocalIsIdempotent(t *testing.T) {
	t.Parallel()
	target := filepath.Join(t.TempDir(), "proj")
	run := func() {
		app := cli.New(&bytes.Buffer{}, &bytes.Buffer{}, cli.Version{})
		if err := app.Run(context.Background(), []string{"init", "--no-git", "--force", "--visibility=local", "go-cli", target}); err != nil {
			t.Fatalf("Run error = %v", err)
		}
	}
	run()
	run()

	content, err := os.ReadFile(filepath.Join(target, ".gitignore"))
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	if n := strings.Count(string(content), "agent-init (private)"); n != 1 {
		t.Errorf("got %d ignore blocks after re-run, want 1:\n%s", n, content)
	}
}

func TestInitVisibilitySharedLeavesNoGitignore(t *testing.T) {
	t.Parallel()
	target := filepath.Join(t.TempDir(), "proj")
	app := cli.New(&bytes.Buffer{}, &bytes.Buffer{}, cli.Version{})

	// Default (no flag) and explicit --visibility=shared must both behave like
	// today: no agent-init block written.
	if err := app.Run(context.Background(), []string{"init", "--no-git", "--visibility=shared", "go-cli", target}); err != nil {
		t.Fatalf("Run(init --visibility=shared) error = %v", err)
	}
	if content, err := os.ReadFile(filepath.Join(target, ".gitignore")); err == nil {
		if strings.Contains(string(content), "agent-init (private)") {
			t.Errorf("shared visibility wrote the ignore block:\n%s", content)
		}
	}
}

func TestInitVisibilityLocalRejectedOnDocCollabFlavor(t *testing.T) {
	t.Parallel()
	app := cli.New(&bytes.Buffer{}, &bytes.Buffer{}, cli.Version{})

	err := app.Run(context.Background(), []string{"init", "--no-git", "--visibility=local", "project-management", t.TempDir()})
	if err == nil {
		t.Fatal("Run(init --visibility=local project-management) error = nil; want rejection")
	}
	if !strings.Contains(err.Error(), "visibility") {
		t.Fatalf("error = %v; want to mention --visibility", err)
	}
}

// All four visibility modes are now implemented, so there is no
// "unimplemented modes rejected" test. Unknown values are covered by
// TestInitVisibilityUnknownValueRejected; per-mode behavior is covered by the
// hidden and global-default tests below.

// hiddenExcludePath is the repo-local exclude file hidden mode writes to.
func hiddenExcludePath(target string) string {
	return filepath.Join(target, ".git", "info", "exclude")
}

func TestInitVisibilityHiddenWritesGitInfoExclude(t *testing.T) {
	t.Parallel()
	target := filepath.Join(t.TempDir(), "proj")
	var out bytes.Buffer
	app := cli.New(&out, &bytes.Buffer{}, cli.Version{})

	if err := app.Run(context.Background(), []string{"init", "--no-git", "--visibility=hidden", "go-cli", target}); err != nil {
		t.Fatalf("Run(init --visibility=hidden go-cli) error = %v", err)
	}

	exclude := hiddenExcludePath(target)
	content, err := os.ReadFile(exclude)
	if err != nil {
		t.Fatalf("read .git/info/exclude: %v", err)
	}
	for _, want := range []string{"agent-init", ".agent/", "/AGENTS.md", ".devcontainer/", "/Justfile"} {
		if !strings.Contains(string(content), want) {
			t.Errorf(".git/info/exclude missing %q:\n%s", want, content)
		}
	}
	// Hidden mode must leave no committed trace. The flavor ships its own
	// .gitignore, so the file exists — but our block must not be in it.
	if gi, err := os.ReadFile(filepath.Join(target, ".gitignore")); err == nil {
		if strings.Contains(string(gi), "agent-init (private)") {
			t.Errorf("hidden mode wrote the block to the committed .gitignore:\n%s", gi)
		}
	}
	// The side effect must be announced with the absolute path.
	if !strings.Contains(out.String(), exclude) {
		t.Errorf("init output did not announce the exclude path %q:\n%s", exclude, out.String())
	}
}

// TestInitPrivateAliasMatchesHidden checks --private is a true alias: it writes
// the same block to the same file as --visibility=hidden and nothing else.
func TestInitPrivateAliasMatchesHidden(t *testing.T) {
	t.Parallel()
	target := filepath.Join(t.TempDir(), "proj")
	app := cli.New(&bytes.Buffer{}, &bytes.Buffer{}, cli.Version{})

	if err := app.Run(context.Background(), []string{"init", "--no-git", "--private", "go-cli", target}); err != nil {
		t.Fatalf("Run(init --private go-cli) error = %v", err)
	}

	exclude := hiddenExcludePath(target)
	content, err := os.ReadFile(exclude)
	if err != nil {
		t.Fatalf("read .git/info/exclude: %v", err)
	}
	if !strings.Contains(string(content), "agent-init (private)") {
		t.Errorf("--private did not write the hidden block:\n%s", content)
	}
	if gi, err := os.ReadFile(filepath.Join(target, ".gitignore")); err == nil {
		if strings.Contains(string(gi), "agent-init (private)") {
			t.Errorf("--private wrote the block to the committed .gitignore:\n%s", gi)
		}
	}
}

// --private and a conflicting --visibility must error rather than silently
// disagree. --private alongside --visibility=hidden is redundant but fine.
func TestInitPrivateConflictsWithVisibility(t *testing.T) {
	t.Parallel()
	// An explicitly-set non-hidden --visibility alongside --private is a
	// conflict — including an explicit --visibility=shared, which must not be
	// silently overridden by the alias.
	for _, mode := range []string{"local", "shared"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			app := cli.New(&bytes.Buffer{}, &bytes.Buffer{}, cli.Version{})
			err := app.Run(context.Background(), []string{"init", "--no-git", "--private", "--visibility=" + mode, "go-cli", t.TempDir()})
			if err == nil {
				t.Fatalf("Run(--private --visibility=%s) error = nil; want conflict rejection", mode)
			}
			if !strings.Contains(err.Error(), "conflict") {
				t.Fatalf("error = %v; want conflict message", err)
			}
		})
	}

	// --private alongside an explicit --visibility=hidden is redundant, not a
	// conflict.
	target := filepath.Join(t.TempDir(), "proj")
	app2 := cli.New(&bytes.Buffer{}, &bytes.Buffer{}, cli.Version{})
	if err := app2.Run(context.Background(), []string{"init", "--no-git", "--private", "--visibility=hidden", "go-cli", target}); err != nil {
		t.Fatalf("Run(--private --visibility=hidden) error = %v; want it to be accepted as redundant", err)
	}
}

func TestInitVisibilityHiddenIsIdempotent(t *testing.T) {
	t.Parallel()
	target := filepath.Join(t.TempDir(), "proj")
	run := func() {
		app := cli.New(&bytes.Buffer{}, &bytes.Buffer{}, cli.Version{})
		if err := app.Run(context.Background(), []string{"init", "--no-git", "--force", "--visibility=hidden", "go-cli", target}); err != nil {
			t.Fatalf("Run error = %v", err)
		}
	}
	run()
	run()

	content, err := os.ReadFile(hiddenExcludePath(target))
	if err != nil {
		t.Fatalf("read .git/info/exclude: %v", err)
	}
	if n := strings.Count(string(content), "agent-init (private)"); n != 1 {
		t.Errorf("got %d ignore blocks after re-run, want 1:\n%s", n, content)
	}
}

// isolateGlobalGitConfig points HOME, XDG_CONFIG_HOME, and GIT_CONFIG_GLOBAL at
// a temp dir so the global-default tests never read or mutate the developer's
// real machine-wide git config. GIT_CONFIG_GLOBAL is what `git config --global`
// reads/writes; pinning it to a temp file is the seam that makes shelling out
// to git safe in tests. Returns the temp HOME.
func isolateGlobalGitConfig(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(home, ".gitconfig"))
	return home
}

func TestInitVisibilityGlobalDefaultWritesExcludesAndWarns(t *testing.T) {
	home := isolateGlobalGitConfig(t)
	var out, errOut bytes.Buffer
	app := cli.New(&out, &errOut, cli.Version{})

	target := filepath.Join(t.TempDir(), "proj")
	if err := app.Run(context.Background(), []string{"init", "--no-git", "--visibility=global-default", "go-cli", target}); err != nil {
		t.Fatalf("Run(--visibility=global-default) error = %v", err)
	}

	// With no core.excludesfile configured, the block lands in the default
	// XDG path under the fake HOME.
	excludes := filepath.Join(home, ".config", "git", "ignore")
	content, err := os.ReadFile(excludes)
	if err != nil {
		t.Fatalf("read global excludes: %v", err)
	}
	if !strings.Contains(string(content), "agent-init (private)") {
		t.Errorf("global excludes missing the block:\n%s", content)
	}
	// The edited path must be announced on stdout.
	if !strings.Contains(out.String(), excludes) {
		t.Errorf("did not announce the global excludes path %q:\n%s", excludes, out.String())
	}
	// The machine-wide warning must be loud (on stderr).
	if !strings.Contains(strings.ToUpper(errOut.String()), "MACHINE-WIDE") || !strings.Contains(errOut.String(), "EVERY") {
		t.Errorf("missing machine-wide warning on stderr:\n%s", errOut.String())
	}
	// The force-add escape hatch is printed.
	if !strings.Contains(out.String(), "git add -f") {
		t.Errorf("missing force-add hint:\n%s", out.String())
	}
}

func TestInitVisibilityGlobalDefaultIsIdempotent(t *testing.T) {
	home := isolateGlobalGitConfig(t)
	run := func() {
		app := cli.New(&bytes.Buffer{}, &bytes.Buffer{}, cli.Version{})
		if err := app.Run(context.Background(), []string{"init", "--no-git", "--force", "--visibility=global-default", "go-cli", filepath.Join(home, "proj")}); err != nil {
			t.Fatalf("Run error = %v", err)
		}
	}
	run()
	run()

	content, err := os.ReadFile(filepath.Join(home, ".config", "git", "ignore"))
	if err != nil {
		t.Fatalf("read global excludes: %v", err)
	}
	if n := strings.Count(string(content), "agent-init (private)"); n != 1 {
		t.Errorf("got %d blocks after re-run, want 1:\n%s", n, content)
	}
}

func TestInitVisibilityHiddenRejectedOnDocCollabFlavor(t *testing.T) {
	t.Parallel()
	app := cli.New(&bytes.Buffer{}, &bytes.Buffer{}, cli.Version{})

	err := app.Run(context.Background(), []string{"init", "--no-git", "--visibility=hidden", "project-management", t.TempDir()})
	if err == nil {
		t.Fatal("Run(init --visibility=hidden project-management) error = nil; want rejection")
	}
	if !strings.Contains(err.Error(), "visibility") {
		t.Fatalf("error = %v; want to mention --visibility", err)
	}
}

func TestInitPrivateRejectedOnDocCollabFlavor(t *testing.T) {
	t.Parallel()
	app := cli.New(&bytes.Buffer{}, &bytes.Buffer{}, cli.Version{})

	err := app.Run(context.Background(), []string{"init", "--no-git", "--private", "project-management", t.TempDir()})
	if err == nil {
		t.Fatal("Run(init --private project-management) error = nil; want rejection")
	}
	if !strings.Contains(err.Error(), "visibility") {
		t.Fatalf("error = %v; want to mention --visibility", err)
	}
}

func TestInitVisibilityHiddenDryRunWritesNothing(t *testing.T) {
	t.Parallel()
	target := filepath.Join(t.TempDir(), "proj")
	var out bytes.Buffer
	app := cli.New(&out, &bytes.Buffer{}, cli.Version{})

	if err := app.Run(context.Background(), []string{"init", "--no-git", "--dry-run", "--private", "go-cli", target}); err != nil {
		t.Fatalf("Run(--dry-run --private) error = %v", err)
	}
	if _, err := os.Stat(hiddenExcludePath(target)); !os.IsNotExist(err) {
		t.Errorf("dry-run wrote a .git/info/exclude, stat err = %v", err)
	}
	if !strings.Contains(out.String(), ".agent/") {
		t.Errorf("dry-run did not preview the ignore block:\n%s", out.String())
	}
}

func TestInitVisibilityGlobalDefaultDryRunWritesNothing(t *testing.T) {
	home := isolateGlobalGitConfig(t)
	var out, errOut bytes.Buffer
	app := cli.New(&out, &errOut, cli.Version{})

	target := filepath.Join(t.TempDir(), "proj")
	if err := app.Run(context.Background(), []string{"init", "--no-git", "--dry-run", "--visibility=global-default", "go-cli", target}); err != nil {
		t.Fatalf("Run(--dry-run --visibility=global-default) error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, ".config", "git", "ignore")); !os.IsNotExist(err) {
		t.Errorf("dry-run wrote the global excludes, stat err = %v", err)
	}
	if !strings.Contains(out.String(), ".agent/") {
		t.Errorf("dry-run did not preview the ignore block:\n%s", out.String())
	}
	// Even dry-run must show the warning so the user understands the scope.
	if !strings.Contains(strings.ToUpper(errOut.String()), "MACHINE-WIDE") {
		t.Errorf("dry-run missing machine-wide warning:\n%s", errOut.String())
	}
}

func TestInitVisibilityGlobalDefaultRejectedOnDocCollabFlavor(t *testing.T) {
	t.Parallel()
	app := cli.New(&bytes.Buffer{}, &bytes.Buffer{}, cli.Version{})

	err := app.Run(context.Background(), []string{"init", "--no-git", "--visibility=global-default", "project-management", t.TempDir()})
	if err == nil {
		t.Fatal("Run(--visibility=global-default project-management) error = nil; want rejection")
	}
	if !strings.Contains(err.Error(), "visibility") {
		t.Fatalf("error = %v; want to mention --visibility", err)
	}
}

func TestInitVisibilityUnknownValueRejected(t *testing.T) {
	t.Parallel()
	app := cli.New(&bytes.Buffer{}, &bytes.Buffer{}, cli.Version{})
	err := app.Run(context.Background(), []string{"init", "--no-git", "--visibility=bogus", "go-cli", t.TempDir()})
	if err == nil {
		t.Fatal("Run(--visibility=bogus) error = nil; want rejection")
	}
	if !strings.Contains(err.Error(), "unknown --visibility") {
		t.Fatalf("error = %v; want unknown-visibility message", err)
	}
}

func TestInitVisibilityLocalDryRunWritesNothing(t *testing.T) {
	t.Parallel()
	target := filepath.Join(t.TempDir(), "proj")
	var out bytes.Buffer
	app := cli.New(&out, &bytes.Buffer{}, cli.Version{})

	if err := app.Run(context.Background(), []string{"init", "--no-git", "--dry-run", "--visibility=local", "go-cli", target}); err != nil {
		t.Fatalf("Run(--dry-run --visibility=local) error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, ".gitignore")); !os.IsNotExist(err) {
		t.Errorf("dry-run wrote a .gitignore, stat err = %v", err)
	}
	if !strings.Contains(out.String(), ".agent/") {
		t.Errorf("dry-run did not preview the ignore block:\n%s", out.String())
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

// TestAddTrackerShipsNoSecretLiteralAndGuidesCredentials guards issue #17:
// the merged .mcp.json must reference credentials from the environment
// (never an empty literal that invites pasting a secret), and the command
// output must tell the user where to set the token and to keep it out of the
// tracked .mcp.json.
func TestAddTrackerShipsNoSecretLiteralAndGuidesCredentials(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		tracker string
		folder  string
		envVar  string
	}{
		{"gh", "github", "GITHUB_TOKEN"},
		{"jira", "jira", "JIRA_API_TOKEN"},
		{"ado", "azure-devops", "ADO_PAT"},
	} {
		tc := tc
		t.Run(tc.tracker, func(t *testing.T) {
			t.Parallel()
			target := filepath.Join(t.TempDir(), "pm")
			var out bytes.Buffer
			app := cli.New(&out, &bytes.Buffer{}, cli.Version{})
			if err := app.Run(context.Background(), []string{"init", "--no-git", "project-management", target}); err != nil {
				t.Fatalf("init: %v", err)
			}
			if err := app.Run(context.Background(), []string{"add-tracker", tc.tracker, target}); err != nil {
				t.Fatalf("add-tracker %s: %v", tc.tracker, err)
			}

			mcp, err := os.ReadFile(filepath.Join(target, ".mcp.json"))
			if err != nil {
				t.Fatalf("read .mcp.json: %v", err)
			}
			if strings.Contains(string(mcp), `: ""`) {
				t.Errorf("%s .mcp.json ships an empty literal (secret invitation):\n%s", tc.tracker, string(mcp))
			}
			if !strings.Contains(string(mcp), "${env:"+tc.envVar+"}") {
				t.Errorf("%s .mcp.json should reference ${env:%s}:\n%s", tc.tracker, tc.envVar, string(mcp))
			}

			if _, err := os.Stat(filepath.Join(target, "integrations", tc.folder, ".env.example")); err != nil {
				t.Errorf("integrations/%s/.env.example should exist: %v", tc.folder, err)
			}

			got := out.String()
			if !strings.Contains(got, tc.envVar) {
				t.Errorf("output should name the env var %s:\n%s", tc.envVar, got)
			}
			if !strings.Contains(got, ".mcp.json") {
				t.Errorf("output should warn about .mcp.json:\n%s", got)
			}
		})
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

func TestTopLevelHelpListsAllSubcommands(t *testing.T) {
	t.Parallel()
	for _, trigger := range [][]string{{"--help"}, {"-h"}, {"help"}} {
		trigger := trigger
		t.Run(strings.Join(trigger, " "), func(t *testing.T) {
			t.Parallel()
			var out bytes.Buffer
			app := cli.New(&out, &bytes.Buffer{}, cli.Version{})
			if err := app.Run(context.Background(), trigger); err != nil {
				t.Fatalf("Run(%v) error = %v; help must exit 0", trigger, err)
			}
			got := out.String()
			for _, sub := range []string{"init", "add-tracker", "list-flavors", "list-trackers", "version"} {
				if !strings.Contains(got, sub) {
					t.Errorf("top-level help missing subcommand %q:\n%s", sub, got)
				}
			}
			if !strings.Contains(got, "--help") {
				t.Errorf("top-level help should point at per-command --help:\n%s", got)
			}
			if !strings.Contains(got, "docs") {
				t.Errorf("top-level help should point at the docs:\n%s", got)
			}
		})
	}
}

func TestSubcommandHelpPrintsFlagsAndExamples(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		want []string
	}{
		{"init", []string{"--force", "--no-git", "--dry-run", "--agents-only", "Examples"}},
		{"add-tracker", []string{"--force", "--dry-run", "gh", "Examples"}},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			// An explicit `<sub> --help` is a success and prints to stdout.
			var out, errOut bytes.Buffer
			app := cli.New(&out, &errOut, cli.Version{})
			if err := app.Run(context.Background(), []string{tc.name, "--help"}); err != nil {
				t.Fatalf("Run(%s --help) error = %v", tc.name, err)
			}
			got := out.String()
			for _, want := range tc.want {
				if !strings.Contains(got, want) {
					t.Errorf("%s --help missing %q:\n%s", tc.name, want, got)
				}
			}

			// `help <sub>` writes the same content to stdout.
			var out2 bytes.Buffer
			app2 := cli.New(&out2, &bytes.Buffer{}, cli.Version{})
			if err := app2.Run(context.Background(), []string{"help", tc.name}); err != nil {
				t.Fatalf("Run(help %s) error = %v", tc.name, err)
			}
			for _, want := range tc.want {
				if !strings.Contains(out2.String(), want) {
					t.Errorf("help %s missing %q:\n%s", tc.name, want, out2.String())
				}
			}
		})
	}
}

func TestFlaglessSubcommandHelp(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"list-flavors", "list-trackers", "version"} {
		name := name
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var out bytes.Buffer
			app := cli.New(&out, &bytes.Buffer{}, cli.Version{})
			if err := app.Run(context.Background(), []string{name, "--help"}); err != nil {
				t.Fatalf("Run(%s --help) error = %v", name, err)
			}
			if !strings.Contains(out.String(), name) {
				t.Errorf("%s --help did not print its own usage:\n%s", name, out.String())
			}
		})
	}
}

func TestUnknownCommandErrorPointsAtHelp(t *testing.T) {
	t.Parallel()
	app := cli.New(&bytes.Buffer{}, &bytes.Buffer{}, cli.Version{})
	err := app.Run(context.Background(), []string{"versoin"})
	if err == nil {
		t.Fatal("Run(unknown command) error = nil; misuse must exit non-zero")
	}
	if !strings.Contains(err.Error(), "--help") {
		t.Fatalf("unknown command error should reference --help; got: %v", err)
	}
}

func TestUnknownFlavorErrorPointsAtHelp(t *testing.T) {
	t.Parallel()
	app := cli.New(&bytes.Buffer{}, &bytes.Buffer{}, cli.Version{})
	err := app.Run(context.Background(), []string{"init", "missing-flavor", t.TempDir()})
	if err == nil {
		t.Fatal("Run(init missing-flavor) error = nil")
	}
	if !strings.Contains(err.Error(), "--help") {
		t.Fatalf("unknown flavor error should reference --help; got: %v", err)
	}
}

// TestHelpFlagsMatchDocs guards the issue-20 "no drift" criterion: every flag
// the binary documents in `<subcommand> --help` must also be described in
// docs/cli.md. Help text is the source of truth; this test fails loudly if the
// docs fall behind it.
func TestHelpFlagsMatchDocs(t *testing.T) {
	t.Parallel()
	docs, err := os.ReadFile(filepath.Join("..", "..", "docs", "cli.md"))
	if err != nil {
		t.Fatalf("read docs/cli.md: %v", err)
	}
	doc := string(docs)

	for _, sub := range []string{"init", "add-tracker"} {
		var out bytes.Buffer
		app := cli.New(&out, &bytes.Buffer{}, cli.Version{})
		if err := app.Run(context.Background(), []string{sub, "--help"}); err != nil {
			t.Fatalf("Run(%s --help) error = %v", sub, err)
		}
		for _, line := range strings.Split(out.String(), "\n") {
			line = strings.TrimSpace(line)
			if !strings.HasPrefix(line, "--") {
				continue
			}
			flag := strings.Fields(line)[0]
			if !strings.Contains(doc, flag) {
				t.Errorf("docs/cli.md does not document flag %q from `%s --help` (drift)", flag, sub)
			}
		}
	}
}

func TestVersion(t *testing.T) {
	t.Parallel()
	var out bytes.Buffer
	app := cli.New(&out, &bytes.Buffer{}, cli.Version{Version: "v1.2.3", Commit: "abc123", BuildDate: "2026-05-12"})

	if err := app.Run(context.Background(), []string{"version"}); err != nil {
		t.Fatalf("Run(version) error = %v", err)
	}
	if got := out.String(); got != "agent-init version=v1.2.3 commit=abc123 buildDate=2026-05-12\n" {
		t.Fatalf("version output = %q", got)
	}
}

func TestVersionDefaultsToDev(t *testing.T) {
	t.Parallel()
	var out bytes.Buffer
	// Mirrors a dev build (`go run`), where ldflags are not applied and the
	// version var keeps its "dev" default.
	app := cli.New(&out, &bytes.Buffer{}, cli.Version{Version: "dev", Commit: "dev", BuildDate: "unknown"})

	if err := app.Run(context.Background(), []string{"version"}); err != nil {
		t.Fatalf("Run(version) error = %v", err)
	}
	if got := out.String(); got != "agent-init version=dev commit=dev buildDate=unknown\n" {
		t.Fatalf("version output = %q", got)
	}
}
