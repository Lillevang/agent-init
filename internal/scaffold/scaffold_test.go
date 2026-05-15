package scaffold_test

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"testing/fstest"

	"github.com/Lillevang/agent-init/internal/flavors"
	"github.com/Lillevang/agent-init/internal/scaffold"
	_ "github.com/Lillevang/agent-init/internal/testflags"
)

func TestRunWritesFullstackScaffold(t *testing.T) {
	t.Parallel()
	target := t.TempDir()
	flavor := mustFlavor(t, "fullstack")
	var out bytes.Buffer

	err := scaffold.Run(context.Background(), scaffold.Options{
		Flavor:  flavor,
		Target:  target,
		Force:   false,
		InitGit: false,
		Out:     &out,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	assertFileContains(t, filepath.Join(target, ".agent", "AGENTS.md"), "Agent Instructions for "+filepath.Base(target))
	assertFileExists(t, filepath.Join(target, "README.agent.md"))
	assertExecutable(t, filepath.Join(target, ".agent", "scripts", "check.sh"))
	assertSymlink(t, filepath.Join(target, ".agent", "CLAUDE.md"), "AGENTS.md")
	assertSymlink(t, filepath.Join(target, "AGENTS.md"), ".agent/AGENTS.md")
	assertSymlink(t, filepath.Join(target, "CLAUDE.md"), ".agent/CLAUDE.md")
}

func TestRunSkipsExistingFilesUnlessForced(t *testing.T) {
	t.Parallel()
	target := t.TempDir()
	flavor := mustFlavor(t, "fullstack")
	path := filepath.Join(target, "README.agent.md")
	if err := os.WriteFile(path, []byte("local edit"), 0o644); err != nil {
		t.Fatalf("write existing file: %v", err)
	}

	err := scaffold.Run(context.Background(), scaffold.Options{
		Flavor:  flavor,
		Target:  target,
		InitGit: false,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	assertFileContains(t, path, "local edit")

	err = scaffold.Run(context.Background(), scaffold.Options{
		Flavor:  flavor,
		Target:  target,
		Force:   true,
		InitGit: false,
	})
	if err != nil {
		t.Fatalf("Run(force) error = %v", err)
	}
	assertFileContains(t, path, "agentic development")
}

func TestRunDryRunDoesNotWriteFiles(t *testing.T) {
	t.Parallel()
	target := filepath.Join(t.TempDir(), "planned")
	flavor := mustFlavor(t, "fullstack")
	var out bytes.Buffer

	err := scaffold.Run(context.Background(), scaffold.Options{
		Flavor:  flavor,
		Target:  target,
		InitGit: false,
		DryRun:  true,
		Out:     &out,
	})
	if err != nil {
		t.Fatalf("Run(dry-run) error = %v", err)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("dry-run created target, stat err = %v", err)
	}
	if !bytes.Contains(out.Bytes(), []byte("(dry-run)")) {
		t.Fatalf("dry-run output did not mention dry-run:\n%s", out.String())
	}
}

func TestRunForceReplacesSymlinkWithoutWritingThroughIt(t *testing.T) {
	t.Parallel()
	target := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside.txt")
	if err := os.WriteFile(outside, []byte("outside"), 0o644); err != nil {
		t.Fatalf("write outside file: %v", err)
	}
	if err := os.Symlink(outside, filepath.Join(target, "README.agent.md")); err != nil {
		t.Fatalf("create symlink: %v", err)
	}

	err := scaffold.Run(context.Background(), scaffold.Options{
		Flavor:  mustFlavor(t, "fullstack"),
		Target:  target,
		Force:   true,
		InitGit: false,
	})
	if err != nil {
		t.Fatalf("Run(force) error = %v", err)
	}
	assertFileContains(t, outside, "outside")
	assertFileContains(t, filepath.Join(target, "README.agent.md"), "agentic development")
}

func TestRunForceRestoresExecutableMode(t *testing.T) {
	t.Parallel()
	target := t.TempDir()
	script := filepath.Join(target, ".agent", "scripts", "check.sh")
	if err := os.MkdirAll(filepath.Dir(script), 0o755); err != nil {
		t.Fatalf("create script dir: %v", err)
	}
	if err := os.WriteFile(script, []byte("#!/usr/bin/env bash\nexit 0\n"), 0o644); err != nil {
		t.Fatalf("write existing script: %v", err)
	}

	err := scaffold.Run(context.Background(), scaffold.Options{
		Flavor:  mustFlavor(t, "fullstack"),
		Target:  target,
		Force:   true,
		InitGit: false,
	})
	if err != nil {
		t.Fatalf("Run(force) error = %v", err)
	}
	assertExecutable(t, script)
}

func TestRunForceRefusesToReplaceDirectory(t *testing.T) {
	t.Parallel()
	target := t.TempDir()
	if err := os.Mkdir(filepath.Join(target, "README.agent.md"), 0o755); err != nil {
		t.Fatalf("create directory: %v", err)
	}

	err := scaffold.Run(context.Background(), scaffold.Options{
		Flavor:  mustFlavor(t, "fullstack"),
		Target:  target,
		Force:   true,
		InitGit: false,
	})
	if err == nil {
		t.Fatal("Run(force over directory) error = nil")
	}
}

func TestRunRendersPathTemplate(t *testing.T) {
	t.Parallel()
	target := filepath.Join(t.TempDir(), "myproj")
	flavor := flavors.Flavor{
		Name:         "test-pathtmpl",
		Templates:    pathTemplateFS(),
		TemplateRoot: "templates",
	}

	err := scaffold.Run(context.Background(), scaffold.Options{
		Flavor:  flavor,
		Target:  target,
		InitGit: false,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	assertFileContains(t, filepath.Join(target, "cmd", "myproj", "main.go"), "package main")
	if _, err := os.Stat(filepath.Join(target, "cmd", "{{.ProjectName}}")); !os.IsNotExist(err) {
		t.Fatalf("literal {{.ProjectName}} directory was not substituted: %v", err)
	}
}

func TestRunLayersFlavorOverCommon(t *testing.T) {
	t.Parallel()
	target := t.TempDir()
	flavor := flavors.Flavor{
		Name: "test-overlay",
		Templates: fstest.MapFS{
			"templates/.agent/scripts/check.sh": &fstest.MapFile{Data: []byte("FLAVOR"), Mode: 0o644},
		},
		TemplateRoot: "templates",
		CommonTemplates: fstest.MapFS{
			"templates/.agent/scripts/check.sh": &fstest.MapFile{Data: []byte("COMMON"), Mode: 0o644},
			"templates/.agent/scripts/extra.sh": &fstest.MapFile{Data: []byte("ONLY-IN-COMMON"), Mode: 0o644},
		},
		CommonRoot: "templates",
	}

	err := scaffold.Run(context.Background(), scaffold.Options{
		Flavor:  flavor,
		Target:  target,
		InitGit: false,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	assertFileContains(t, filepath.Join(target, ".agent", "scripts", "check.sh"), "FLAVOR")
	assertFileContains(t, filepath.Join(target, ".agent", "scripts", "extra.sh"), "ONLY-IN-COMMON")
}

func pathTemplateFS() fstest.MapFS {
	return fstest.MapFS{
		"templates/cmd/{{.ProjectName}}/main.go": &fstest.MapFile{
			Data: []byte("package main\n"),
			Mode: 0o644,
		},
	}
}

func TestRunAgentsOnlySkipsFreshOnlyPathsAndPrefersVariant(t *testing.T) {
	t.Parallel()
	target := filepath.Join(t.TempDir(), "myproj")
	flavor := flavors.Flavor{
		Name: "test-agents-only",
		Templates: fstest.MapFS{
			"templates/cmd/{{.ProjectName}}/main.go": &fstest.MapFile{Data: []byte("FRESH-MAIN"), Mode: 0o644},
			"templates/go.mod":                       &fstest.MapFile{Data: []byte("FRESH-GOMOD"), Mode: 0o644},
			"templates/Justfile.tmpl":                &fstest.MapFile{Data: []byte("FRESH-JUSTFILE-{{.ProjectName}}"), Mode: 0o644},
			"templates/Justfile.agents-only.tmpl":    &fstest.MapFile{Data: []byte("AGENTS-JUSTFILE-{{.ProjectName}}"), Mode: 0o644},
			"templates/README.md":                    &fstest.MapFile{Data: []byte("UNCHANGED"), Mode: 0o644},
		},
		TemplateRoot:       "templates",
		SupportsAgentsOnly: true,
		FreshOnlyPaths: []string{
			"cmd/{{.ProjectName}}/main.go",
			"go.mod",
		},
	}

	err := scaffold.Run(context.Background(), scaffold.Options{
		Flavor:     flavor,
		Target:     target,
		InitGit:    false,
		AgentsOnly: true,
	})
	if err != nil {
		t.Fatalf("Run(agents-only) error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, "cmd")); !os.IsNotExist(err) {
		t.Fatalf("FreshOnlyPaths leaked: cmd/ should not exist, stat err = %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, "go.mod")); !os.IsNotExist(err) {
		t.Fatalf("FreshOnlyPaths leaked: go.mod should not exist, stat err = %v", err)
	}
	assertFileContains(t, filepath.Join(target, "Justfile"), "AGENTS-JUSTFILE-myproj")
	if _, err := os.Stat(filepath.Join(target, "Justfile.agents-only")); !os.IsNotExist(err) {
		t.Fatalf("variant suffix leaked into destination, stat err = %v", err)
	}
	assertFileContains(t, filepath.Join(target, "README.md"), "UNCHANGED")
}

func TestRunFreshModeIgnoresAgentsOnlyVariants(t *testing.T) {
	t.Parallel()
	target := filepath.Join(t.TempDir(), "myproj")
	flavor := flavors.Flavor{
		Name: "test-agents-only-fresh",
		Templates: fstest.MapFS{
			"templates/cmd/{{.ProjectName}}/main.go": &fstest.MapFile{Data: []byte("FRESH-MAIN"), Mode: 0o644},
			"templates/Justfile.tmpl":                &fstest.MapFile{Data: []byte("FRESH-JUSTFILE"), Mode: 0o644},
			"templates/Justfile.agents-only.tmpl":    &fstest.MapFile{Data: []byte("AGENTS-JUSTFILE"), Mode: 0o644},
		},
		TemplateRoot:       "templates",
		SupportsAgentsOnly: true,
		FreshOnlyPaths:     []string{"cmd/{{.ProjectName}}/main.go"},
	}

	err := scaffold.Run(context.Background(), scaffold.Options{
		Flavor:  flavor,
		Target:  target,
		InitGit: false,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	assertFileContains(t, filepath.Join(target, "cmd", "myproj", "main.go"), "FRESH-MAIN")
	assertFileContains(t, filepath.Join(target, "Justfile"), "FRESH-JUSTFILE")
	if _, err := os.Stat(filepath.Join(target, "Justfile.agents-only")); !os.IsNotExist(err) {
		t.Fatalf("agents-only variant leaked in fresh mode, stat err = %v", err)
	}
}

func TestRunDoesNotInitNestedGitRepo(t *testing.T) {
	t.Parallel()
	parent := t.TempDir()
	runGit(t, parent, "init", "-q")
	target := filepath.Join(parent, "child")

	err := scaffold.Run(context.Background(), scaffold.Options{
		Flavor:  mustFlavor(t, "fullstack"),
		Target:  target,
		InitGit: true,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, ".git")); !os.IsNotExist(err) {
		t.Fatalf("nested .git stat err = %v", err)
	}
}

func mustFlavor(t *testing.T, name string) flavors.Flavor {
	t.Helper()
	flavor, err := flavors.DefaultRegistry().Get(name)
	if err != nil {
		t.Fatalf("Get(%q) error = %v", name, err)
	}
	return flavor
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, string(output))
	}
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file %s: %v", path, err)
	}
}

func assertFileContains(t *testing.T, path, want string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if !bytes.Contains(content, []byte(want)) {
		t.Fatalf("%s does not contain %q", path, want)
	}
}

func assertExecutable(t *testing.T, path string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		return
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	if info.Mode()&0o111 == 0 {
		t.Fatalf("%s is not executable: %v", path, info.Mode())
	}
}

func assertSymlink(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.Readlink(path)
	if err != nil {
		t.Fatalf("readlink %s: %v", path, err)
	}
	if got != want {
		t.Fatalf("readlink %s = %q, want %q", path, got, want)
	}
}
