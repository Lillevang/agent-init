---
name: sync-tracker
description: Push the local plan (epics + work items) to the active tracker(s) via MCP, or pull tracker changes back into the local plan. The only skill that writes to external systems — always show a diff before pushing. Invoke when an epic is ready to materialize as tracker items, or when the local plan looks out of date with the tracker.
---

# sync-tracker

You're being invoked because the local plan needs to be reconciled with one or more external trackers (Jira, Azure DevOps, GitHub). Unlike the other skills in this workspace, this one writes to external systems via MCP. Treat that with care: every other skill is reversible by editing a local file, this one isn't.

## Inputs you need

If the user didn't say, ask:

1. **Direction.** Push (local → tracker), pull (tracker → local), or both?
2. **Scope.** Which epic(s) or work items? "Everything" is acceptable for the first sync of a brand-new project; on an ongoing project, scope tightly.
3. **Which tracker(s).** Read `AGENTS.md`'s "Active trackers" line. If multiple are active, push to which one? Pull from which one? (Trackers don't merge: don't push to two trackers in one run unless the user explicitly asks.)

## Preflight

### 1. Verify MCP is wired

Check `.mcp.json` at the workspace root. The target tracker must have an entry under `mcpServers` with real credentials (not the placeholder values from `agent-init add-tracker`). If credentials are placeholder or missing:

- Tell the user what's missing.
- Point at `integrations/<tracker>/README.md` for setup instructions.
- Stop. Do not attempt to call the MCP server.

### 2. Read the local plan

For push: the source of truth is `epics/<name>.md` (or whichever files the user scoped). Parse the **Breakdown** section. Identify which items already have a tracker link (idempotency — don't recreate items that already exist) and which are new.

For pull: query the tracker for the items linked from the scoped local files. Identify what changed since the last sync.

### 3. Show the diff

This is the single most important step. Before any write, print a diff:

- **Push diff.** "I will create N new items under Feature 'X'. I will update M existing items where the local description changed. I will *not* touch P items that are unchanged."
- **Pull diff.** "I will update the local breakdown of epic Y because items A, B, C have new statuses in the tracker. I will *not* touch the business value or assumptions sections."

Wait for the user to confirm. **Do not write without explicit confirmation.**

## Push

After confirmation:

1. Create new items in the tracker via the MCP server. Use the tracker's terminology (User Story / PBI / Issue per the cheatsheet at `integrations/<tracker>/README.md`).
2. Capture the returned URL or ID for each created item.
3. Update `epics/<name>.md`'s **Breakdown** section to add a link to each newly-created item.
4. If an existing item's title or description was changed locally and you're pushing the update, note the previous tracker state in case it needs reverting.

When the push completes, summarize what was created/updated, and link each result.

## Pull

After confirmation:

1. Fetch the current state of the scoped items.
2. Update `epics/<name>.md` to reflect the new statuses, completed counts, or moved items.
3. If a tracker item was newly created on the tracker side (someone added it through Jira/ADO/GH directly), surface it as a question: "Item Z exists in the tracker but not in the local plan. Add it to the epic, archive it, or ignore?"

Don't silently add tracker-side items to local epics; that hides who's making changes where.

## Both directions in one run

Sometimes the user wants reconciliation in both directions. Do **pull first**, then **push**:

1. Pull: bring local file state up to date with what's actually in the tracker.
2. Re-evaluate the push diff against the now-current local state.
3. Confirm again, then push.

This avoids the situation where you'd push changes based on a stale local view, overwriting work someone else did in the tracker.

## When to stop and ask

- If the diff includes destructive operations (delete a tracker item, close an issue, move work to "won't do"), ask for explicit confirmation per-item, not bulk.
- If the local plan references decisions that aren't in `decisions.md`, push pause — the breakdown might be missing context that affects how items should be created.
- If you discover the tracker has work items that aren't in the local plan AND aren't trivially mappable to existing epics, stop. Surface them as a question rather than guessing which epic they belong to.
- If the MCP server returns an authentication error or unexpected schema, stop and report. Don't retry blindly.

## Style

- The diff output should be skim-able. Bullet lists by action: created, updated, unchanged, skipped (with reason).
- Don't editorialize ("this is a clean sync!"). State the facts.
- Link every result. After a push, the user should be able to click straight into the tracker for each item.
