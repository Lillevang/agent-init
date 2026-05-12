# Agent Instructions for go-backend

You are working inside a sandboxed dev container. Your filesystem access is bounded but not zero — be deliberate. This file is the canonical source of truth for how to work in this repo. Both Codex (`AGENTS.md`) and Claude Code (`CLAUDE.md`, symlinked here) read it.

> **Edit me.** Replace project-generic prose with project-specific facts. The discipline of writing this well is what makes the agent useful.

---

## The done-gate

Before declaring **any** task complete, you must run:

```bash
./.agent/scripts/check.sh
```

If it fails, fix the failure. Do not declare success. Do not skip steps. Do not work around lints by adding ignore directives unless the code itself genuinely warrants it (and if so, justify in a comment).

The check script runs (in order): codemap regeneration, format (`gofmt`/`goimports`), lint (`golangci-lint`), type check (`go vet`), tests (`go test ./...`). All configured checks must pass.

## Project context

> **Replace this section.** Describe:
> - What this CLI does (one paragraph, concrete).
> - The binary entry point (`cmd/<name>/main.go`) and what it dispatches to.
> - Where domain logic lives (`internal/...`) and how packages depend on each other.
> - How releases are produced (the `cross-build` Justfile recipe; ldflags for `commit`/`buildDate`).

## File map

See [`.agent/CODEBASE.md`](./CODEBASE.md). Read it before exploring the tree blindly. The auto-generated section lists the directory tree and public Go API surface; the hand-written section explains *why* things are where they are.

If you find yourself running `find` or `tree` to understand the codebase, stop and check `CODEBASE.md` first. If `CODEBASE.md` is wrong or missing context, propose an update at the end of your turn.

## Corrections

See [`.agent/CORRECTIONS.md`](./CORRECTIONS.md). This file lists patterns you (or previous agent runs) have gotten wrong, with examples of the preferred form. Read it before starting work and after any review.

## Conventions

### Go style

- Run `gofmt` + `goimports`. Don't fight the formatter.
- Errors are values. Wrap with `fmt.Errorf("doing X: %w", err)` so the cause chain survives. No `panic` outside `main` and `init`, and only for genuinely unrecoverable startup states.
- Prefer small interfaces defined at the consumer side ("accept interfaces, return structs"). Don't define an interface until you have two implementations or a clear test reason.
- No global mutable state. Pass dependencies through constructors. `os.Stdout`/`os.Stderr` should be injectable for tests (`io.Writer` parameters, not direct `fmt.Println`).
- Context-aware operations take `context.Context` as the first parameter, even if not used yet — adding it later is churn.
- Keep functions under ~40 lines. If longer, justify in a comment.

### Project layout

- `cmd/<binary>/main.go` — entry point. Thin. Wires flags to internal packages.
- `internal/` — domain logic. Not importable from outside this module.
- `pkg/` — only if a library API is exposed publicly. Default: don't create `pkg/`.

### Naming

- Files: snake_case for Go files (`flag_parser.go`).
- Exported Go identifiers: idiomatic CamelCase, no stutter (`flags.Parser` not `flags.FlagParser`).
- CLI flags: kebab-case (`--no-git`, `--force`).

### Commits

- Conventional commits: `feat:`, `fix:`, `refactor:`, `docs:`, `test:`, `chore:`.
- One logical change per commit. If the commit message contains "and", split it.
- Never `--force-push` to shared branches.

### Dependencies

- Standard library first. Adding a dependency requires justification in the PR.
- Pin via `go.mod`.

## Testing

- Unit tests live next to the code (`foo.go` + `foo_test.go`).
- Table tests are idiomatic; use them.
- Use `t.TempDir()` and `t.Setenv()` so tests clean up after themselves.
- New functionality requires new tests. Bug fixes require a regression test that fails before the fix.

## What you should NOT do

- Do not commit secrets, API keys, or tokens.
- Do not modify `.git/`, `.devcontainer/`, or `.agent/` files unless explicitly asked.
- Do not run `git push --force` on `main` or any branch you didn't create yourself.
- Do not install global packages that aren't in the Dockerfile. If you need one, propose adding it.
- Sibling repos may be mounted as peers of this workspace at `/workspaces/<other-repo>`. Treat them as read-only context unless explicitly told they're in scope.
- Do not bypass `check.sh` failures with `--no-verify` or by editing the script.

## When you're stuck

Stop and ask. Specifically:

- If two parts of the codebase contradict each other, ask which is canonical.
- If the test suite is flaky, report it; don't retry until it passes.
- If a check fails for a reason you don't understand, report the full output rather than guessing.

## Reviewer agent

After completing a non-trivial change, run:

```bash
./.agent/scripts/review.sh
```

This invokes a separate agent that reviews your diff against this file, `CORRECTIONS.md`, and `CODEBASE.md`. Output lands in `.agent/REVIEW.md`. Read it. Address legitimate findings.

## Project-specific notes

> **Add anything here that doesn't fit elsewhere.** Quirks, gotchas, "the X service is unreliable on Tuesdays," etc.
