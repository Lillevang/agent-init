# agent-init

Scaffolds a repository for sandboxed agentic development with Claude Code, Codex, or similar tools. Drops in a devcontainer, an `AGENTS.md`/`CLAUDE.md`, a check-gate script, codemap scaffolding, a corrections file, a reviewer-agent invocation, OpenAPI client generation hooks, Playwright recording, and pre-commit wiring.

The goal is: clone a repo, run `agent-init`, edit a few files, then let the agent loose inside a container with bounded blast radius.

## What it sets up

```
your-project/
├── .devcontainer/
│   ├── devcontainer.json        # Podman/Docker compatible
│   ├── Dockerfile               # base image + agents installed
│   └── init-firewall.sh         # egress allowlist (optional)
├── .agent/
│   ├── AGENTS.md                # canonical agent instructions (Codex)
│   ├── CLAUDE.md                # symlink → AGENTS.md
│   ├── CODEBASE.md              # codemap (partly auto-generated)
│   ├── CORRECTIONS.md           # "don't do this" file
│   ├── REVIEW.md                # latest reviewer output (gitignored)
│   └── scripts/
│       ├── check.sh             # the agent's "am I done" gate
│       ├── review.sh            # invoke reviewer agent
│       ├── gen-codemap.sh       # regenerate auto sections of CODEBASE.md
│       └── record-feature.sh    # playwright + ffmpeg helper
├── apis/                        # OpenAPI specs go here
├── clients/                     # generated clients land here
├── vendor/                      # mount-point for sibling repos
├── AGENTS.md → .agent/AGENTS.md # top-level symlinks so tools find them
├── CLAUDE.md → .agent/CLAUDE.md
├── .pre-commit-config.yaml
├── Justfile
├── .gitignore
└── README.agent.md              # project-level doc (this is for the user)
```

## Dependencies

You install these on the **host** (Fedora WSL or Fedora bare-metal). Inside the container, the Dockerfile handles its own dependencies.

### Required

**1. Podman** (or Docker, but Podman is the Fedora-native choice)

```bash
sudo dnf install -y podman podman-docker
systemctl --user enable --now podman.socket
echo 'export DOCKER_HOST=unix://$XDG_RUNTIME_DIR/podman/podman.sock' >> ~/.bashrc
```

On Fedora WSL, ensure systemd is enabled. Add to `/etc/wsl.conf`:
```ini
[boot]
systemd=true
```
Then `wsl --shutdown` from Windows and restart the distro.

**2. Node.js + the devcontainer CLI**

```bash
sudo dnf install -y nodejs npm
npm config set prefix ~/.local
echo 'export PATH=$HOME/.local/bin:$PATH' >> ~/.bashrc
npm install -g @devcontainers/cli
```

Verify: `devcontainer --version`

**3. just** (command runner)

```bash
sudo dnf install -y just
```
If your Fedora is older and doesn't have `just` in dnf:
```bash
curl --proto '=https' --tlsv1.2 -sSf https://just.systems/install.sh | bash -s -- --to ~/.local/bin
```

**4. git** (you have it, just listing)

```bash
sudo dnf install -y git
```

### Optional but recommended

**5. pre-commit** — runs locally only if you want host-side hooks too. Inside the container it's installed automatically.

```bash
pipx install pre-commit
# or: sudo dnf install -y pre-commit
```

**6. GitHub CLI** — if you want the agent to interact with PRs/issues.

```bash
sudo dnf install -y gh
gh auth login
```

**7. Helix** — your editor of choice for inspecting code from the host terminal.

```bash
sudo dnf install -y helix
```

## Installing agent-init itself

```bash
git clone <this-repo> ~/.local/share/agent-init
ln -s ~/.local/share/agent-init/agent-init ~/.local/bin/agent-init
```

Make sure `~/.local/bin` is on your `PATH`.

## Usage

```bash
# In an existing or empty repo:
cd my-project
agent-init

# Or scaffold into a fresh directory:
agent-init my-new-project
cd my-new-project
```

Flags:
- `--force` — overwrite existing files (default: skip)
- `--no-git` — don't `git init` if not already a repo

After scaffolding:

1. Read `README.agent.md` in the project for project-side documentation.
2. Edit `.agent/AGENTS.md` to describe **this project's** stack, conventions, and quirks. The template is generic.
3. Edit `.devcontainer/Dockerfile` to add language toolchains you need (Rust, Go, Python, etc.).
4. Optionally enable the firewall: in `devcontainer.json` flip `"postCreateCommand"` to call `init-firewall.sh`.
5. `devcontainer up --workspace-folder .`
6. `devcontainer exec --workspace-folder . bash` and run `just check` to confirm the gate works.
7. Inside the container, set your API keys and start the agent:
   ```bash
   export ANTHROPIC_API_KEY=...
   claude
   ```

## Mounting sibling repos (fake monorepo)

Edit `.devcontainer/devcontainer.json` and add to `mounts`:

```json
"mounts": [
  "source=${localEnv:HOME}/code/sibling-repo,target=/workspaces/vendor/sibling-repo,type=bind,readonly"
]
```

The `readonly` is important if you don't want the agent to commit there by mistake. Remove it if the agent legitimately needs to edit.

## What this tool does NOT do

- It doesn't pick a language stack for you. Edit the Dockerfile.
- It doesn't generate the *prose* of `CODEBASE.md` for an existing codebase — only the directory tree section. You write the rest, or have an agent do an initial pass.
- It doesn't configure SpecKit. Add it manually if/when a feature is large enough to warrant it.
- It doesn't install agent CLIs on the host. They run inside the container.

## Updating an existing scaffolded project

`agent-init --force` overwrites all template files, including any local edits to `AGENTS.md`. **Don't do this casually.** The intended workflow:

1. Keep your `agent-init` template repo as the source of truth.
2. When you improve a template, copy the relevant file into your project manually.
3. Or: keep project-specific overrides at the bottom of `AGENTS.md` under a clearly marked section, and only re-pull the top half.

A future version may do three-way merging. It does not today.

## License

Whatever you want. Make it yours.
