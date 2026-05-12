# agent-init

A Go CLI that scaffolds repositories for sandboxed agentic development with Codex, Claude Code, or similar tools. It writes a devcontainer, an `AGENTS.md`/`CLAUDE.md` pair, a codemap, correction notes, check and review scripts, a `Justfile`, and pre-commit wiring — tuned to a chosen flavor.

## Flavors

| Flavor | What it scaffolds |
|--------|-------------------|
| `fullstack` | TypeScript/Node frontend + backend with Playwright recording and OpenAPI client generation. |
| `go-cli` | Go command-line tool with `cmd/{{.ProjectName}}/` (path-templated), `internal/`, cross-build via Justfile, and `golangci-lint`. |
| `go-backend` | Go HTTP backend with `cmd/server`, `internal/api` router, a `/healthz` handler, and `run-dev` / `cross-build` recipes. |

Planned: `terraform`, `ansible`.

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
agent-init list-flavors
agent-init version
```

Defaults: flavor `fullstack`, target `.`.

Examples:

```bash
# Scaffold the current repo with the default flavor.
agent-init init

# Scaffold a new Go CLI repo. The cmd/ subdirectory is named after the target.
agent-init init go-cli ./my-tool

# Path syntax also works without `init`.
agent-init ./my-tool
```

Flags for `init`:

- `--force` — overwrite existing files instead of skipping them.
- `--no-git` — skip `git init` when the target is not already a repository.
- `--dry-run` — print planned writes without changing files.

## What It Writes

Every flavor produces this skeleton:

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

On top of that skeleton, each flavor adds its own files:

- `fullstack` — `apis/`, `clients/`, an OpenAPI-aware Justfile, and a Playwright `record-feature.sh` script.
- `go-cli` — `cmd/{{.ProjectName}}/main.go` (rendered to your target dir name), `internal/version/`, `go.mod`, and a Justfile with `build`, `cross-build`.
- `go-backend` — `cmd/server/main.go`, `internal/api/handlers.go` + tests, `go.mod`, and a Justfile with `run-dev`, `build`, `cross-build`.

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

Pushes to `main` run the same gate, then build Linux binaries for `amd64` and `arm64`, publish tarballs plus SHA-256 checksums, and create a GitHub release tagged `build-<run-number>`. The release matrix is structured so Windows can be added when that target enters scope.
