# CLI

`agent-init` is a small CLI with six subcommands. Source: [internal/cli/cli.go](../internal/cli/cli.go).

```
agent-init init [flavor] [target-dir]
agent-init add-tracker <tracker> <target-dir>
agent-init status [target]
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
| `--visibility` | How the scaffold is tracked by git. Four modes, all implemented: `shared`, `local`, `hidden`, `global-default`. `shared` (default) commits the scaffold normally. `local` appends a fenced, idempotent block to the committed `.gitignore` so the team sees the scaffold is ignored but does not carry the files. `hidden` writes the same block to the never-committed `.git/info/exclude`, leaving zero committed trace (per-repo, local to your clone). `global-default` writes the same block to your **machine-wide** git excludes file, ignoring the scaffold in **every** repository on the machine. Code flavors only; rejected on doc-collab flavors. |
| `--private` | Alias for `--visibility=hidden`. Passing it alongside a conflicting `--visibility` errors. |

### Behavior

- The scaffold engine walks the flavor's templates, then the common overlay (if the flavor has one). Existing files are skipped unless `--force` is set.
- After file writes: creates the flavor's declared symlinks (code flavors get the AGENTS.md/CLAUDE.md trio; doc-collab flavors get none), then runs `git init` unless `--no-git`, then prints the flavor's `NextSteps` message.
- With `--agents-only`: paths listed in the flavor's `FreshOnlyPaths` are skipped, and any template named `<file>.agents-only.<ext>` (e.g. `Justfile.agents-only.tmpl`) is written as `<file>.<ext>` in place of the base. See [docs/flavors/go-cli.md](./flavors/go-cli.md) for a worked example.
- With `--visibility=local`: after the scaffold is written (symlinks and `git init` included — visibility controls tracking, not creation), a fenced block is appended to the committed `.gitignore`, creating it if absent. The block covers the agentic envelope (`.agent/`, `/AGENTS.md`, `/CLAUDE.md`, `.devcontainer/`, `/Justfile`, `.pre-commit-config.yaml`). It is delimited by `# >>> agent-init (private) >>>` / `# <<< agent-init <<<` markers, so re-running replaces it in place (never duplicates) and it can be removed by hand to undo. `init` prints the absolute path it edited. `--dry-run` previews the path and block, writing nothing. Block management lives in [internal/gitignore](../internal/gitignore/gitignore.go).
- With `--visibility=hidden` (or `--private`): the identical block is written to `.git/info/exclude` instead of `.gitignore`, creating the `.git/info` directory if absent. `.git/info/exclude` is git's per-repo, never-committed ignore file, so a teammate cloning the repo sees no agent-init trace. The mode is otherwise the same as `local`: idempotent in-place replacement, the absolute path is announced, `--dry-run` previews and writes nothing, and the symlink trio is still created (visibility controls tracking, not creation). Because `.git/info/exclude` does not appear in `git diff`, remember to remove the fenced block by hand to undo.
- With `--visibility=global-default`: the **same** fenced block is written to your machine-wide git excludes file instead of a repo file, so the scaffold is ignored in **every** git repository on the machine. **This is action-at-a-distance.** The command prints a loud machine-wide warning on stderr and always announces the absolute path it edited. The target file is `git config --global core.excludesfile` if set (honored even when it points somewhere unusual); otherwise `${XDG_CONFIG_HOME:-~/.config}/git/ignore`, which is created and set as `core.excludesfile` only when no global excludes is configured. No other global-config key is touched. Idempotent (the marked block is replaced in place) and reversible (remove the block by hand). `--dry-run` resolves and prints the target path and the block but writes nothing and touches no git config. To commit the scaffold openly in a specific repo despite the global default, force-add it there — `git add -f .agent AGENTS.md CLAUDE.md .devcontainer Justfile .pre-commit-config.yaml` — since git never re-ignores a tracked file (gitignore negation cannot re-include a file under an excluded directory, so force-add is the documented override). The global excludes-file resolution lives in [internal/gitconfig](../internal/gitconfig/gitconfig.go); the block content is shared from [internal/gitignore](../internal/gitignore/gitignore.go).
- Source: [scaffold.go:31](../internal/scaffold/scaffold.go#L31) (`Run`), [cli.go:applyVisibility](../internal/cli/cli.go).

### Output

- When the output stream is a TTY, `init` colorizes its output verbs (`write`, `skip`, `link`) and prints a final `Done.` summary. Color is disabled when `NO_COLOR` is set, `TERM=dumb`, or the output is not a TTY (e.g. a pipe or file).
- Symlink paths are displayed relative to the scaffolded project root, even when the target directory is specified with a relative path (`./foo`) or via a symlink.
- The `NextSteps` message for code-based flavors explains that `AGENTS.md` and `CLAUDE.md` in the root are symlinks to a canonical file under `.agent/`.

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
- Overlays the tracker's templates onto the target (writes `integrations/<tracker>/README.md` and `integrations/<tracker>/.env.example`).
- Merges an entry into the target's `.mcp.json` under `mcpServers`. **Idempotent**: if the entry already exists, the merge is a no-op with a notice — existing config is *not* overwritten with the new default.
- Source: [cli.go:runAddTracker](../internal/cli/cli.go) and [trackers/mcp.go:MergeMCPServer](../internal/trackers/mcp.go).

### Credentials

The merged `.mcp.json` entry references every credential from the environment via `${env:VAR}` (e.g. `"GITHUB_PERSONAL_ACCESS_TOKEN": "${env:GITHUB_TOKEN}"`). No empty literal is ever written, so there is no field inviting a pasted secret into the tracked file. Set the vars in your shell or a gitignored `.env`; each tracker ships `integrations/<tracker>/.env.example` listing what it needs. For the GitHub tracker, `export GITHUB_TOKEN="$(gh auth token)"` reuses the devcontainer's existing `gh` login. `add-tracker` prints the variable names and this guidance after merging. Changing `.mcp.json` requires restarting the MCP client (or session) to reconnect. See [`trackerEnvVars`](../internal/cli/cli.go) and [`internal/trackers/registry.go`](../internal/trackers/registry.go).

### Removing a tracker

There is no `remove-tracker` subcommand yet. Manual cleanup:

1. Delete `integrations/<tracker>/`.
2. Remove the entry from `.mcp.json` under `mcpServers`.
3. Remove the tracker name from `AGENTS.md`'s "Active trackers" line.

## `status`

Reports how the scaffold's agentic envelope is currently tracked by git. Read-only — `status` writes no files and touches no git configuration.

```bash
agent-init status            # report status of the current directory
agent-init status ./my-tool  # report status of ./my-tool
```

The optional positional argument defaults to `.`. The target is resolved to an absolute path before reporting, so the printed paths are unambiguous.

### Output

Each line is `<label>: <value>`. The fields:

| Field | Meaning |
|-------|---------|
| `mode` | One of `shared` (no agent-init ignore block found), `local` (block in the committed `.gitignore`), `hidden` (block in `.git/info/exclude`), or `shadowed-by-global` (no repo-local block, but a block exists in your machine-wide git excludes file). |
| `target` | Absolute path of the directory `status` was run against. |
| `ignore` | Absolute path of the file carrying the agent-init ignore block. Omitted in `shared` mode. Annotated `(machine-wide)` for `shadowed-by-global`. |
| `undo` | A portable `sed` invocation that deletes the fenced block from the carrier, plus a fallback instruction for users who would rather edit by hand. The markers come from [internal/gitignore](../internal/gitignore/gitignore.go) (`MarkerStart` / `MarkerEnd`) so they cannot drift from what is on disk. |
| `note` | Only printed for `shadowed-by-global`. Explains that already-tracked files stay tracked (a global ignore does not retroactively untrack files in the index) and prints the same force-add line `init --visibility=global-default` shows, for repos where the scaffold is being newly added but should be committed openly. |

### Detection precedence

Git's ignore precedence is `.gitignore` > `.git/info/exclude` > `core.excludesfile`, so the block can in principle exist in more than one file at once. `status` reports the most-local (highest-precedence) carrier it finds. `shadowed-by-global` is reported only when neither `.gitignore` nor `.git/info/exclude` carries the block but the machine-wide excludes file does — that is the state where the scaffold is committed openly here yet ignored everywhere else.

The carrier-path helpers come from [internal/gitignore](../internal/gitignore/gitignore.go) (`LocalPath`, `HiddenPath`) and the global-excludes path from [internal/gitconfig](../internal/gitconfig/gitconfig.go) (`GlobalPath`, which is read-only — it never sets `core.excludesfile`). The presence check uses the same managed-block markers the writers do, exposed via `gitignore.HasBlock`.

### Behavior

- Reads at most three files: the target's `.gitignore`, its `.git/info/exclude`, and the machine-wide excludes file. May invoke `git config --global --get core.excludesfile` to locate the machine-wide file. A missing file is not an error — it is the normal `shared` case.
- Writes nothing. Never sets a git config key. Safe to run in CI or against a repo you do not own.
- Detection is marker-based, not effective-behavior. `status` looks for the managed fenced block; it does not evaluate the surrounding `.gitignore`. A user who adds a manual `!.agent/` rule below the block to opt back in still gets `mode: local`/`hidden`/`shadowed-by-global`.
- No flags. `--help` and `-h` print the usage as on every subcommand.
- Source: [internal/cli/status.go](../internal/cli/status.go) (`runStatus`, `detectStatus`, `writeStatusReport`).

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

Prints the embedded version, commit, and build date set via `-ldflags` at build
time. On release binaries `version` is the pushed semver tag (`github.ref_name`,
e.g. `v1.2.3`); see [`docs/engine/releases.md`](engine/releases.md).

```
$ agent-init version
agent-init version=v1.2.3 commit=abc123 buildDate=2026-05-14T10:00:00Z
```

In dev builds (`go run ./cmd/agent-init version`), prints
`version=dev commit=dev buildDate=unknown` — the ldflags only apply to release
builds.

## Help

The binary documents its own usage. Help text is generated from a single data
table in [cli.go](../internal/cli/cli.go) (`commands`), so it cannot drift from
the dispatched subcommands. `TestHelpFlagsMatchDocs` also fails if a flag shown
in `--help` is missing from this page.

```
agent-init --help        # or -h, or `agent-init help` — top-level overview
agent-init init --help   # per-subcommand: usage form, flags, examples
agent-init help init     # same content as `init --help`, printed to stdout
agent-init add-tracker --help
agent-init list-flavors --help
```

- **Top-level help** lists every subcommand with a one-line summary, the global
  usage form, a pointer to per-command help, and the documentation URL.
- **Per-subcommand help** prints that subcommand's usage form, its flags with
  descriptions, and one or two worked examples.
- `--help` exits 0 and prints to stdout. `-h` and `--help` are accepted on
  every subcommand; the flagless ones (`list-flavors`, `list-trackers`,
  `version`) recognize them too.
- A genuine parse error (e.g. an unknown flag) prints the same command help to
  stderr and exits non-zero.

## Error messages

Invalid input prints a short hint and points the user at `--help`, then exits
non-zero. Specific cases worth knowing:

- **Unknown subcommand** prints `unknown command "foo"` followed by `Run 'agent-init --help' for usage`.
- **Unknown flavor** prints the list of known flavors: `unknown flavor "foo" (known: claude-cowork, fullstack, go-backend, go-cli, project-management)`, then the init `--help` hint.
- **Unknown tracker** prints the list of known trackers, then the add-tracker `--help` hint.
- **`add-tracker` on a target without `.mcp.json`** suggests the corresponding `init` command.

Source: [`unknownFlavorError`](../internal/cli/cli.go) and the registry `Get` methods in [trackers/registry.go](../internal/trackers/registry.go) and [flavors/registry.go](../internal/flavors/registry.go).
