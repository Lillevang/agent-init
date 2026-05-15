# Corrections

Patterns the agent has gotten wrong, and the preferred form. Read this before starting work.

> **Maintenance note.** Keep this file under ~30 entries. When it grows beyond that, refactor recurring lessons into `AGENTS.md` or `CODEBASE.md` and remove them here. An overlong corrections file is ignored by the agent (and rightly so — it stops being useful).

## Format

Each entry is a heading + a bad example + a good example + (optional) a one-line rationale. Be concrete. "Be more careful" is not an entry.

---

## Example: Don't use `count` for distinct resources

**Bad:**
```hcl
resource "aws_iam_user" "team" {
  count = length(var.team_members)
  name  = var.team_members[count.index]
}
```
Removing a name from the middle of `team_members` causes Terraform to destroy and recreate every user after it.

**Good:**
```hcl
resource "aws_iam_user" "team" {
  for_each = toset(var.team_members)
  name     = each.key
}
```
`for_each` keys resources by value, so removing one only destroys that one.

---

## Example: Use fully-qualified module names in Ansible

**Bad:**
```yaml
- name: install nginx
  package:
    name: nginx
    state: present
```

**Good:**
```yaml
- name: install nginx
  ansible.builtin.package:
    name: nginx
    state: present
```

Bare `package` can be shadowed by a collection that ships its own `package` module, leading to surprising behaviour.

---

<!-- ADD NEW ENTRIES BELOW -->
