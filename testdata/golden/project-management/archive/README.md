# archive/

Superseded versions of project documents. When a spec, time-plan, epic, or meeting summary is no longer the current version, move it here rather than overwriting or deleting.

## Conventions

- **Date-prefix archived files.** `2026-05-12_dashboard-auth-v1.md`. Sorts chronologically and the origin survives the move.
- **Mirror the origin folder's depth.** A spec from `specs/auth.md` archived becomes `archive/2026-05-12_specs_auth.md` or `archive/specs/2026-05-12_auth.md` — pick one convention and stick with it on the project. The former is flatter; the latter survives if many specs exist.
- **Don't archive things from `integrations/`.** Those track the active toolchain. To stop using a tracker, follow the cleanup steps in `integrations/README.md`.

## Why archive instead of delete

- **Recovery.** Six weeks later, a coworker asks "what was in the v1 spec?". Archive answers in seconds; deletion forces you to dig through git history or wonder if it ever existed.
- **Audit.** When a decision is questioned, the archived materials show what was on the table at the time.
- **Cheap.** Disk is cheap; reconstructing context isn't.
