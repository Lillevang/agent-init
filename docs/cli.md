# CLI

`agent-init` is a small CLI with five subcommands. Source: [internal/cli/cli.go](../internal/cli/cli.go).

```
agent-init init [flavor] [target-dir]
agent-init add-tracker <tracker> <target-dir>
agent-init list-flavors
agent-init list-trackers
agent-init version
```

If no subcommand is given, the binary defaults to `init` with the default flavor. So `agent-init` and `agent-init init` are equivalent.

## `init`

Scaffolds a project. Default flavor: `fullstack`. Default target: current directory.

```bash
agent-init init                           # scaffold fullstack into .
agent-init init go-cli ./my-tool          # scaffold go-cli into ./my-tool
agent-init init project-management ~/work/pm
agent-init ./my-tool                      # path-only form; implies fullstack
```

### Flags

| Flag | Effect |
|------|--------|
| `--force` | Overwrite existing files instead of skipping them. Default: skip with a notice. |
| `--no-git` | Skip `git init` when the target is not already a repo. |
| `--dry-run` | Print planned writes without changing files. |
| `--agents-only` | Skip the flavor's fresh-project files; ship only the agentic envelope (AGENTS.md, scripts, devcontainer, Justfile, pre-commit). For adding agents to an existing project. Supported on every code flavor: `fullstack`, `go-cli`, `go-backend`, `iac`. Rejected on doc-collab flavors (`claude-cowork`, `project-management`) since they don't bootstrap a project layout. |

### Behavior

- The scaffold engine walks the flavor's templates, then the common overlay (if the flavor has one). Existing files are skipped unless `--force` is set.
- After file writes: creates the flavor's declared symlinks (code flavors get the AGENTS.md/CLAUDE.md trio; doc-collab flavors get none), then runs `git init` unless `--no-git`, then prints the flavor's `NextSteps` message.
- With `--agents-only`: paths listed in the flavor's `FreshOnlyPaths` are skipped, and any template named `<file>.agents-only.<ext>` (e.g. `Justfile.agents-only.tmpl`) is written as `<file>.<ext>` in place of the base. See [docs/flavors/go-cli.md](./flavors/go-cli.md) for a worked example.
- Source: [scaffold.go:31](../internal/scaffold/scaffold.go#L31) (`Run`).

## `add-tracker`

Adds a work-tracker integration (Jira, Azure DevOps, or GitHub) to an existing `project-management` scaffold. Only meaningful for that flavor — the subcommand errors if the target lacks an `.mcp.json` file (the scaffold-presence marker).

```bash
agent-init add-tracker gh    ~/work/pm
agent-init add-tracker jira  ~/work/pm
agent-init add-tracker ado   ~/work/pm
```

Multiple trackers can be added to the same workspace. Useful during migrations (e.g., ADO → Jira: add Jira, migrate epic-by-epic via `/sync-tracker`, then remove ADO manually).

### Flags

| Flag | Effect |
|------|--------|
| `--force` | Overwrite the tracker's integration files if they already exist. |
| `--dry-run` | Print what would happen without writing files or modifying `.mcp.json`. |

### Behavior

- Verifies target has a `.mcp.json` (errors with a usage hint if missing).
- Overlays the tracker's templates onto the target (writes `integrations/<tracker>/README.md`).
- Merges an entry into the target's `.mcp.json` under `mcpServers`. **Idempotent**: if the entry already exists, the merge is a no-op with a notice — existing config is *not* overwritten with the new default.
- Source: [cli.go:runAddTracker](../internal/cli/cli.go) and [trackers/mcp.go:MergeMCPServer](../internal/trackers/mcp.go).

### Removing a tracker

There is no `remove-tracker` subcommand yet. Manual cleanup:

1. Delete `integrations/<tracker>/`.
2. Remove the entry from `.mcp.json` under `mcpServers`.
3. Remove the tracker name from `AGENTS.md`'s "Active trackers" line.

## `list-flavors`

Prints registered flavors with one-line descriptions, sorted by name.

```
$ agent-init list-flavors
claude-cowork        Shared document-collaboration folder ...
fullstack            TypeScript/Node frontend and backend ...
go-backend           Go HTTP backend scaffold ...
go-cli               Go command-line tool scaffold ...
project-management   Project-management workspace ...
```

Format: `<name>\t<description>\n`. Stable enough that the `Justfile` smoke-test recipe parses it via `awk '{print $1}'`.

## `list-trackers`

Prints registered trackers with one-line descriptions, sorted by name.

```
$ agent-init list-trackers
ado    Azure DevOps (Epic → Feature → PBI). MCP server: @azure-devops/mcp ...
gh     GitHub Issues (flat or grouped via labels/milestones). MCP server: ...
jira   Jira (Epic → Feature → User Story). MCP server: mcp-atlassian (community).
```

Same format as `list-flavors`.

## `version`

Prints the embedded commit + build date set via `-ldflags` at build time.

```
$ agent-init version
agent-init commit=abc123 buildDate=2026-05-14T10:00:00Z
```

In dev builds (`go run ./cmd/agent-init version`), prints `commit=dev buildDate=unknown` — the ldflags only apply to release builds.

## Help

```
agent-init help          # or -h, --help
agent-init init --help   # init-specific help via flag.ErrHelp
agent-init add-tracker --help
```

The `--help` flag on subcommands returns cleanly (exit 0, no error printed) — `flag.ErrHelp` is recognized and swallowed.

## Error messages

Two specific cases are worth knowing:

- **Unknown flavor** prints the list of known flavors: `unknown flavor "foo" (known: claude-cowork, fullstack, go-backend, go-cli, project-management)`.
- **Unknown tracker** prints the list of known trackers.
- **`add-tracker` on a target without `.mcp.json`** suggests the corresponding `init` command.

Source: [`unknownFlavorError`](../internal/cli/cli.go) and the registry `Get` methods in [trackers/registry.go](../internal/trackers/registry.go) and [flavors/registry.go](../internal/flavors/registry.go).
