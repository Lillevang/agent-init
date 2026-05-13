# Decision Log

Append-only log of decisions made in this workspace. Each entry captures the choice, the options considered, and the reasoning. Future coworkers (and future-you) read this when wondering *why* something is the way it is.

## When to add an entry

A decision warrants an entry when:

- A choice was made between two or more plausible options.
- The choice affects future work — a structure adopted, a vendor selected, a scope cut.
- The reasoning is non-obvious from the resulting state alone.

Internal back-and-forth ("we picked the blue color over red because it looked nicer") doesn't need an entry. Decisions that will be re-litigated in three months do.

## Format

Each entry uses this shape:

```
## YYYY-MM-DD — Short title of the decision

**Context.** One or two sentences: what triggered the choice.

**Options considered.**
- Option A: brief description, who proposed.
- Option B: brief description.

**Decision.** Which option won.

**Reasoning.** Why. Be specific — "cheaper" alone isn't enough; "cheaper by ~30% because we already license X" is.

**Implications.** What changes downstream. Who needs to know.
```

Keep entries chronological. Don't edit old entries — if a decision is reversed, write a new entry referencing the old one.

---

<!-- ADD NEW ENTRIES BELOW -->

## YYYY-MM-DD — Example: We adopted the agent-init claude-cowork scaffold for this folder

**Context.** This workspace previously had no formal structure. New coworkers had no way to know what belonged where, and Claude's behavior drifted across sessions.

**Options considered.**
- A: Free-form folder structure with a single README.
- B: agent-init `claude-cowork` scaffold (this one) — fixed shape for `reference/`, `templates/`, `archive/`, plus AGENTS.md + decisions.md + corrections.md.

**Decision.** B.

**Reasoning.** The scaffold encodes the conventions that were previously tribal knowledge. New coworkers can self-onboard via `README.md`. Claude has consistent instructions across sessions via `AGENTS.md`.

**Implications.** Existing materials need to be sorted into `reference/` and project folders. One-time migration cost; ongoing benefit.

> *Delete this example entry once you've added your first real one.*
