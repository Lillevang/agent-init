# Docs

Per-feature documentation for `agent-init`. The convention is defined in [`.agent/AGENTS.md`](../.agent/AGENTS.md#documentation); this file is the index.

## Rule

Every user-visible feature has an entry here. A "feature" is a CLI subcommand, a flag, a flavor, an engine capability, or any behavior a downstream user could rely on. Internal refactors don't need a doc.

Before declaring a feature task complete:

- **Add** an entry if the feature has none.
- **Verify** the existing entry still matches current behavior.

Out-of-date docs mislead worse than missing docs. Keep entries short (one page or less) and link into source rather than restating it.

## File layout

One markdown file per topic. Group flavors under `flavors/`, engine capabilities under `engine/`, CLI surface flat.

```
docs/
├── README.md                   # this file
├── cli.md                      # subcommands + flags (DONE)
├── engine/
│   ├── flavor-hooks.md         # Symlinks, NextSteps, CommonTemplates (DONE)
│   ├── templates.md            # .tmpl content substitution (DONE)
│   ├── path-templating.md      # {{.ProjectName}} in file paths (DONE)
│   ├── common-overlay.md       # common/ fallback layer (DONE)
│   ├── done-gate.md            # what check.sh runs downstream (DONE)
│   └── releases.md             # tag-driven release flow (DONE)
└── flavors/
    ├── fullstack.md            # (DONE)
    ├── go-cli.md               # (DONE — worked --agents-only example)
    ├── go-backend.md           # (DONE)
    ├── iac.md                  # (DONE — Terraform + Ansible)
    ├── claude-cowork.md        # (DONE)
    └── project-management.md   # (DONE — worked skill examples)
```

Every subcommand, flag, flavor, and engine capability is documented — the backlog is empty. When you add a new feature, add its doc in the same change and link it from the layout above.

## Style

- Plain prose. No emojis. No marketing adjectives ("powerful", "elegant", "seamless").
- Lead with what the feature *is*. Then how to use it. Then edge cases / gotchas. Then source pointers.
- Code blocks for shell invocations and config snippets.
- Link to source with `file:line` so the reference survives refactors better than a paraphrase would.

## Skills

Two project-scoped Claude Code skills automate the workflows around this directory:

- `/feature-doc` — bootstraps a new doc entry or refreshes an existing one. Routes the doc to the right subdirectory and updates the layout index above. See [`.claude/skills/feature-doc/SKILL.md`](../.claude/skills/feature-doc/SKILL.md).
- `/add-flavor` — walks the seven-step flavor-authoring checklist, including the `docs/flavors/<name>.md` entry. See [`.claude/skills/add-flavor/SKILL.md`](../.claude/skills/add-flavor/SKILL.md).
