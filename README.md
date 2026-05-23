# agent-init

`agent-init` adds an agentic development workflow to a Go, TypeScript, or Infrastructure-as-Code project — primarily by enhancing one you already have, optionally by bootstrapping a new repo from scratch.

It also supports project management workspaces, by bootstrapping agents and folder structures as well as MCP integrations for Azure DevOps, Jira or GitHub.

In addition you can bootstrap a workspace for claude cowork, useful for projects centered around documents, analysis and design rather than implementation.

## What it's for

The primary use case is **adding agents to existing projects**. Most code already exists. Run:

```bash
agent-init init <flavor> --agents-only ./your-existing-repo
```

…and the scaffold drops in just the agentic envelope — a devcontainer, an `AGENTS.md`/`CLAUDE.md` pair (the agent's house rules), helper scripts, a `Justfile`, and pre-commit hooks — **without touching** your `go.mod`, `package.json`, `main.go`, or any existing source. The "fresh project" mode (omit `--agents-only`) is supported and useful when you genuinely are starting from zero, but it's the secondary path.

Two things this is trying to enforce:

- **Sandbox the agent.** The devcontainer is the work surface. Agents run inside it with a bounded toolchain (Go, Node, lint, security scanners — whatever the chosen flavor needs) and write only into the mounted workspace. Credentials, host SSH keys, and cloud configs are explicit opt-in mounts — commented out in `devcontainer.json` by default so the agent doesn't pick them up unless you choose to share them.
- **Force agents to verify.** Every scaffolded project ships a `check.sh` "done-gate" that runs codemap regeneration, formatting, linting, type-checking, tests, and (per flavor) cross-builds, vulnerability scans, or security scans. The agent's contract — encoded in the generated `AGENTS.md` — is: don't declare a task complete until `./.agent/scripts/check.sh` passes. This is the simplest mechanism we've found that stops agents from cutting work short with "I think it's done."

## Flavors

| Flavor | What it scaffolds |
|--------|-------------------|
| `fullstack` | TypeScript/Node frontend + backend with Playwright recording and OpenAPI client generation. |
| `go-cli` | Go command-line tool with `cmd/{{.ProjectName}}/` (path-templated), `internal/`, cross-build via Justfile, and `golangci-lint`. |
| `go-backend` | Go HTTP backend with `cmd/server`, `internal/api` router, a `/healthz` handler, and `run-dev` / `cross-build` recipes. |
| `claude-cowork` | OneDrive-backed document collaboration folder for Claude Cowork. No devcontainer / Justfile / symlinks; root-level `AGENTS.md` + `decisions.md` + `corrections.md` + `reference/`, `templates/`, `archive/`. |
| `project-management` | Project-management workspace (epics, meetings, decisions, stakeholders, time plans). Ships five skills (`/intake-meeting`, `/break-down-epic`, `/log-decision`, `/track-stakeholder`, `/sync-tracker`) and supports MCP tracker integrations via `agent-init add-tracker {jira\|ado\|gh}`. |
| `iac` | Combined Terraform + Ansible scaffold. Ships `terraform/` (root module, `modules/`) and `ansible/` (`inventory/`, `playbooks/`, `roles/`) trees, a devcontainer with `terraform` + `tflint` + `tfsec` + `trivy` + `ansible-core` + `ansible-lint` + `yamllint`, and a Justfile whose recipes auto-detect which toolchain is present. Cloud-credential and `~/.ssh` mounts are commented out by default with a warning. |

## Build

```bash
go build -o agent-init ./cmd/agent-init
```

For a local install:

```bash
go install ./cmd/agent-init
```

Release builds set version metadata via `-ldflags`:

```bash
go build \
  -ldflags "-X main.commit=$(git rev-parse --short HEAD) -X main.buildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  -o agent-init ./cmd/agent-init
```

## Usage

```bash
agent-init init [flavor] [target-dir]
agent-init add-tracker <tracker> <target-dir>
agent-init list-flavors
agent-init list-trackers
agent-init version
```

Defaults: flavor `fullstack`, target `.`.

Run `agent-init --help` for the subcommand list, and `agent-init <command> --help` for a command's flags and examples. See [`docs/cli.md`](./docs/cli.md) for the full reference.

Examples:

```bash
# Primary use case: add the agentic envelope to an existing project.
# Drops in .devcontainer/, .agent/, Justfile, .pre-commit-config.yaml, etc.
# Does NOT touch go.mod, package.json, main.go, or anything else you already have.
agent-init init go-cli --agents-only ~/repos/my-existing-cli
agent-init init fullstack --agents-only ~/repos/my-existing-app
agent-init init iac --agents-only ~/repos/my-existing-infra

# Bootstrap mode: scaffold a brand-new repo with the flavor's project layout.
agent-init init go-cli ./my-new-tool
agent-init init fullstack            # current directory, default flavor
agent-init ./my-new-tool             # path syntax, implies fullstack

# Preview without writing anything.
agent-init init go-backend --agents-only --dry-run ~/repos/my-service
```

Flags for `init`:

- `--force` — overwrite existing files instead of skipping them.
- `--no-git` — skip `git init` when the target is not already a repository.
- `--dry-run` — print planned writes without changing files.
- `--agents-only` — skip the flavor's fresh-project files; ship only the agentic envelope (AGENTS.md, scripts, devcontainer, Justfile, pre-commit). For adding `agent-init` to an existing project. Supported on every code flavor: `fullstack`, `go-cli`, `go-backend`, `iac`. See [`docs/flavors/go-cli.md`](./docs/flavors/go-cli.md) for a worked example.

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

Pushes to `main` run the same gate, then build binaries for Linux `amd64`, Linux `arm64`, macOS `arm64`, and Windows `amd64` — published as `.tar.gz` (Linux and macOS) and `.zip` (Windows) with a `checksums.txt`, attached to a GitHub release tagged `build-<run-number>`.
