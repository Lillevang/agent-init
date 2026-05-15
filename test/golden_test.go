package test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Lillevang/agent-init/internal/testflags"
)

type goldenCase struct {
	name       string
	flavor     string
	agentsOnly bool
}

func TestFlavorGolden(t *testing.T) {
	cases := []goldenCase{
		{name: "claude-cowork", flavor: "claude-cowork"},
		{name: "fullstack", flavor: "fullstack"},
		{name: "fullstack-agents-only", flavor: "fullstack", agentsOnly: true},
		{name: "go-backend", flavor: "go-backend"},
		{name: "go-backend-agents-only", flavor: "go-backend", agentsOnly: true},
		{name: "go-cli", flavor: "go-cli"},
		{name: "go-cli-agents-only", flavor: "go-cli", agentsOnly: true},
		{name: "iac", flavor: "iac"},
		{name: "iac-agents-only", flavor: "iac", agentsOnly: true},
		{name: "project-management", flavor: "project-management"},
	}
	binary := buildAgentInit(t)
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			target := filepath.Join(t.TempDir(), c.name)
			args := []string{"init", "--no-git"}
			if c.agentsOnly {
				args = append(args, "--agents-only")
			}
			args = append(args, c.flavor, target)
			runAgentInit(t, binary, args...)
			runGeneratedCodemap(t, target)

			golden := filepath.Join("..", "testdata", "golden", c.name)
			if *testflags.Update {
				if err := os.RemoveAll(golden); err != nil {
					t.Fatalf("remove golden: %v", err)
				}
				if err := copyTree(target, golden); err != nil {
					t.Fatalf("update golden: %v", err)
				}
				return
			}
			if err := compareTrees(golden, target); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func buildAgentInit(t *testing.T) string {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "agent-init")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	cmd := exec.Command("go", "build", "-o", binary, "../cmd/agent-init")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go build: %v\n%s", err, string(output))
	}
	return binary
}

func runAgentInit(t *testing.T, binary string, args ...string) {
	t.Helper()
	cmd := exec.Command(binary, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("agent-init %v: %v\n%s", args, err, string(output))
	}
}

func runGeneratedCodemap(t *testing.T, target string) {
	t.Helper()
	script := filepath.Join(target, ".agent", "scripts", "gen-codemap.sh")
	if _, err := os.Stat(script); err != nil {
		// Flavor doesn't ship a codemap script (e.g. doc-collab flavors); nothing to regenerate.
		return
	}
	cmd := exec.Command("./.agent/scripts/gen-codemap.sh")
	cmd.Dir = target
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("gen-codemap.sh: %v\n%s", err, string(output))
	}
}

func compareTrees(wantRoot, gotRoot string) error {
	wantEntries, err := collectTree(wantRoot)
	if err != nil {
		return fmt.Errorf("collect golden: %w", err)
	}
	gotEntries, err := collectTree(gotRoot)
	if err != nil {
		return fmt.Errorf("collect scaffold: %w", err)
	}
	for path, want := range wantEntries {
		got, ok := gotEntries[path]
		if !ok {
			return fmt.Errorf("missing %s", path)
		}
		if err := compareEntry(path, want, got); err != nil {
			return err
		}
		delete(gotEntries, path)
	}
	for path := range gotEntries {
		return fmt.Errorf("unexpected %s", path)
	}
	return nil
}

type treeEntry struct {
	mode    os.FileMode
	content []byte
	link    string
}

func collectTree(root string) (map[string]treeEntry, error) {
	entries := map[string]treeEntry{}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == root {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		info, err := os.Lstat(path)
		if err != nil {
			return err
		}
		item := treeEntry{mode: info.Mode()}
		switch {
		case info.Mode()&os.ModeSymlink != 0:
			item.link, err = os.Readlink(path)
		case info.Mode().IsRegular():
			item.content, err = os.ReadFile(path)
		}
		if err != nil {
			return err
		}
		entries[rel] = item
		return nil
	})
	return entries, err
}

func compareEntry(path string, want, got treeEntry) error {
	if want.mode.Type() != got.mode.Type() {
		return fmt.Errorf("%s type = %v, want %v", path, got.mode.Type(), want.mode.Type())
	}
	if runtime.GOOS != "windows" && want.mode.Perm() != got.mode.Perm() {
		return fmt.Errorf("%s mode = %v, want %v", path, got.mode.Perm(), want.mode.Perm())
	}
	if want.link != got.link {
		return fmt.Errorf("%s link = %q, want %q", path, got.link, want.link)
	}
	if !bytes.Equal(want.content, got.content) {
		return fmt.Errorf("%s content differs:\n%s", path, contentDiff(want.content, got.content))
	}
	return nil
}

// contentDiff returns a short unified diff between want and got, or a
// fallback message if the system has no `diff` binary. The output is capped
// so a wildly-divergent file doesn't drown the test log.
func contentDiff(want, got []byte) string {
	if _, err := exec.LookPath("diff"); err != nil {
		return fmt.Sprintf("(want %d bytes, got %d bytes; install `diff` for a unified view)", len(want), len(got))
	}
	wantFile, err := writeTempFile(want)
	if err != nil {
		return fmt.Sprintf("(diff unavailable: %v)", err)
	}
	defer func() { _ = os.Remove(wantFile) }()
	gotFile, err := writeTempFile(got)
	if err != nil {
		return fmt.Sprintf("(diff unavailable: %v)", err)
	}
	defer func() { _ = os.Remove(gotFile) }()
	cmd := exec.Command("diff", "-u", "--label", "golden", "--label", "scaffolded", wantFile, gotFile)
	output, _ := cmd.CombinedOutput()
	lines := strings.Split(strings.TrimRight(string(output), "\n"), "\n")
	const maxLines = 60
	if len(lines) > maxLines {
		lines = append(lines[:maxLines], fmt.Sprintf("... (%d more diff lines suppressed)", len(lines)-maxLines))
	}
	return strings.Join(lines, "\n")
}

func writeTempFile(data []byte) (string, error) {
	f, err := os.CreateTemp("", "golden-*.txt")
	if err != nil {
		return "", err
	}
	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		_ = os.Remove(f.Name())
		return "", err
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(f.Name())
		return "", err
	}
	return f.Name(), nil
}

func copyTree(srcRoot, dstRoot string) error {
	return filepath.WalkDir(srcRoot, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(srcRoot, path)
		if err != nil {
			return err
		}
		dst := filepath.Join(dstRoot, rel)
		info, err := os.Lstat(path)
		if err != nil {
			return err
		}
		switch {
		case path == srcRoot:
			return os.MkdirAll(dst, 0o755)
		case info.Mode()&os.ModeSymlink != 0:
			link, err := os.Readlink(path)
			if err != nil {
				return err
			}
			return os.Symlink(link, dst)
		case entry.IsDir():
			return os.MkdirAll(dst, info.Mode().Perm())
		default:
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
				return err
			}
			return os.WriteFile(dst, content, info.Mode().Perm())
		}
	})
}
