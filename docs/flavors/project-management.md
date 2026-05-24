# project-management

A flavor for running the *business side* of a software project — requirements, decisions, stakeholders, meetings, time plans, and breaking work down into trackable items (Jira / Azure DevOps / GitHub Issues). Most projects don't fail because the code is hard. They fail because the requirements drift, the decisions evaporate, and nobody remembers who agreed to what. This flavor exists to fight that.

Structurally it's closest to `claude-cowork`: no devcontainer, no `Justfile`, no done-gate. But where `claude-cowork` is generic doc work, this flavor is opinionated about *what* lives where and ships five purpose-built skills, plus a separate subcommand for wiring in tracker integrations.

## What it scaffolds

```
<workspace>/
├── AGENTS.md                   # canonical multi-role instructions
├── README.md                   # human onboarding
├── decisions.md                # append-only decision log (Context/Options/Decision/Authority/Reasoning/Implications)
├── open-questions.md           # things to clarify; each entry has owner + date raised + status
├── stakeholders.md             # who can decide what (with anchors for cross-linking from decisions.md)
├── .gitignore                  # standard hygiene + .env exclusion
├── .mcp.json                   # empty mcpServers map; extended by add-tracker
├── epics/_template_.md         # shape of an epic file
├── meetings/_template_.md      # shape of a meeting summary
├── specs/_template_.md         # shape of a spec / requirement doc
├── time-plans/_template_.md    # shape of a milestone plan
├── integrations/README.md      # explains the folder; populated by add-tracker
├── archive/README.md
└── .claude/skills/
    ├── intake-meeting/
    ├── break-down-epic/
    ├── log-decision/
    ├── track-stakeholder/
    └── sync-tracker/
```

Source: [internal/flavors/projectmgmt/](../../internal/flavors/projectmgmt/).

## Usage

```bash
agent-init init --no-git project-management ~/work/myproject
```

`--no-git` is the typical choice — the workspace works both as a git repo and as a OneDrive folder; the scaffold doesn't init git. Add `git init` yourself if you want version history.

After scaffolding, wire one or more trackers:

```bash
agent-init add-tracker gh    ~/work/myproject
agent-init add-tracker jira  ~/work/myproject
agent-init add-tracker ado   ~/work/myproject
```

Each call adds a `.mcp.json` entry and writes `integrations/<tracker>/README.md`. Multiple trackers can coexist — useful during migrations.

## Worked examples

Each of the five skills has a specific job. Below are concrete walk-throughs showing the kind of input it takes and the artifacts it leaves behind.

### 1. `/intake-meeting` — meeting notes → structured outputs

**You paste this into Claude:**

> /intake-meeting
>
> Just got out of the kickoff with Acme. Notes:
>
> - Bob said we should adopt ADO. Customer's other vendors use it. Carol agreed but flagged budget impact.
> - Alice raised: which auth provider? Entra is in place at the customer, but Okta is on the table. Nobody could decide — need PMO.
> - Bob committed: Alice will draft the auth spec by Friday.
> - We agreed: phase 1 doesn't touch reporting; that's phase 2.
> - Carol mentioned a new person — Diana, customer PMO Director. She's the one to talk to about scope changes.

**What lands on disk:**

`meetings/2026-05-14_kickoff.md`:

```markdown
# Meeting: Acme kickoff

- **Date.** 2026-05-14
- **Attendees.** Alice (Solutions Architect), Bob (Eng Lead), Carol (Account Manager), Diana (Customer PMO Director)
- **Type.** kickoff

## Decisions

- Adopt Azure DevOps as the work tracker → [`decisions.md`](../decisions.md#2026-05-14-tracker).
- Phase 1 scope excludes reporting; reporting moves to phase 2 → [`decisions.md`](../decisions.md#2026-05-14-phase-1-scope).

## Action items

- [ ] Owner: Alice — draft the auth spec — by 2026-05-17 (Friday)

## Open questions surfaced

- [Q: Which auth provider should the customer use?](../open-questions.md#auth-provider)

## Discussion

Bob proposed Azure DevOps based on the customer's existing vendor ecosystem...
```

`decisions.md` gets two new entries appended. `open-questions.md` gets one. `stakeholders.md` gets a new section for Diana (the previously-unknown PMO Director).

After the skill finishes, Claude tells you what it wrote and what it surfaced for you to confirm.

### 2. `/break-down-epic` — vague description → ordered work items + assumptions

**You paste this into Claude:**

> /break-down-epic
>
> Epic: customer dashboard MVP. Customer wants a dashboard for their service desk team to see SLA status across our managed services. Phase 1 only. Should integrate with our existing API.

**What lands on disk:**

`epics/customer-dashboard-mvp.md`:

```markdown
# Epic: customer-dashboard-mvp

## Status
`broken-down` — 2026-05-14, initial scoping run.

## Business value
The service desk team currently runs SQL queries against the support DB to check SLA status; takes ~15 min per check, done 6-10 times a day. A dashboard reduces that to seconds and removes the dependency on DB read access (which the team shouldn't have anyway). Target: 80% of SLA queries served by the dashboard within 3 months of GA.

## Scope
In: SLA status (green/amber/red), trend over last 30 days, drill-down to ticket level.
Out: ticket modification, customer-facing views, reporting export — reporting was moved to phase 2 per [`decisions.md`](../decisions.md#2026-05-14-phase-1-scope).

## Acceptance criteria
1. A service desk operator can see SLA status for all managed services on one screen.
2. The 30-day trend chart loads in under 2 seconds for a customer with 500 active tickets.
3. The drill-down lists tickets ordered by time-to-breach, ascending.

## Assumptions
- The existing /sla API endpoint is performant enough for direct UI binding (see open question).
- The customer is OK with us using Entra ID for auth (depends on auth-provider decision; see open question).
- Phase 1 is single-tenant.

## Breakdown
- **Feature: SLA overview screen**
  - PBI: Implement SLA status grid component
  - PBI: Wire grid to /sla API endpoint
  - PBI: Define color coding for green/amber/red states
- **Feature: 30-day trend**
  - PBI: Implement trend chart component
  - PBI: Add /sla/history endpoint to API
  - PBI: Wire chart to history endpoint with caching
- **Feature: Ticket drill-down**
  - PBI: Implement ticket list view
  - PBI: Sort by time-to-breach

## Risks
- API endpoint may not be performant — mitigation: load-test before committing UI binding.

## Stakeholders
- [Alice](../stakeholders.md#alice) — technical architecture
- [Bob](../stakeholders.md#bob) — delivery commitment
- [Diana](../stakeholders.md#diana) — scope authority on customer side
```

Claude also appends to `open-questions.md`:

```markdown
## Q: Is the existing /sla API endpoint performant enough for direct UI binding?

- **Raised.** 2026-05-14
- **Owner.** Alice
- **Status.** open
- **Blocks.** [`epics/customer-dashboard-mvp.md`](./epics/customer-dashboard-mvp.md) — assumption underlying the dashboard breakdown.
- **Question.** The breakdown assumes /sla can serve direct UI calls. If response times exceed 500ms p95 under typical load, we need a caching layer in front, which adds 2 PBIs and a Redis dependency.
```

The breakdown uses **PBI** because the active tracker (per `AGENTS.md`'s "Active trackers" line) is ADO. If you'd added Jira instead, it would say **User Story**; for GH, **Issue**.

### 3. `/log-decision` — append a structured entry to `decisions.md`

**You paste this into Claude:**

> /log-decision
>
> We're going with Entra ID for auth. Decided in this morning's PMO sync. Diana authorized it. The decision was driven by the fact that the customer's IT team already manages Entra and adding Okta would require a 2-week procurement cycle we don't have time for.

**What lands on disk:**

A new entry appended below the `<!-- ADD NEW ENTRIES BELOW -->` marker in `decisions.md`:

```markdown
## 2026-05-14 — Adopted Entra ID as the auth provider for the dashboard

**Context.** The customer dashboard MVP needs an auth provider; both Entra ID and Okta were viable. The choice was raised in the kickoff meeting (see `meetings/2026-05-14_kickoff.md`) and parked in [`open-questions.md`](./open-questions.md#auth-provider) pending PMO authorization.

**Options considered.**
- A: Entra ID. Customer's IT team already manages it; no new procurement.
- B: Okta. Better federation features but would require ~2 weeks of customer procurement.

**Decision.** A — Entra ID.

**Authorized by.** [Diana](./stakeholders.md#diana)

**Reasoning.** Procurement timing was the deciding factor. The 2-week Okta procurement cycle would push the dashboard MVP delivery past the customer's committed quarter-end demo. Federation features that Okta would add aren't needed for phase 1; they could be added in a later phase if the customer requests them.

**Implications.** Phase 1 SSO setup uses Entra. Alice updates the auth spec accordingly. The "Auth provider" open question is marked answered.

**Related.** Open question: [`open-questions.md`](./open-questions.md#auth-provider). Meeting: [`meetings/2026-05-14_pmo-sync.md`](./meetings/2026-05-14_pmo-sync.md).
```

Claude also updates `stakeholders.md` (Diana's **Past decisions** list gains this entry) and `open-questions.md` (the auth-provider question moves to `answered`, with the answer added below the question).

### 4. `/track-stakeholder` — add or refine a profile

**You paste this into Claude:**

> /track-stakeholder
>
> Add Diana. She's the customer's PMO Director. She authorizes scope changes on the customer side. Email is diana@acme.example. Based in London — GMT. Mention she's vacationing the second half of June.

**What lands on disk:**

A new entry in `stakeholders.md`:

```markdown
## Diana Mukherjee {#diana}

- **Role.** PMO Director at Acme (customer side).
- **Decision authority.** Scope changes on the customer side. Authorizes commitments that affect the customer's delivery expectations or contracted scope.
- **Communication.** diana@acme.example. Based in London (GMT/BST).
- **Constraints.** Out 2026-06-15 through 2026-06-30 (vacation).
- **Past decisions.** *(empty so far)*
- **Notes.** *(empty so far)*
```

The `{#diana}` anchor lets future `decisions.md` entries link her in **Authorized by** as `[Diana](./stakeholders.md#diana)`.

If you later run `/track-stakeholder` for Diana again with an update — say, you learn she also has budget authority up to $100k — the skill *updates* the entry (doesn't duplicate). Don't replace **Past decisions**; append.

### 5. `/sync-tracker` — push the local plan to Jira/ADO/GH

This is the only skill that writes to external systems. It always shows a diff before pushing and waits for confirmation.

**You ask:**

> /sync-tracker push the customer-dashboard-mvp epic to ADO

**What Claude does:**

1. Reads `epics/customer-dashboard-mvp.md`.
2. Reads `.mcp.json` to find the ADO server config (added by `add-tracker ado`).
3. Calls the ADO MCP server to see which items already exist (matching by title or stored URL).
4. Prints a diff:

```
Push diff (Azure DevOps):
  Create 1 Epic:
    - customer-dashboard-mvp
  Create 3 Features under that epic:
    - SLA overview screen
    - 30-day trend
    - Ticket drill-down
  Create 8 PBIs under those Features:
    - Implement SLA status grid component
    - Wire grid to /sla API endpoint
    ... (6 more)
  Update 0 existing items.
  Skip 0 unchanged items.

Confirm to proceed?
```

5. After confirmation, calls the MCP server to create each item. Stores the returned URL/ID in the epic file:

```markdown
## Breakdown
- **Feature: SLA overview screen** [ADO #1234](https://dev.azure.com/...)
  - PBI: Implement SLA status grid component [#1235](https://...)
  ...
```

If you later edit the local breakdown and re-run `/sync-tracker`, only the changed items are pushed; unchanged items are skipped.

## How the five skills talk to each other

`/intake-meeting` typically triggers:

- `/log-decision` (for any decisions captured in the meeting)
- `/track-stakeholder` (for any new stakeholders mentioned)

`/break-down-epic` typically triggers:

- Writes to `open-questions.md` (for assumptions worth confirming)
- Eventually: `/sync-tracker` (when the breakdown is ready to materialize)

`/log-decision` does not push to a tracker. Decisions are local-first; the tracker is a downstream consumer.

`/sync-tracker` is the only skill that writes externally. Treat it accordingly — every other skill is reversible by editing a local file.

## Storage modes

Works as either a **git repo** or a **OneDrive folder**, identically.

- **Git mode.** Init git after scaffold (`cd <workspace> && git init`). Decisions and stakeholder history become tracked via commits. Cross-team review goes through PRs. Use this when the workspace lives alongside the code repo or as a sibling.
- **OneDrive mode.** Don't init git; let OneDrive sync the folder. Same files, no version history. Use this when non-developer collaborators need access without learning git.

The scaffold doesn't create symlinks — OneDrive's symlink handling is unreliable on Windows. If you want `CLAUDE.md` alongside `AGENTS.md`, create it manually:

- Linux/macOS + git: `ln -s AGENTS.md CLAUDE.md`
- Windows / OneDrive: `copy AGENTS.md CLAUDE.md`, keep in sync manually.

## Tracker integration via `add-tracker`

Trackers are added incrementally with the [`add-tracker` subcommand](../cli.md#add-tracker), not at scaffold time. This means:

- A workspace starts tracker-agnostic. You can begin filling in epics, stakeholders, and decisions before deciding which tracker to use.
- Multiple trackers can coexist. Common during migrations (ADO → Jira), or for projects whose wiki lives in one system and tasks in another.
- `add-tracker` is idempotent. Running `add-tracker gh` twice is a no-op the second time.

Each tracker integration writes a terminology cheatsheet and an MCP server config:

- **GitHub** (`gh`): Issues with `epic`/`feature`/`task` labels. MCP server: `@modelcontextprotocol/server-github`.
- **Jira** (`jira`): Epic → Feature → User Story. MCP server: `mcp-atlassian` (community).
- **Azure DevOps** (`ado`): Epic → Feature → PBI. MCP server: `@azure-devops/mcp` (official; verify name).

> **MCP servers are moving targets.** Each tracker's `integrations/<tracker>/README.md` includes a note saying "verify against the upstream project before activating". The default `.mcp.json` entries are reasonable starting points but assume nothing about the current state of the MCP ecosystem.

## Edge cases and gotchas

- **MCP server credentials.** `add-tracker` writes `.mcp.json` entries whose env values are `${env:VAR_NAME}` references, never literals — the secret is read from your shell environment, so it never lands in the tracked file. Set the vars in your shell or a gitignored `.env` (each tracker ships `integrations/<tracker>/.env.example`). For the GitHub tracker, `export GITHUB_TOKEN="$(gh auth token)"` reuses the devcontainer's existing login with no separate PAT. Changing `.mcp.json` needs an MCP/session restart to reconnect. See [credential setup in `docs/cli.md`](../cli.md#add-tracker).
- **One MCP entry per tracker instance.** If you work with two GitHub repos in the same workspace, you need two entries (e.g., `github-repo1`, `github-repo2`). The `add-tracker` command only writes a single canonical entry per tracker; copy-and-edit by hand for the second.
- **Removing a tracker.** No `remove-tracker` subcommand yet. Manual cleanup: delete `integrations/<tracker>/`, remove the entry from `.mcp.json`, remove the name from `AGENTS.md`'s "Active trackers" line.
- **Stakeholder anchors are forever.** Once `decisions.md` links to `stakeholders.md#alice`, don't rename the anchor — it silently breaks the link. If you need to disambiguate two Alices later, add a new entry with a different anchor rather than renaming.
- **Decisions are append-only.** Reversed decisions get a *new* entry that references the old one. Don't edit the old entry.

## Source pointers

- Flavor package: [`internal/flavors/projectmgmt/`](../../internal/flavors/projectmgmt/)
- Flavor templates: [`internal/flavors/projectmgmt/templates/`](../../internal/flavors/projectmgmt/templates/)
- Five skills: [`internal/flavors/projectmgmt/templates/.claude/skills/`](../../internal/flavors/projectmgmt/templates/.claude/skills/)
- `add-tracker` subcommand: [`internal/cli/cli.go`](../../internal/cli/cli.go) (`runAddTracker`)
- Tracker registry: [`internal/trackers/registry.go`](../../internal/trackers/registry.go)
- MCP JSON merge: [`internal/trackers/mcp.go`](../../internal/trackers/mcp.go) (`MergeMCPServer`)
- Tracker packages: [`internal/trackers/gh/`](../../internal/trackers/gh/), [`internal/trackers/jira/`](../../internal/trackers/jira/), [`internal/trackers/ado/`](../../internal/trackers/ado/)

## Tests

- Golden test: [`test/golden_test.go`](../../test/golden_test.go) — `TestFlavorGolden/project-management`.
- MCP merge: [`internal/trackers/mcp_test.go`](../../internal/trackers/mcp_test.go) — covers add-new, idempotency, missing-file, malformed-JSON, missing-mcpServers-key.
- `add-tracker` CLI: [`internal/cli/cli_test.go`](../../internal/cli/cli_test.go) — `TestAddTrackerWritesFilesAndMergesMCP`, `TestAddTrackerIsIdempotent`, `TestAddTrackerMultipleCoexist`, `TestAddTrackerRejectsMissingScaffold`, `TestAddTrackerRejectsUnknownTracker`.
