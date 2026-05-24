# integrations/jira/

Jira integration. Activated by `agent-init add-tracker jira .`.

## Terminology

Jira has a built-in hierarchy. The standard breakdown:

| Level | Jira issue type | Use for |
|-------|-----------------|---------|
| Top | Epic | One per file in `epics/` |
| Mid | Feature (or "Initiative" in some configs) | Grouping of stories that deliver a coherent capability |
| Work | User Story | The unit a team actually picks up in a sprint |
| Sub | Sub-task | Implementation breakdown within a story; optional |

Jira instances vary — some teams use Initiative → Epic → Story instead of Epic → Feature → Story. The `/break-down-epic` skill reads this file to pick the right terminology. If your instance differs, edit the mapping above.

## MCP server

The `agent-init add-tracker jira` command added this entry to `.mcp.json`:

```json
"atlassian": {
  "command": "uvx",
  "args": ["--from", "mcp-atlassian", "mcp-atlassian"],
  "env": {
    "JIRA_URL": "${env:JIRA_URL}",
    "JIRA_USERNAME": "${env:JIRA_USERNAME}",
    "JIRA_API_TOKEN": "${env:JIRA_API_TOKEN}"
  }
}
```

This calls the [mcp-atlassian community server](https://github.com/sooperset/mcp-atlassian) via `uvx` (which fetches and runs Python tools without permanent installation). **Verify the package name and arg shape against the upstream README before activating** — the Atlassian MCP ecosystem has multiple competing servers with different config shapes (community vs official Atlassian effort vs forks).

### Credentials

`.mcp.json` reads these from your environment via `${env:...}` references, so the
token never lands in the tracked file. Set them in your shell or in a gitignored
`.env` (copy from `.env.example` in this folder); never paste a literal into
`.mcp.json`.

1. **JIRA_URL** — your Atlassian Cloud or Data Center URL. For Cloud: `https://yourdomain.atlassian.net`. For Data Center: the base URL of your installation.
2. **JIRA_USERNAME** — your Atlassian account email.
3. **JIRA_API_TOKEN** — generate at [id.atlassian.com → Security → API tokens](https://id.atlassian.com/manage-profile/security/api-tokens).

```bash
cp integrations/jira/.env.example .env   # then fill in, .env is gitignored
```

After setting credentials, restart your MCP client (or session) so the server reconnects.

### Server-specific gotchas

- Some Jira MCP servers require an explicit project key per call. Others infer from URL. Read the upstream README.
- API tokens are per-account; if you switch accounts, the token must be regenerated.
- Self-hosted Jira (Data Center) instances may need basic-auth or OAuth instead of API tokens; check the server's docs.

## Conventions for `/sync-tracker`

When pushing to Jira:

- Issue summaries start with a verb. "Implement OIDC login flow" not "OIDC login".
- Issue descriptions follow the work-item shape from `epics/_template_.md`'s Breakdown section, formatted in Jira's wiki markup (or ADF if the MCP server supports it).
- The parent Epic link is set on each child issue.
- Components and Labels can be applied if your project has them; the `/sync-tracker` skill will ask before adding new ones.

When pulling from Jira:

- The pull updates the local epic file's **Breakdown** statuses.
- Issues created in Jira directly (not via sync) are surfaced as open questions ("Issue PROJ-1234 exists in Jira but isn't in the local plan").

## Gotchas

- **Project scope.** Most Jira MCP servers operate on one project at a time. If your work spans projects, configure one entry per project (use distinct mcpServers keys: `atlassian-proj1`, `atlassian-proj2`).
- **Custom fields.** Jira instances accumulate custom fields. The MCP server may not surface all of them. Test with a small push before bulk-creating issues.
- **Permissions.** API tokens inherit the account's permissions. If `/sync-tracker` returns 403 errors, the account doesn't have create/edit permission on the project.

## Removing this tracker

To stop using Jira for this workspace:

1. Delete this folder (`integrations/jira/`).
2. Remove the `"atlassian"` entry from `.mcp.json` under `mcpServers`.
3. Remove `jira` from the "Active trackers" line at the top of `AGENTS.md`.

`agent-init` doesn't ship a `remove-tracker` subcommand — the three steps above are the manual procedure.
