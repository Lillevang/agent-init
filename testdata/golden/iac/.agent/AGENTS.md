# Agent Instructions for iac

You are working inside a sandboxed dev container on an Infrastructure-as-Code repo that may contain Terraform, Ansible, or both. Your filesystem access is bounded but not zero — be deliberate. This file is the canonical source of truth for how to work in this repo. Both Codex (`AGENTS.md`) and Claude Code (`CLAUDE.md`, symlinked here) read it.

> **Edit me.** Replace project-generic prose with project-specific facts (which clouds, which environments, which inventories, who owns the state backend).

---

## The done-gate

Before declaring **any** task complete, you must run:

```bash
./.agent/scripts/check.sh
```

The script runs `fmt`, `lint`, `typecheck`, `test` via Just. Missing recipes are skipped silently — but if a recipe exists and fails, the task is not done.

For this flavor the recipes do:

- `just fmt` — `terraform fmt -recursive` on any `*.tf` present.
- `just lint` — `tflint` (when configured), `ansible-lint` (when playbooks/roles exist), `yamllint .`.
- `just typecheck` — `terraform validate` per module (requires `terraform init -backend=false` first; `post-create.sh` does that), `ansible-playbook --syntax-check` on every playbook.
- `just test` — `terraform test` if `tests/` is present, `molecule test` if `molecule/` is present.
- `just security` — `tfsec` and `trivy config` (advisory; not part of the done-gate by default).

## Hard rules (IaC-specific)

These dangers are why this flavor exists. Violate any of them and you have shipped a security incident.

### State files

- **Never commit `*.tfstate`, `*.tfstate.backup`, or `.terraform/`**. The `.gitignore` blocks them; if you find yourself about to override that, stop.
- Configure a **remote backend** (S3+DynamoDB, GCS, azurerm, Terraform Cloud) before the first `terraform apply` in any shared environment.
- State files contain plaintext secrets even when you used `sensitive = true` in your config. Treat them as credentials.

### Cloud credentials

- `~/.aws`, `~/.config/gcloud`, and the Azure CLI cache are mounted **read-only** into the container so `terraform apply` and Ansible cloud modules can use them. Do **not** ask the agent to write credentials to disk.
- Prefer per-environment named profiles (`AWS_PROFILE`, `gcloud --configuration`) over editing default credentials.
- Service-account JSON keys and IAM access keys must never enter the repo. If a tutorial tells you to commit one, the tutorial is wrong.

### SSH keys (Ansible)

- `~/.ssh` is mounted **read-only** into the container. `ansible-playbook` from inside the container reaches the host's SSH targets.
- For production runs, prefer host-side `ssh-agent` forwarding over copying keys into the workspace.
- Never commit `id_rsa`, `id_ed25519`, or any private key. The `.gitignore` blocks `*.pem` and `*.key`; do not override.

### Vault passwords (Ansible)

- `ansible-playbook --vault-password-file=...` must point at a file **outside the workspace** (e.g. `~/.ansible/vault-pass`).
- `.vault_pass*` is gitignored as a safety net; treat that as belt-and-braces, not as your only line of defense.
- Vault-encrypted vars files (`group_vars/*/vault.yml`) are fine to commit when properly encrypted. Plain-text vars files containing secrets are not.

### Real inventory

- `inventory/hosts.yml.example` is committed; `inventory/hosts.yml` (with real hostnames, IPs, jump hosts) is gitignored.
- Don't put real production hostnames into module fixtures or playbook examples either — they leak topology.

### Apply discipline

- For Terraform: **always `terraform plan` first**, read the diff, then `apply`. Never `apply -auto-approve` interactively.
- For Ansible: use `--check --diff` first on any non-trivial play. Use `--limit` aggressively when developing against shared inventory.
- Destructive resource recreations (anything Terraform shows as `-/+`) require explicit operator confirmation. Surface them in your turn output, do not silently apply.

## Project context

> **Replace this section.** Describe:
> - Which providers/clouds this manages (AWS account IDs at high level, GCP projects, Azure subscriptions).
> - Which environments exist (`dev/staging/prod`) and how they map to workspaces or directories.
> - Where state lives (S3 bucket, GCS bucket, Terraform Cloud workspace).
> - The on-call / change-management process for `apply` in production.

## File map

See [`.agent/CODEBASE.md`](./CODEBASE.md). The auto-generated section enumerates Terraform modules, Ansible roles, and playbooks. The hand-written section explains the environment topology and module dependencies.

## Corrections

See [`.agent/CORRECTIONS.md`](./CORRECTIONS.md). Read it before starting work and after any review.

## Conventions

### Terraform style

- Run `terraform fmt -recursive`. Don't fight the formatter.
- One resource per `.tf` file is overkill; group by logical concern (network, iam, compute). Files over ~300 lines should be split.
- Module inputs declared in `variables.tf`, outputs in `outputs.tf`, provider/backend in `versions.tf`. Don't scatter `variable {}` blocks across files.
- Pin provider versions in `versions.tf` (`>= 5.0, < 6.0`-style). Bare `>= 5.0` is a rollback hazard.
- Prefer `for_each` over `count` for collections of distinct resources; `count` makes the diff lie when items are reordered.

### Ansible style

- Roles are the unit of reuse. A playbook is a thin orchestration layer.
- Variables: `defaults/main.yml` for role defaults, `vars/main.yml` for role-fixed values, `group_vars/` / `host_vars/` for inventory-side overrides.
- Use `ansible.builtin.*` fully-qualified module names. Bare `copy`, `template`, `shell` collide with collection-shadowed modules.
- Templates use Jinja2 `{{ var }}` — that's why the playbook/inventory files in this scaffold stay as plain `.yml` (Go's templater would otherwise eat the braces).
- Avoid `command`/`shell` when a proper module exists. They break idempotency.

### Naming

- Terraform resources: `<provider>_<resource>.<purpose>` — e.g. `aws_iam_role.lambda_exec`.
- Ansible roles: kebab-case directory names (`roles/web-server/`).
- Files: `snake_case.tf` for Terraform, `kebab-case.yml` for Ansible playbooks/inventory.

### Commits

- Conventional commits: `feat:`, `fix:`, `refactor:`, `docs:`, `test:`, `chore:`.
- Tag plan-affecting changes clearly. A commit message saying "tidy" should not change the plan output.
- Never `--force-push` to shared branches.

### Dependencies

- Terraform: pin providers in `versions.tf`. Use `.terraform.lock.hcl` to lock provider hashes; commit it.
- Ansible: declare collection/role dependencies in `requirements.yml`. `post-create.sh` installs them.

## Testing

- Terraform: prefer `terraform validate` for syntax and `terraform plan` against a real backend for behaviour. `terraform test` (1.6+) for module unit tests.
- Ansible: `ansible-playbook --syntax-check`, `--check --diff`, and (for non-trivial roles) Molecule.
- Don't ship a module or role without at least a syntax-check passing in CI.

## What you should NOT do

- Do not commit secrets, API keys, SSH keys, vault passwords, or state files.
- Do not modify `.git/`, `.devcontainer/`, or `.agent/` files unless explicitly asked.
- Do not run `terraform apply` or `ansible-playbook` against production without operator confirmation in your turn.
- Do not run `git push --force` on `main` or any branch you didn't create yourself.
- Do not install global packages that aren't in the Dockerfile. If you need one, propose adding it.
- Do not bypass `check.sh` failures with `--no-verify` or by editing the script.
- Do not put real production hostnames or IPs into the committed inventory or fixtures.

## When you're stuck

Stop and ask. Specifically:

- If a `terraform plan` shows resource destruction you didn't intend, surface the full plan and stop.
- If a vault decrypt fails, do not commit unencrypted vars to "diagnose".
- If a check fails for a reason you don't understand, report the full output rather than guessing.

## Reviewer agent

After completing a non-trivial change:

```bash
./.agent/scripts/review.sh
```

This invokes a separate agent that reviews your diff against this file, `CORRECTIONS.md`, and `CODEBASE.md`. Output lands in `.agent/REVIEW.md`. Read it. Address legitimate findings.

## Project-specific notes

> **Add anything here that doesn't fit elsewhere.** Which backend bucket, which inventory file is canonical, the bastion's address, etc.
