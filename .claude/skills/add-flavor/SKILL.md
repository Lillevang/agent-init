---
name: add-flavor
description: Add a new flavor to agent-init. Walks the seven-step authoring checklist (create internal/flavors/<name>/, write flavor.go, populate templates/, register in registry.go, add to golden test, regenerate goldens, write docs entry). Invoke when the user wants to add a project flavor — e.g. "add a terraform flavor", "create a python-cli flavor".
---

# add-flavor

You're being invoked because the user wants to add a new flavor to `agent-init`. Flavors are the central abstraction in this repo — read [`.agent/AGENTS.md`](../../../.agent/AGENTS.md) for the project conventions if you haven't already.

## Inputs you need

If the user didn't say, ask for these before writing any files:

1. **Flavor name** — kebab-case (`python-cli`, `terraform`, `rust`). This is both the directory name under `internal/flavors/` and the CLI argument users will type.
2. **One-line description** — appears in `agent-init list-flavors` output.
3. **What's unique to this flavor** — language, build tool, lint/format toolchain, any flavor-specific scripts. If any of this is fuzzy, ask before drafting templates.

Use `go-cli` as your structural reference — it's the simplest live flavor and demonstrates path templating.

## Steps

### 1. Create the flavor package

```bash
mkdir -p internal/flavors/<name>/templates
```

The Go package name is the directory name with hyphens stripped: `python-cli` → `package pythoncli`.

### 2. Write `flavor.go`

```go
package <pkgname>

import "embed"

//go:embed all:templates
var templates embed.FS

func Templates() embed.FS {
	return templates
}

func ExecutablePaths() []string {
	return []string{
		// List flavor-specific executable scripts only.
		// Do NOT list .agent/scripts/{check,gen-codemap,review}.sh — they're in common.
		".devcontainer/post-create.sh",
	}
}
```

### 3. Populate `templates/`

Add **only files unique to this flavor**. Do not copy anything from `internal/flavors/common/templates/` — the engine layers it in for you.

Typical contents:

- `.agent/AGENTS.md.tmpl` — flavor-specific agent instructions (language conventions, layout rules).
- `.agent/CODEBASE.md` and `.agent/CORRECTIONS.md` — usually identical placeholder copies; consider whether they belong in common eventually.
- `.devcontainer/Dockerfile`, `devcontainer.json.tmpl`, `post-create.sh` — the toolchain image.
- `Justfile.tmpl` — the flavor's `fmt`, `lint`, `typecheck`, `test`, `build` recipes. Keep recipe names stable (`fmt`/`lint`/etc.) so the common `check.sh` finds them via `maybe_step`.
- `README.agent.md.tmpl` — downstream user-facing readme.
- `.gitignore` and `.pre-commit-config.yaml`.
- Language scaffolding (`go.mod.tmpl`, `pyproject.toml.tmpl`, etc.) plus a minimal entry point and a passing test so `just check` passes on a fresh scaffold.

**Path templating**: if a directory name should reflect the project, use `cmd/{{.ProjectName}}/main.go.tmpl`. The `.tmpl` extension is required for files inside `{{...}}`-named directories so Go tooling doesn't try to parse the literal `{` as a package path.

### 4. Register in `internal/flavors/registry.go`

Add the import and a new `Flavor{...}` entry inside `DefaultRegistry()`. Match the pattern of the existing entries: wire `commonTemplates` as `CommonTemplates`, prepend `commonExec` to the flavor's `ExecutablePaths()`.

### 5. Add to the golden test

In [`test/golden_test.go`](../../../test/golden_test.go), append the flavor name to the `flavors := []string{...}` slice in `TestFlavorGolden`.

### 6. Regenerate the golden snapshot

```bash
just smoke-test-update
```

The recipe iterates `agent-init list-flavors`, so your new flavor is picked up automatically. The command also runs `just check` inside the scaffolded output — that catches missing test files, broken build, lint failures, etc. before the golden is written.

Review `testdata/golden/<name>/` for surprises before staging.

### 7. Write the docs entry

Create `docs/flavors/<name>.md`. Invoke the `/feature-doc` skill, or copy structure from an existing flavor doc once one is written. Then remove the matching item from the backlog in [`docs/README.md`](../../../docs/README.md) (or add a new backlog item if none existed).

## Verify

```bash
./.agent/scripts/check.sh
```

Every step in the done-gate must pass — including the multi-flavor smoke test that exercises your new flavor end-to-end. If the smoke test fails on `just check` inside the scaffolded output, the bug is almost always in one of these:

- Missing recipe in `Justfile.tmpl` that `check.sh` expects (`fmt`, `lint`, `typecheck`, `test`).
- A scaffolded source file that doesn't compile or has a failing test.
- An import path that doesn't match the `go.mod.tmpl` module path (Go flavors only).

Fix and re-run; don't disable smoke-test for your flavor.

## When to stop and ask

- If the flavor's "unique" content overlaps heavily with common, ask whether the duplication should move into common instead.
- If you need a new engine capability (e.g. a new template variable beyond `{{.ProjectName}}`), surface that as a separate task before continuing. Engine changes go through `internal/scaffold/`, not through a flavor.
