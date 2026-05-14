# Open Questions

Things that need clarification before the work can progress. Each entry has an owner (the person who can answer or who's chasing the answer), a date raised, and a status.

## When to add an entry

- An epic breakdown surfaced an assumption that needs confirmation.
- A meeting raised a question that wasn't answered in the meeting.
- A spec is ambiguous and the ambiguity blocks implementation.
- A decision was almost made but stalled because nobody in the room could authorize it.

## Format

Each entry:

```
## <Short title — the question itself, ideally>

- **Raised.** YYYY-MM-DD
- **Owner.** Who is chasing the answer (not the same as who can answer).
- **Status.** open | in-progress | answered | abandoned
- **Blocks.** What work is blocked by this — link to an epic, a PBI, a spec.
- **Question.** One paragraph describing what needs to be clarified, why it matters, and what would change based on the answer.
- **Notes.** Optional running log of who's been asked, what they said, where the conversation went.

When answered: change status to `answered`, add the answer below the question, link to the matching entry in `decisions.md` if a decision flowed from it. Don't delete answered questions — they're context for future agents.
```

---

<!-- ADD NEW ENTRIES BELOW -->

## Example: Which auth provider should the customer use for the new dashboard?

- **Raised.** 2026-05-12
- **Owner.** Alice (Solutions Architect)
- **Status.** open
- **Blocks.** [`epics/customer-dashboard-mvp.md`](./epics/customer-dashboard-mvp.md) — auth-related stories can't be broken down without this.
- **Question.** The customer's other systems use a mix of Entra ID and Okta. The dashboard could federate either, or both via OIDC. Picking one affects the SSO setup, the per-user license cost, and the timeline (Entra is in place today; Okta would add a 2-week procurement cycle). The customer hasn't named a preferred provider — the question is who at the customer side can authorize this.
- **Notes.**
  - 2026-05-13: Alice asked customer-pm@acme.example; awaiting response.

> *Delete this example entry once you've added your first real one.*
