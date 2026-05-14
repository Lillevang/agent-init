---
name: track-stakeholder
description: Add or update an entry in stakeholders.md. Captures role, decision authority, communication preferences, and past decisions authorized. Invoke when a new stakeholder surfaces or an existing one's authority becomes clearer.
---

# track-stakeholder

You're being invoked because a stakeholder has surfaced (in a meeting, in a decision, in an email) and needs a profile, or because an existing stakeholder's authority has become clearer and the profile needs updating. The goal is a who-do-I-ask cheatsheet that gets sharper over time.

## Inputs you need

If the user didn't already provide them, ask for:

1. **Name.** Full or "Firstname Lastname"; pick a consistent style across the file.
2. **Role.** Title and which side they're on (vendor / customer / partner / consultant).
3. **Decision authority.** What can they authorize? Be specific. "Budget up to $50k", "anything in the data domain", "scope changes for phase 1 only" — much more useful than "decisions".
4. **Communication.** Email, Slack channel, Teams, timezone. How and when to reach them.
5. **Constraints.** Known vacations, off-days, anything that affects when to engage them.

Most fields can start empty or fuzzy and sharpen over time. Better to log a stub now than to wait for perfect info.

## Steps

### 1. Check whether they already exist

Read [`stakeholders.md`](../../../stakeholders.md). If the person is already listed:

- This is an *update* run. Find their entry (by name or anchor) and update fields.
- Don't replace the **Past decisions** list — append.
- If their decision authority changed (promoted, role shifted, scope of authority widened/narrowed), reflect that and note the change in **Notes** with a date.

If the person is *not* listed, this is a new entry.

### 2. Pick an anchor

The entry header uses `## <Name> {#<anchor>}` so [`decisions.md`](../../../decisions.md) can link via `[Name](./stakeholders.md#anchor)`.

- Default: lowercase first name. `## Alice Chen {#alice}`.
- If two stakeholders share a first name, disambiguate with their side: `{#alice-vendor}`, `{#alice-customer}`.
- Anchors are stable. Don't change an anchor after `decisions.md` references it — broken links rot silently.

### 3. Write or update the entry

Use the format from [`stakeholders.md`](../../../stakeholders.md) (the file's intro shows the shape). Order:

1. **Role.** Their title + side.
2. **Decision authority.** What they can authorize.
3. **Communication.** Where to reach them; timezone.
4. **Constraints.** Vacations, language preference, communication style.
5. **Past decisions.** Cross-referenced to `decisions.md`. Initially empty; grows over time.
6. **Notes.** Free-form. Style preferences, who they delegate to, what they care about, etc.

### 4. Cross-link the latest decision

If `/log-decision` triggered this run (i.e., a decision was just recorded and the authorizing stakeholder didn't have a profile yet), add the new decision to **Past decisions**. Link goes both ways:

- `decisions.md` entry links to the stakeholder anchor.
- `stakeholders.md` entry's **Past decisions** section links to the decision anchor.

### 5. Don't speculate about authority

If you're not sure what a person can authorize, ask the user or leave the field with a placeholder like *"to be confirmed — Alice mentioned X but full scope unclear"*. Don't invent authority. A stakeholder profile that overstates someone's authority is worse than no profile — agents will route decisions to the wrong place.

## Style

- One paragraph per field, max. The file should remain skim-able.
- Be concrete in **Decision authority**. Vague entries don't help future agents route decisions.
- **Notes** is where personality goes. "Prefers async written discussion; will push back hard on scope creep — bring data" is the kind of thing that makes the next agent's life easier.

## When to stop and ask

- If you don't know the person's side (vendor/customer/partner) and the meeting didn't make it obvious, ask.
- If decision authority is unclear and the user doesn't know either, leave a placeholder and surface as an open question in [`open-questions.md`](../../../open-questions.md).
- If the user asks you to *demote* a stakeholder (their authority was narrower than the entry says), confirm the change and add a dated note in **Notes** — don't silently rewrite history.
