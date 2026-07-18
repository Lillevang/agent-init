# Common overlay

Code flavors don't carry their own copy of the shared scaffold files (`check.sh`, `gen-codemap.sh`, `review.sh`, the `.pre-commit-config.yaml`, and so on). Those live once in `internal/flavors/common/templates/` and are layered under every code flavor at scaffold time. A flavor ships only what is *unique* to it; the engine fills in the rest from common.

This keeps shared behavior in one place. To change something for every code flavor, edit the file in `common/`. To change it for a single flavor, ship that flavor's own copy at the same relative path — the flavor's copy wins.

## How the layering works

`writeTemplates` ([scaffold.go:138-161](../../internal/scaffold/scaffold.go#L138-L161)) builds an ordered list of layers and walks them in sequence:

1. The flavor's own `Templates` (at `TemplateRoot`).
2. The flavor's `CommonTemplates` (at `CommonRoot`), **only if** the flavor sets it.

A single `claimed` set — `map[string]bool` — is shared across both walks. The first layer to produce a given destination path claims it; any later layer that would write the same path is skipped ([walkLayer:221-224](../../internal/scaffold/scaffold.go#L221-L224)). Because the flavor layer is walked first, **the flavor overrides common** on any path collision, with no special-casing.

```
flavor layer   →  claims cmd/main.go, its own Justfile.tmpl, README.agent.md.tmpl
common layer   →  fills in .agent/scripts/check.sh, gen-codemap.sh, review.sh, ...
                  skips anything the flavor already claimed
```

## Opting in and out

The overlay is driven by two `Flavor` fields (see [flavor-hooks.md](./flavor-hooks.md#commontemplates)):

- **Code flavors** set `CommonTemplates: common.Templates()` ([common/flavor.go:13](../../internal/flavors/common/flavor.go#L13)), so they share the `.agent/scripts/` tooling and other common files.
- **Doc-collab flavors** (`claude-cowork`, `project-management`) leave `CommonTemplates` nil. They get no overlay — a document folder has no use for `check.sh` or `gen-codemap.sh`. When the field is nil the second layer is simply not appended ([scaffold.go:146](../../internal/scaffold/scaffold.go#L146)), and the walk covers the flavor's templates alone.

## Don't copy common files into a flavor

If a shared file needs to change for everyone, change it in `common/` — don't copy it into a flavor "to be safe." A flavor should contain only the paths where it genuinely diverges from common. An unnecessary copy silently pins that flavor to a stale version of a file the rest of the tree keeps evolving.

## `Overlay`: a single-layer writer for subcommands

`Overlay` ([scaffold.go:102-113](../../internal/scaffold/scaffold.go#L102-L113)) is a public entry point that walks **one** template layer onto an already-scaffolded target, using the same write/skip/dry-run semantics but without symlinks, `git init`, or a next-steps message. `add-tracker` uses it to merge a tracker's `integrations/<tracker>/` files into an existing `project-management` workspace. It is the incremental counterpart to the full `Run` scaffold.

## Source

- Layer assembly and the shared `claimed` set: [scaffold.go:138-161](../../internal/scaffold/scaffold.go#L138-L161) (`writeTemplates`).
- Per-layer walk and the skip-if-claimed rule: [scaffold.go:163-235](../../internal/scaffold/scaffold.go#L163-L235) (`walkLayer`), collision skip at [221-224](../../internal/scaffold/scaffold.go#L221-L224).
- Single-layer overlay for subcommands: [scaffold.go:102-113](../../internal/scaffold/scaffold.go#L102-L113) (`Overlay`).
- The common layer itself: [internal/flavors/common/flavor.go](../../internal/flavors/common/flavor.go) (`Templates`).
