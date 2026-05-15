# Base image: Debian slim is the safest default for devcontainer compatibility.
# Switch to fedora:40 or similar if you prefer; just adjust the package commands.
FROM mcr.microsoft.com/devcontainers/base:debian-12

# The base image ships a `vscode` user (uid 1000) with passwordless sudo.
ARG USERNAME=vscode
ARG GO_VERSION=1.26.3
ARG GOLANGCI_LINT_VERSION=v2.12.2

# System packages
RUN apt-get update && apt-get install -y --no-install-recommends \
        build-essential \
        ca-certificates \
        curl \
        git \
        gnupg \
        jq \
        less \
        ripgrep \
        shellcheck \
        tree \
        fd-find \
        unzip \
        xz-utils \
    && rm -rf /var/lib/apt/lists/*

# Go toolchain for this repository.
RUN curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" -o /tmp/go.tgz \
    && rm -rf /usr/local/go \
    && tar -C /usr/local -xzf /tmp/go.tgz \
    && rm /tmp/go.tgz
ENV PATH="/usr/local/go/bin:${PATH}"

# Go lint/format/vuln helpers used by the local done gate.
RUN curl -fsSL https://golangci-lint.run/install.sh \
        | sh -s -- -b /usr/local/bin "${GOLANGCI_LINT_VERSION}" \
    && GOBIN=/usr/local/bin go install golang.org/x/tools/cmd/goimports@v0.30.0 \
    && GOBIN=/usr/local/bin go install golang.org/x/vuln/cmd/govulncheck@v1.1.4

# Node.js 20 — only for the agent CLIs (claude-code, codex, devcontainers).
RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
    && apt-get install -y --no-install-recommends nodejs \
    && rm -rf /var/lib/apt/lists/*

# GitHub CLI (for `gh pr ...` from agents and humans alike)
RUN curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg \
        -o /usr/share/keyrings/githubcli-archive-keyring.gpg \
    && chmod go+r /usr/share/keyrings/githubcli-archive-keyring.gpg \
    && echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" \
        > /etc/apt/sources.list.d/github-cli.list \
    && apt-get update \
    && apt-get install -y --no-install-recommends gh \
    && rm -rf /var/lib/apt/lists/*

# Python (for pre-commit and tooling). Adjust versions as needed.
RUN apt-get update && apt-get install -y --no-install-recommends \
        python3 python3-pip python3-venv pipx \
    && rm -rf /var/lib/apt/lists/*

# just (command runner)
RUN curl --proto '=https' --tlsv1.2 -sSf https://just.systems/install.sh \
    | bash -s -- --to /usr/local/bin

# Agent CLIs
RUN npm install -g \
        @anthropic-ai/claude-code \
        @openai/codex \
        @devcontainers/cli

# post-create helper
COPY post-create.sh /usr/local/bin/post-create.sh
RUN chmod +x /usr/local/bin/post-create.sh

USER ${USERNAME}
ENV PATH="/home/${USERNAME}/go/bin:/home/${USERNAME}/.local/bin:${PATH}"

# Pre-commit, installed for the dev user so it lands on PATH for the agent.
RUN pipx install pre-commit && pipx ensurepath

# Persist agent CLI auth/config across rebuilds (bind-mount these from the host
# in devcontainer.json if you want auth to survive container rebuilds).
RUN mkdir -p /home/${USERNAME}/.config/claude /home/${USERNAME}/.config/codex
