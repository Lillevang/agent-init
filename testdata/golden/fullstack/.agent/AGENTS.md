# Agent Instructions for fullstack

You are working inside a sandboxed dev container. Your filesystem access is bounded but not zero — be deliberate. This file is the canonical source of truth for how to work in this repo. Both Codex (`AGENTS.md`) and Claude Code (`CLAUDE.md`, symlinked here) read it.

> **Edit me.** Replace project-generic prose with project-specific facts. The discipline of writing this well is what makes the agent useful.

---

## The done-gate

Before declaring **any** task complete, you must run:

```bash
./.agent/scripts/check.sh
```

If it fails, fix the failure. Do not declare success. Do not skip steps. Do not work around lints by adding ignore directives unless the code itself genuinely warrants it (and if so, justify in a comment).

The check script runs (in order): codemap regeneration, API client generation, formatting, linting, type checking, unit tests, and Playwright tests when the project has Playwright coverage. All configured checks must pass.

## Project context

> **Replace this section.** Describe:
> - What this project does (one paragraph, concrete).
> - The frontend stack (React/Vite/Next/etc.) and backend stack (Node/Express/Fastify/etc.).
> - Where frontend, backend, shared packages, and tests live.
> - How local services are started and which ports they use.

## File map

See [`.agent/CODEBASE.md`](./CODEBASE.md). Read it before exploring the tree blindly. The auto-generated section lists the directory tree and public API surface; the hand-written section explains *why* things are where they are.

If you find yourself running `find` or `tree` to understand the codebase, stop and check `CODEBASE.md` first. If `CODEBASE.md` is wrong or missing context, propose an update at the end of your turn.

## Corrections

See [`.agent/CORRECTIONS.md`](./CORRECTIONS.md). This file lists patterns you (or previous agent runs) have gotten wrong, with examples of the preferred form. Read it before starting work and after any review.

## Conventions

> **Replace these with project-specific conventions.**

### Code style
- Follow whatever the formatter outputs. If you disagree with the formatter, do not fight it; raise the disagreement in chat.
- Prefer small, single-purpose functions. If a function exceeds ~40 lines, justify it.
- Errors are returned, not swallowed. No bare `catch`/`except` without re-raise or explicit handling.

### Naming
- Files: kebab-case (`user-service.ts`) unless the framework convention says otherwise.
- Functions/variables: TypeScript-idiomatic camelCase; exported React components use PascalCase.
- Tests live next to code (`foo.ts` + `foo.test.ts`) or in `tests/` if the project does it that way — check existing structure.

### Commits
- Conventional commits format: `feat:`, `fix:`, `refactor:`, `docs:`, `test:`, `chore:`.
- One logical change per commit. If you find yourself writing "and" in a commit message, split it.
- Never `--force-push` to shared branches. On feature branches, fine.

### Dependencies
- Don't add a dependency without explicit approval. Ask first.
- If you do add one: pin the version, justify it in the PR description.
- Prefer the standard library over a dependency for anything trivial.

## API generation

If the repo has OpenAPI specs in `apis/`, the check script regenerates clients into `clients/`. Do not hand-edit files in `clients/` — they will be overwritten. If the generated client is wrong, fix the spec or the generator config.

## Testing

> **Replace this with project-specific testing instructions.**

- Unit tests must pass before declaring done.
- For frontend feature changes, run the Playwright recording: `./.agent/scripts/record-feature.sh <feature-name>`. The mp4 lands in `.agent/recordings/`. Review it before declaring done.
- New functionality requires new tests. Bug fixes require a regression test.

## What you should NOT do

- Do not commit secrets, API keys, or tokens. The repo's `.gitignore` covers common cases but not all.
- Do not modify `.git/`, `.devcontainer/`, or `.agent/` files unless explicitly asked.
- Do not run `git push --force` on `main` or any branch you didn't create yourself.
- Do not install global packages that aren't in the Dockerfile. If you need one, propose adding it.
- Sibling repos may be mounted as peers of this workspace at `/workspaces/<other-repo>` (i.e. `cd ../<other-repo>` from the workspace root). Treat them as read-only context for understanding integrations or patterns, not as code to edit, unless explicitly told they're in scope.
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
