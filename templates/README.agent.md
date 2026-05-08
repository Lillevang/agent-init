# __PROJECT_NAME__ — Agentic Development Setup

This repo was scaffolded with `agent-init`. It's configured for sandboxed agentic development: agents run inside a devcontainer, gated by a check script, with a codemap and corrections file to keep them on-rails.

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

You need these on the **host** (Fedora WSL or Fedora bare-metal). The container handles its own internals.

### Required

| Tool | Install |
|------|---------|
| **Podman** | `sudo dnf install -y podman podman-docker` |
| **Node.js + npm** | `sudo dnf install -y nodejs npm` |
| **devcontainer CLI** | `npm install -g @devcontainers/cli` |
| **just** | `sudo dnf install -y just` |
| **git** | `sudo dnf install -y git` |

Configure Podman as a Docker drop-in (devcontainer CLI talks to the Docker socket):

```bash
systemctl --user enable --now podman.socket
echo 'export DOCKER_HOST=unix://$XDG_RUNTIME_DIR/podman/podman.sock' >> ~/.bashrc
```

On **Fedora WSL**, ensure systemd is enabled. In `/etc/wsl.conf`:
```ini
[boot]
systemd=true
```
Then `wsl --shutdown` from Windows PowerShell and re-open the distro.

### Optional

| Tool | Why | Install |
|------|-----|---------|
| **Helix** | Inspect code from the host terminal | `sudo dnf install -y helix` |
| **GitHub CLI** | Agent can interact with PRs/issues | `sudo dnf install -y gh && gh auth login` |
| **pre-commit** | Run hooks on the host too | `pipx install pre-commit` |

## Layout

```
.
├── .devcontainer/         # container definition (Dockerfile, devcontainer.json)
├── .agent/                # everything the agent reads
│   ├── AGENTS.md          # instructions (Codex)
│   ├── CLAUDE.md          # symlink → AGENTS.md (Claude Code)
│   ├── CODEBASE.md        # codemap (auto + hand-written sections)
│   ├── CORRECTIONS.md     # known anti-patterns
│   └── scripts/           # check.sh, review.sh, gen-codemap.sh, record-feature.sh
├── apis/                  # OpenAPI specs — clients generated from these
├── clients/               # generated clients (don't hand-edit)
├── vendor/                # mount-point for sibling repos (fake monorepo)
├── Justfile               # check, fmt, lint, typecheck, test, etc.
├── .pre-commit-config.yaml
└── README.agent.md        # this file
```

## The done-gate

The agent considers itself done only when `just check` (a.k.a. `.agent/scripts/check.sh`) passes. This runs:

1. Codemap regeneration
2. API client generation (if `apis/*.yaml` exist)
3. Format
4. Lint
5. Type check
6. Tests

Edit the `Justfile` recipes to plug in the actual tools for your stack. Recipes that don't exist are skipped silently — but **CI must run the same gate**, so don't leave it empty.

## Reviewer agent

After non-trivial changes:

```bash
just review
# or: REVIEWER=codex just review
```

Output lands in `.agent/REVIEW.md` (gitignored). It's a separate agent reading the diff against `main`, with read-only tool access. It catches violations of `AGENTS.md` and `CORRECTIONS.md` — useful, but not a substitute for you reading the diff yourself.

Override the base ref:
```bash
BASE_REF=develop just review
```

## Frontend recordings

For frontend features, record the agent's Playwright run to mp4:

```bash
just record my-feature
# expects tests/my-feature.spec.ts
```

Watch the recording before declaring done. The video lives in `.agent/recordings/<feature>/`.

## Mounting sibling repos (monorepo simulation)

Edit `.devcontainer/devcontainer.json`, add to `mounts`:

```json
"mounts": [
  "source=${localEnv:HOME}/code/sibling-repo,target=/workspaces/vendor/sibling-repo,type=bind,readonly"
]
```

Use `,readonly` if the agent shouldn't write there. Drop it if cross-repo edits are legitimate.

The agent sees these under `vendor/`. Mention them explicitly in `.agent/AGENTS.md` so the agent knows what's read-only context vs. in-scope code.

## Egress firewall (optional, recommended)

`.devcontainer/init-firewall.sh` drops all egress except an allowlist (Anthropic API, OpenAI API, npm, pypi, github, etc.). To enable:

1. Review the allowlist in `init-firewall.sh` and adjust.
2. In `devcontainer.json`, change `postCreateCommand` to:
   ```json
   "postCreateCommand": "sudo /usr/local/bin/init-firewall.sh && /usr/local/bin/post-create.sh"
   ```
3. Rebuild the container.

DNS is allowed; only HTTP(S) egress to listed hosts succeeds. Dropped packets are logged via iptables — check with `sudo dmesg | grep 'FIREWALL DROP'` to see what got blocked.

## Helix from the host

The repo is bind-mounted into the container, so editing from the host with Helix (or anything else) Just Works. The agent inside the container and you outside see the same bytes. No SSH, no remote editing dance.

```bash
hx .
```

## SpecKit (for larger features)

Not installed by default. When you have a feature big enough to warrant a spec-driven flow:

```bash
# In the container
npm install -g @githubnext/spec-kit
# then follow SpecKit's own setup
```

Don't reach for this on small changes. It's overhead.

## Updating the scaffold

`agent-init --force` overwrites template files including local edits. Don't run it casually. The recommended approach: when you improve a template, copy the file manually, or keep project-specific overrides clearly marked at the bottom of `AGENTS.md`.

## Troubleshooting

**"podman: command not found" or socket errors on WSL**
Make sure systemd is enabled (`/etc/wsl.conf`) and the user socket is running:
```bash
systemctl --user status podman.socket
```

**SELinux permission denied on bind mounts (bare-metal Fedora)**
The devcontainer CLI usually handles this, but if you mount things manually, add `:Z` (private label) or `:z` (shared label) to the mount.

**Agent can't reach the API**
If you enabled the firewall, check `sudo dmesg | grep DROP`. If a host you need is being blocked, add it to `init-firewall.sh` and re-run the container.

**Pre-commit hook is slow**
The `pre-push` `just check` hook runs the full gate. Move it to `pre-commit` if you want faster feedback per commit, or drop it entirely and rely on CI.
