# claude-cowork

A flavor for document-collaboration workspaces backed by OneDrive (or any cloud sync) and loaded into Claude Cowork. Different from the code flavors: no devcontainer, no `Justfile`, no `.agent/` subdirectory, no symlinks. All agent-facing files live at the workspace root so coworkers see them when they open the folder.

## What it scaffolds

```
<workspace>/
├── AGENTS.md           # canonical agent instructions
├── README.md           # human onboarding
├── decisions.md        # append-only decision log
├── corrections.md      # patterns the agent has gotten wrong
├── .gitignore          # Office lock files (~$*), OS metadata
├── reference/README.md # "source materials, read-only"
├── templates/README.md # ".potx/.dotx/.xltx for new documents"
└── archive/README.md   # "move here instead of deleting"
```

Eight files. No symlinks created — OneDrive (and Dropbox, and most desktop sync clients) mangle symlinks on Windows. Source: [internal/flavors/claudecowork/](../../internal/flavors/claudecowork/).

## Usage

```bash
agent-init init --no-git claude-cowork ~/OneDrive/myteam
```

`--no-git` is strongly recommended — these aren't code repositories. The scaffold honours `--force` and `--dry-run` like the other flavors.

After scaffolding:

1. **Edit `AGENTS.md`** — replace "What this workspace is" with one or two sentences specific to your team.
2. **Create `CLAUDE.md`** (the file Claude Cowork actually reads):
   - Linux/macOS: `ln -s AGENTS.md CLAUDE.md`
   - Windows: `copy AGENTS.md CLAUDE.md` — OneDrive corrupts symlinks on Windows, so keep the two in sync manually.
3. **Drop materials**: vendor docs into `reference/`, document templates into `templates/`.
4. **Share the folder** with coworkers and load it into Claude Cowork.

The post-scaffold message ([flavor.go:25-43](../../internal/flavors/claudecowork/flavor.go#L25-L43)) walks through the same steps with the actual target path filled in.

## What's different from the code flavors

| | code flavors | `claude-cowork` |
|---|---|---|
| `.agent/` subdirectory | yes | no — files at root |
| `AGENTS.md`/`CLAUDE.md` symlinks | yes (`createSymlinks` ships three) | none — host sets it up |
| `Justfile` + done-gate | yes | none |
| `gen-codemap.sh` | yes (via `common/`) | none |
| `CommonTemplates` overlay | yes | omitted |
| Post-scaffold message | "devcontainer up + just check" | edit AGENTS.md, set up CLAUDE.md, share the OneDrive link |

These differences are flavor-controlled, not engine-hardcoded. The relevant engine hooks:

- **Symlinks** are declarative via `Flavor.Symlinks` ([flavor.go:25](../../internal/flavors/flavor.go#L25); type definition at [flavor.go:37-40](../../internal/flavors/flavor.go#L37-L40)); claude-cowork sets it to nil.
- **Next-steps message** is a per-flavor `NextSteps func(target string) string` hook ([flavor.go:26-30](../../internal/flavors/flavor.go#L26-L30)); claude-cowork supplies its own.
- **`CommonTemplates`** is optional; flavors that don't set it skip the overlay.

The full reference for these hooks lives at [`docs/engine/flavor-hooks.md`](../engine/flavor-hooks.md).

If you're adding another non-code flavor (Notion export bundle, Obsidian vault, etc.), copy this flavor's structure rather than fighting the code-flavor defaults.

## Decision log convention

The shipped `decisions.md` includes a format template plus one example entry. The intent: every time a real choice is made (vendor selection, structural change, scope cut), append an entry with **Context**, **Options considered**, **Decision**, **Reasoning**, **Implications**. Append-only — if a decision is reversed, write a new entry referencing the old one.

The example entry tells users to delete it once they've added their first real decision. Source: [decisions.md:41-55](../../internal/flavors/claudecowork/templates/decisions.md#L41-L55).

## Corrections file

[corrections.md](../../internal/flavors/claudecowork/templates/corrections.md) follows the same pattern as the code-flavor `CORRECTIONS.md`: heading + bad example + good example + one-line rationale. Different example (meeting summary format) to match the doc-collab context.

## Edge cases

- **The smoke-test recipe and golden test** need to handle this flavor specially. Both check `AGENTS.md` at root or `.agent/AGENTS.md`, and skip `just check` if there's no `Justfile`. See [Justfile:61-67](../../Justfile#L61-L67) and [test/golden_test.go:66-79](../../test/golden_test.go#L66-L79).
- **`agent-init list-flavors`** shows `claude-cowork` alongside the code flavors. The description explicitly calls out the OneDrive/no-symlinks framing so users picking a flavor see the difference upfront.
- **`gen-codemap.sh`** is not shipped, intentionally. The codemap concept (auto-discover code structure) doesn't apply to a folder of `.docx` files. If you want some kind of "what's in this folder" overview, write it by hand in `README.md` or use `decisions.md`.

## Tests

[test/golden_test.go](../../test/golden_test.go) covers `claude-cowork` in the table-driven `TestFlavorGolden`. The golden snapshot at [testdata/golden/claude-cowork/](../../testdata/golden/claude-cowork/) is byte-compared against a fresh scaffold; regenerate via `just smoke-test-update` after intentional template changes.
