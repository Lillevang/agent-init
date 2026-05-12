# Agent Instructions for agent-init

You are working inside a sandboxed dev container. Your filesystem access is bounded but not zero — be deliberate. This file is the canonical source of truth for how to work in this repo. Both Codex (`AGENTS.md`) and Claude Code (`CLAUDE.md`, symlinked here) read it.

---

## Project context

`agent-init` is a Go CLI that scaffolds repositories for sandboxed agentic development. A user runs `agent-init [flavor]` inside a repo; the tool drops in a devcontainer, a `Justfile`, a `.pre-commit-config.yaml`, an `AGENTS.md`/`CLAUDE.md`, a `CODEBASE.md`, a `CORRECTIONS.md`, helper scripts (`check.sh`, `review.sh`, `gen-codemap.sh`, `record-feature.sh`), and supporting files — all tuned to a chosen project flavor.

**Flavors** are the central abstraction. A flavor is a named bundle of templates + metadata that says "this is what a $TYPE project needs." Initial flavors:

- `fullstack` — TypeScript/Node frontend + backend, Playwright recording, OpenAPI client generation.
- `terraform` — Terraform-heavy IaC repos. Different lint/format tooling (`terraform fmt`, `tflint`, `tfsec`/`trivy`), no Playwright, codemap surfaces modules and variables instead of code symbols.
- `ansible` — Ansible-heavy IaC repos. `ansible-lint`, `yamllint`, role/playbook discovery in the codemap.

**Stack:** Go (toolchain version pinned in `go.mod`). Standard library first; `cobra` for CLI argument parsing is acceptable if subcommand complexity warrants it, otherwise `flag` is fine. Templates are embedded into the binary via `embed.FS` — the shipped binary must be self-contained.

**Entry point:** `cmd/agent-init/main.go`. Subcommand handlers live in `internal/cli/`. Flavor definitions and templates live in `internal/flavors/<flavor-name>/`.

**Current phase:** Phase 1 — Go CLI with the `fullstack` flavor. The old bash prototype has been removed; the Go implementation and embedded flavor templates are the source of truth. Phase 2 adds `terraform` and `ansible` flavors.

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
- `agent-init list-flavors` — print available flavors with descriptions.
- `agent-init version` — print version info (commit + build date, embedded via `-ldflags`).

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

- **Token substitution uses `{{.ProjectName}}`-style Go template syntax for `.tmpl` files only.** Files without `.tmpl` extension are copied verbatim. This matters because many shell scripts and config files legitimately contain `{{ }}` (Ansible, Helm, GitHub Actions). Renaming to `.tmpl` is the explicit opt-in to templating.
- **Executable bits are not preserved by `embed.FS`.** The scaffold engine has an explicit list of paths that should be marked executable on write. Add new scripts to that list.
- **Symlinks aren't representable in `embed.FS`.** Symlinks (`AGENTS.md` ↔ `CLAUDE.md`, top-level → `.agent/*`) are created by the scaffold engine after file write, not stored as files.

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
2. Add a `flavor.go` exposing the flavor's metadata (display name, description, required tools, recommended Justfile recipes).
3. Add `templates/` with the file tree to scaffold.
4. Register the flavor in `internal/flavors/registry.go`.
5. Add a golden snapshot under `testdata/golden/<name>/`.
6. Add a smoke-test entry that scaffolds + runs `just check` on the output.

Flavors should share what's genuinely common (the `.agent/` skeleton, the devcontainer base) via composition, not copy-paste. The current pattern: a `common/` template tree that all flavors include, plus a flavor-specific overlay. Resolve conflicts in favor of the flavor-specific file.

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

## Project-specific notes

- **Recursion gotcha:** this repo's own `.agent/`, `.devcontainer/`, `Justfile`, etc. are local working files for developing `agent-init`. They are not authoritative downstream templates — those live under `internal/flavors/*/templates/`. When you're tempted to "fix" this repo's top-level scaffolding, ask whether the fix belongs in a flavor template instead.
- **IaC flavors (phase 2) will need different defaults**, especially: no Playwright, codemap based on Terraform modules / Ansible roles, different mount conventions in the devcontainer (state files, SSH keys for Ansible — handle carefully and document the security implications in the flavor's own README).
- **The binary must be cross-compilable.** `GOOS=linux GOARCH=amd64 go build` and `GOOS=darwin GOARCH=arm64 go build` both have to work. Don't reach for syscalls or platform-specific paths without a build tag.
- **Global gitignore is a footgun.** Do not expose `--gitignore-global` until it is implemented and tested. When implementing it, the help text and any printed output must make clear that the change is machine-wide and affects every repo. Default is `off` for this reason. Don't be clever about defaults.
