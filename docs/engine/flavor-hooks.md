# Flavor hooks

Per-flavor customization points on the `Flavor` struct that let one engine support both code-project scaffolds and non-code shapes (doc collaboration, IaC, etc.). The hooks exist because the scaffold engine used to hardcode things like "always create the `AGENTS.md` symlink trio" and "always print the devcontainer-up next-steps message" — those assumptions broke as soon as `claude-cowork` arrived. Each hook is optional; flavors that don't set it get the engine's default code-project behavior.

## The hooks

| Field on `Flavor` | Type | Default when nil/empty | Source |
|---|---|---|---|
| `Symlinks` | `[]Symlink` | No symlinks created | [flavor.go:25](../../internal/flavors/flavor.go#L25) (field); [flavor.go:55-58](../../internal/flavors/flavor.go#L55-L58) (type) |
| `NextSteps` | `func(target string) string` | Default code-project message (devcontainer + just check) | [flavor.go:30](../../internal/flavors/flavor.go#L30) |
| `CommonTemplates` | `fs.FS` | No common-overlay layer walked | [flavor.go:22](../../internal/flavors/flavor.go#L22) |

## Symlinks

```go
type Symlink struct {
    Path   string // relative to the scaffold target
    Target string // symlink destination, written verbatim
}
```

`createSymlinks` ([scaffold.go:335](../../internal/scaffold/scaffold.go#L335)) iterates `opts.Flavor.Symlinks` and creates each in order. `Target` is written verbatim — relative-path conventions like `.agent/AGENTS.md` survive into the scaffolded tree. The engine creates parent directories as needed but won't replace existing directories with symlinks (returns an error instead).

**Code-flavor convention.** All four code flavors (`fullstack`, `go-cli`, `go-backend`, `iac`) share `codeFlavorSymlinks()` in [registry.go:145-151](../../internal/flavors/registry.go#L145-L151):

```go
{Path: "AGENTS.md", Target: ".agent/AGENTS.md"},
{Path: "CLAUDE.md", Target: ".agent/CLAUDE.md"},
{Path: ".agent/CLAUDE.md", Target: "AGENTS.md"},
```

Four entry points (root `AGENTS.md`, root `CLAUDE.md`, `.agent/AGENTS.md`, `.agent/CLAUDE.md`) resolve to one underlying file. Codex finds `AGENTS.md` at the root; Claude Code finds `CLAUDE.md` at the root; both end up at `.agent/AGENTS.md` (the real file).

**When to omit.** Skip symlinks when the target filesystem doesn't preserve them reliably. The `claude-cowork` flavor leaves `Symlinks` nil because OneDrive (and other desktop sync clients) mangle symlinks on Windows. The flavor's `NextSteps` then explains how the user creates `CLAUDE.md` on their host (`ln -s` on Linux, `copy` on Windows).

## NextSteps

```go
NextSteps func(target string) string
```

If set, the engine calls `flavor.NextSteps(target)` after writing and prints the returned string verbatim (`printNextSteps`, [scaffold.go:470](../../internal/scaffold/scaffold.go#L470)). If nil, the engine prints its default code-project message — devcontainer up, devcontainer exec, just check.

The signature takes the scaffold target so flavors can interpolate paths into shell commands the user is about to copy:

```go
// internal/flavors/claudecowork/flavor.go
func NextSteps(target string) string {
    return fmt.Sprintf(`
Done.

Next steps:
  1. Edit %s/AGENTS.md ...
  2. Create CLAUDE.md alongside AGENTS.md (run inside %s):
       - Linux/macOS:  ln -s AGENTS.md CLAUDE.md
       - Windows:      copy AGENTS.md CLAUDE.md
...
`, target, target, ...)
}
```

Keep it short and actionable. Numbered steps with concrete commands are easier than prose; the user is about to act on it.

## CommonTemplates

```go
CommonTemplates fs.FS
CommonRoot      string
```

Optional fallback layer. If set, the scaffold engine walks `CommonTemplates` after the flavor's own `Templates`, claiming any relative path the flavor didn't already produce. Code flavors all set `CommonTemplates: common.Templates()` so they share `.agent/scripts/check.sh`, `gen-codemap.sh`, and `review.sh`. Non-code flavors leave it nil — `claude-cowork` doesn't want `check.sh` or `gen-codemap.sh` in a document folder.

See [common-overlay.md](./common-overlay.md) for the layering semantics and conflict resolution.

## Adding a non-code flavor

When the flavor breaks code-project assumptions:

1. Leave `Symlinks` nil (or list only the symlinks your filesystem can preserve).
2. Provide `NextSteps` — the default message will mention devcontainer and `just check`, which won't make sense.
3. Omit `CommonTemplates` if the flavor doesn't want `.agent/scripts/` ⇒ doc-collab flavors usually don't.
4. The smoke-test recipe ([Justfile:60-65](../../Justfile#L60-L65)) and golden test ([test/golden_test.go:64-73](../../test/golden_test.go#L64-L73)) already handle flavors without `.agent/scripts/` or `Justfile` — no changes needed there.

## Tests

The hooks themselves aren't unit-tested directly. They're exercised by:

- `TestRunLayersFlavorOverCommon` ([scaffold_test.go:183](../../internal/scaffold/scaffold_test.go#L183)) covers the `CommonTemplates` overlay semantics.
- `TestFlavorGolden` ([test/golden_test.go:15](../../test/golden_test.go#L15)) covers every registered flavor end-to-end, which exercises the symlink list and `NextSteps` produces a byte-identical output relative to the golden.
