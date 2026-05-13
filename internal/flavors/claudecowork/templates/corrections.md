# Corrections

Patterns Claude has gotten wrong in this workspace, and the preferred form. Read this before starting work and after any review.

> **Maintenance note.** Keep this file under ~30 entries. When it grows beyond that, refactor recurring lessons into `AGENTS.md` and remove them here. An overlong corrections file gets ignored (and rightly so — it stops being useful).

## Format

Each entry: heading + a bad example + a good example + (optional) one-line rationale. Be concrete. "Be more careful" is not an entry.

---

## Example: Don't summarize meetings with bullet-only outputs when prose is asked for

**Bad:**
```
- Decision: adopt X
- Action: Alice to draft proposal
- Action: Bob to review by Friday
```

**Good:**
```
The team decided to adopt X. Alice will draft the proposal; Bob has agreed
to review it by Friday. The decision turned on cost — option Y was 30%
more expensive given our existing licenses.
```

Bullets discard reasoning. When a coworker asks "what did we agree?", a paragraph carries the *why*, which is what they'll want six weeks later.

---

<!-- ADD NEW ENTRIES BELOW -->
