# Stakeholders

Who can make what kinds of decisions on this project. Built up over time as the project unfolds. Future agents (and future humans joining the project) read this when wondering "who do I ask about X?".

## How to use this file

- One section per person, ordered by first appearance on the project.
- When a decision is recorded in [`decisions.md`](./decisions.md), cross-reference the stakeholder who authorized it.
- When a stakeholder is mentioned in a meeting or surfaces in an `/intake-meeting` run, add them here.
- Update **Decision authority** as you learn what each person actually controls — don't guess; let observation refine it over time.

## Format

Each entry:

```
## <Name> {#<lowercase-firstname-or-slug>}

- **Role.** Their title or function on this project. ("Engineering Lead at Vendor X", "PMO Director at Customer Acme".)
- **Decision authority.** What they can authorize. Be specific — "vendor selection up to $50k", "anything in the data domain", "scope changes for phase 1 only".
- **Communication.** Where to reach them — email, Slack, Teams. Include timezone if relevant.
- **Constraints.** Holiday schedules, "out until X", language preferences, anything that affects how/when to engage them.
- **Past decisions.** Decisions they authorized, linked to [`decisions.md`](./decisions.md). Built up over time; usually empty at first.
- **Notes.** Anything else useful — communication style, who they delegate to, what they care about.
```

The `{#anchor}` lets [`decisions.md`](./decisions.md) link directly to a stakeholder via `[Bob](./stakeholders.md#bob)`. Use the person's first name as the anchor by default; if multiple people share a first name, disambiguate (e.g., `bob-customer`, `bob-vendor`).

---

<!-- ADD NEW ENTRIES BELOW -->

## Alice Chen {#alice}

- **Role.** Solutions Architect, our side. Owns end-to-end architecture for the dashboard project.
- **Decision authority.** Technical architecture choices (frameworks, infrastructure patterns, integration patterns). Can authorize work-item priority shuffles within the architecture domain. Cannot authorize scope changes that affect customer commitments.
- **Communication.** alice@vendor.example. Based in Copenhagen (CET). Slack: @alice in #project-acme.
- **Constraints.** Out 2026-06-15 through 2026-06-30 (vacation).
- **Past decisions.**
  - 2026-05-12: chose Entra ID over Okta for SSO — [`decisions.md`](./decisions.md#2026-05-12-sso).
- **Notes.** Prefers async written discussion over synchronous calls. Will push back hard on scope creep — bring data when you propose changes.

## Bob Larsen {#bob}

- **Role.** Engineering Lead, our side. Manages the implementation team and owns delivery commitments.
- **Decision authority.** Anything inside the delivery domain: team capacity allocation, hiring, sub-vendor selection (within budget envelope set by Carol). Can authorize scope changes for in-flight epics where the cost is internal-only.
- **Communication.** bob@vendor.example. Based in Aarhus (CET).
- **Past decisions.**
  - 2026-05-08: adopted Azure DevOps as the work tracker — [`decisions.md`](./decisions.md#2026-05-08-tracker).

## Carol Martinez {#carol}

- **Role.** Account Manager (vendor side) and customer-facing budget owner.
- **Decision authority.** Budget envelope for the engagement. Can authorize commitments that change the contract terms. Anything that affects the customer's invoice goes through her.
- **Communication.** carol@vendor.example. Based in Madrid (CET). Prefers email; calls before noon CET.
- **Constraints.** Friday afternoons reserved for customer escalation calls.
- **Notes.** Loops in legal on any new data-handling commitments — bring those proactively.

> *Replace these example stakeholders with the people actually on your project.*
