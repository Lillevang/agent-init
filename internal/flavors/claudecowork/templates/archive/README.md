# archive/

Superseded drafts, old briefs, retired reference material. Anything you'd otherwise delete, move here instead.

## Why archive instead of delete

- **Recovery.** Coworkers occasionally need to recover a version that "looked fine last month". Archive preserves the option.
- **Audit.** When a decision is questioned later, the archived materials show what was on the table at the time.
- **Cheap.** Disk is cheap; explaining a deleted file isn't.

## Conventions

- **Date-prefix archived files**: `2026-05-12_meeting_brief.md`. Sorts chronologically, survives moves.
- **Mirror the original location's depth**. If a brief lived at `hoto/brief.md`, archive it as `archive/2026-05-12_hoto_brief.md` so the origin is recoverable from the filename.
- **Don't archive things from `reference/`** unless they're genuinely retired (vendor sunset, contract expired). Reference materials are usually still relevant even when stale.

## Claude's rules for this folder

Claude treats `archive/` as read-only context. It may quote or compare against archived versions but won't modify them. To restore a file from archive, you (the human) move it back — Claude will not silently un-archive.
