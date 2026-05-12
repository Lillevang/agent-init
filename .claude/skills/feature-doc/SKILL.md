---
name: feature-doc
description: Add or refresh a per-feature doc under ./docs/. Use when the user wants to document a feature, drain an item from the docs/README.md backlog, or verify an existing doc still matches its source. Routes the doc to docs/cli.md, docs/engine/<topic>.md, or docs/flavors/<name>.md and updates the backlog.
---

# feature-doc

You're being invoked because the user wants to add or refresh a per-feature doc under `./docs/`. The convention is defined in [`.agent/AGENTS.md`](../../../.agent/AGENTS.md#documentation) and the current backlog lives in [`docs/README.md`](../../../docs/README.md).

## Inputs you need

If the user named a specific feature, work on that. Otherwise read the backlog in `docs/README.md` and either pick the most urgent or ask the user which one to drain.

If the user is *refreshing* an existing doc, treat the on-disk doc as suspect — read source first, then reconcile.

## Identify the destination

| Feature type | Lives at |
|--------------|----------|
| CLI subcommand or flag | `docs/cli.md` (one file covering them all, sectioned by subcommand) |
| Engine capability (templates, path-templating, common-overlay, done-gate, scaffold semantics) | `docs/engine/<topic>.md` |
| Flavor | `docs/flavors/<name>.md` |

If the feature doesn't fit any of these, stop and ask the user where it belongs before writing — getting the layout wrong now means moving files later.

## Read the source

Every backlog entry in `docs/README.md` points at a source file with a `file:line` link. Open it. Read the actual code. Skim adjacent tests — they often tell you about edge cases the source alone won't reveal.

Do not write a doc from memory of what you think the feature does. If you can't find the source, ask.

## Write the doc

Structure each entry:

1. **What it is** — one paragraph, concrete. Define the feature in user terms, not implementation terms. "`--force` overwrites existing files instead of skipping them" beats "`--force` is a boolean flag that controls overwrite behavior".
2. **How to use it** — minimum example first as a code block, then complex example if there is one. Show real invocations, not pseudocode.
3. **Edge cases and gotchas** — only the ones a user could hit. Internal-to-agent-init gotchas (devcontainer locale issues, embed.FS path quirks) belong in `.agent/AGENTS.md`, not here.
4. **Source** — one or two `file:line` links to the canonical implementation. These survive refactors better than paraphrased descriptions do.

Style rules from [`docs/README.md`](../../../docs/README.md):

- One page or less.
- No emojis. No marketing adjectives (*powerful, elegant, seamless, robust, comprehensive*).
- Short sentences. Plain prose. "Foo does X" beats "Foo is a tool that performs X".
- Code over prose where they convey the same information.
- Link to source with `file:line` syntax so the reference survives refactors.

## Update the backlog

In `docs/README.md`'s "Backlog" section, find the matching TODO and delete the line. Don't tick a box and leave the line — that just makes the backlog longer over time. The backlog should shrink as you work.

If your work uncovered new doc-worthy follow-ups (a related feature you noticed lacks docs, a sub-behavior worth its own entry), append fresh TODOs at the bottom of the backlog.

## Verify

```bash
./.agent/scripts/check.sh
```

The done-gate doesn't lint docs, but it confirms you didn't break the build along the way.

Then read your new doc top to bottom as if you've never seen the project. If anything reads like a press release, or like the writing version of throat-clearing, rewrite. If you can't picture the user it's helping, the doc isn't doing its job yet.

## When to stop and ask

- If the source contradicts an existing doc and the user is the only one who knows which is correct, ask before rewriting.
- If the feature is half-implemented (PR open, behind a flag, etc.), ask whether to document current behavior or skip until it lands.
- If documenting one feature reveals a design issue (e.g. the CLI surface is more confusing than expected), surface that — don't paper over it in the doc.
