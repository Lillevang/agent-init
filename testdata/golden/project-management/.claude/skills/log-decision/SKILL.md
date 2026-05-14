---
name: log-decision
description: Append a structured decision entry to decisions.md. The entry captures context, options considered, the choice, who authorized it, and the implications. Invoke after any non-trivial decision is made, or when /intake-meeting surfaces one.
---

# log-decision

You're being invoked because a decision has been made (or is about to be) and needs to land in the project's decision log. The log is append-only; the value is in being able to read it six months from now and understand *why* something is the way it is.

## Inputs you need

If the user didn't already provide them, ask for:

1. **What was decided.** One sentence is fine if the decision is clear.
2. **Authorizing stakeholder.** Who in [`stakeholders.md`](../../../stakeholders.md) had the authority to make this call? If the answer is "I'm not sure", the decision isn't ready — surface it as an open question instead.
3. **Options considered.** What else was on the table.
4. **Reasoning.** Why this option won.
5. **Implications.** What changes downstream as a result.

If any of these are missing, ask before writing. A decision entry missing **Authorized by** or **Reasoning** is worse than no entry — it looks documented but isn't actually defensible.

## Steps

### 1. Verify the decision is real

If the user says "we're leaning toward X" or "we'll probably go with Y", that's not a decision. It belongs in [`open-questions.md`](../../../open-questions.md) as "leaning toward X — to be confirmed by Z". Tell the user, point them at `/intake-meeting` or `open-questions.md`, and stop.

A decision is recordable when:

- A choice has been made.
- The authority to make that choice was present (either in the room, or by delegation traceable in `stakeholders.md`).
- The decision will affect future work.

### 2. Verify the stakeholder

Check [`stakeholders.md`](../../../stakeholders.md) for the authorizing person. If they're not listed, run `/track-stakeholder` first (or write a stub entry inline) — the decision log shouldn't reference unknown authorities.

### 3. Write the entry

Append to [`decisions.md`](../../../decisions.md), below the `<!-- ADD NEW ENTRIES BELOW -->` marker, in the format already defined in the file:

```
## YYYY-MM-DD — Short title

**Context.** ...
**Options considered.** ...
**Decision.** ...
**Authorized by.** [<Name>](./stakeholders.md#<anchor>)
**Reasoning.** ...
**Implications.** ...
**Related.** (optional)
```

Date is today's date in YYYY-MM-DD. Title is short and descriptive ("Adopted Azure DevOps as work tracker"), not vague ("Tracker decision"). Anchor in **Authorized by** must match a real entry in `stakeholders.md`.

### 4. Cross-reference

Decisions rarely happen in isolation. Update:

- **The triggering artifact.** If this decision came out of a meeting, link the decision from `meetings/<file>.md`. If from a spec, link from `specs/<file>.md`. Two-way links survive refactors.
- **`stakeholders.md`.** Add the decision to the **Past decisions** section of the authorizing stakeholder's entry.
- **`open-questions.md`.** If this decision answers an open question, mark that question `answered` and link to this decision.

### 5. Don't edit older entries

The log is append-only. If you've recorded a decision that's later reversed:

- Don't edit the old entry.
- Write a new entry with today's date, link to the old entry in **Related**, and explain in **Reasoning** what changed and why.

The old entry stays. The chain of reasoning is the audit trail.

## Style

- Lead **Reasoning** with the deciding factor. "Cheaper" alone isn't a reason — "cheaper by ~30% because we already license X" is.
- **Implications** is for the *consequences* not the *plan*. "We need to migrate 200 existing tickets" is an implication. "Alice will do the migration by Friday" is an action item, belongs in a meeting summary, not here.
- No emojis. No marketing language. No "AI-assisted decision-making" or "synergy".

## When to stop and ask

- If the user can't name the authorizing stakeholder confidently, the decision isn't ready. Don't write the entry.
- If the user describes implications that sound speculative ("we might need to..."), push for certainty or carve those out as separate open questions.
- If the decision is reversing a recent prior decision (less than a few weeks old), make sure the user knows — sometimes that's intentional, sometimes the prior decision was forgotten.
