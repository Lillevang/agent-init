# Decision Log

Append-only log of decisions made on this project. Each entry captures the choice, the options considered, the reasoning, the stakeholder who authorized it, and the implications. Future-you (and the next person who joins the project) read this when wondering *why* something is the way it is.

## When to add an entry

A decision warrants an entry when:

- A choice was made between two or more plausible options.
- The choice affects future work — a vendor selected, a structure adopted, a scope cut, a deadline moved.
- The reasoning is not obvious from the resulting state.

Things that don't need an entry: routine implementation choices, format/style preferences, anything that would be re-decided trivially. Things that *do* need an entry: anything you'd want to defend in a stakeholder review meeting six months from now.

## Format

Each entry uses this shape:

```
## YYYY-MM-DD — Short title of the decision

**Context.** One or two sentences: what triggered the choice.

**Options considered.**
- Option A: brief description, who proposed.
- Option B: brief description.

**Decision.** Which option won.

**Authorized by.** Stakeholder name and a link to [`stakeholders.md`](./stakeholders.md#<anchor>).

**Reasoning.** Why. Be specific — "cheaper" alone isn't enough; "cheaper by ~30% because we already license X" is.

**Implications.** What changes downstream. Who needs to know. What follow-up actions land in `open-questions.md` or as work items.

**Related.** Optional. Links to meeting notes, specs, prior decisions.
```

Keep entries chronological. Don't edit old entries — if a decision is reversed, write a new entry that links to the old one in **Related**.

---

<!-- ADD NEW ENTRIES BELOW -->

## YYYY-MM-DD — Example: We adopted Azure DevOps as the work tracker

**Context.** The team had been tracking work in a mix of Jira and spreadsheets. New compliance requirements made traceability between requirements and code essential.

**Options considered.**
- A: Keep Jira (familiar; integrates with our existing dashboards).
- B: Move to Azure DevOps (proposed by Alice; tight VS Code integration; the customer's other vendors use ADO).

**Decision.** B.

**Authorized by.** Bob (Engineering Lead) — see [`stakeholders.md`](./stakeholders.md#bob).

**Reasoning.** Customer's other vendors are already on ADO; cross-vendor work-item linkage is a hard requirement from the customer's PMO. Jira would require maintaining a custom mirroring pipeline.

**Implications.** All existing Jira tickets need migration; estimated two-week one-way migration window. Custom dashboards in Jira need rebuilding in ADO Analytics — see `open-questions.md` for who owns that work.

**Related.** Meeting notes: `meetings/2026-05-08_tracker-decision.md`.

> *Delete this example entry once you've added your first real one.*
