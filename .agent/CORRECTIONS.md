# Corrections

Patterns the agent has gotten wrong, and the preferred form. Read this before starting work.

> **Maintenance note.** Keep this file under ~30 entries. When it grows beyond that, refactor recurring lessons into `AGENTS.md` or `CODEBASE.md` and remove them here. An overlong corrections file is ignored by the agent (and rightly so — it stops being useful).

## Format

Each entry is a heading + a bad example + a good example + (optional) a one-line rationale. Be concrete. "Be more careful" is not an entry.

---

## Example: Don't use `any` in TypeScript

**Bad:**
```ts
function process(data: any) { ... }
```

**Good:**
```ts
function process(data: UserPayload) { ... }
```

If you genuinely don't know the type, use `unknown` and narrow.

---

<!-- ADD NEW ENTRIES BELOW -->
