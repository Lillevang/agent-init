# CLI

`agent-init` is a small CLI with six subcommands. Source: [internal/cli/cli.go](../internal/cli/cli.go).

```
agent-init init [flavor] [target-dir]
agent-init add-tracker <tracker> <target-dir>
agent-init list-flavors
agent-init list-trackers
agent-init version
agent-init upgrade [--check] [--dry-run] [--force]
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

## `upgrade`

Updates `agent-init` in place to the latest GitHub release. There is **no
automatic background check**: a release is only contacted when you run this
command, so normal invocations make no network calls.

```bash
agent-init upgrade           # install the latest release, replacing this binary
agent-init upgrade --check   # only report whether a newer version exists
agent-init upgrade --dry-run # download and verify, but do not replace the binary
```

### Flags

| Flag | Effect |
|------|--------|
| `--check` | Report whether a newer release exists and exit, without downloading or installing anything. |
| `--dry-run` | Download the latest archive and verify its checksum, but stop before replacing the binary. |
| `--force` | Install the latest release even when the current version is already newest. Also required to upgrade a dev build, which has no release version to compare against. |

### Behavior

- Makes a network call to GitHub's releases API and downloads the OS/arch-specific asset plus `checksums.txt`. Honors `GITHUB_TOKEN` / `GH_TOKEN` to lift the anonymous rate limit.
- Verifies the archive's SHA-256 against the published checksum and swaps the binary in place. **Fails closed on checksum mismatch** — the existing binary is left untouched.
- Requires write permission on the binary's install directory. A root-owned install path (e.g. `/usr/local/bin`) cannot be upgraded without elevated access; `upgrade` reports the error rather than escalating itself.
- A dev build (`version=dev`) cannot be compared to a release; `upgrade` refuses unless `--force` is passed.
- Source: [internal/selfupdate/selfupdate.go](../internal/selfupdate/selfupdate.go) (verify + replace), [internal/selfupdate/github.go](../internal/selfupdate/github.go) (releases client), [cli.go:runUpgrade](../internal/cli/cli.go). The release asset names this matches are cut by [.github/workflows/release.yml](../.github/workflows/release.yml).

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
