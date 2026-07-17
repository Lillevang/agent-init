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
│   ├── releases.md             # tag-driven release flow (DONE)
│   ├── templates.md            # .tmpl content substitution (TODO)
│   ├── path-templating.md      # {{.ProjectName}} in file paths (TODO)
│   ├── common-overlay.md       # common/ fallback layer (TODO)
│   └── done-gate.md            # what check.sh runs downstream (TODO)
└── flavors/
    ├── fullstack.md            # (DONE)
    ├── go-cli.md               # (DONE — worked --agents-only example)
    ├── go-backend.md           # (DONE)
    ├── iac.md                  # (DONE — Terraform + Ansible)
    ├── claude-cowork.md        # (DONE)
    └── project-management.md   # (DONE — worked skill examples)
```

Every flavor and subcommand is documented. The four engine docs marked `TODO` are the remaining gap; the backlog below tracks them.

## Backlog

Features that exist in the code but don't yet have a doc entry. The next agent to touch one of these is on the hook to write its doc. Every flavor is documented; the remaining gap is four engine-capability docs.

### Engine

- [ ] `engine/templates.md` — `.tmpl` opt-in for content substitution. Source: [scaffold.go:306-318](../internal/scaffold/scaffold.go#L306-L318) (content render) and the `.tmpl` strip in [walkLayer, scaffold.go:163](../internal/scaffold/scaffold.go#L163).
- [ ] `engine/path-templating.md` — `{{.ProjectName}}` in file paths; the `.tmpl` workaround for `cmd/{{.ProjectName}}/` directories. Source: [renderPath, scaffold.go:320](../internal/scaffold/scaffold.go#L320).
- [ ] `engine/common-overlay.md` — `internal/flavors/common/` as a fallback layer; flavor-first conflict resolution. Source: [Overlay, scaffold.go:102](../internal/scaffold/scaffold.go#L102) and [walkLayer, scaffold.go:163](../internal/scaffold/scaffold.go#L163).
- [ ] `engine/done-gate.md` — what `check.sh` runs in scaffolded projects and how `maybe_step` skips missing recipes. Source: [internal/flavors/common/templates/.agent/scripts/check.sh](../internal/flavors/common/templates/.agent/scripts/check.sh).

## Style

- Plain prose. No emojis. No marketing adjectives ("powerful", "elegant", "seamless").
- Lead with what the feature *is*. Then how to use it. Then edge cases / gotchas. Then source pointers.
- Code blocks for shell invocations and config snippets.
- Link to source with `file:line` so the reference survives refactors better than a paraphrase would.

## Skills

Two project-scoped Claude Code skills automate the workflows around this directory:

- `/feature-doc` — bootstraps a new doc entry against a backlog item or refreshes an existing one. Routes the doc to the right subdirectory and removes the matching TODO above. See [`.claude/skills/feature-doc/SKILL.md`](../.claude/skills/feature-doc/SKILL.md).
- `/add-flavor` — walks the seven-step flavor-authoring checklist, including the `docs/flavors/<name>.md` entry. See [`.claude/skills/add-flavor/SKILL.md`](../.claude/skills/add-flavor/SKILL.md).
