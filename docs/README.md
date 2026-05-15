# Docs

Per-feature documentation for `agent-init`. The convention is defined in [`.agent/AGENTS.md`](../.agent/AGENTS.md#documentation); this file is the index.

## Rule

Every user-visible feature has an entry here. A "feature" is a CLI subcommand, a flag, a flavor, an engine capability, or any behavior a downstream user could rely on. Internal refactors don't need a doc.

Before declaring a feature task complete:

- **Add** an entry if the feature has none.
- **Verify** the existing entry still matches current behavior.

Out-of-date docs mislead worse than missing docs. Keep entries short (one page or less) and link into source rather than restating it.

## File layout

One markdown file per topic. Group flavors under `flavors/`, engine capabilities under `engine/`, CLI surface flat. Suggested layout:

```
docs/
├── README.md                   # this file
├── cli.md                      # subcommands + flags (DONE)
├── engine/
│   ├── flavor-hooks.md         # Symlinks, NextSteps, CommonTemplates (DONE)
│   ├── templates.md            # .tmpl content substitution
│   ├── path-templating.md
│   ├── common-overlay.md
│   └── done-gate.md
└── flavors/
    ├── claude-cowork.md        # (DONE)
    ├── project-management.md   # (DONE — with worked examples)
    ├── fullstack.md
    ├── go-cli.md
    └── go-backend.md
```

Not all of these exist yet. The backlog below lists what's needed.

## Backlog

Features that exist in the code but don't yet have a doc entry. The next agent to touch one of these is on the hook to write its doc.

### Engine

- [ ] `engine/templates.md` — `.tmpl` opt-in for content substitution. Source: [internal/scaffold/scaffold.go:155-167](../internal/scaffold/scaffold.go#L155-L167).
- [ ] `engine/path-templating.md` — `{{.ProjectName}}` in file paths; the `.tmpl` workaround for `cmd/{{.ProjectName}}/` directories. Source: [internal/scaffold/scaffold.go:169-182](../internal/scaffold/scaffold.go#L169-L182).
- [ ] `engine/common-overlay.md` — `internal/flavors/common/` as a fallback layer; flavor-first conflict resolution. Source: [internal/flavors/common/flavor.go](../internal/flavors/common/flavor.go), [internal/scaffold/scaffold.go:80-114](../internal/scaffold/scaffold.go#L80-L114).
- [ ] `engine/done-gate.md` — what `check.sh` runs in scaffolded projects and how `maybe_step` skips missing recipes. Source: [internal/flavors/common/templates/.agent/scripts/check.sh](../internal/flavors/common/templates/.agent/scripts/check.sh).

### Flavors

- [ ] `flavors/fullstack.md` — TypeScript/Node frontend + backend, Playwright recording, OpenAPI client generation. Source: [internal/flavors/fullstack/](../internal/flavors/fullstack/).
- [x] `flavors/go-cli.md` — fresh-project Go CLI scaffold + `--agents-only` mode for existing projects.
- [ ] `flavors/go-backend.md` — `cmd/server`, `internal/api`, `/healthz`, `run-dev`. Source: [internal/flavors/gobackend/](../internal/flavors/gobackend/).

## Style

- Plain prose. No emojis. No marketing adjectives ("powerful", "elegant", "seamless").
- Lead with what the feature *is*. Then how to use it. Then edge cases / gotchas. Then source pointers.
- Code blocks for shell invocations and config snippets.
- Link to source with `file:line` so the reference survives refactors better than a paraphrase would.

## Skills

Two project-scoped Claude Code skills automate the workflows around this directory:

- `/feature-doc` — bootstraps a new doc entry against a backlog item or refreshes an existing one. Routes the doc to the right subdirectory and removes the matching TODO above. See [`.claude/skills/feature-doc/SKILL.md`](../.claude/skills/feature-doc/SKILL.md).
- `/add-flavor` — walks the seven-step flavor-authoring checklist, including the `docs/flavors/<name>.md` entry. See [`.claude/skills/add-flavor/SKILL.md`](../.claude/skills/add-flavor/SKILL.md).
