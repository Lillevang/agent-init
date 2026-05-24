# integrations/azure-devops/

Azure DevOps integration. Activated by `agent-init add-tracker ado .`.

## Terminology

Azure DevOps uses a fixed Work Item hierarchy:

| Level | ADO Work Item type | Use for |
|-------|--------------------|---------|
| Top | Epic | One per file in `epics/` |
| Mid | Feature | Grouping of PBIs that deliver a coherent capability |
| Work | Product Backlog Item (PBI) | The unit a team picks up in a sprint (Scrum); equivalent is User Story in Agile process template |
| Sub | Task | Implementation breakdown within a PBI; tracks remaining hours |

Note: ADO's hierarchy depends on the **process template** (Basic, Agile, Scrum, CMMI). The mapping above is for **Scrum**. For **Agile**, replace PBI with User Story. The `/break-down-epic` skill reads this file to pick the right terminology — edit the mapping above if your process is different.

## MCP server

The `agent-init add-tracker ado` command added this entry to `.mcp.json`:

```json
"azure-devops": {
  "command": "npx",
  "args": ["-y", "@azure-devops/mcp"],
  "env": {
    "ADO_ORG_URL": "${env:ADO_ORG_URL}",
    "ADO_PROJECT": "${env:ADO_PROJECT}",
    "ADO_PAT": "${env:ADO_PAT}"
  }
}
```

> **The ADO MCP ecosystem is the least mature of the three trackers** at the time of writing. Microsoft has published multiple servers under different package names; the community has also forked. **Always check the upstream README before activating.** The package name above may need updating to match whatever the current authoritative server is.

### Credentials

`.mcp.json` reads these from your environment via `${env:...}` references, so the
PAT never lands in the tracked file. Set them in your shell or in a gitignored
`.env` (copy from `.env.example` in this folder); never paste a literal into
`.mcp.json`. After setting credentials, restart your MCP client (or session) so
the server reconnects.

1. **ADO_ORG_URL** — your Azure DevOps organization URL: `https://dev.azure.com/yourorg`.
2. **ADO_PROJECT** — the project name within that org. Often required even when the MCP server "could" infer it.
3. **ADO_PAT** — a Personal Access Token. Generate at https://dev.azure.com/yourorg/_usersSettings/tokens.

```bash
cp integrations/azure-devops/.env.example .env   # then fill in, .env is gitignored
```

#### PAT scopes needed

For `/sync-tracker` to push/pull work items, the PAT needs:

- **Work Items**: Read & Write
- **Project and Team**: Read (to resolve area/iteration paths)

Set the PAT to expire on a reasonable schedule (90 days is the ADO default); rotate before expiry or you'll wonder why sync starts failing on a random Tuesday.

## Conventions for `/sync-tracker`

When pushing to Azure DevOps:

- Work item titles start with a verb. "Implement OIDC login flow" not "OIDC login".
- Descriptions follow the work-item shape from `epics/_template_.md`, formatted in HTML (ADO descriptions are HTML, not Markdown).
- Area Path defaults to the project root unless the local epic file specifies one (under a `**Area path**` field in the epic).
- Iteration Path is not set by default; the team's planner assigns iterations.

When pulling from Azure DevOps:

- The pull updates the local epic file's **Breakdown** statuses.
- Work items created in ADO directly (not via sync) are surfaced as open questions.

## Gotchas

- **Process template differences.** Scrum uses PBI; Agile uses User Story; Basic uses Issue. Make sure the terminology mapping above matches your project.
- **Area paths are tree-shaped.** If your team uses deep area paths, the MCP server needs to know which path to assign. Configure via `ADO_AREA_PATH_DEFAULT` if the server supports it, or specify per-epic in the local file.
- **Wiki integration.** ADO has a separate wiki product (Project Wikis). The MCP server typically doesn't touch the wiki; if your specs live there, treat the wiki as a separate sync target.
- **Rate limits.** ADO's REST API is generally permissive but enforces a per-user concurrent request limit. Bulk pushes work fine for typical project sizes (hundreds of items, not thousands).

## Removing this tracker

To stop using Azure DevOps for this workspace:

1. Delete this folder (`integrations/azure-devops/`).
2. Remove the `"azure-devops"` entry from `.mcp.json` under `mcpServers`.
3. Remove `ado` from the "Active trackers" line at the top of `AGENTS.md`.

`agent-init` doesn't ship a `remove-tracker` subcommand — the three steps above are the manual procedure.

## Migration scenarios

ADO ↔ Jira migrations are common (in both directions). Run with **both trackers active** during the migration:

```bash
agent-init add-tracker ado .   # if not already
agent-init add-tracker jira .  # add the migration target
```

Now both `integrations/azure-devops/` and `integrations/jira/` exist, and `.mcp.json` lists both servers. `/sync-tracker` will ask which tracker to push to per epic. When the migration completes (no items left in the source tracker), remove the source via the three-step cleanup above.
