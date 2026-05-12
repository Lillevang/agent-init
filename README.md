# agent-init

`agent-init` is a Go CLI that scaffolds repositories for sandboxed agentic development with Codex, Claude Code, or similar tools. It writes an agent-ready devcontainer, `AGENTS.md`/`CLAUDE.md`, a project codemap, correction notes, check/review scripts, a `Justfile`, and pre-commit wiring tuned to a selected flavor.

The current implemented flavor is:

- `fullstack` — TypeScript/Node fullstack projects with OpenAPI client generation and Playwright recording support.

## Build

```bash
go build -o agent-init ./cmd/agent-init
```

For a local install:

```bash
go install ./cmd/agent-init
```

Release builds should set version metadata:

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

Defaults:

- Flavor: `fullstack`
- Target directory: `.`

Examples:

```bash
# Scaffold the current repo with the default fullstack flavor.
agent-init init

# Scaffold a new fullstack repo.
agent-init init fullstack ./my-app

# Backward-compatible shorthand also works.
agent-init ./my-app
```

Flags for `init`:

- `--force` — overwrite existing files instead of skipping them.
- `--no-git` — skip `git init` when the target is not already a repository.
- `--dry-run` — print planned writes without changing files.

## What It Writes

For the `fullstack` flavor, the CLI writes:

```text
your-project/
├── .devcontainer/
├── .agent/
│   ├── AGENTS.md
│   ├── CLAUDE.md -> AGENTS.md
│   ├── CODEBASE.md
│   ├── CORRECTIONS.md
│   └── scripts/
├── apis/
├── clients/
├── AGENTS.md -> .agent/AGENTS.md
├── CLAUDE.md -> .agent/CLAUDE.md
├── .pre-commit-config.yaml
├── Justfile
├── .gitignore
└── README.agent.md
```

The fullstack `Justfile` runs npm-based formatting, linting, type checking, unit tests, OpenAPI generation, and Playwright tests when the target project has the corresponding files or npm scripts. Empty repos remain stub-friendly so the scaffold can be installed before application code exists.

## Development

Inside this repo:

```bash
just check
just smoke-test
```

The devcontainer installs Go, `goimports`, `golangci-lint`, `just`, pre-commit, Node tooling for the embedded fullstack flavor, and the agent CLIs.

The old bash implementation has been removed. The embedded flavor templates under `internal/flavors/<flavor>/templates/` are now the source of truth for scaffolded output.

## CI And Releases

Pull requests to `main` run `just check`, which covers codemap regeneration, formatting, linting, `go vet`, tests, cross-builds, and the fullstack scaffold smoke test.

Pushes to `main` run the same gate, then build Linux binaries for `amd64` and `arm64`, publish tarballs plus SHA-256 checksums, and create a GitHub release tagged as `build-<run-number>`. The release build matrix is ready for Windows entries when that target enters scope.
