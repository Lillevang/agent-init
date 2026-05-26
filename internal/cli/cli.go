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
	"sort"
	"strings"

	"github.com/Lillevang/agent-init/internal/flavors"
	"github.com/Lillevang/agent-init/internal/gitconfig"
	"github.com/Lillevang/agent-init/internal/gitignore"
	"github.com/Lillevang/agent-init/internal/scaffold"
	"github.com/Lillevang/agent-init/internal/trackers"
)

// visibility selects how the scaffold's agentic envelope is tracked by git.
// The flag has four values, all implemented: "shared" (default, committed),
// "local" (committed .gitignore), "hidden" (.git/info/exclude, no committed
// trace), and "global-default" (machine-wide git excludes, every repo).
type visibility string

const (
	visibilityShared        visibility = "shared"
	visibilityLocal         visibility = "local"
	visibilityHidden        visibility = "hidden"
	visibilityGlobalDefault visibility = "global-default"
)

// parseVisibility validates the --visibility value. All four modes are
// implemented; an unknown value errors with the list of known modes.
func parseVisibility(v string) (visibility, error) {
	switch visibility(v) {
	case visibilityShared, visibilityLocal, visibilityHidden, visibilityGlobalDefault:
		return visibility(v), nil
	default:
		return "", fmt.Errorf("unknown --visibility %q (known: shared, local, hidden, global-default)", v)
	}
}

// resolveVisibility folds the --private boolean alias into the --visibility
// value. --private means --visibility=hidden. Passing it alongside an
// explicitly-set --visibility that isn't "hidden" is a conflict and errors, so
// the two flags can't silently disagree. visibilitySet distinguishes an
// explicit --visibility=shared (a real conflict with --private) from the
// unset default (no conflict — --private simply selects hidden). The default
// (no --private) just parses --visibility.
func resolveVisibility(v string, visibilitySet, private bool) (visibility, error) {
	if !private {
		return parseVisibility(v)
	}
	if visibilitySet && visibility(v) != visibilityHidden {
		return "", fmt.Errorf("--private conflicts with --visibility=%s; --private is an alias for --visibility=hidden", v)
	}
	return visibilityHidden, nil
}

// docsURL points users at the full documentation. Surfaced in top-level help.
const docsURL = "https://github.com/Lillevang/agent-init/tree/main/docs"

type Version struct {
	Version   string
	Commit    string
	BuildDate string
}

// commandHelp is the single source of truth for a subcommand's help text.
// Both the rendered --help output and docs/cli.md are kept in sync against
// these values (see TestHelpFlagsMatchDocs), so help is plain data, not prose
// scattered across handlers.
type commandHelp struct {
	name     string
	summary  string
	usage    string
	flags    []flagHelp
	examples []string
}

type flagHelp struct {
	name string
	desc string
}

// commands lists every subcommand in display order. Order matters: it drives
// the top-level help listing.
var commands = []commandHelp{
	{
		name:    "init",
		summary: "scaffold a project from a flavor",
		usage:   "agent-init init [flavor] [target-dir]",
		flags: []flagHelp{
			{"--force", "overwrite existing files instead of skipping them"},
			{"--no-git", "skip git init when the target is not already a repo"},
			{"--dry-run", "print planned writes without changing files"},
			{"--agents-only", "ship only the agentic envelope (skip fresh-project files); rejected on claude-cowork and project-management"},
			{"--visibility", "shared (default, committed), local (ignore in committed .gitignore), hidden (ignore in .git/info/exclude, no committed trace), or global-default (ignore in machine-wide git excludes — affects EVERY repo); code flavors only"},
			{"--private", "alias for --visibility=hidden (hide the scaffold in this repo with no committed trace)"},
		},
		examples: []string{
			"agent-init init                                    # scaffold fullstack into .",
			"agent-init init go-cli ./my-tool                   # scaffold go-cli into ./my-tool",
			"agent-init init --agents-only go-cli               # add agents to an existing project",
			"agent-init init --visibility=local go-cli          # ignore the scaffold in .gitignore",
			"agent-init init --private go-cli                   # hide the scaffold via .git/info/exclude",
			"agent-init init --visibility=global-default go-cli # ignore in machine-wide git excludes (all repos)",
		},
	},
	{
		name:    "add-tracker",
		summary: "overlay a tracker integration onto a project-management scaffold",
		usage:   "agent-init add-tracker <tracker> <target-dir>",
		flags: []flagHelp{
			{"--force", "overwrite existing tracker files"},
			{"--dry-run", "print what would happen without writing files or modifying .mcp.json"},
		},
		examples: []string{
			"agent-init add-tracker gh   ~/work/pm   # valid trackers: gh, jira, ado",
			"agent-init add-tracker jira ~/work/pm",
		},
	},
	{
		name:    "list-flavors",
		summary: "print available flavors with descriptions",
		usage:   "agent-init list-flavors",
	},
	{
		name:    "list-trackers",
		summary: "print available trackers with descriptions",
		usage:   "agent-init list-trackers",
	},
	{
		name:    "version",
		summary: "print version info (version + commit + build date)",
		usage:   "agent-init version",
	},
}

func lookupCommand(name string) (commandHelp, bool) {
	for _, c := range commands {
		if c.name == name {
			return c, true
		}
	}
	return commandHelp{}, false
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
	// Top-level help flags must be caught before the bare-flag fast path
	// (which otherwise treats anything starting with "-" as init flags).
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		a.printHelp()
		return nil
	}
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
	case "help":
		// `help <subcommand>` prints that subcommand's help; bare help is
		// the top-level overview. (`-h` / `--help` are caught earlier.)
		if len(args) > 1 {
			if cmd, ok := lookupCommand(args[1]); ok {
				a.printCommandHelp(cmd)
				return nil
			}
		}
		a.printHelp()
		return nil
	default:
		if _, err := a.registry.Get(args[0]); err == nil || looksLikeTarget(args[0]) {
			return a.runInit(ctx, args)
		}
		return fmt.Errorf("unknown command %q\nRun 'agent-init --help' for usage", args[0])
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
	return fmt.Errorf("unknown flavor %q (known: %s)\nRun 'agent-init init --help' for usage", name, strings.Join(known, ", "))
}

// newFlagSet builds a flag.FlagSet whose Usage prints the structured help for
// the named command to stderr, so a parse error (e.g. an unknown flag) is
// followed by that command's usage. An explicitly requested --help is handled
// separately (see wantsHelp) and goes to stdout.
func (a App) newFlagSet(name string) *flag.FlagSet {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.SetOutput(a.errOut)
	if cmd, ok := lookupCommand(name); ok {
		flags.Usage = func() { a.fprintCommandHelp(a.errOut, cmd) }
	}
	return flags
}

// wantsHelp reports whether args contains an explicit help flag. An explicit
// --help is a successful request and its output belongs on stdout, unlike the
// usage printed alongside a parse error (which the flag.FlagSet sends to
// stderr).
func wantsHelp(args []string) bool {
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			return true
		}
	}
	return false
}

func (a App) runInit(ctx context.Context, args []string) error {
	if wantsHelp(args) {
		cmd, _ := lookupCommand("init")
		a.printCommandHelp(cmd)
		return nil
	}
	flags := a.newFlagSet("init")
	force := flags.Bool("force", false, "overwrite existing files")
	noGit := flags.Bool("no-git", false, "skip git init when target is not already a repo")
	dryRun := flags.Bool("dry-run", false, "print what would happen without writing files")
	agentsOnly := flags.Bool("agents-only", false, "ship only the agentic envelope (skip fresh-project files); for adding agents to an existing project")
	visibilityFlag := flags.String("visibility", string(visibilityShared), "scaffold visibility: shared (committed), local (ignore in committed .gitignore), hidden (ignore in .git/info/exclude), or global-default (ignore in machine-wide git excludes; affects every repo)")
	private := flags.Bool("private", false, "alias for --visibility=hidden")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	visibilitySet := false
	flags.Visit(func(f *flag.Flag) {
		if f.Name == "visibility" {
			visibilitySet = true
		}
	})
	vis, err := resolveVisibility(*visibilityFlag, visibilitySet, *private)
	if err != nil {
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
		return fmt.Errorf("usage: agent-init init [flavor] [target-dir]\nRun 'agent-init init --help' for usage")
	}
	flavor, err := a.registry.Get(flavorName)
	if err != nil {
		return a.unknownFlavorError(flavorName)
	}
	if *agentsOnly && !flavor.SupportsAgentsOnly {
		return fmt.Errorf("flavor %q does not support --agents-only", flavor.Name)
	}
	if vis != visibilityShared && !flavor.SupportsVisibility {
		return fmt.Errorf("flavor %q does not support --visibility (code flavors only)", flavor.Name)
	}
	if err := scaffold.Run(ctx, scaffold.Options{
		Flavor:     flavor,
		Target:     target,
		Force:      *force,
		InitGit:    !*noGit,
		DryRun:     *dryRun,
		AgentsOnly: *agentsOnly,
		Out:        a.out,
	}); err != nil {
		return err
	}
	return a.applyVisibility(vis, target, *dryRun)
}

// applyVisibility writes the scaffold's ignore envelope according to the chosen
// visibility mode. "shared" (the default) is a no-op: the scaffold is committed
// normally. "local" and "hidden" append a fenced, idempotent block to a
// repo-local ignore file (the committed .gitignore, or the never-committed
// .git/info/exclude respectively) and share the file-writer path.
// "global-default" writes the same block to the user's machine-wide git
// excludes file, affecting every repo, and is routed separately because it goes
// through internal/gitconfig and prints the machine-wide warning and force-add
// hint. Every mutating mode announces the absolute path it edited.
func (a App) applyVisibility(vis visibility, target string, dryRun bool) error {
	switch vis {
	case visibilityLocal, visibilityHidden:
		return a.applyFileVisibility(vis, target, dryRun)
	case visibilityGlobalDefault:
		return a.applyGlobalVisibility(dryRun)
	default:
		return nil
	}
}

// applyFileVisibility appends the ignore block to a repo-local ignore file
// selected by the visibility mode (committed .gitignore for "local",
// .git/info/exclude for "hidden"). --dry-run previews the path and block,
// writing nothing.
func (a App) applyFileVisibility(vis visibility, target string, dryRun bool) error {
	pathFn, ensureFn, ok := visibilityWriters(vis)
	if !ok {
		return nil
	}
	if dryRun {
		path, err := pathFn(target)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintf(a.out, "  ignore %s (dry-run):\n%s", path, indentBlock(gitignore.Block()))
		return nil
	}
	path, err := ensureFn(target)
	if err != nil {
		return fmt.Errorf("applying --visibility=%s: %w", vis, err)
	}
	_, _ = fmt.Fprintf(a.out, "  ignore %s (agent-init scaffold block)\n", path)
	return nil
}

// visibilityWriters maps a repo-local visibility mode to the gitignore
// functions that compute its target path (dry-run) and write its block. The
// third return is false for modes that don't touch a repo-local ignore file
// (shared, and global-default, which goes through internal/gitconfig instead).
func visibilityWriters(vis visibility) (path func(string) (string, error), ensure func(string) (string, error), ok bool) {
	switch vis {
	case visibilityLocal:
		return gitignore.LocalPath, gitignore.EnsureLocal, true
	case visibilityHidden:
		return gitignore.HiddenPath, gitignore.EnsureHidden, true
	default:
		return nil, nil, false
	}
}

// applyGlobalVisibility writes the ignore block to the user's machine-wide git
// excludes file (core.excludesfile, or ~/.config/git/ignore if unset). This is
// action-at-a-distance: it ignores the agentic envelope in EVERY repository on
// the machine, so the warning and the edited path are printed loudly. To commit
// the scaffold openly in a specific repo despite this default, force-add it
// (see the printed hint). --dry-run resolves and prints the target path and the
// block but writes nothing and touches no git config.
func (a App) applyGlobalVisibility(dryRun bool) error {
	runner := gitconfig.NewExecRunner()
	env := gitconfig.OSEnv{}
	a.warnGlobalVisibility()
	if dryRun {
		path, err := gitconfig.GlobalPath(runner, env)
		if err != nil {
			return fmt.Errorf("resolving global excludes path: %w", err)
		}
		_, _ = fmt.Fprintf(a.out, "  ignore %s (machine-wide, dry-run):\n%s", path, indentBlock(gitignore.Block()))
		return nil
	}
	path, err := gitconfig.EnsureGlobal(runner, env, gitignore.Upsert)
	if err != nil {
		return fmt.Errorf("applying --visibility=global-default: %w", err)
	}
	_, _ = fmt.Fprintf(a.out, "  ignore %s (machine-wide git excludes — affects EVERY repo)\n", path)
	a.printForceAddHint()
	return nil
}

// warnGlobalVisibility prints the unmissable machine-wide warning before the
// global excludes file is touched (or previewed). A global write affects every
// repository on the machine, so it must never happen silently.
func (a App) warnGlobalVisibility() {
	_, _ = fmt.Fprintln(a.errOut, "WARNING: --visibility=global-default edits your MACHINE-WIDE git excludes.")
	_, _ = fmt.Fprintln(a.errOut, "         The agent-init scaffold will be ignored in EVERY git repository on this machine.")
}

// printForceAddHint tells the user how to commit the scaffold openly in a repo
// that should override the global default. Git never re-ignores a tracked file,
// so force-add is the documented escape hatch (gitignore negation cannot
// re-include a file under an excluded directory).
func (a App) printForceAddHint() {
	_, _ = fmt.Fprintln(a.out, "  To commit the scaffold openly in a specific repo, force-add it there:")
	_, _ = fmt.Fprintln(a.out, "    git add -f .agent AGENTS.md CLAUDE.md .devcontainer Justfile .pre-commit-config.yaml")
}

// indentBlock prefixes each line of the ignore block for the dry-run preview so
// it reads as nested detail under the "ignore" line.
func indentBlock(block string) string {
	lines := strings.Split(strings.TrimRight(block, "\n"), "\n")
	for i, l := range lines {
		lines[i] = "         " + l
	}
	return strings.Join(lines, "\n") + "\n"
}

func (a App) runListFlavors(args []string) error {
	if handled := a.handleNoArgHelp("list-flavors", args); handled {
		return nil
	}
	if len(args) > 0 {
		return fmt.Errorf("usage: agent-init list-flavors\nRun 'agent-init --help' for usage")
	}
	for _, flavor := range a.registry.List() {
		_, _ = fmt.Fprintf(a.out, "%s\t%s\n", flavor.Name, flavor.Description)
	}
	return nil
}

func (a App) runListTrackers(args []string) error {
	if handled := a.handleNoArgHelp("list-trackers", args); handled {
		return nil
	}
	if len(args) > 0 {
		return fmt.Errorf("usage: agent-init list-trackers\nRun 'agent-init --help' for usage")
	}
	for _, t := range a.trackers.List() {
		_, _ = fmt.Fprintf(a.out, "%s\t%s\n", t.Name, t.Description)
	}
	return nil
}

// handleNoArgHelp prints the command's help when the first argument is a help
// flag. It exists because the flagless subcommands (list-flavors,
// list-trackers, version) don't construct a flag.FlagSet and so wouldn't
// otherwise recognize --help / -h.
func (a App) handleNoArgHelp(name string, args []string) bool {
	if len(args) == 0 {
		return false
	}
	switch args[0] {
	case "-h", "--help", "help":
		if cmd, ok := lookupCommand(name); ok {
			a.printCommandHelp(cmd)
		}
		return true
	default:
		return false
	}
}

func (a App) runAddTracker(ctx context.Context, args []string) error {
	_ = ctx
	if wantsHelp(args) {
		cmd, _ := lookupCommand("add-tracker")
		a.printCommandHelp(cmd)
		return nil
	}
	flags := a.newFlagSet("add-tracker")
	force := flags.Bool("force", false, "overwrite existing tracker files")
	dryRun := flags.Bool("dry-run", false, "print what would happen without writing files")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() != 2 {
		return fmt.Errorf("usage: agent-init add-tracker <tracker> <target-dir>\nRun 'agent-init add-tracker --help' for usage")
	}
	trackerName := flags.Arg(0)
	target := flags.Arg(1)

	tracker, err := a.trackers.Get(trackerName)
	if err != nil {
		return fmt.Errorf("%w\nRun 'agent-init add-tracker --help' for usage", err)
	}
	// Target must be an existing project-management scaffold. We use the
	// presence of .mcp.json as the marker — `init project-management`
	// always ships it.
	if _, err := os.Stat(filepath.Join(target, ".mcp.json")); err != nil {
		return fmt.Errorf("target %q does not look like a project-management scaffold (no .mcp.json found). Run `agent-init init project-management %s` first", target, target)
	}

	_, _ = fmt.Fprintf(a.out, "-> Adding %s tracker integration to: %s\n", tracker.DisplayName, target)
	if err := scaffold.Overlay(scaffold.Options{
		Target: target,
		Force:  *force,
		DryRun: *dryRun,
		Out:    a.out,
	}, tracker.Templates, tracker.TemplateRoot); err != nil {
		return fmt.Errorf("writing tracker templates: %w", err)
	}

	if *dryRun {
		_, _ = fmt.Fprintf(a.out, "  merge  .mcp.json: would add %q under mcpServers (dry-run)\n", tracker.MCPServerKey)
		return nil
	}
	changed, err := trackers.MergeMCPServer(target, tracker.MCPServerKey, tracker.MCPServer)
	if err != nil {
		return fmt.Errorf("merging .mcp.json: %w", err)
	}
	if changed {
		_, _ = fmt.Fprintf(a.out, "  merge  .mcp.json: added %q under mcpServers\n", tracker.MCPServerKey)
	} else {
		_, _ = fmt.Fprintf(a.out, "  skip   .mcp.json: %q already present under mcpServers\n", tracker.MCPServerKey)
	}
	// The integrations subfolder is named by the tracker package's template
	// layout (e.g. integrations/github/, integrations/jira/), not the
	// MCPServerKey. Resolve it from the tracker's templates so the printed
	// path matches what was actually written.
	folder := trackerIntegrationFolder(tracker)
	if folder != "" {
		_, _ = fmt.Fprintf(a.out, "\nDone. Review %s/integrations/%s/README.md for setup notes.\n", target, folder)
	} else {
		_, _ = fmt.Fprintf(a.out, "\nDone. Review the new files under %s/integrations/.\n", target)
	}
	a.printTrackerCredentialHelp(tracker, target, folder)
	return nil
}

// printTrackerCredentialHelp tells the user where to put tracker credentials.
// The .mcp.json entry references secrets via ${env:VAR}, so the token must
// live in the environment (or a gitignored .env), never in the tracked
// .mcp.json. Changing .mcp.json needs an MCP/session restart to reconnect.
func (a App) printTrackerCredentialHelp(t trackers.Tracker, target, folder string) {
	envVars := trackerEnvVars(t)
	if len(envVars) == 0 {
		return
	}
	_, _ = fmt.Fprintf(a.out, "\nCredentials: set %s in your environment (do NOT paste secrets into .mcp.json — it is tracked).\n", strings.Join(envVars, ", "))
	if folder != "" {
		_, _ = fmt.Fprintf(a.out, "  See %s/integrations/%s/.env.example for the full list and a gitignored .env you can source.\n", target, folder)
	}
	if t.Name == "gh" {
		_, _ = fmt.Fprintf(a.out, "  GitHub tip: reuse the devcontainer's gh login with `export GITHUB_TOKEN=$(gh auth token)` — no separate PAT needed.\n")
	}
	_, _ = fmt.Fprintf(a.out, "  Restart your MCP client (or session) after setting credentials so the server reconnects.\n")
}

// trackerEnvVars returns the environment variable names referenced (via
// ${env:VAR}) by the tracker's MCP env block, sorted for stable output. This
// includes non-secret config (e.g. ADO_ORG_URL) as well as secrets; all are
// supplied from the environment so none ends up as a literal in .mcp.json.
func trackerEnvVars(t trackers.Tracker) []string {
	env, ok := t.MCPServer["env"].(map[string]any)
	if !ok {
		return nil
	}
	var names []string
	for _, raw := range env {
		val, ok := raw.(string)
		if !ok {
			continue
		}
		if name, found := strings.CutPrefix(val, "${env:"); found {
			names = append(names, strings.TrimSuffix(name, "}"))
		}
	}
	sort.Strings(names)
	return names
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
	if handled := a.handleNoArgHelp("version", args); handled {
		return nil
	}
	if len(args) > 0 {
		return fmt.Errorf("usage: agent-init version\nRun 'agent-init --help' for usage")
	}
	_, _ = fmt.Fprintf(a.out, "agent-init version=%s commit=%s buildDate=%s\n", a.version.Version, a.version.Commit, a.version.BuildDate)
	return nil
}

// printHelp renders the top-level overview from the commands table so the
// subcommand list never drifts from what the binary actually dispatches.
func (a App) printHelp() {
	var b strings.Builder
	b.WriteString("agent-init scaffolds repositories for sandboxed agentic development.\n\n")
	b.WriteString("Usage:\n  agent-init <command> [arguments]\n")
	b.WriteString("  agent-init [flavor] [target-dir]   # shorthand for 'init'\n\n")
	b.WriteString("Commands:\n")
	for _, c := range commands {
		fmt.Fprintf(&b, "  %-14s %s\n", c.name, c.summary)
	}
	b.WriteString("\nThe default flavor is fullstack. The default target directory is the current directory.\n\n")
	b.WriteString("Run 'agent-init <command> --help' for command-specific flags and examples.\n")
	fmt.Fprintf(&b, "Documentation: %s\n", docsURL)
	_, _ = fmt.Fprint(a.out, b.String())
}

// printCommandHelp writes a subcommand's help to stdout (used by explicit
// `help <cmd>` / `<cmd> --help` on flagless subcommands).
func (a App) printCommandHelp(c commandHelp) {
	a.fprintCommandHelp(a.out, c)
}

// fprintCommandHelp renders one subcommand's structured help. It backs both the
// stdout help paths and the flag.FlagSet Usage hook (which writes to errOut).
func (a App) fprintCommandHelp(w io.Writer, c commandHelp) {
	var b strings.Builder
	fmt.Fprintf(&b, "%s — %s\n\n", c.name, c.summary)
	fmt.Fprintf(&b, "Usage:\n  %s\n", c.usage)
	if len(c.flags) > 0 {
		b.WriteString("\nFlags:\n")
		for _, f := range c.flags {
			fmt.Fprintf(&b, "  %-14s %s\n", f.name, f.desc)
		}
	}
	if len(c.examples) > 0 {
		b.WriteString("\nExamples:\n")
		for _, ex := range c.examples {
			fmt.Fprintf(&b, "  %s\n", ex)
		}
	}
	_, _ = fmt.Fprint(w, b.String())
}
