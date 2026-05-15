# iac — Agentic Development Setup

This repo was scaffolded with `agent-init` (`iac` flavor). It's configured for sandboxed agentic development of Infrastructure-as-Code: Terraform, Ansible, or both. Agents run inside a devcontainer that ships `terraform`, `tflint`, `tfsec`, `trivy`, `ansible-core`, `ansible-lint`, and `yamllint`, gated by a check script with a codemap and corrections file.

## Quick start

```bash
# 1. (Once) install host dependencies — see "Host dependencies" below

# 2. Set API keys in your host shell
export ANTHROPIC_API_KEY=...
export OPENAI_API_KEY=...

# 3. Bring up the container
devcontainer up --workspace-folder .

# 4. Open a shell in it
devcontainer exec --workspace-folder . bash

# 5. Inside the container — run the agent
claude
# or: codex
```

## Host dependencies

You need these on the **host**. The container handles its own internals.

### Required

| Tool | Install |
|------|---------|
| **Podman** or Docker | `sudo dnf install -y podman podman-docker` (Fedora/WSL) |
| **Node.js + npm** | needed only for the devcontainer CLI on the host |
| **devcontainer CLI** | `npm install -g @devcontainers/cli` |
| **just** | `sudo dnf install -y just` |
| **git** | `sudo dnf install -y git` |

### Optional but recommended for IaC work

| Tool | Why |
|------|-----|
| **AWS CLI / gcloud / Azure CLI** | host-side `aws sso login` etc. before bringing the container up |
| **ssh-agent** | preferred over bind-mounting `~/.ssh` into the container |
| **pre-commit** | run hooks on the host too |

## Credential mounts

The devcontainer is shipped with **all credential mounts commented out**. Uncomment the ones you actually need in `.devcontainer/devcontainer.json` — every one of them gives the container read access to host credentials that `terraform apply` and Ansible can use.

The available mounts:

| Mount | Purpose |
|-------|---------|
| `~/.aws` | AWS CLI credentials and profiles |
| `~/.config/gcloud` | Google Cloud SDK credentials |
| `~/.azure` | Azure CLI cache |
| `~/.ssh` | SSH keys for Ansible — prefer ssh-agent forwarding over this in production |

All mounts are `,readonly`. Tools that need to write credentials (`aws sso login`, `gcloud auth login`) run on the host.

## Layout

```
.
├── .devcontainer/         # container definition (Dockerfile, devcontainer.json)
├── .agent/                # everything the agent reads
│   ├── AGENTS.md          # instructions (Codex)
│   ├── CLAUDE.md          # symlink → AGENTS.md (Claude Code)
│   ├── CODEBASE.md        # IaC-aware codemap (modules + roles + playbooks)
│   ├── CORRECTIONS.md     # known anti-patterns
│   └── scripts/           # check.sh, review.sh, gen-codemap.sh (iac-aware)
├── terraform/
│   ├── main.tf            # root module
│   ├── variables.tf
│   ├── outputs.tf
│   ├── versions.tf        # provider version pins
│   └── modules/           # reusable child modules
├── ansible/
│   ├── inventory/
│   │   └── hosts.yml.example   # commit this; real hosts.yml is gitignored
│   ├── playbooks/
│   │   └── site.yml
│   ├── roles/
│   └── requirements.yml   # collections + role deps
├── ansible.cfg            # default inventory, roles_path, ssh args
├── .tflint.hcl            # tflint config
├── .yamllint.yml          # yamllint config
├── .ansible-lint          # ansible-lint config
├── Justfile               # fmt / lint / typecheck / test / security
├── .pre-commit-config.yaml
├── .gitignore             # blocks tfstate, vault passwords, real inventory
└── README.agent.md        # this file
```

After scaffolding:

1. Edit `terraform/versions.tf` to pin the providers you actually use.
2. Edit `.devcontainer/devcontainer.json` and uncomment the credential mounts you need.
3. Edit `.agent/AGENTS.md` to describe THIS project's environments, state backend, and on-call process.
4. Configure a remote state backend in `terraform/versions.tf` (or per-environment) before any real `apply`.
5. Run `just check` to confirm the gate passes on a fresh tree.

## The done-gate

The agent considers itself done only when `just check` (a.k.a. `.agent/scripts/check.sh`) passes. For this flavor that's:

1. Codemap regeneration (IaC-aware: lists Terraform modules and Ansible roles)
2. `just fmt` — `terraform fmt -recursive`
3. `just lint` — `tflint` + `ansible-lint` + `yamllint`
4. `just typecheck` — `terraform validate` per module, `ansible-playbook --syntax-check` per playbook
5. `just test` — `terraform test` and/or `molecule test` if those directories exist

Missing recipes and missing tools are skipped silently, so the same Justfile works in Terraform-only, Ansible-only, and mixed repos.

### Security scans

`just security` runs `tfsec` and `trivy config`. It is **not** part of the default done-gate because the noise/signal ratio is project-specific. Wire it into CI as a separate job and triage findings there.

## Reviewer agent

After non-trivial changes:

```bash
just review
# or: REVIEWER=codex just review
```

Output lands in `.agent/REVIEW.md` (gitignored). It's a separate agent reading the diff against `main`, with read-only tool access. Catches violations of `AGENTS.md` and `CORRECTIONS.md`.

## State, secrets, and other footguns

This flavor's `AGENTS.md` enumerates them. The short version:

- `*.tfstate*` and `.terraform/` are gitignored — never override.
- `*.pem`, `*.key`, `.vault_pass*` are gitignored — never override.
- `inventory/hosts.yml` is gitignored; commit `inventory/hosts.yml.example`.
- Configure a remote state backend before the first shared-environment `apply`.
- Always `terraform plan` before `apply`; always `--check --diff` before a non-trivial Ansible run.

## Updating the scaffold

`agent-init --force` overwrites template files including local edits. Don't run it casually. When you improve a template, copy the file manually, or keep project-specific overrides clearly marked at the bottom of `AGENTS.md`.
