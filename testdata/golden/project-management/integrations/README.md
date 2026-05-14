# integrations/

One subfolder per active tracker, added by `agent-init add-tracker <name> .` from the workspace root.

Each tracker subfolder contains:

- `README.md` — terminology cheatsheet for that tracker. Jira uses Epic → Feature → User Story; Azure DevOps uses Epic → Feature → PBI; GitHub uses just Issues (optionally grouped via labels/milestones). The skills (`/break-down-epic`, `/sync-tracker`) read these to use the right terminology when talking to the tracker.
- Optional: tracker-specific scripts or templates the team chose to add.

## Adding a tracker

```bash
agent-init add-tracker jira .
agent-init add-tracker ado .
agent-init add-tracker gh .
```

Each call:

1. Writes `integrations/<name>/` if not already present.
2. Adds an entry to `.mcp.json` under `mcpServers` so Claude can talk to the tracker via MCP. The entry uses placeholder credentials — edit before activating.

Calls are idempotent. Running the same tracker twice is a no-op with a notice.

## Multiple trackers simultaneously

Common during migrations: source tracker still owns the open work, target tracker is the new home for new work. Both folders coexist here, both entries coexist in `.mcp.json`, AGENTS.md lists both as active.

When the migration is complete, remove the old tracker:

1. Delete `integrations/<old>/`.
2. Remove the corresponding entry from `.mcp.json` under `mcpServers`.
3. Remove the tracker name from the "Active trackers" line at the top of `AGENTS.md`.

There's no `remove-tracker` subcommand in agent-init yet — the cleanup is manual. The three steps above are all you need.
