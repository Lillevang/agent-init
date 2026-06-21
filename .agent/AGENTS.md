# Agent Instructions for agent-init

You are working inside a sandboxed dev container. Your filesystem access is bounded but not zero — be deliberate. This file is the canonical source of truth for how to work in this repo. Both Codex (`AGENTS.md`) and Claude Code (`CLAUDE.md`, symlinked here) read it.

---

## Project context

`agent-init` is a Go CLI that scaffolds repositories for sandboxed agentic development. A user runs `agent-init [flavor]` inside a repo; the tool drops in a devcontainer, a `Justfile`, a `.pre-commit-config.yaml`, an `AGENTS.md`/`CLAUDE.md`, a `CODEBASE.md`, a `CORRECTIONS.md`, helper scripts (`check.sh`, `review.sh`, `gen-codemap.sh`), and supporting files — all tuned to a chosen project flavor.

**Flavors** are the central abstraction. A flavor is a named bundle of templates + metadata that says "this is what a $TYPE project needs." Live flavors:

- `fullstack` — TypeScript/Node frontend + backend, Playwright recording, OpenAPI client generation.
- `go-cli` — Go command-line tool. `cmd/{{.ProjectName}}/main.go` (path-templated), `internal/`, cross-build via Justfile, `golangci-lint`.
- `go-backend` — Go HTTP backend. `cmd/server` + `internal/api` with a `/healthz` handler, `run-dev` and `cross-build` recipes.
- `claude-cowork` — OneDrive-backed document-collaboration folder for Claude Cowork. No devcontainer, no Justfile, no symlinks, no `.agent/` subdirectory; `AGENTS.md` + `decisions.md` + `corrections.md` at root plus `reference/`, `templates/`, `archive/`. The non-code flavor — it exercises the per-flavor `Symlinks` and `NextSteps` engine hooks.
- `project-management` — workspace for running the business side of a project: epics, meetings, decisions, stakeholders, open questions, time plans. Ships five skills (`/intake-meeting`, `/break-down-epic`, `/log-decision`, `/track-stakeholder`, `/sync-tracker`) and is extended via the `agent-init add-tracker {jira|ado|gh}` subcommand which merges entries into `.mcp.json` for MCP-based tracker integration.

Planned:

- `terraform` — Terraform-heavy IaC repos. Different lint/format tooling (`terraform fmt`, `tflint`, `tfsec`/`trivy`), no Playwright, codemap surfaces modules and variables instead of code symbols.
- `ansible` — Ansible-heavy IaC repos. `ansible-lint`, `yamllint`, role/playbook discovery in the codemap.

**Stack:** Go (toolchain version pinned in `go.mod`). Standard library first; `cobra` for CLI argument parsing is acceptable if subcommand complexity warrants it, otherwise `flag` is fine. Templates are embedded into the binary via `embed.FS` — the shipped binary must be self-contained.

**Entry point:** `cmd/agent-init/main.go`. Subcommand handlers live in `internal/cli/`. Flavor definitions and templates live in `internal/flavors/<flavor-name>/`. Files shared across every flavor live in `internal/flavors/common/templates/` and are layered in by the scaffold engine.

## The done-gate

Before declaring **any** task complete, you must run:

```bash
./.agent/scripts/check.sh
```

If it fails, fix the failure. Do not declare success. Do not skip steps. Do not work around lints by adding ignore directives unless the code itself genuinely warrants it (and if so, justify in a comment).

The check script runs (in order): codemap regeneration, format (`gofmt`/`goimports`), lint (`golangci-lint`), type check (`go vet`), tests (`go test ./...`). All must pass.

**Additional gate specific to this project:** after any change to templates under `internal/flavors/*/templates/`, run the scaffold smoke test:

```bash
just smoke-test
```

This scaffolds each flavor into a temp dir, runs `just check` inside the scaffolded output (using a stub-friendly mode), and verifies the resulting tree matches a golden snapshot. If you changed templates intentionally, regenerate the golden with `just smoke-test-update` and review the diff carefully.

## File map

See [`.agent/CODEBASE.md`](./CODEBASE.md). Read it before exploring the tree blindly. The auto-generated section lists the directory tree and public Go API surface; the hand-written section explains module boundaries and the flavor-plugin pattern.

If `CODEBASE.md` is wrong or missing context, propose an update at the end of your turn.

## Corrections

See [`.agent/CORRECTIONS.md`](./CORRECTIONS.md). Read it before starting work and after any review.

## CLI surface

The binary's interface is the product. Keep it small and stable.

### Subcommands

- `agent-init init [flavor] [target-dir]` — scaffold a project. Default flavor: `fullstack`. Default target: `.`.
- `agent-init add-tracker <tracker> <target-dir>` — overlay a tracker integration (`jira`, `ado`, or `gh`) onto an existing `project-management` scaffold. Writes `integrations/<tracker>/README.md` and merges an entry into the target's `.mcp.json`. Idempotent.
- `agent-init list-flavors` — print available flavors with descriptions.
- `agent-init list-trackers` — print available trackers with descriptions.
- `agent-init version` — print version info (commit + build date, embedded via `-ldflags`).
- `agent-init upgrade [--check] [--dry-run] [--force]` — update the binary in place from the latest GitHub release. Verifies SHA-256 against the published `checksums.txt` and fails closed on mismatch. See [`docs/cli.md`](../docs/cli.md#upgrade).

### Releases

Releases are tag-driven. Pushing a semver tag (`vX.Y.Z`) is the only trigger that publishes a public GitHub Release; pushes and merges to `main` run CI (check + build) and never publish. The release tag, name, and body derive from the pushed tag (`github.ref_name`), not `github.run_number`. The release job is the only job granted `contents: write`; the workflow defaults to `contents: read`. See [`.github/workflows/release.yml`](../.github/workflows/release.yml) and [`docs/engine/releases.md`](../docs/engine/releases.md). The version-bump / who-tags convention is a separate open follow-up.

### Flags on `init`

| Flag | Purpose |
|------|---------|
| `--force` | Overwrite existing files. Default: skip with a notice. |
| `--no-git` | Skip `git init` if target isn't already a repo. |
| `--dry-run` | Print what would happen without writing anything. |

### Planned `--gitignore-global`

Global gitignore support is planned but not exposed until implemented. When added, it must be idempotent, use a clearly marked managed block, include tests in `internal/gitconfig/`, and warn that global excludes affect every repository on the user's machine.

## Conventions

### Go style

- Run `gofmt` + `goimports`. Don't fight the formatter.
- Errors are values. Wrap with `fmt.Errorf("doing X: %w", err)` so the cause chain survives. No `panic` outside `main` and `init`, and only for genuinely unrecoverable startup states.
- Prefer small interfaces defined at the consumer side (Go idiom: "accept interfaces, return structs"). Don't define an interface until you have two implementations or a clear test reason.
- No global mutable state. Pass dependencies through constructors. The CLI's `os.Stdout`/`os.Stderr` should be injectable for tests (`io.Writer` parameters, not direct `fmt.Println`).
- Context-aware operations take `context.Context` as the first parameter, even if not used yet — adding it later is churn.
- Keep functions under ~40 lines. If longer, justify in a comment.

### Project layout (Go convention)

- `cmd/agent-init/` — `main` package. Thin. Wires flags to `internal/cli`.
- `internal/cli/` — subcommand handlers. One file per subcommand.
- `internal/flavors/` — flavor definitions and embedded templates.
- `internal/scaffold/` — the engine that walks a flavor and writes files. Template substitution, skip/overwrite logic, symlink creation.
- `internal/gitconfig/` — read/write the global excludes file. Isolated module so it's easy to test with a fake HOME.
- `internal/codemap/` — codemap generation logic (so the scaffolded projects' `gen-codemap.sh` has a Go-callable equivalent if we want it later; for now, the shell script is shipped as-is).
- `pkg/` — only if we expose a library API. Default: don't create `pkg/`.

### Templates

Templates are embedded via `//go:embed all:templates`. Important constraints:

- **Content substitution uses `{{.ProjectName}}`-style Go template syntax for `.tmpl` files only.** Files without `.tmpl` extension are copied verbatim. This matters because many shell scripts and config files legitimately contain `{{ }}` (Ansible, Helm, GitHub Actions). Renaming to `.tmpl` is the explicit opt-in to content templating.
- **Path templating** applies to every file, regardless of `.tmpl` extension. A template path like `cmd/{{.ProjectName}}/main.go.tmpl` renders to `cmd/myproject/main.go`. Use this when a directory name should reflect the project name. Gotcha: a `cmd/{{.ProjectName}}/` directory containing a plain `.go` file breaks `go build ./...` in this repo because Go tooling tries to read the literal `{` as a package path. The fix is to give the file a `.tmpl` extension — that hides it from Go tooling, and `text/template` parses the no-op content fine.
- **Common overlay.** Every flavor is layered on top of `internal/flavors/common/templates/`. The scaffold engine walks the flavor first, then walks common as a fallback. A flavor overrides a common file by shipping its own copy at the same relative path. Don't copy a common file into a flavor "just to be safe" — if you need to change shared behavior for everyone, change it in `common/`.
- **Executable bits are not preserved by `embed.FS`.** The scaffold engine has an explicit list of paths that should be marked executable on write. Common scripts are listed in `common.ExecutablePaths()`; flavor-specific scripts go in the flavor's own list.
- **Symlinks aren't representable in `embed.FS`.** Symlinks are declared per-flavor via `Flavor.Symlinks` and created by the scaffold engine after file write. Code flavors get the canonical trio (`AGENTS.md` → `.agent/AGENTS.md`, `CLAUDE.md` → `.agent/CLAUDE.md`, `.agent/CLAUDE.md` → `AGENTS.md`) via `codeFlavorSymlinks()` in `registry.go`. Non-code flavors (e.g. `claude-cowork`) set `Symlinks` to nil; the engine just skips the step.
- **Post-scaffold message is per-flavor.** `Flavor.NextSteps func(target string) string` returns the "what to do next" text printed after writing. If nil, the engine prints the default code-project message (devcontainer up + just check). Doc-collab flavors override this.

### Naming

- Files: snake_case for Go files (`flavor_registry.go`), kebab-case for shell scripts (`gen-codemap.sh`), as conventions dictate.
- Exported Go identifiers: idiomatic CamelCase, no stutter (`flavors.Registry` not `flavors.FlavorRegistry`).
- CLI flags: kebab-case (`--no-git`, `--force`, `--gitignore-global`).

### Commits

- Conventional commits: `feat:`, `fix:`, `refactor:`, `docs:`, `test:`, `chore:`.
- One logical change per commit. Splitting "Go port" into many commits is encouraged.
- Never `--force-push` to `main`. On feature branches, fine.

### Dependencies

- Standard library first. Keep the Go CLI lean unless a dependency clearly pays for itself.
- Acceptable additions without asking: `github.com/spf13/cobra`, `github.com/spf13/pflag` (if cobra is chosen).
- Anything else: ask first. Justify in the PR. Pin via `go.mod`.

## Documentation

This repo has two documentation surfaces. Both are part of the implementation; neither is optional.

### Per-feature docs (`./docs/`)

Every user-visible feature has an entry under `./docs/`. A "feature" here means a CLI subcommand, a flag, a flavor, an engine capability (path templating, layer overlay, the done-gate), or any behavior a downstream user could rely on. Internal refactors don't need a docs entry.

Before declaring a feature task complete, you must:

- **Add** a doc entry if the feature has none.
- **Verify** the entry still matches current behavior if one exists. Out-of-date docs are worse than missing ones — they actively mislead.

Doc files are short. One page or less. Reference source files and tests rather than duplicating them; the source is the truth, the doc is a map into it. Use `file:line` references where helpful so links stay machine-checkable.

The convention and the current backlog of features needing docs live in [`./docs/README.md`](../docs/README.md).

### Project README

The root `README.md` is the first thing a new visitor reads. Keep it:

- **Accurate** — when you ship a flavor or a flag, update the README in the same commit. Stale READMEs are silent bugs.
- **Plain prose** — no emojis. No marketing adjectives (*powerful, elegant, robust, seamless, comprehensive, lightweight, blazing fast*). No filler verbs (*leverage, delve into, embark on, facilitate, unlock*). No tagline phrasing like *"X isn't just Y, it's Z"* or *"the modern way to ..."*.
- **Short sentences** — direct, not hedged. "Foo does X" beats "Foo is a tool that can be used to perform X".
- **Scannable structure** — headings match the user's mental model: Build, Usage, What It Writes, Development, CI.

After editing the README, read it top to bottom as if you'd never seen the project. If it reads like a press release or like an LLM trying to sound enthusiastic, rewrite it.

## Testing

- Unit tests for `internal/scaffold/` (substitution, skip-existing, force, symlink creation) — these are the engine's correctness guarantees.
- Unit tests for `internal/gitconfig/` with a fake `HOME` via `t.Setenv("HOME", t.TempDir())`. Cover: no existing excludes file, existing file without our block, existing file with our block (idempotency), the case where `core.excludesfile` points somewhere unusual.
- Table tests are idiomatic; use them.
- Integration tests live in `test/` and run the compiled binary against a tempdir. Use `t.TempDir()` so cleanup is automatic.
- For each flavor: a golden-file test that scaffolds into a tempdir and diffs against `testdata/golden/<flavor>/`. Update goldens via `go test ./... -update` (implement an `-update` flag in the test).
- New functionality requires new tests. Bug fixes require a regression test that fails before the fix.

There is no Playwright in this repo. The `record-feature.sh` script is something we *ship* to scaffolded projects (in the `fullstack` flavor). It is not something we run here.

## API generation

This project does not consume external APIs and has no `apis/` directory. The `apis/` and `clients/` directories are templates we ship to downstream scaffolded projects.

## Flavor authoring

When adding a new flavor:

1. Create `internal/flavors/<name>/`.
2. Add a `flavor.go` exposing `Templates()` (via `//go:embed all:templates`) and `ExecutablePaths()` for any flavor-specific scripts.
3. Add `templates/` with files that are *unique* to this flavor. For **code flavors**, do not duplicate anything that already lives in `internal/flavors/common/templates/` — the engine layers common in. For **doc-collab flavors** (like `claude-cowork`), omit `CommonTemplates` entirely; common's `.agent/scripts/` don't apply.
4. Register the flavor in `internal/flavors/registry.go`:
   - Code flavors: wire `commonTemplates` as `CommonTemplates`, prepend `commonExec` to executables, set `Symlinks: codeFlavorSymlinks()`.
   - Doc-collab flavors: omit `CommonTemplates`, leave `Symlinks` nil, provide a `NextSteps func(target string) string` that explains the post-scaffold setup (since `just check` doesn't apply).
5. Add the flavor name to the slice in `test/golden_test.go` so the golden test exercises it.
6. Generate the golden snapshot with `just smoke-test-update` — the recipe iterates `agent-init list-flavors`, so no Justfile change is needed. The recipe checks `AGENTS.md` at root *or* in `.agent/`, and runs `just check` only when a `Justfile` exists.
7. Add a docs entry under `./docs/` describing the new flavor (see [Documentation](#documentation)).

Resolve common-overlay conflicts in favor of the flavor-specific file. The engine handles that automatically by tracking which relative paths each layer claimed.

The per-flavor hooks (`Symlinks`, `NextSteps`, optional `CommonTemplates`) are documented in [`docs/engine/flavor-hooks.md`](../docs/engine/flavor-hooks.md).

## What you should NOT do

- Do not commit secrets, API keys, or tokens.
- Do not modify `.git/`, `.devcontainer/`, or `.agent/` files (the ones in *this* repo) unless explicitly asked. Edits to *template* files under `internal/flavors/*/templates/` are normal work.
- Do not run `git push --force` on `main` or any branch you didn't create yourself.
- Do not install global packages that aren't in the Dockerfile. If you need one, propose adding it.
- Sibling repos may be mounted as peers of this workspace at `/workspaces/<other-repo>`. Treat them as read-only context unless explicitly told they're in scope.
- Do not bypass `check.sh` failures with `--no-verify` or by editing the script.
- Do not introduce a templating engine other than `text/template` (or `html/template` for HTML files, if that ever applies). The `.tmpl` extension convention exists specifically to avoid escaping wars with files that contain `{{` natively.
- Do not add a flavor without a golden-file test. Untested templates rot.
- Do not modify the user's global git config without it being an explicit user-requested action (i.e. only the `--gitignore-global` flag may touch it, and only with the documented managed-block pattern). Never silently mutate any other `git config --global` keys.

## When you're stuck

Stop and ask. Specifically:

- If two parts of the codebase contradict each other, ask which is canonical.
- If the test suite is flaky, report it; don't retry until it passes.
- If a check fails for a reason you don't understand, report the full output rather than guessing.
- If a templating decision feels like it'll force a design choice on every downstream user of a flavor, surface it before implementing — these decisions are sticky.

## Reviewer agent

After completing a non-trivial change, run:

```bash
./.agent/scripts/review.sh
```

This invokes a separate agent that reviews your diff against this file, `CORRECTIONS.md`, and `CODEBASE.md`. Output lands in `.agent/REVIEW.md`. Read it. Address legitimate findings.

## Skills

Project-scoped Claude Code skills live under `.claude/skills/<name>/SKILL.md` and are invoked with `/<name>`. Current skills:

- `/add-flavor` — walks the flavor-authoring checklist end-to-end (create package, write `flavor.go`, populate templates, register, golden, docs).
- `/feature-doc` — adds or refreshes a doc entry under `./docs/` and updates the backlog. The mechanism that drains the backlog one feature at a time.

If you find yourself doing repetitive work that neither skill covers, propose a new skill rather than encoding the procedure as ad-hoc instructions in this file. A skill is justified when the workflow is multi-step, error-prone, and likely to recur.

## Project-specific notes

- **Recursion gotcha:** this repo's own `.agent/`, `.devcontainer/`, `Justfile`, etc. are local working files for developing `agent-init`. They are not authoritative downstream templates — those live under `internal/flavors/*/templates/`. When you're tempted to "fix" this repo's top-level scaffolding, ask whether the fix belongs in a flavor template instead.
- **IaC flavors (phase 2) will need different defaults**, especially: no Playwright, codemap based on Terraform modules / Ansible roles, different mount conventions in the devcontainer (state files, SSH keys for Ansible — handle carefully and document the security implications in the flavor's own README).
- **The binary must be cross-compilable.** `GOOS=linux GOARCH=amd64 go build` and `GOOS=darwin GOARCH=arm64 go build` both have to work. Don't reach for syscalls or platform-specific paths without a build tag.
- **Global gitignore is a footgun.** Do not expose `--gitignore-global` until it is implemented and tested. When implementing it, the help text and any printed output must make clear that the change is machine-wide and affects every repo. Default is `off` for this reason. Don't be clever about defaults.
