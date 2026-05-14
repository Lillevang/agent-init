# integrations/github/

GitHub Issues integration. Activated by `agent-init add-tracker gh .`.

## Terminology

GitHub doesn't have a built-in Epic/Feature/Story hierarchy — it has Issues, optionally grouped via:

- **Labels** for type (`epic`, `feature`, `task`, `bug`).
- **Milestones** for time-bound groupings (e.g., one milestone per release).
- **Projects (v2)** for board-style tracking with custom fields.
- **Sub-issues** (newer feature) for parent/child relationships.

For this workspace's epic-breakdown convention, the recommended mapping is:

| Local concept | GitHub representation |
|---------------|------------------------|
| Epic (file under `epics/`) | Issue with `epic` label, listed in a parent milestone |
| Feature | Issue with `feature` label, linked from the epic as a task list |
| Work item (Story/PBI equivalent) | Issue with `task` label, linked from the parent feature |

Use sub-issues when available — they survive label renames and integrate with GitHub's UI better than task-list links.

## MCP server

The `agent-init add-tracker gh` command added this entry to `.mcp.json`:

```json
"github": {
  "command": "npx",
  "args": ["-y", "@modelcontextprotocol/server-github"],
  "env": {
    "GITHUB_PERSONAL_ACCESS_TOKEN": "${env:GITHUB_TOKEN}"
  }
}
```

This calls the official MCP server at [@modelcontextprotocol/server-github](https://github.com/modelcontextprotocol/servers/tree/main/src/github). **Verify the package name and arg shape against the upstream project before activating** — MCP server names and CLI flags evolve.

### Credentials

The default config reads `GITHUB_TOKEN` from your host environment via the `${env:GITHUB_TOKEN}` interpolation. Set it via:

```bash
# Option 1: a fine-grained personal access token (preferred for org repos)
gh auth token
# Option 2: a classic PAT with `repo` and `read:org` scopes
```

If your token is named differently (e.g. `GH_TOKEN`), edit the env interpolation in `.mcp.json` accordingly.

### Scopes needed

For the `/sync-tracker` skill to push/pull issues, the token needs:

- `repo` (or fine-grained equivalent: Issues read+write, Pull requests read for cross-linking).
- `read:org` if you reference org-level projects or milestones owned by a team.

## Conventions for `/sync-tracker`

When pushing to GitHub:

- Issue titles start with a verb. "Implement OIDC login flow" not "OIDC login".
- Issue bodies follow the work-item shape from `epics/_template_.md`'s Breakdown section: description, acceptance criteria, assumptions.
- Labels are applied automatically: `epic`, `feature`, `task` per the mapping above.
- The parent epic's link is included in each child issue's body for traceability.

When pulling from GitHub:

- The pull updates the local epic file's **Breakdown** status indicators.
- Issues created on GitHub directly (not via sync) are surfaced as open questions ("Issue #123 exists in GitHub but isn't in the local plan").

## Gotchas

- **Repo selection.** GitHub MCP servers usually operate on a single repo at a time. If your project spans multiple repos, you'll need either one tracker entry per repo (use distinct mcpServers keys like `github-repo1`, `github-repo2`) or a wrapper that routes by label.
- **Token rotation.** PATs expire. If `/sync-tracker` starts returning auth errors, rotate the token and restart your MCP client.
- **Rate limits.** GitHub's REST API has per-token rate limits. A first-time bulk push of a large epic can hit them; the MCP server should back off, but if pushes start failing partway, retry after the rate-limit window resets.

## Removing this tracker

To stop using GitHub for this workspace:

1. Delete this folder (`integrations/github/`).
2. Remove the `"github"` entry from `.mcp.json` under `mcpServers`.
3. Remove `github` (or `gh`) from the "Active trackers" line at the top of `AGENTS.md`.

`agent-init` doesn't ship a `remove-tracker` subcommand — the three steps above are the manual procedure.
