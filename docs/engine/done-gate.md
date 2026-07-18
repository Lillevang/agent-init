# The done-gate (`check.sh`)

Every scaffolded code project ships `.agent/scripts/check.sh` — the agent's "am I done" gate. The contract, encoded in the generated `AGENTS.md`, is: **don't declare a task complete until `./.agent/scripts/check.sh` passes.** Every step must succeed; a failing step means the work is not done, and the agent's job is to fix the failure, not bypass it.

The script is shipped from the common layer ([internal/flavors/common/templates/.agent/scripts/check.sh](../../internal/flavors/common/templates/.agent/scripts/check.sh)), so every code flavor gets the same gate. What actually runs is driven by the flavor's `Justfile`, which is how one script serves flavors as different as `go-cli` and `iac`.

## What it runs

The script runs `set -euo pipefail`, resolves the repo root with `git rev-parse`, then runs, in order:

1. **codemap** — regenerates `.agent/CODEBASE.md` via `gen-codemap.sh`, if that script is present and executable. Always run so the committed map reflects current state.
2. **generate-clients** — OpenAPI client generation (the `fullstack` flavor).
3. **fmt**
4. **lint**
5. **typecheck**
6. **test**
7. **playwright** — end-to-end tests (the `fullstack` flavor).

## Two kinds of step

The behavior that lets one script fit every flavor is the split between hard and optional steps:

| Helper | Behavior |
|--------|----------|
| `step` | Runs the command. On a non-zero exit it prints `✗ <name> failed` and **exits 1** — the gate fails. Used for the codemap regeneration. |
| `maybe_step` | Looks the recipe up in `just --list`. If the flavor's `Justfile` defines it, it runs as a hard `step`. If not, it prints `⊘ skipping '<recipe>' (no Just recipe defined)` and moves on. |

So the gate adapts to the project:

- `go-cli` / `go-backend` define `fmt`, `lint`, `typecheck`, `test` — those run; `generate-clients` and `playwright` are skipped.
- `fullstack` additionally defines `generate-clients` and `playwright` — those run too.
- `iac` defines `fmt` / `lint` / `typecheck` / `test`, each guarded internally with `command -v` so they no-op when the toolchain isn't installed.
- A brand-new empty scaffold skips every `maybe_step` and still passes — the project is installable before any application code exists.

The distinction is important: `maybe_step` skips only when the **recipe is absent**. A recipe that exists and *fails* is a hard failure that stops the gate.

## Example

```
$ ./.agent/scripts/check.sh
Running done-gate checks for my-tool

→ codemap
✓ codemap passed
⊘ skipping 'generate-clients' (no Just recipe defined)

→ fmt
✓ fmt passed

→ lint
✓ lint passed

→ typecheck
✓ typecheck passed

→ test
✓ test passed
⊘ skipping 'playwright' (no Just recipe defined)

✓ all checks passed
```

## The Justfile is the source of truth for steps

`check.sh` never hard-codes tools; it only knows recipe names. The actual commands live in the flavor's `Justfile`, and each flavor doc lists its recipes — see [go-cli](../flavors/go-cli.md#justfile-recipes), [go-backend](../flavors/go-backend.md#justfile-recipes), [fullstack](../flavors/fullstack.md#justfile-recipes), and [iac](../flavors/iac.md#justfile-recipes). To add a gate step for a flavor, add the recipe to its `Justfile`; `check.sh` picks it up automatically if it is one of the names above.

> This repository's *own* `check.sh` (used to develop `agent-init`) is a superset of the shipped one — it adds a soft `vulncheck` step and the scaffold smoke test. Downstream projects get the common gate described here.

## Source

- The shipped gate: [internal/flavors/common/templates/.agent/scripts/check.sh](../../internal/flavors/common/templates/.agent/scripts/check.sh) (`step` and `maybe_step`).
- Codemap regeneration it invokes: [common/templates/.agent/scripts/gen-codemap.sh](../../internal/flavors/common/templates/.agent/scripts/gen-codemap.sh).
- Recipes come from each flavor's `Justfile.tmpl` under `internal/flavors/<flavor>/templates/`.
