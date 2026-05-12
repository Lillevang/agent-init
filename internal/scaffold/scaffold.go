package scaffold

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/mikeschinkel/agent-init/internal/flavors"
)

type Options struct {
	Flavor  flavors.Flavor
	Target  string
	Force   bool
	InitGit bool
	DryRun  bool
	Out     io.Writer
}

type templateData struct {
	ProjectName string
}

func Run(ctx context.Context, opts Options) error {
	out := opts.Out
	if out == nil {
		out = io.Discard
	}
	target, err := prepareTarget(opts.Target, opts.DryRun)
	if err != nil {
		return err
	}
	data := templateData{ProjectName: filepath.Base(target)}
	fmt.Fprintf(out, "-> Scaffolding %s agentic dev environment in: %s\n", opts.Flavor.Name, target)
	if err := writeTemplates(opts, target, data, out); err != nil {
		return err
	}
	if err := createSymlinks(opts, target, out); err != nil {
		return err
	}
	if opts.InitGit {
		if err := initGit(ctx, target, opts.DryRun, out); err != nil {
			return err
		}
	}
	printNextSteps(out, target)
	return nil
}

func prepareTarget(target string, dryRun bool) (string, error) {
	if target == "" {
		target = "."
	}
	if !dryRun {
		if err := os.MkdirAll(target, 0o755); err != nil {
			return "", fmt.Errorf("creating target directory: %w", err)
		}
	}
	abs, err := filepath.Abs(target)
	if err != nil {
		return "", fmt.Errorf("resolving target directory: %w", err)
	}
	if dryRun {
		return filepath.Clean(abs), nil
	}
	real, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", fmt.Errorf("resolving target symlinks: %w", err)
	}
	return real, nil
}

func writeTemplates(opts Options, target string, data templateData, out io.Writer) error {
	rootFS, err := fs.Sub(opts.Flavor.Templates, opts.Flavor.TemplateRoot)
	if err != nil {
		return fmt.Errorf("opening template root: %w", err)
	}
	return fs.WalkDir(rootFS, ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		content, err := fs.ReadFile(rootFS, path)
		if err != nil {
			return fmt.Errorf("reading template %s: %w", path, err)
		}
		rel := strings.TrimSuffix(path, ".tmpl")
		rendered, err := render(path, content, data)
		if err != nil {
			return err
		}
		return writeFile(opts, target, rel, rendered, out)
	})
}

func writeFile(opts Options, target, rel string, content []byte, out io.Writer) error {
	dst := filepath.Join(target, filepath.FromSlash(rel))
	info, exists, err := lstat(dst)
	if err != nil {
		return fmt.Errorf("checking %s: %w", rel, err)
	}
	if exists && !opts.Force {
		fmt.Fprintf(out, "  skip   %s (exists, use --force to overwrite)\n", rel)
		return nil
	}
	if opts.DryRun {
		fmt.Fprintf(out, "  write  %s (dry-run)\n", rel)
		return nil
	}
	if exists {
		if info.IsDir() {
			return fmt.Errorf("refusing to overwrite directory %s", rel)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			if err := os.Remove(dst); err != nil {
				return fmt.Errorf("removing existing symlink %s: %w", rel, err)
			}
		}
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("creating parent directory for %s: %w", rel, err)
	}
	mode := fs.FileMode(0o644)
	if executable(rel, opts.Flavor.ExecutablePaths) {
		mode = 0o755
	}
	if err := os.WriteFile(dst, content, mode); err != nil {
		return fmt.Errorf("writing %s: %w", rel, err)
	}
	if err := os.Chmod(dst, mode); err != nil {
		return fmt.Errorf("setting permissions on %s: %w", rel, err)
	}
	fmt.Fprintf(out, "  write  %s\n", rel)
	return nil
}

func render(path string, content []byte, data templateData) ([]byte, error) {
	if !strings.HasSuffix(path, ".tmpl") {
		return content, nil
	}
	tmpl, err := template.New(filepath.Base(path)).Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("parsing template %s: %w", path, err)
	}
	var out bytes.Buffer
	if err := tmpl.Execute(&out, data); err != nil {
		return nil, fmt.Errorf("rendering template %s: %w", path, err)
	}
	return out.Bytes(), nil
}

func createSymlinks(opts Options, target string, out io.Writer) error {
	if err := link(opts, target, "AGENTS.md", ".agent/AGENTS.md", out); err != nil {
		return err
	}
	if err := link(opts, target, "CLAUDE.md", ".agent/CLAUDE.md", out); err != nil {
		return err
	}
	return link(opts, filepath.Join(target, ".agent"), "CLAUDE.md", "AGENTS.md", out)
}

func link(opts Options, dir, name, dest string, out io.Writer) error {
	path := filepath.Join(dir, name)
	display := strings.TrimPrefix(filepath.ToSlash(strings.TrimPrefix(path, opts.Target)), "/")
	if display == "" || strings.HasPrefix(display, "..") {
		display = name
	}
	info, exists, err := lstat(path)
	if err != nil {
		return fmt.Errorf("checking %s: %w", display, err)
	}
	if exists && !opts.Force {
		return nil
	}
	if opts.DryRun {
		fmt.Fprintf(out, "  link   %s -> %s (dry-run)\n", display, dest)
		return nil
	}
	if exists {
		if info.IsDir() {
			return fmt.Errorf("refusing to replace directory %s with symlink", display)
		}
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("removing existing %s: %w", display, err)
		}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating parent directory for symlink %s: %w", display, err)
	}
	if err := os.Symlink(dest, path); err != nil {
		return fmt.Errorf("creating symlink %s: %w", display, err)
	}
	fmt.Fprintf(out, "  link   %s -> %s\n", display, dest)
	return nil
}

func initGit(ctx context.Context, target string, dryRun bool, out io.Writer) error {
	if insideGitWorkTree(ctx, target) {
		return nil
	}
	if dryRun {
		fmt.Fprintln(out, "  init   git repository (dry-run)")
		return nil
	}
	cmd := exec.CommandContext(ctx, "git", "init", "-q")
	cmd.Dir = target
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("initializing git repository: %w", err)
	}
	fmt.Fprintln(out, "  init   git repository")
	return nil
}

func printNextSteps(out io.Writer, target string) {
	fmt.Fprintf(out, `
Done.

Next steps:
  1. Read %s/README.agent.md for dependency install instructions
  2. Edit .agent/AGENTS.md to describe THIS project's specifics
  3. Edit .agent/CODEBASE.md once you have code to map
  4. Run:  devcontainer up --workspace-folder %s
  5. Run:  devcontainer exec --workspace-folder %s bash
  6. Inside the container: just check
`, target, target, target)
}

func insideGitWorkTree(ctx context.Context, target string) bool {
	cmd := exec.CommandContext(ctx, "git", "-C", target, "rev-parse", "--is-inside-work-tree")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

func lstat(path string) (fs.FileInfo, bool, error) {
	info, err := os.Lstat(path)
	if err == nil {
		return info, true, nil
	}
	if os.IsNotExist(err) {
		return nil, false, nil
	}
	return nil, false, err
}

func executable(path string, executablePaths []string) bool {
	for _, item := range executablePaths {
		if path == item {
			return true
		}
	}
	return false
}
