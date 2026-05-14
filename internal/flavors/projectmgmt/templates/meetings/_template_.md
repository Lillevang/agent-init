# Meeting: <topic>

> Filename convention: `YYYY-MM-DD_<topic>.md` (dated, kebab-case topic). Sorts chronologically; the topic is recoverable from the filename alone.

- **Date.** YYYY-MM-DD
- **Attendees.** Comma-separated; mark anyone authoritative with their role. ("Alice (Solutions Architect), Bob (Eng Lead), Customer-PMO-Director")
- **Type.** kickoff | status | decision | working-session | demo | retrospective | other
- **Source.** Where the raw notes came from — paste of chat transcript, hand-written notes, recording link, etc. Useful when re-reading later and the summary loses context.

## Decisions

Lead with these. Bullet list, one line each. If a decision needs more context, mention the full entry now lives in [`decisions.md`](../decisions.md#<anchor>).

- Decision: <one line> → [`decisions.md`](../decisions.md#<anchor>)
- Decision: <one line> → [`decisions.md`](../decisions.md#<anchor>)

If no decisions were made, write "None — see Discussion and Action items".

## Action items

What needs doing, by whom, by when. If an action item maps to a work item in the tracker, link it (or note "to be created via `/sync-tracker`").

- [ ] Owner: <name> — what they'll do — by <date>
- [ ] Owner: <name> — what they'll do — by <date>

## Open questions surfaced

Anything raised that couldn't be answered in the meeting. Each goes in [`open-questions.md`](../open-questions.md) as a separate entry; link the questions here:

- [Q: <short>](../open-questions.md#<anchor>)

## Discussion

Prose. The "why" behind the decisions and action items. Future-you reads this when wondering "why did we agree to X".

Keep it tight — paragraphs, not transcripts. The transcript itself goes in the **Source** field above or under `archive/` if it's a big paste.

## Follow-up meetings

If this meeting triggered the need for another, link it (or note that it needs scheduling).
