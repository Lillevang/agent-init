# go-backend-agents-only — Agentic Development Setup

This repo was scaffolded with `agent-init` (`go-backend` flavor). It's configured for sandboxed agentic development of a Go HTTP backend: agents run inside a devcontainer, gated by a check script, with a codemap and corrections file to keep them on-rails.

## Quick start

```bash
# 1. (Once) install host dependencies — see "Host dependencies" below

# 2. Set API keys in your host shell
export ANTHROPIC_API_KEY=...
export OPENAI_API_KEY=...

# 3. Bring up the container
devcontainer up --workspace-folder .

# 4. Open a shell in it
devcontainer exec --workspace-folder . bash

# 5. Inside the container — run the server, or the agent
just run-dev
# or:
claude
```

## Host dependencies

You need these on the **host**. The container handles its own internals.

### Required

| Tool | Install |
|------|---------|
| **Podman** or Docker | `sudo dnf install -y podman podman-docker` (Fedora/WSL) |
| **Node.js + npm** | needed only for the devcontainer CLI on the host |
| **devcontainer CLI** | `npm install -g @devcontainers/cli` |
| **just** | `sudo dnf install -y just` |
| **git** | `sudo dnf install -y git` |

### Optional

| Tool | Why |
|------|-----|
| **GitHub CLI** | Agent can interact with PRs/issues from the container |
| **pre-commit** | Run hooks on the host too |
| **psql / redis-cli** | If your backend touches a database; add to `.devcontainer/Dockerfile` |

## Layout

```
.
├── .devcontainer/         # container definition (Dockerfile, devcontainer.json)
├── .agent/                # everything the agent reads
│   ├── AGENTS.md          # instructions (Codex)
│   ├── CLAUDE.md          # symlink → AGENTS.md (Claude Code)
│   ├── CODEBASE.md        # codemap (auto + hand-written sections)
│   ├── CORRECTIONS.md     # known anti-patterns
│   └── scripts/           # check.sh, review.sh, gen-codemap.sh
├── cmd/server/            # HTTP server entry point
├── internal/api/          # router and handlers
├── Justfile               # check, fmt, lint, typecheck, test, run-dev, build, cross-build
├── go.mod                 # module path — edit the `example.com/...` prefix
├── .pre-commit-config.yaml
└── README.agent.md        # this file
```

After scaffolding, you should:

1. Edit `go.mod` and change `module example.com/go-backend-agents-only` to your actual module path (e.g. `github.com/yourorg/go-backend-agents-only`).
2. Update the import in `cmd/server/main.go` from `example.com/go-backend-agents-only/internal/api` to match.
3. Edit `.agent/AGENTS.md` to describe THIS project's specifics.
4. Run `just check` to confirm the gate passes on a fresh tree.
5. Run `just run-dev` and `curl localhost:8080/healthz` to confirm the server boots.

## The done-gate

The agent considers itself done only when `just check` (a.k.a. `.agent/scripts/check.sh`) passes. This runs:

1. Codemap regeneration
2. Format (`gofmt` + `goimports`)
3. Lint (`golangci-lint`)
4. Type check (`go vet`)
5. Tests (`go test ./...`)

Recipes that don't exist are skipped silently — but **CI must run the same gate**, so don't leave it empty.

## Running the server

```bash
just run-dev          # default :8080
ADDR=:9090 just run-dev
```

The default scaffold ships a `/healthz` endpoint that returns commit + buildDate metadata. Wire your real routes in [`internal/api/handlers.go`](./internal/api/handlers.go).

## Cross-compilation

```bash
just cross-build
```

Produces `dist/server-linux-amd64` and `dist/server-linux-arm64` — the typical container deploy targets. The `-ldflags "-s -w"` strip is on by default; add `-X main.commit=$(git rev-parse HEAD) -X main.buildDate=...` if your release pipeline injects version metadata.

## Reviewer agent

After non-trivial changes:

```bash
just review
# or: REVIEWER=codex just review
```

Output lands in `.agent/REVIEW.md` (gitignored). It's a separate agent reading the diff against `main`, with read-only tool access. It catches violations of `AGENTS.md` and `CORRECTIONS.md` — useful, but not a substitute for you reading the diff yourself.

## Mounting sibling repos

Sibling repos are mounted as **peers** of this workspace inside the container:

```
/workspaces/
├── go-backend-agents-only/      ← this repo (workspace root, read-write)
├── shared-lib/            ← sibling, mounted read-only
```

Edit `.devcontainer/devcontainer.json`, add to `mounts`:

```json
"mounts": [
  "source=${localEnv:HOME}/repos/tools/shared-lib,target=/workspaces/shared-lib,type=bind,readonly"
]
```

Use `,readonly` unless cross-repo edits are legitimate.

## Adding a database

The scaffold deliberately ships without DB clients. When you add one:

1. Add the Go driver to `go.mod` (`go get github.com/jackc/pgx/v5`, etc.).
2. Add the matching CLI to `.devcontainer/Dockerfile` so the agent can poke at the DB (`postgresql-client`, `redis-tools`, …).
3. Document the connection string convention in `.agent/AGENTS.md`.
4. Consider a `docker-compose.yml` if the dev DB should boot alongside the devcontainer.

## Updating the scaffold

`agent-init --force` overwrites template files including local edits. Don't run it casually. When you improve a template, copy the file manually, or keep project-specific overrides clearly marked at the bottom of `AGENTS.md`.
