# Spec: <short-title>

> Filename convention: `<kebab-case-topic>.md`. For versioned specs, suffix the version: `dashboard-auth-v2.md`. Move the older version to `archive/` with the original date prefix.

## Status

`draft` | `review` | `approved` | `superseded`

If superseded, link to the spec that replaced this one.

## Authors and reviewers

- Author(s): names + roles.
- Required reviewers: who must sign off before status moves to `approved`.
- Approved by: filled in when status becomes `approved`. Cross-reference `stakeholders.md`.

## Summary

Two to four sentences: what this spec covers and what it doesn't. The first sentence should make sense in isolation, e.g. linked from an epic or a decision.

## Context

Why are we writing this spec? What problem are we solving? Reference the epic, the prior decisions, the meetings that surfaced the need. One to two paragraphs.

## Requirements

The functional and non-functional requirements. Numbered so they can be referenced from work items, decisions, and acceptance criteria.

### Functional

1. **R-F1.** <Requirement statement.>
2. **R-F2.** <Requirement statement.>

### Non-functional

1. **R-N1.** Performance: <target>.
2. **R-N2.** Compliance: <constraint>.
3. **R-N3.** Operability: <constraint>.

## Design

The proposed approach. Diagrams in mermaid or linked images. Trade-offs called out. Alternatives considered (briefly — full alternative analyses go in their own spec or in `decisions.md`).

## Open questions

Things this spec depends on that aren't yet decided. Each goes in [`open-questions.md`](../open-questions.md) as well; link them:

- [Q: <short>](../open-questions.md#<anchor>)

## Acceptance criteria

What does "spec done" look like? When can we move status to `approved`?

## Related

- Epic: `epics/<file>.md`
- Decisions: `decisions.md#<anchor>`
- Prior spec (superseded): `archive/YYYY-MM-DD_<file>.md`
