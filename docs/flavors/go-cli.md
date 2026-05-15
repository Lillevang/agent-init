# `go-cli` flavor

A Go command-line tool scaffold. Ships a `cmd/{{.ProjectName}}/main.go` entry point, an `internal/version/` package, a `go.mod`, and a Justfile with `build` / `cross-build` recipes wired to that layout. Standard agentic envelope on top (`.agent/`, devcontainer, pre-commit, AGENTS.md / CLAUDE.md symlinks).

Source: [internal/flavors/gocli/](../../internal/flavors/gocli/).

## Two modes

### Fresh project (default)

```bash
agent-init init go-cli ./my-tool
```

Writes the full skeleton, including the Go entry point. `cmd/my-tool/main.go` is path-templated to match the target directory name.

Best for: starting a new Go CLI from zero.

### Add to existing project

```bash
agent-init init go-cli --agents-only ~/repos/my-existing-tool
```

Writes only the agentic envelope — no `cmd/`, no `go.mod`, no `internal/version/`. Ships a layout-agnostic Justfile that drops the `build` / `cross-build` recipes (which would otherwise reference a `./cmd/{{.ProjectName}}/` directory the existing project doesn't have).

Best for: adding `agent-init`'s tooling to a Go CLI you already maintain.

## What `--agents-only` writes

```
your-tool/
├── .agent/
│   ├── AGENTS.md
│   ├── CLAUDE.md -> AGENTS.md
│   ├── CODEBASE.md
│   ├── CORRECTIONS.md
│   └── scripts/
│       ├── check.sh
│       ├── gen-codemap.sh
│       └── review.sh
├── .devcontainer/
│   ├── Dockerfile
│   ├── devcontainer.json
│   └── post-create.sh
├── .gitignore
├── .pre-commit-config.yaml
├── AGENTS.md -> .agent/AGENTS.md
├── CLAUDE.md -> .agent/CLAUDE.md
├── Justfile                 # layout-agnostic variant
└── README.agent.md
```

What's *not* written (the `FreshOnlyPaths` for `go-cli`):

- `cmd/{{.ProjectName}}/main.go`
- `go.mod`
- `internal/version/version.go`
- `internal/version/version_test.go`

Declared at [internal/flavors/registry.go](../../internal/flavors/registry.go) on the `go-cli` Flavor.

## Justfile recipes

Fresh and agents-only mode ship slightly different Justfiles.

| Recipe | Fresh | Agents-only |
|---|:---:|:---:|
| `check` | ✓ | ✓ |
| `codemap` | ✓ | ✓ |
| `fmt`, `lint`, `typecheck`, `test`, `vulncheck` | ✓ | ✓ |
| `review` | ✓ | ✓ |
| `build`, `cross-build` | ✓ | — |

In agents-only mode `build` / `cross-build` are dropped because they hardcode `./cmd/{{.ProjectName}}/`. Your existing project has its own entry-point layout — copy a `build` recipe from the fresh-mode Justfile and adapt the path. The agents-only Justfile is shipped from [internal/flavors/gocli/templates/Justfile.agents-only.tmpl](../../internal/flavors/gocli/templates/Justfile.agents-only.tmpl).

## After scaffolding into an existing project

The `--agents-only` scaffold *adds* files; it doesn't touch your existing `go.mod`, your `main.go`, your tests, or anything else. But there are a few collisions worth watching:

- **`.gitignore`** — scaffold ships one; if you already have one, `--agents-only` skips it (you'll see a `skip` notice). Use `--force` if you want the scaffold's version.
- **`Justfile`** — same. If you already have a Justfile, scaffold skips it. With `--force` it's replaced wholesale; you'll lose any custom recipes.
- **`.pre-commit-config.yaml`** — same. If you have hooks configured, scaffold skips. If you want the agentic hooks plus your own, merge by hand.
- **`AGENTS.md` / `CLAUDE.md`** — scaffold creates these as symlinks pointing into `.agent/`. If you had real files at those paths, they're skipped.

Run with `--dry-run` first to see what would land:

```bash
agent-init init go-cli --agents-only --dry-run ~/repos/my-existing-tool
```

## Why this design

Two real use cases motivated the split:

1. **Starting fresh**: agent wants a Go CLI scaffolded from zero, complete with module, entry point, and version package.
2. **Adopting agents in an existing tool**: maintainer has a Go CLI with its own layout (often *not* `cmd/<name>/main.go`) and wants the agent envelope (devcontainer, pre-commit, done-gate, AGENTS.md) without their existing files getting overwritten.

A flag is the natural seam — same flavor, two modes — rather than a parallel `go-cli-agents` flavor that would duplicate most of the template tree.

## Engine internals

The flavor declares two fields that govern agents-only behavior:

```go
SupportsAgentsOnly: true,
FreshOnlyPaths: []string{
    "cmd/{{.ProjectName}}/main.go",
    "go.mod",
    "internal/version/version.go",
    "internal/version/version_test.go",
},
```

- `SupportsAgentsOnly` is required for the CLI to accept `--agents-only` against this flavor. Without it the CLI returns an error.
- `FreshOnlyPaths` is consulted only when `--agents-only` is set. Paths are matched after `.tmpl` is stripped from the source filename, before `{{.ProjectName}}` is rendered.
- The `<file>.agents-only.<ext>` naming convention lets a flavor ship a *variant* of a file (e.g. a layout-agnostic Justfile) that's used only in agents-only mode. The base file is shadowed automatically when the variant exists in the same layer.

Source: [scaffold.go:walkLayer](../../internal/scaffold/scaffold.go).

## Tests

- Engine: [internal/scaffold/scaffold_test.go](../../internal/scaffold/scaffold_test.go) — `TestRunAgentsOnlySkipsFreshOnlyPathsAndPrefersVariant` and `TestRunFreshModeIgnoresAgentsOnlyVariants`.
- CLI: [internal/cli/cli_test.go](../../internal/cli/cli_test.go) — `TestInitAgentsOnlyDropsFreshOnlyFiles` and `TestInitAgentsOnlyRejectsUnsupportedFlavor`.
- Golden snapshots: [testdata/golden/go-cli/](../../testdata/golden/go-cli/) (fresh) and [testdata/golden/go-cli-agents-only/](../../testdata/golden/go-cli-agents-only/).
