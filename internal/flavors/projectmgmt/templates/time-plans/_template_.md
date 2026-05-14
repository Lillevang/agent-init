# Time plan: <phase or release name>

> Filename convention: `<short-name>.md` for the live plan; `archive/YYYY-MM-DD_<short-name>.md` for snapshots taken when the plan changed materially.

## Status

`draft` | `committed` | `slipped` | `delivered` | `superseded`

If the plan is `slipped`, add a one-line note: what slipped and by how much.

## Scope

What's in this plan. What's out (link to other time-plans where relevant).

## Key dates

| Date | Milestone | Owner | Status |
|------|-----------|-------|--------|
| YYYY-MM-DD | <Milestone> | [<Name>](../stakeholders.md#<anchor>) | pending / in-progress / done / slipped |
| YYYY-MM-DD | <Milestone> | [<Name>](../stakeholders.md#<anchor>) | pending / in-progress / done / slipped |

Keep dates absolute (YYYY-MM-DD) not relative ("next week"). Relative dates rot.

## Epic-to-date mapping

Which epics feed which milestones. The link goes both ways: this file links to `epics/`, each epic links here from its **Related** section.

- Milestone <Name> (YYYY-MM-DD)
  - [`epics/<name-1>.md`](../epics/<name-1>.md)
  - [`epics/<name-2>.md`](../epics/<name-2>.md)

## Critical path

The chain of work items that drives the delivery date. If any of these slips, the milestone slips. Identify them so you know where to focus when things drift.

1. <Item> — blocks: <next item>
2. <Item> — blocks: <next item>

## Assumptions

The plan assumes these to be true. If any becomes false, the plan needs revision. Mirror to `open-questions.md` if any assumption is uncertain.

- Assumption 1
- Assumption 2

## Risks to the plan

Known risks with their likelihood and impact on the date. Each with a mitigation if one exists.

| Risk | Likelihood | Impact on date | Mitigation |
|------|------------|----------------|------------|
| <Risk> | low / med / high | <number of days> | <Mitigation> |

## Change history

When the plan changed materially (a milestone moved more than a few days, scope shifted, owner changed), log it here. Don't rewrite the dates above silently — paper trail matters.

- YYYY-MM-DD: <what changed> — see [`decisions.md`](../decisions.md#<anchor>).
