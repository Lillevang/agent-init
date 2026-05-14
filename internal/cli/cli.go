package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/mikeschinkel/agent-init/internal/flavors"
	"github.com/mikeschinkel/agent-init/internal/scaffold"
	"github.com/mikeschinkel/agent-init/internal/trackers"
)

type Version struct {
	Commit    string
	BuildDate string
}

type App struct {
	out      io.Writer
	errOut   io.Writer
	version  Version
	registry flavors.Registry
	trackers trackers.Registry
}

func New(out, errOut io.Writer, version Version) App {
	return App{
		out:      out,
		errOut:   errOut,
		version:  version,
		registry: flavors.DefaultRegistry(),
		trackers: trackers.DefaultRegistry(),
	}
}

func (a App) Run(ctx context.Context, args []string) error {
	if len(args) == 0 || strings.HasPrefix(args[0], "-") {
		return a.runInit(ctx, args)
	}
	switch args[0] {
	case "init":
		return a.runInit(ctx, args[1:])
	case "list-flavors":
		return a.runListFlavors(args[1:])
	case "list-trackers":
		return a.runListTrackers(args[1:])
	case "add-tracker":
		return a.runAddTracker(ctx, args[1:])
	case "version":
		return a.runVersion(args[1:])
	case "help", "-h", "--help":
		a.printHelp()
		return nil
	default:
		if _, err := a.registry.Get(args[0]); err == nil || looksLikeTarget(args[0]) {
			return a.runInit(ctx, args)
		}
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func looksLikeTarget(arg string) bool {
	return filepath.IsAbs(arg) ||
		strings.HasPrefix(arg, ".") ||
		strings.ContainsAny(arg, `/\`)
}

func (a App) unknownFlavorError(name string) error {
	flavors := a.registry.List()
	known := make([]string, 0, len(flavors))
	for _, f := range flavors {
		known = append(known, f.Name)
	}
	return fmt.Errorf("unknown flavor %q (known: %s)", name, strings.Join(known, ", "))
}

func (a App) runInit(ctx context.Context, args []string) error {
	flags := flag.NewFlagSet("init", flag.ContinueOnError)
	flags.SetOutput(a.errOut)
	force := flags.Bool("force", false, "overwrite existing files")
	noGit := flags.Bool("no-git", false, "skip git init when target is not already a repo")
	dryRun := flags.Bool("dry-run", false, "print what would happen without writing files")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	flavorName := "fullstack"
	target := "."
	switch flags.NArg() {
	case 0:
	case 1:
		arg := flags.Arg(0)
		if _, err := a.registry.Get(arg); err == nil {
			flavorName = arg
		} else if looksLikeTarget(arg) {
			target = arg
		} else {
			return a.unknownFlavorError(arg)
		}
	case 2:
		flavorName = flags.Arg(0)
		target = flags.Arg(1)
	default:
		return fmt.Errorf("usage: agent-init init [flavor] [target-dir]")
	}
	flavor, err := a.registry.Get(flavorName)
	if err != nil {
		return a.unknownFlavorError(flavorName)
	}
	return scaffold.Run(ctx, scaffold.Options{
		Flavor:  flavor,
		Target:  target,
		Force:   *force,
		InitGit: !*noGit,
		DryRun:  *dryRun,
		Out:     a.out,
	})
}

func (a App) runListFlavors(args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("usage: agent-init list-flavors")
	}
	for _, flavor := range a.registry.List() {
		fmt.Fprintf(a.out, "%s\t%s\n", flavor.Name, flavor.Description)
	}
	return nil
}

func (a App) runListTrackers(args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("usage: agent-init list-trackers")
	}
	for _, t := range a.trackers.List() {
		fmt.Fprintf(a.out, "%s\t%s\n", t.Name, t.Description)
	}
	return nil
}

func (a App) runAddTracker(ctx context.Context, args []string) error {
	_ = ctx
	flags := flag.NewFlagSet("add-tracker", flag.ContinueOnError)
	flags.SetOutput(a.errOut)
	force := flags.Bool("force", false, "overwrite existing tracker files")
	dryRun := flags.Bool("dry-run", false, "print what would happen without writing files")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() != 2 {
		return fmt.Errorf("usage: agent-init add-tracker <tracker> <target-dir>")
	}
	trackerName := flags.Arg(0)
	target := flags.Arg(1)

	tracker, err := a.trackers.Get(trackerName)
	if err != nil {
		return err
	}
	// Target must be an existing project-management scaffold. We use the
	// presence of .mcp.json as the marker — `init project-management`
	// always ships it.
	if _, err := os.Stat(filepath.Join(target, ".mcp.json")); err != nil {
		return fmt.Errorf("target %q does not look like a project-management scaffold (no .mcp.json found). Run `agent-init init project-management %s` first", target, target)
	}

	fmt.Fprintf(a.out, "-> Adding %s tracker integration to: %s\n", tracker.DisplayName, target)
	if err := scaffold.Overlay(scaffold.Options{
		Target: target,
		Force:  *force,
		DryRun: *dryRun,
		Out:    a.out,
	}, tracker.Templates, tracker.TemplateRoot); err != nil {
		return fmt.Errorf("writing tracker templates: %w", err)
	}

	if *dryRun {
		fmt.Fprintf(a.out, "  merge  .mcp.json: would add %q under mcpServers (dry-run)\n", tracker.MCPServerKey)
		return nil
	}
	changed, err := trackers.MergeMCPServer(target, tracker.MCPServerKey, tracker.MCPServer)
	if err != nil {
		return fmt.Errorf("merging .mcp.json: %w", err)
	}
	if changed {
		fmt.Fprintf(a.out, "  merge  .mcp.json: added %q under mcpServers\n", tracker.MCPServerKey)
	} else {
		fmt.Fprintf(a.out, "  skip   .mcp.json: %q already present under mcpServers\n", tracker.MCPServerKey)
	}
	// The integrations subfolder is named by the tracker package's template
	// layout (e.g. integrations/github/, integrations/jira/), not the
	// MCPServerKey. Resolve it from the tracker's templates so the printed
	// path matches what was actually written.
	folder := trackerIntegrationFolder(tracker)
	if folder != "" {
		fmt.Fprintf(a.out, "\nDone. Review %s/integrations/%s/README.md for setup notes.\n", target, folder)
	} else {
		fmt.Fprintf(a.out, "\nDone. Review the new files under %s/integrations/.\n", target)
	}
	return nil
}

// trackerIntegrationFolder returns the single subdirectory under
// integrations/ that the tracker ships templates for. Returns "" if the
// tracker doesn't ship an integrations/ subfolder (unusual but possible).
func trackerIntegrationFolder(t trackers.Tracker) string {
	entries, err := fs.ReadDir(t.Templates, "templates/integrations")
	if err != nil {
		return ""
	}
	for _, e := range entries {
		if e.IsDir() {
			return e.Name()
		}
	}
	return ""
}

func (a App) runVersion(args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("usage: agent-init version")
	}
	fmt.Fprintf(a.out, "agent-init commit=%s buildDate=%s\n", a.version.Commit, a.version.BuildDate)
	return nil
}

func (a App) printHelp() {
	fmt.Fprintln(a.out, `agent-init scaffolds repositories for sandboxed agentic development.

Usage:
  agent-init init [flavor] [target-dir]
  agent-init add-tracker <tracker> <target-dir>
  agent-init list-flavors
  agent-init list-trackers
  agent-init version

The default flavor is fullstack. The default target directory is the current directory.

Flags for init:
  --force                    overwrite existing files
  --no-git                   skip git init when target is not already a repo
  --dry-run                  print what would happen without writing files

Flags for add-tracker:
  --force                    overwrite existing tracker files
  --dry-run                  print what would happen without writing files`)
}
