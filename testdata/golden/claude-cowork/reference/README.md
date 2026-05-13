# reference/

Source materials read by Claude as context: vendor documentation, exported wikis, contracts, PDFs, large reference `.docx` files.

## What goes here

- **Vendor / external docs** — anything you're consuming but not authoring. Vendor manuals, partner specs, RFC PDFs.
- **Exported wikis** — Azure DevOps / Confluence / GitHub wiki dumps as nested directories of markdown.
- **Reference materials** — handbooks, policy docs, glossaries, anything Claude needs to read to do its job but shouldn't rewrite.

## What does NOT go here

- **Your own drafts.** Put those in a project folder at the workspace root.
- **Templates** (`.potx`, `.dotx`). Those go in `templates/`.
- **Superseded copies.** Those go in `archive/`.

## Conventions

- Group by source or topic. Sub-folders are encouraged when a single source has many files.
- Don't rename files coming from exports (wikis, git dumps). The original names usually encode metadata Claude can use.
- If a file is sensitive (NDA-covered, customer-specific), name the parent folder so coworkers know — e.g. `reference/customer-acme-nda/`.

## Claude's rules for this folder

Claude treats `reference/` as read-only by default. It will quote and summarize but won't edit files here. If you want a file rewritten or condensed, say so explicitly and Claude will produce the new version *outside* `reference/` (typically in a project folder).
