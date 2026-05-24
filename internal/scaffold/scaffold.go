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

	"github.com/Lillevang/agent-init/internal/flavors"
)

type Options struct {
	Flavor  flavors.Flavor
	Target  string
	Force   bool
	InitGit bool
	DryRun  bool
	// AgentsOnly drops the fresh-project files declared in
	// Flavor.FreshOnlyPaths and prefers any ".agents-only" variant template
	// files. Used to add the agentic envelope to an existing project.
	AgentsOnly bool
	Out        io.Writer
	counts     *operationCounts
	style      outputStyle
}

type operationCounts struct {
	written int
	skipped int
	linked  int
}

type outputStyle struct {
	enabled bool
}

const (
	ansiReset     = "\x1b[0m"
	ansiBold      = "\x1b[1m"
	ansiGreen     = "\x1b[32m"
	ansiYellow    = "\x1b[33m"
	ansiCyan      = "\x1b[36m"
	ansiBoldGreen = "\x1b[1;32m"
)

// agentsOnlySuffix marks a template file as the variant to use in
// agents-only mode. The suffix is stripped from the destination path before
// writing, so `Justfile.agents-only.tmpl` writes as `Justfile`. In fresh
// mode the variant is skipped entirely.
const agentsOnlySuffix = ".agents-only"

type templateData struct {
	ProjectName string
}

func Run(ctx context.Context, opts Options) error {
	out := opts.Out
	if out == nil {
		out = io.Discard
	}
	counts := &operationCounts{}
	opts.counts = counts
	style := outputStyle{enabled: colorEnabled(out)}
	opts.style = style
	target, err := prepareTarget(opts.Target, opts.DryRun)
	if err != nil {
		return err
	}
	data := templateData{ProjectName: filepath.Base(target)}
	_, _ = fmt.Fprintf(out, "%s\n", style.header(fmt.Sprintf("-> Scaffolding %s agentic dev environment in: %s", opts.Flavor.Name, target)))
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
	printNextSteps(out, opts.Flavor, target)
	printSummary(out, opts.DryRun, counts, style)
	return nil
}

// Overlay writes a single template layer onto an existing target directory.
// Unlike Run, it does not init git, create symlinks, or print a next-steps
// message — it just walks the layer and writes (or skips, or dry-runs) the
// files using the same engine semantics as a normal scaffold. Use this for
// incremental subcommands like add-tracker that augment an already-scaffolded
// project.
//
// opts.Flavor is ignored except for opts.Force, opts.DryRun, opts.Target,
// opts.Out — the other Flavor fields (Symlinks, NextSteps, etc.) are not used.
func Overlay(opts Options, fsys fs.FS, root string) error {
	out := opts.Out
	if out == nil {
		out = io.Discard
	}
	target, err := prepareTarget(opts.Target, opts.DryRun)
	if err != nil {
		return err
	}
	data := templateData{ProjectName: filepath.Base(target)}
	return walkLayer(opts, fsys, root, target, data, map[string]bool{}, out)
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
	claimed := map[string]bool{}
	layers := []struct {
		fsys fs.FS
		root string
	}{
		{opts.Flavor.Templates, opts.Flavor.TemplateRoot},
	}
	if opts.Flavor.CommonTemplates != nil {
		layers = append(layers, struct {
			fsys fs.FS
			root string
		}{opts.Flavor.CommonTemplates, opts.Flavor.CommonRoot})
	}
	for _, layer := range layers {
		if layer.fsys == nil {
			continue
		}
		if err := walkLayer(opts, layer.fsys, layer.root, target, data, claimed, out); err != nil {
			return err
		}
	}
	return nil
}

func walkLayer(opts Options, fsys fs.FS, root, target string, data templateData, claimed map[string]bool, out io.Writer) error {
	rootFS, err := fs.Sub(fsys, root)
	if err != nil {
		return fmt.Errorf("opening template root %q: %w", root, err)
	}
	// First pass: in agents-only mode, identify destination rels that have a
	// .agents-only variant in this layer so the base file gets shadowed.
	coveredByVariant := map[string]bool{}
	if opts.AgentsOnly {
		if err := fs.WalkDir(rootFS, ".", func(path string, entry fs.DirEntry, err error) error {
			if err != nil || entry.IsDir() {
				return err
			}
			destRel := strings.TrimSuffix(path, ".tmpl")
			if !strings.HasSuffix(destRel, agentsOnlySuffix) {
				return nil
			}
			baseRel := strings.TrimSuffix(destRel, agentsOnlySuffix)
			baseRel, err = renderPath(baseRel, data)
			if err != nil {
				return fmt.Errorf("rendering path %s: %w", path, err)
			}
			coveredByVariant[baseRel] = true
			return nil
		}); err != nil {
			return err
		}
	}
	// Pre-render FreshOnlyPaths once so the per-file check is a plain
	// string comparison.
	freshOnly := renderedFreshOnly(opts, data)
	return fs.WalkDir(rootFS, ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		destRel := strings.TrimSuffix(path, ".tmpl")
		isVariant := strings.HasSuffix(destRel, agentsOnlySuffix)
		if isVariant {
			if !opts.AgentsOnly {
				return nil
			}
			destRel = strings.TrimSuffix(destRel, agentsOnlySuffix)
		}
		destRel, err = renderPath(destRel, data)
		if err != nil {
			return fmt.Errorf("rendering path %s: %w", path, err)
		}
		if opts.AgentsOnly {
			if freshOnly[destRel] {
				return nil
			}
			if !isVariant && coveredByVariant[destRel] {
				return nil
			}
		}
		if claimed[destRel] {
			return nil
		}
		claimed[destRel] = true
		content, err := fs.ReadFile(rootFS, path)
		if err != nil {
			return fmt.Errorf("reading template %s: %w", path, err)
		}
		rendered, err := render(path, content, data)
		if err != nil {
			return err
		}
		return writeFile(opts, target, destRel, rendered, out)
	})
}

// renderedFreshOnly resolves Flavor.FreshOnlyPaths into a set of rendered
// destination paths for the active scaffold. Only meaningful when
// opts.AgentsOnly is set; callers check that before consulting the result.
func renderedFreshOnly(opts Options, data templateData) map[string]bool {
	if !opts.AgentsOnly || len(opts.Flavor.FreshOnlyPaths) == 0 {
		return nil
	}
	out := make(map[string]bool, len(opts.Flavor.FreshOnlyPaths))
	for _, p := range opts.Flavor.FreshOnlyPaths {
		rendered, err := renderPath(p, data)
		if err != nil {
			continue
		}
		out[rendered] = true
	}
	return out
}

func writeFile(opts Options, target, rel string, content []byte, out io.Writer) error {
	dst := filepath.Join(target, filepath.FromSlash(rel))
	info, exists, err := lstat(dst)
	if err != nil {
		return fmt.Errorf("checking %s: %w", rel, err)
	}
	if exists && !opts.Force {
		if opts.counts != nil {
			opts.counts.skipped++
		}
		printOperation(out, opts.style, "skip", "%s (exists, use --force to overwrite)", rel)
		return nil
	}
	if opts.DryRun {
		if opts.counts != nil {
			opts.counts.written++
		}
		printOperation(out, opts.style, "write", "%s (dry-run)", rel)
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
	if opts.counts != nil {
		opts.counts.written++
	}
	printOperation(out, opts.style, "write", "%s", rel)
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

func renderPath(rel string, data templateData) (string, error) {
	if !strings.Contains(rel, "{{") {
		return rel, nil
	}
	tmpl, err := template.New("path").Parse(rel)
	if err != nil {
		return "", err
	}
	var out bytes.Buffer
	if err := tmpl.Execute(&out, data); err != nil {
		return "", err
	}
	return out.String(), nil
}

func createSymlinks(opts Options, target string, out io.Writer) error {
	for _, sl := range opts.Flavor.Symlinks {
		dir, name := filepath.Split(sl.Path)
		linkDir := filepath.Join(target, filepath.FromSlash(dir))
		if err := link(opts, linkDir, name, sl.Target, filepath.ToSlash(sl.Path), out); err != nil {
			return err
		}
	}
	return nil
}

func link(opts Options, dir, name, dest, display string, out io.Writer) error {
	path := filepath.Join(dir, name)
	if display == "" || strings.HasPrefix(display, "..") {
		display = name
	}
	info, exists, err := lstat(path)
	if err != nil {
		return fmt.Errorf("checking %s: %w", display, err)
	}
	if exists && !opts.Force {
		if opts.counts != nil {
			opts.counts.skipped++
		}
		return nil
	}
	if opts.DryRun {
		if opts.counts != nil {
			opts.counts.linked++
		}
		printOperation(out, opts.style, "link", "%s -> %s (dry-run)", display, dest)
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
	if opts.counts != nil {
		opts.counts.linked++
	}
	printOperation(out, opts.style, "link", "%s -> %s", display, dest)
	return nil
}

func initGit(ctx context.Context, target string, dryRun bool, out io.Writer) error {
	if insideGitWorkTree(ctx, target) {
		return nil
	}
	if dryRun {
		_, _ = fmt.Fprintln(out, "  init   git repository (dry-run)")
		return nil
	}
	cmd := exec.CommandContext(ctx, "git", "init", "-q")
	cmd.Dir = target
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("initializing git repository: %w", err)
	}
	_, _ = fmt.Fprintln(out, "  init   git repository")
	return nil
}

func printSummary(out io.Writer, dryRun bool, counts *operationCounts, style outputStyle) {
	if counts == nil {
		return
	}
	if dryRun {
		_, _ = fmt.Fprintf(out, "\nDry run: %d would be written, %d skipped, %d would be linked.\n", counts.written, counts.skipped, counts.linked)
		return
	}
	_, _ = fmt.Fprintf(out, "\n%s\n", style.done(fmt.Sprintf("Done. %d written, %d skipped, %d linked.", counts.written, counts.skipped, counts.linked)))
}

func printOperation(out io.Writer, style outputStyle, op, format string, args ...any) {
	_, _ = fmt.Fprintf(out, "  %s%s%s\n", style.verb(op), strings.Repeat(" ", 7-len(op)), fmt.Sprintf(format, args...))
}

func colorEnabled(out io.Writer) bool {
	return colorEnabledWith(out, os.Getenv, isTerminal)
}

func colorEnabledWith(out io.Writer, getenv func(string) string, isTerm func(*os.File) bool) bool {
	if getenv("NO_COLOR") != "" || getenv("TERM") == "dumb" {
		return false
	}
	file, ok := out.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	if err != nil || info.Mode()&os.ModeCharDevice == 0 {
		return false
	}
	return isTerm(file)
}

func (s outputStyle) header(text string) string {
	if !s.enabled {
		return text
	}
	return ansiBold + text + ansiReset
}

func (s outputStyle) done(text string) string {
	if !s.enabled {
		return text
	}
	return ansiBoldGreen + text + ansiReset
}

func (s outputStyle) verb(op string) string {
	if !s.enabled {
		return op
	}
	switch op {
	case "write":
		return ansiGreen + op + ansiReset
	case "skip":
		return ansiYellow + op + ansiReset
	case "link":
		return ansiCyan + op + ansiReset
	default:
		return op
	}
}

func printNextSteps(out io.Writer, flavor flavors.Flavor, target string) {
	if flavor.NextSteps != nil {
		message := flavor.NextSteps(target)
		message = strings.TrimPrefix(message, "\nDone.\n\n")
		message = strings.TrimPrefix(message, "Done.\n\n")
		_, _ = fmt.Fprint(out, message)
		return
	}
	_, _ = fmt.Fprintf(out, `
Next steps:
  1. Read %s/README.agent.md for dependency install instructions
  2. Edit .agent/AGENTS.md to describe THIS project's specifics
  3. AGENTS.md and CLAUDE.md are symlinks to .agent/AGENTS.md; edit that one file
  4. Edit .agent/CODEBASE.md once you have code to map
  5. Run:  devcontainer up --workspace-folder %s
  6. Run:  devcontainer exec --workspace-folder %s bash
  7. Inside the container: just check
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
