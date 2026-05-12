package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/mikeschinkel/agent-init/internal/flavors"
	"github.com/mikeschinkel/agent-init/internal/scaffold"
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
}

func New(out, errOut io.Writer, version Version) App {
	return App{
		out:      out,
		errOut:   errOut,
		version:  version,
		registry: flavors.DefaultRegistry(),
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
  agent-init list-flavors
  agent-init version

The default flavor is fullstack. The default target directory is the current directory.

Flags for init:
  --force                    overwrite existing files
  --no-git                   skip git init when target is not already a repo
  --dry-run                  print what would happen without writing files`)
}
