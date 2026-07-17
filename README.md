# agent-init

`agent-init` sets a repository up for sandboxed agentic development. It drops in a devcontainer, an `AGENTS.md`/`CLAUDE.md` house-rules pair, a `check.sh` "done-gate", a `Justfile`, pre-commit hooks, and a codebase map — all tuned to a chosen project flavor.

The primary use is **adding this envelope to a project you already have**; it can also bootstrap a fresh repo from scratch. Beyond code, it scaffolds two non-code workspaces: a **project-management** workspace (epics, decisions, stakeholders, plus MCP tracker integrations for Jira, Azure DevOps, or GitHub) and a **Claude Cowork** folder for document-centred analysis and design work.

## Quickstart

```bash
# Add the agentic envelope to an existing Go CLI — none of your source is touched.
agent-init init go-cli --agents-only ~/repos/my-cli

# Or scaffold a brand-new project from a flavor.
agent-init init go-cli ./my-new-tool

# See every flavor, or drill into any command.
agent-init list-flavors
agent-init --help
```

## What it's for

Adding agents to a project you already have is the main path, and the scaffold is careful about it. `--agents-only` drops in just the envelope — devcontainer, `AGENTS.md`/`CLAUDE.md` house rules, helper scripts, `Justfile`, pre-commit hooks — **without touching** your `go.mod`, `package.json`, `main.go`, or any existing source. Omit the flag and you also get the flavor's full fresh-project layout, useful when you're starting from zero.

Two things the envelope is built to enforce:

- **Sandbox the agent.** The devcontainer is the work surface. Agents run inside it with a bounded toolchain (Go, Node, lint, security scanners — whatever the chosen flavor needs) and write only into the mounted workspace. Credentials, host SSH keys, and cloud configs are explicit opt-in mounts — commented out in `devcontainer.json` by default so the agent doesn't pick them up unless you choose to share them.
- **Force agents to verify.** Every scaffolded project ships a `check.sh` "done-gate" that runs codemap regeneration, formatting, linting, type-checking, tests, and (per flavor) cross-builds, vulnerability scans, or security scans. The agent's contract — encoded in the generated `AGENTS.md` — is: don't declare a task complete until `./.agent/scripts/check.sh` passes. This is the simplest mechanism we've found that stops agents from cutting work short with "I think it's done."

## Flavors

| Flavor | What it scaffolds |
|--------|-------------------|
| [`fullstack`](./docs/flavors/fullstack.md) | TypeScript/Node frontend + backend with Playwright recording and OpenAPI client generation. |
| [`go-cli`](./docs/flavors/go-cli.md) | Go command-line tool with `cmd/{{.ProjectName}}/` (path-templated), `internal/`, cross-build via Justfile, and `golangci-lint`. |
| [`go-backend`](./docs/flavors/go-backend.md) | Go HTTP backend with `cmd/server`, `internal/api` router, a `/healthz` handler, and `run-dev` / `cross-build` recipes. |
| [`claude-cowork`](./docs/flavors/claude-cowork.md) | OneDrive-backed document collaboration folder for Claude Cowork. No devcontainer / Justfile / symlinks; root-level `AGENTS.md` + `decisions.md` + `corrections.md` + `reference/`, `templates/`, `archive/`. |
| [`project-management`](./docs/flavors/project-management.md) | Project-management workspace (epics, meetings, decisions, stakeholders, time plans). Ships five skills (`/intake-meeting`, `/break-down-epic`, `/log-decision`, `/track-stakeholder`, `/sync-tracker`) and supports MCP tracker integrations via `agent-init add-tracker {jira\|ado\|gh}`. |
| [`iac`](./docs/flavors/iac.md) | Combined Terraform + Ansible scaffold. Ships `terraform/` (root module, `modules/`) and `ansible/` (`inventory/`, `playbooks/`, `roles/`) trees, a devcontainer with `terraform` + `tflint` + `tfsec` + `trivy` + `ansible-core` + `ansible-lint` + `yamllint`, and a Justfile whose recipes auto-detect which toolchain is present. Cloud-credential and `~/.ssh` mounts are commented out by default with a warning. |

## Documentation

This README is the front door. The full reference lives under [`docs/`](./docs/):

- [`docs/cli.md`](./docs/cli.md) — full CLI reference: every subcommand, flag, and exit behavior.
- [`docs/engine/flavor-hooks.md`](./docs/engine/flavor-hooks.md) — the per-flavor engine hooks (`Symlinks`, `NextSteps`, `CommonTemplates`) that let one engine serve code and non-code flavors.
- [`docs/engine/releases.md`](./docs/engine/releases.md) — the tag-driven release and build-provenance flow.
- [`docs/README.md`](./docs/README.md) — the docs index and per-feature documentation convention.

Each flavor has its own page under [`docs/flavors/`](./docs/flavors/), linked from the table above.

## Build

```bash
go build -o agent-init ./cmd/agent-init
```

For a local install:

```bash
go install ./cmd/agent-init
```

Release builds set version metadata via `-ldflags`. On tag-driven release builds
`main.version` is the pushed semver tag, so `agent-init version` reports the
release version. Local builds can set it too:

```bash
go build \
  -ldflags "-X main.version=$(git describe --tags --always) -X main.commit=$(git rev-parse --short HEAD) -X main.buildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  -o agent-init ./cmd/agent-init
```

## Usage

```bash
agent-init init [flavor] [target-dir]
agent-init add-tracker <tracker> <target-dir>
agent-init status [target]
agent-init list-flavors
agent-init list-trackers
agent-init version
agent-init upgrade [--check] [--dry-run] [--force]
```

Defaults: flavor `fullstack`, target `.`.

`agent-init upgrade` replaces the binary in place with the latest GitHub release, verifying its checksum first. There is no automatic update check — run `agent-init upgrade --check` to see if a newer version exists. See [`docs/cli.md`](./docs/cli.md#upgrade).

Run `agent-init --help` for the subcommand list, and `agent-init <command> --help` for a command's flags and examples. See [`docs/cli.md`](./docs/cli.md) for the full reference.

Examples. Each block states what the command writes and what it leaves alone.

```bash
# Primary use case: add the agentic envelope to an existing project.
agent-init init go-cli --agents-only ~/repos/my-existing-cli
# Writes: .devcontainer/, .agent/ (AGENTS.md, CODEBASE.md, CORRECTIONS.md, scripts/),
#   AGENTS.md and CLAUDE.md symlinks, Justfile, .pre-commit-config.yaml, .gitignore.
# Leaves untouched: go.mod, main.go, internal/, and all your existing source.
# Verify: `ls .agent && cat AGENTS.md`. See docs/flavors/go-cli.md for a full walkthrough.
```

```bash
# Bootstrap mode: scaffold a brand-new repo with the flavor's project layout.
agent-init init go-cli ./my-new-tool
# Writes: the agentic envelope above PLUS the fresh-project files —
#   cmd/my-new-tool/main.go, internal/version/, go.mod, build/cross-build Justfile recipes.
# Verify: `cd my-new-tool && just check` should pass on the empty scaffold.

agent-init init fullstack            # current directory, default flavor (fullstack)
agent-init ./my-new-tool             # path syntax with no flavor, implies fullstack
```

```bash
# Preview without writing anything.
agent-init init go-backend --agents-only --dry-run ~/repos/my-service
# Prints the planned writes (one line per file) and exits. No files change on disk.
```

Flags for `init`:

- `--force` — overwrite existing files instead of skipping them.
- `--no-git` — skip `git init` when the target is not already a repository.
- `--dry-run` — print planned writes without changing files.
- `--agents-only` — skip the flavor's fresh-project files; ship only the agentic envelope (AGENTS.md, scripts, devcontainer, Justfile, pre-commit). For adding `agent-init` to an existing project. Supported on every code flavor: `fullstack`, `go-cli`, `go-backend`, `iac`. See [`docs/flavors/go-cli.md`](./docs/flavors/go-cli.md) for a worked example.
- `--visibility` — how the scaffold is tracked by git. `shared` (default) commits it. `local` appends a fenced, idempotent block to the committed `.gitignore` so the scaffold is ignored while the ignore rule stays visible to the team. `hidden` writes the same block to the never-committed `.git/info/exclude`, leaving no trace for other contributors. `global-default` writes the same block to your machine-wide git excludes file, ignoring the scaffold in every repository on the machine; it prints a warning and the path it edited, and you force-add the scaffold to commit it openly in a given repo. Code flavors only. See [`docs/cli.md`](./docs/cli.md).
- `--private` — alias for `--visibility=hidden`.

The `add-tracker` subcommand extends a `project-management` scaffold with a Jira / Azure DevOps / GitHub integration. Each call writes an `integrations/<tracker>/` cheatsheet and merges an entry into the target's `.mcp.json`. Idempotent and additive — multiple trackers can coexist (useful during migrations). See [`docs/cli.md`](./docs/cli.md) and [`docs/flavors/project-management.md`](./docs/flavors/project-management.md) for details.

## What It Writes

The **code flavors** (`fullstack`, `go-cli`, `go-backend`, `iac`) all produce this skeleton:

```text
your-project/
├── .devcontainer/
├── .agent/
│   ├── AGENTS.md
│   ├── CLAUDE.md -> AGENTS.md
│   ├── CODEBASE.md
│   ├── CORRECTIONS.md
│   └── scripts/
│       ├── check.sh
│       ├── gen-codemap.sh
│       └── review.sh
├── AGENTS.md -> .agent/AGENTS.md
├── CLAUDE.md -> .agent/CLAUDE.md
├── .pre-commit-config.yaml
├── Justfile
├── .gitignore
└── README.agent.md
```

On top of that skeleton, each code flavor adds its own fresh-project files:

- `fullstack` — `apis/`, `clients/`, an OpenAPI-aware Justfile, and a Playwright `record-feature.sh` script.
- `go-cli` — `cmd/{{.ProjectName}}/main.go` (rendered to your target dir name), `internal/version/`, `go.mod`, and a Justfile with `build`, `cross-build`.
- `go-backend` — `cmd/server/main.go`, `internal/api/handlers.go` + tests, `go.mod`, and a Justfile with `run-dev`, `build`, `cross-build`.
- `iac` — `terraform/` (root module + `modules/`) and `ansible/` (`inventory/`, `playbooks/`, `roles/`, `requirements.yml`) trees, `ansible.cfg`, `.tflint.hcl`, `.yamllint.yml`, `.ansible-lint`, and a Justfile whose `fmt` / `lint` / `typecheck` / `test` / `security` recipes auto-detect whether Terraform, Ansible, or both are present. Ships a flavor-local `gen-codemap.sh` that surfaces TF modules, root `.tf` `variable` / `output` / `resource` declarations, Ansible roles, and playbook task counts.

Every code flavor supports `--agents-only` for **adding the envelope to a project that already exists** — the flavor-specific files above are skipped, and (for `go-cli` and `go-backend`) the Justfile is replaced with a layout-agnostic variant that drops the `build` / `cross-build` recipes those flavors otherwise tie to a fixed `cmd/` path. Per-flavor `FreshOnlyPaths` are declared in [`internal/flavors/registry.go`](./internal/flavors/registry.go); see [`docs/flavors/go-cli.md`](./docs/flavors/go-cli.md) for a worked example.

The `claude-cowork` flavor uses a deliberately different shape — no devcontainer, no symlinks, no `.agent/` subdirectory:

```text
your-workspace/
├── AGENTS.md           # canonical agent instructions, at root
├── README.md           # human onboarding
├── decisions.md        # append-only decision log
├── corrections.md
├── .gitignore
├── reference/          # source materials (read-only context)
├── templates/          # .potx / .dotx / .xltx
└── archive/            # superseded work
```

The Justfile `check` recipe runs whatever steps the scaffolded project supports; missing recipes are skipped silently. Empty repos remain installable before any application code exists.

## How templating works

- **Content templating** — files with a `.tmpl` extension pass through `text/template`. `{{.ProjectName}}` is the project's directory name. Files without `.tmpl` are copied verbatim, so configs that legitimately contain `{{ }}` (Helm, Ansible, GitHub Actions) work unchanged.
- **Path templating** — file paths also pass through `text/template`. A template path like `cmd/{{.ProjectName}}/main.go.tmpl` renders to `cmd/my-tool/main.go`.
- **Common overlay** — files shared across every flavor live in `internal/flavors/common/templates/`. The scaffold engine walks the flavor first, then layers common in. A flavor overrides a common file by shipping its own copy at the same relative path.

See [`docs/`](./docs/) for per-feature details.

## Development

Inside this repo:

```bash
just check        # full done-gate: fmt, lint, vet, test, cross-build, smoke-test
just smoke-test   # scaffold every flavor + run its check.sh + diff against the golden
```

`just smoke-test-update` regenerates golden snapshots under `testdata/golden/<flavor>/` after intentional template changes.

The devcontainer installs Go, `goimports`, `golangci-lint`, `just`, pre-commit, the agent CLIs, and the Node tooling the fullstack flavor smoke-tests against.

## CI and releases

Pull requests to `main` run `just check`, which covers codemap regeneration, formatting, linting, `go vet`, tests, cross-builds, and the full multi-flavor smoke test.

Pushes to `main` run the same gate and build the binaries, but do not publish a release.

Releases are tag-driven. Pushing a semver tag (`vX.Y.Z`) runs the gate, builds binaries for Linux `amd64`, Linux `arm64`, macOS `arm64`, and Windows `amd64`, and publishes a GitHub release named and tagged after the pushed tag. Artifacts are `.tar.gz` (Linux and macOS) and `.zip` (Windows) with a `checksums.txt`. To cut a release:

```bash
git tag v1.2.3
git push origin v1.2.3
```

Each published archive carries a signed build provenance attestation. Verify a download before running it:

```bash
gh attestation verify agent-init-linux-amd64.tar.gz --repo Lillevang/agent-init
```
