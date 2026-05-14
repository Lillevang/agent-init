---
name: break-down-epic
description: Convert a vague epic into ordered Features and work items (Jira User Stories / ADO PBIs / GitHub Issues) using the active tracker's terminology. Surfaces assumptions and unanswered questions explicitly rather than papering over them. Invoke when the user wants to scope or refine an epic.
---

# break-down-epic

You're being invoked because the user wants to take an epic — usually described in a paragraph and a couple of meeting references — and turn it into something a team can start implementing. The goal is *not* a perfect breakdown. The goal is a breakdown plus a clear list of what's still unknown, so the team can either start work or know exactly what they're waiting on.

## Inputs you need

If the user didn't already provide them, ask for:

1. **The epic.** A paragraph describing what it is, who benefits, what success looks like. Or a pointer to where this lives (meeting notes, spec, an existing rough draft).
2. **The active tracker** — read `AGENTS.md`'s "Active trackers" line. This determines terminology: Jira (Feature → User Story), ADO (Feature → PBI), GitHub (Issues). If multiple trackers are active, ask which one this epic targets.
3. **Constraints, if known** — deadline, team size, hard scope limits.

## Steps

### 1. Read what's already there

Check whether `epics/<short-name>.md` already exists for this work. If yes — open it, this is a refinement task, not a fresh breakdown. Update the existing file; don't create a duplicate.

Also check `decisions.md` for any decisions that constrain the breakdown (chosen tech stack, vendor selections, deadlines committed). Check `open-questions.md` for unresolved blockers that touch this epic.

### 2. Write or update `epics/<short-name>.md`

Use [`epics/_template_.md`](../../../epics/_template_.md) as the shape. Filename is kebab-case, short, descriptive (`customer-dashboard-mvp`, not `epic-1`).

Order the sections:

1. **Business value** — must be concrete. "Reduce time-to-onboard by 50%" beats "Improve onboarding".
2. **Scope** — in and out. The out-of-scope list is more important than it looks; it pre-answers "is X in this?" questions.
3. **Acceptance criteria** — testable conditions for done. Each criterion should be checkable by someone not on the team.
4. **Assumptions** — see below; this is the most important section.
5. **Breakdown** — Features then work items, using tracker-correct terminology.
6. **Risks** — known risks, with mitigations if known.
7. **Stakeholders** — link to [`stakeholders.md`](../../../stakeholders.md).

### 3. Make assumptions explicit

Every breakdown is built on assumptions. List them. For each assumption, decide: is it safe enough to proceed without confirmation, or does it need to go to [`open-questions.md`](../../../open-questions.md) before the team commits?

Examples of assumptions worth surfacing:

- "We can reuse the existing auth library." → If the auth library doesn't fit, the breakdown changes significantly. Open question.
- "The customer is OK with us pulling work from sprint 7 into sprint 8." → Schedule assumption that affects commitments. Open question.
- "All work items will be in English." → Probably safe. Don't surface.

When unsure, surface it. The cost of confirming a non-issue is low; the cost of building on a wrong assumption is high.

### 4. Draft the breakdown

Each Feature should be 2–5 work items. If a feature has 1, it's not really a feature; collapse it into a sibling. If it has 10+, it's another epic; split it.

Each work item (Story / PBI / Issue):

- Title that starts with a verb. "Implement OIDC login flow" beats "OIDC login".
- One-paragraph description.
- Acceptance criteria — observable, not internal.

Don't estimate effort yet. Estimation is a separate conversation; the breakdown is what we'd want regardless.

### 5. Don't push to the tracker

Push is `/sync-tracker`'s job. End your run by either invoking that skill (if the breakdown is ready) or telling the user "ready for review; run `/sync-tracker` when you want to push".

## Style

- Concrete. Avoid management filler ("leverage", "align", "drive synergy").
- Use the tracker's actual terminology in the breakdown. If active tracker is ADO, write "PBI", not "story" or "ticket".
- Don't invent acceptance criteria. If you don't know what success looks like for a work item, the work item needs more thinking before it's ready.

## When to stop and ask

- If the epic is too vague to break down (no business value statement; scope wildly open), stop. Surface the gaps in `open-questions.md` and tell the user the epic isn't ready.
- If the breakdown depends on a tech choice that hasn't been made (database, framework, deployment target), surface the choice as an open question and write the breakdown's affected sections placeholder-style with "depends on the resolution of Q: ...".
- If two existing decisions in `decisions.md` contradict each other in their impact on this epic, surface the conflict — don't pick a side.
