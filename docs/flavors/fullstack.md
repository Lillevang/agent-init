# `fullstack` flavor

A TypeScript/Node fullstack scaffold. Ships an OpenAPI client-generation hook (`apis/` + `clients/`), Playwright feature recording, and a Justfile whose recipes auto-detect what the project ships (`package.json`, `tsconfig.json`, `playwright.config.*`, etc.) and no-op when nothing matches.

Source: [internal/flavors/fullstack/](../../internal/flavors/fullstack/).

## Two modes

### Fresh project (default)

```bash
agent-init init fullstack ./my-app
```

Writes the full skeleton including `apis/README.md` (OpenAPI specs go here) and `clients/README.md` (generated TypeScript clients land here). The flavor doesn't ship a `package.json` — you bring your own when you initialize the frontend/backend toolchains. The Justfile is wired so once `package.json` exists, `npm run lint`/`typecheck`/`test`/`format` start working automatically.

### Add to existing project

```bash
agent-init init fullstack --agents-only ~/repos/my-existing-app
```

Skips `apis/README.md` and `clients/README.md` — those are starter documentation for the OpenAPI workflow you may or may not be using. Writes only the agentic envelope.

The Justfile is unchanged between modes. Every recipe is layout-agnostic by design — for example, `lint`:

```bash
if [[ ! -f package.json ]]; then echo "no package.json, skipping lint"; exit 0; fi
if npm run | grep -qE '^  lint($|:)'; then npm run lint; ...
```

So you don't lose any functionality going agents-only; the recipes simply wake up if your existing project meets the conditions they check for.

## What `--agents-only` writes

```
your-app/
├── .agent/
│   ├── AGENTS.md, CLAUDE.md -> AGENTS.md, CODEBASE.md, CORRECTIONS.md
│   └── scripts/{check,gen-codemap,review,record-feature}.sh
├── .devcontainer/{Dockerfile, devcontainer.json, post-create.sh}
├── .gitignore
├── .pre-commit-config.yaml
├── AGENTS.md -> .agent/AGENTS.md
├── CLAUDE.md -> .agent/CLAUDE.md
├── Justfile
└── README.agent.md
```

Not written (the `FreshOnlyPaths` for `fullstack`):

- `apis/README.md`
- `clients/README.md`

Declared at [internal/flavors/registry.go](../../internal/flavors/registry.go) on the `fullstack` Flavor.

## Justfile recipes

| Recipe | Behavior |
|---|---|
| `check` | Runs `.agent/scripts/check.sh`, which calls `maybe_step` for each recipe below. |
| `codemap` | Regenerates `.agent/CODEBASE.md`. |
| `generate-clients` | `npx openapi-typescript apis/*.{yaml,yml,json}` → `clients/*.d.ts`. No-op if `apis/` is empty. |
| `fmt`, `lint`, `typecheck`, `test` | Run the matching `npm run <name>` script if defined, else fall back (`prettier` for fmt, `tsc --noEmit` for typecheck), else skip. |
| `playwright` | `npx playwright test` if `playwright.config.*` or `tests/e2e/` exists; else skip. |
| `record FEATURE` | `.agent/scripts/record-feature.sh` — Playwright video capture, available in both modes. |
| `review` | Invoke the reviewer agent. |

Recipe definitions: [internal/flavors/fullstack/templates/Justfile.tmpl](../../internal/flavors/fullstack/templates/Justfile.tmpl).

## After scaffolding into an existing project

The `--agents-only` scaffold *adds* the envelope; it doesn't touch your existing source, `package.json`, or build configuration. Collisions to watch:

- **`Justfile` / `.gitignore` / `.pre-commit-config.yaml`** — if you already have these, scaffold skips with a notice. `--force` overwrites; you'd lose customizations.
- **`AGENTS.md` / `CLAUDE.md`** — scaffold creates these as symlinks pointing into `.agent/`. If you had real files at those paths, they're skipped.

Run `--dry-run` first to preview:

```bash
agent-init init fullstack --agents-only --dry-run ~/repos/my-existing-app
```

## Engine internals

The flavor's registry entry:

```go
SupportsAgentsOnly: true,
FreshOnlyPaths: []string{
    "apis/README.md",
    "clients/README.md",
},
```

No `Justfile.agents-only.tmpl` variant exists for this flavor — the base Justfile is already layout-agnostic, so the engine writes it as-is in both modes.

## Tests

- Golden snapshots: [testdata/golden/fullstack/](../../testdata/golden/fullstack/) (fresh) and [testdata/golden/fullstack-agents-only/](../../testdata/golden/fullstack-agents-only/).
- Engine tests cover the FreshOnlyPaths skip behavior generically: [internal/scaffold/scaffold_test.go](../../internal/scaffold/scaffold_test.go).
