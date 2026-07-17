# Path templating

The scaffold engine runs every file's **destination path** through Go's `text/template`, regardless of whether the file's content is templated. This lets a directory or file name reflect the project name — most importantly `cmd/{{.ProjectName}}/main.go`, which renders to `cmd/my-tool/main.go` when scaffolding into `./my-tool`.

Path templating uses the same one-field data model as [content templating](./templates.md): `{{.ProjectName}}` is the target directory's base name.

## How it works

`renderPath` ([scaffold.go:320-333](../../internal/scaffold/scaffold.go#L320-L333)) renders a relative path string:

- If the path contains no `{{`, it is returned unchanged (a fast path that skips the template engine for the common case).
- Otherwise the path is parsed and executed against the scaffold data, and the result is the on-disk destination.

It is called on every walked file after the `.tmpl` suffix is stripped ([walkLayer:209](../../internal/scaffold/scaffold.go#L209)), and also on the flavor's `FreshOnlyPaths` ([scaffold.go:246](../../internal/scaffold/scaffold.go#L246)) and `.agents-only` variant base names ([scaffold.go:181](../../internal/scaffold/scaffold.go#L181)) so those comparisons match the rendered layout.

## Example

The `go-cli` flavor ships its entry point as:

```
internal/flavors/gocli/templates/cmd/{{.ProjectName}}/main.go.tmpl
```

Scaffolded into `./my-tool`, the path renders and the `.tmpl` suffix drops, producing:

```
cmd/my-tool/main.go
```

Here **both** kinds of templating apply: the path is rendered (`{{.ProjectName}}` → `my-tool`) and, because the source ends in `.tmpl`, the content is rendered too.

## The `.tmpl` gotcha for path-templated directories

A file that lives under a `{{.ProjectName}}` directory should carry a `.tmpl` extension **even if its content needs no substitution**. The reason is Go tooling, not the scaffold engine: a real `cmd/{{.ProjectName}}/main.go` in this repository's template tree makes `go build ./...` fail, because the Go toolchain tries to read the literal `{` as a package path.

Adding `.tmpl` fixes it two ways at once:

- Go tooling ignores non-`.go` files, so `cmd/{{.ProjectName}}/main.go.tmpl` no longer breaks the build of `agent-init` itself.
- `text/template` parses the file (a no-op when there is nothing to substitute) and the suffix is stripped on write.

So path-templated Go sources are always shipped as `.tmpl`. This is why `go-cli`'s entry point is `main.go.tmpl` rather than `main.go`.

## Relationship to content templating

Path templating always runs; [content templating](./templates.md) only runs for `.tmpl` files. The two are resolved separately during the walk — the path is rendered, then the content is rendered if the suffix calls for it.

## Source

- Path render + the no-`{{` fast path: [scaffold.go:320-333](../../internal/scaffold/scaffold.go#L320-L333) (`renderPath`).
- Where it is applied during the walk: [scaffold.go:209](../../internal/scaffold/scaffold.go#L209), and for `FreshOnlyPaths` / variant matching at [scaffold.go:246](../../internal/scaffold/scaffold.go#L246) and [scaffold.go:181](../../internal/scaffold/scaffold.go#L181).
