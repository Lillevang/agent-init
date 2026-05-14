# Epic: <short-title>

> Filename convention: `<kebab-case-name>.md`. Don't keep `_template_.md` named that — copy it, rename, fill in. Leave `_template_.md` in place as a reference.

## Status

`discovery` | `broken-down` | `in-progress` | `done` | `cancelled`

Plus a one-line note: when the status last changed and why.

## Business value

Why does this epic exist? Who benefits, by how much, measured how? One paragraph. If you can't answer in one paragraph, the epic isn't ready.

## Scope

What's in scope. What's explicitly out of scope. Bullets are fine.

## Acceptance criteria

Concrete, observable conditions for "done". Each one should be testable by someone who isn't on the team.

## Assumptions

Things you're assuming to be true. These are the most important field. The breakdown below depends on these being correct. Whenever you log an assumption, ask whether it should also be in `open-questions.md` so a stakeholder can confirm it.

- Assumption 1
- Assumption 2

## Breakdown

The work below the epic. Use the terminology that matches the active tracker:

- **Jira**: Features → User Stories
- **Azure DevOps**: Features → PBIs
- **GitHub**: Issues (flat or grouped by labels/milestones)

Format the breakdown as a nested list. After `/sync-tracker` runs, each item gains a link to the tracker URL.

- **Feature: <name>**
  - Story/PBI/Issue: <title>
  - Story/PBI/Issue: <title>
- **Feature: <name>**
  - Story/PBI/Issue: <title>

## Risks

Things that could go wrong. Each with a one-line mitigation if you have one.

## Stakeholders

People with skin in this epic. Cross-reference `stakeholders.md` so the authority trail is clear:

- [Alice](../stakeholders.md#alice) — technical authority
- [Carol](../stakeholders.md#carol) — budget authority

## Related

Links to specs, decisions, meeting notes, prior epics.

- Spec: `specs/<file>.md`
- Decisions: `decisions.md#<anchor>`
- Meetings: `meetings/YYYY-MM-DD_<topic>.md`
