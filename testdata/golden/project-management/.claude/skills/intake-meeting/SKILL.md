---
name: intake-meeting
description: Process raw meeting notes or a transcript into a structured meeting summary under meetings/, with decisions, action items, and open questions extracted. Invoke when the user pastes meeting notes, a transcript, or says "intake this meeting".
---

# intake-meeting

You're being invoked because the user wants to turn raw meeting material (notes, paste of chat transcript, hand-written summary) into a structured artifact this workspace can use. The goal is to leave behind a meeting summary plus the things derived from it: decisions, action items, open questions, stakeholders.

## Inputs you need

If the user didn't already provide them, ask for:

1. **The raw material.** Paste of notes, transcript, or a path to a file already in the workspace (typical for a meeting recording's auto-generated transcript that landed in `meetings/raw/`).
2. **Meeting metadata** — when, who, what type. If the raw material has a date and attendee list, you can pull from that; otherwise ask.

## Steps

### 1. Write the meeting summary

Create `meetings/<YYYY-MM-DD>_<short-topic>.md` following the structure in [`meetings/_template_.md`](../../../meetings/_template_.md). The template's order — Decisions, then Action items, then Open questions, then Discussion — is the order to write in. The reader cares about the conclusions first.

If the meeting was long (say 30+ minutes of transcript), trim aggressively. A meeting summary that's longer than 1.5 pages is a transcript in disguise.

### 2. Decisions

Anything that was definitively decided in the meeting → append to [`decisions.md`](../../../decisions.md) via the `/log-decision` skill (or write the entry directly using the same format). Cross-link from the meeting summary to the decision entry.

If something was *almost* decided but stalled because nobody in the room had authority → don't put it in `decisions.md`. Put it in `open-questions.md` instead, with the missing authority as the blocker.

### 3. Action items

Convert agreed-upon next steps into checklist items with **Owner** and **Due date** (or "no date" if unstated). If an action item maps directly to a tracker work item, note "to be created via `/sync-tracker`" — don't push to the tracker from here.

### 4. Open questions

Anything raised that wasn't answered → append to [`open-questions.md`](../../../open-questions.md) as separate entries. Cross-link from the meeting summary. Be specific about *what* needs to be clarified and *why it matters* — vague entries rot.

### 5. New stakeholders

If the meeting mentions a new person who's been making decisions or has authority over something — flag them. Either invoke `/track-stakeholder` or note at the end of your summary "consider adding <Name> to stakeholders.md — they authorized X in this meeting".

### 6. Archive the raw material (optional)

If the raw notes were a paste (and there's no file source), the meeting summary IS the record. If the user pasted a long transcript, ask whether they want the raw paste preserved under `archive/YYYY-MM-DD_<topic>_raw.md`.

## Style

- Lead with decisions and action items. Discussion comes last.
- Plain prose. No bullet lists of three words each — sentences carry the reasoning.
- Don't quote attendees by name unless attribution matters (e.g., a stakeholder authorized a decision).
- Don't editorialize. "The team disagreed about X" is fine; "the team failed to align on X" is not.

## When to stop and ask

- If the raw material is too thin to summarize (e.g., a couple of one-line bullets), ask the user for more context before writing a summary you'd have to invent.
- If two attendees clearly contradicted each other and the meeting didn't resolve it, surface the disagreement as an open question — don't pick a side.
- If a decision was implied but not stated explicitly, ask which side prevailed before you write it as a decision.
