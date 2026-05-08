# Base image: Debian slim is the safest default for devcontainer compatibility.
# Switch to fedora:40 or similar if you prefer; just adjust the package commands.
FROM mcr.microsoft.com/devcontainers/base:debian-12

# The base image ships a `vscode` user (uid 1000) with passwordless sudo.
ARG USERNAME=vscode

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
        fd-find \
        ffmpeg \
        unzip \
    && rm -rf /var/lib/apt/lists/*

# Node.js 20 (for the agent CLIs and Playwright)
RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
    && apt-get install -y --no-install-recommends nodejs \
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

# OpenAPI generators (pick what you need; both are fine to have)
RUN npm install -g \
        @openapitools/openapi-generator-cli \
        openapi-typescript

# Playwright with browsers (heavy; comment out if you don't need frontend testing)
RUN npm install -g playwright && npx playwright install --with-deps chromium

# post-create helper
COPY post-create.sh /usr/local/bin/post-create.sh
RUN chmod +x /usr/local/bin/post-create.sh

USER ${USERNAME}
ENV PATH="/home/${USERNAME}/.local/bin:${PATH}"

# Pre-commit, installed for the dev user so it lands on PATH for the agent.
RUN pipx install pre-commit && pipx ensurepath

# Persist agent CLI auth/config across rebuilds (bind-mount these from the host
# in devcontainer.json if you want auth to survive container rebuilds).
RUN mkdir -p /home/${USERNAME}/.config/claude /home/${USERNAME}/.config/codex
