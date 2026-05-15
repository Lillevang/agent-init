# `go-backend` flavor

A Go HTTP backend scaffold. Ships `cmd/server/main.go`, an `internal/api/` router with a `/healthz` handler + table-test, a `go.mod`, and a Justfile with `run-dev` / `build` / `cross-build` recipes wired to `./cmd/server`.

Source: [internal/flavors/gobackend/](../../internal/flavors/gobackend/).

## Two modes

### Fresh project (default)

```bash
agent-init init go-backend ./my-service
```

Writes the full skeleton including the server entry point, a sample router with `/healthz`, and a table test. The Justfile's `build` and `cross-build` recipes target `./cmd/server` (not project-name-templated like `go-cli` is, because backends are conventionally `cmd/server/`).

Best for: starting a new HTTP service from zero.

### Add to existing project

```bash
agent-init init go-backend --agents-only ~/repos/my-existing-service
```

Skips the Go bootstrap files entirely. Writes only the agentic envelope, with a layout-agnostic Justfile that drops `run-dev` / `build` / `cross-build` — those hardcode `./cmd/server` and your existing service may have a different entry-point layout.

## What `--agents-only` writes

```
your-service/
├── .agent/{AGENTS.md, CLAUDE.md -> AGENTS.md, CODEBASE.md, CORRECTIONS.md, scripts/{check,gen-codemap,review}.sh}
├── .devcontainer/{Dockerfile, devcontainer.json, post-create.sh}
├── .gitignore
├── .pre-commit-config.yaml
├── AGENTS.md -> .agent/AGENTS.md
├── CLAUDE.md -> .agent/CLAUDE.md
├── Justfile               # layout-agnostic variant
└── README.agent.md
```

Not written (the `FreshOnlyPaths` for `go-backend`):

- `cmd/server/main.go`
- `go.mod`
- `internal/api/handlers.go`
- `internal/api/handlers_test.go`

Declared at [internal/flavors/registry.go](../../internal/flavors/registry.go) on the `go-backend` Flavor.

## Justfile recipes

| Recipe | Fresh | Agents-only |
|---|:---:|:---:|
| `check` | ✓ | ✓ |
| `codemap` | ✓ | ✓ |
| `fmt`, `lint`, `typecheck`, `test`, `vulncheck` | ✓ | ✓ |
| `review` | ✓ | ✓ |
| `run-dev`, `build`, `cross-build` | ✓ | — |

In agents-only mode the layout-specific recipes are dropped. Copy them from the fresh-mode Justfile and adapt the path if you want `just run-dev` against your own entry point. The variant Justfile is shipped from [internal/flavors/gobackend/templates/Justfile.agents-only.tmpl](../../internal/flavors/gobackend/templates/Justfile.agents-only.tmpl).

## After scaffolding into an existing project

Standard collisions — scaffold skips files that already exist, `--force` overwrites:

- `Justfile`, `.gitignore`, `.pre-commit-config.yaml` — skipped if present.
- `AGENTS.md` / `CLAUDE.md` — scaffold creates these as symlinks into `.agent/`. Existing real files are skipped.

Preview first:

```bash
agent-init init go-backend --agents-only --dry-run ~/repos/my-existing-service
```

## Engine internals

The flavor's registry entry:

```go
SupportsAgentsOnly: true,
FreshOnlyPaths: []string{
    "cmd/server/main.go",
    "go.mod",
    "internal/api/handlers.go",
    "internal/api/handlers_test.go",
},
```

The `Justfile.agents-only.tmpl` variant is automatically preferred over `Justfile.tmpl` in agents-only mode (see [docs/cli.md](../cli.md) → `--agents-only` → "Behavior" for the suffix convention).

## Tests

- Golden snapshots: [testdata/golden/go-backend/](../../testdata/golden/go-backend/) (fresh) and [testdata/golden/go-backend-agents-only/](../../testdata/golden/go-backend-agents-only/).
- Engine tests for the FreshOnlyPaths + variant-preference mechanism: [internal/scaffold/scaffold_test.go](../../internal/scaffold/scaffold_test.go).
