# Content templating (`.tmpl`)

The scaffold engine substitutes template variables into file **content** only for files whose name ends in `.tmpl`. Every other file is copied byte-for-byte. Renaming a file to `.tmpl` is the explicit opt-in to content substitution; verbatim copy is the safe default.

This opt-in exists because many config files legitimately contain `{{ }}` — Helm charts, Ansible playbooks, GitHub Actions expressions. If the engine templated everything, those files would break or need every brace escaped. Keeping content templating behind the `.tmpl` suffix means those files ship unchanged.

## How it works

`render` ([scaffold.go:305-318](../../internal/scaffold/scaffold.go#L305-L318)) is the whole mechanism:

- If the source path does **not** end in `.tmpl`, the bytes are returned as-is.
- If it does, the content is parsed and executed with Go's [`text/template`](https://pkg.go.dev/text/template) against the scaffold's data.

The `.tmpl` suffix is stripped from the destination path ([walkLayer:201](../../internal/scaffold/scaffold.go#L201)), so `Justfile.tmpl` is written as `Justfile` and `README.agent.md.tmpl` as `README.agent.md`.

## The data model

Templates receive a single field, `templateData` ([scaffold.go:58-60](../../internal/scaffold/scaffold.go#L58-L60)):

| Field | Value |
|-------|-------|
| `{{.ProjectName}}` | The target directory's base name (`filepath.Base(target)`, [scaffold.go:75](../../internal/scaffold/scaffold.go#L75)). Scaffolding into `~/repos/my-tool` yields `my-tool`. |

That is the entire surface. There are no conditionals, ranges, or helper functions used in the shipped templates today; a `.tmpl` file is almost always a plain file with `{{.ProjectName}}` in a comment header or an example identifier.

## Example

`internal/flavors/gocli/templates/README.agent.md.tmpl`:

```markdown
# {{.ProjectName}} — agent notes
```

Scaffolded into `./my-tool`, this renders to `README.agent.md`:

```markdown
# my-tool — agent notes
```

## Escaping a literal `{{` inside a `.tmpl` file

If a `.tmpl` file needs a literal `{{ }}` in its output — for example prose in the `iac` flavor's `AGENTS.md.tmpl` that talks about Ansible's Jinja syntax — escape it with a template action that emits the braces:

```gotemplate
The one literal reference is written as {{"{{"}} var {{"}}"}}.
```

When a whole file is full of native `{{ }}` (an Ansible `site.yml`, a Helm template), don't fight the escaping — leave the file **without** a `.tmpl` extension so it is copied verbatim. The `iac` flavor does exactly this for its playbooks and inventory; see [docs/flavors/iac.md](../flavors/iac.md#template-files-and-the-jinja-gotcha).

## Relationship to path templating

Content templating (this page) and [path templating](./path-templating.md) are independent. A file's **path** is always run through the template engine; its **content** is only rendered when the name ends in `.tmpl`. A file can have a templated path and verbatim content, or the reverse.

## Source

- Content render + the `.tmpl` gate: [scaffold.go:305-318](../../internal/scaffold/scaffold.go#L305-L318) (`render`).
- Suffix strip on the destination path: [scaffold.go:201](../../internal/scaffold/scaffold.go#L201).
- The data type: [scaffold.go:58-60](../../internal/scaffold/scaffold.go#L58-L60) (`templateData`).
