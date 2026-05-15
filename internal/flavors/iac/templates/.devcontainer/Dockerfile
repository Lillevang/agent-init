# Base image: Debian slim is the safest default for devcontainer compatibility.
FROM mcr.microsoft.com/devcontainers/base:debian-12

# The base image ships a `vscode` user (uid 1000) with passwordless sudo.
ARG USERNAME=vscode
ARG TERRAFORM_VERSION=1.9.8
ARG TFLINT_VERSION=v0.55.0
ARG TFSEC_VERSION=v1.28.11
ARG TRIVY_VERSION=0.58.0
ARG TERRAFORM_DOCS_VERSION=v0.18.0

# System packages: build tools, jq/curl, ripgrep/tree for the agent, ssh client
# for Ansible, sshpass for password-auth fallback (use ssh keys in real life).
RUN apt-get update && apt-get install -y --no-install-recommends \
        build-essential \
        ca-certificates \
        curl \
        git \
        gnupg \
        jq \
        less \
        openssh-client \
        ripgrep \
        sshpass \
        tree \
        unzip \
        xz-utils \
    && rm -rf /var/lib/apt/lists/*

# Python (for ansible + pre-commit + yamllint).
RUN apt-get update && apt-get install -y --no-install-recommends \
        python3 python3-pip python3-venv pipx \
    && rm -rf /var/lib/apt/lists/*

# Node.js 20 — only for the agent CLIs (claude-code, codex, devcontainers).
RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
    && apt-get install -y --no-install-recommends nodejs \
    && rm -rf /var/lib/apt/lists/*

# GitHub CLI
RUN curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg \
        -o /usr/share/keyrings/githubcli-archive-keyring.gpg \
    && chmod go+r /usr/share/keyrings/githubcli-archive-keyring.gpg \
    && echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" \
        > /etc/apt/sources.list.d/github-cli.list \
    && apt-get update \
    && apt-get install -y --no-install-recommends gh \
    && rm -rf /var/lib/apt/lists/*

# Terraform (HashiCorp APT repo for the version pinning).
RUN ARCH=$(dpkg --print-architecture) \
    && curl -fsSL "https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_${ARCH}.zip" -o /tmp/tf.zip \
    && unzip -q /tmp/tf.zip -d /usr/local/bin \
    && chmod +x /usr/local/bin/terraform \
    && rm /tmp/tf.zip

# tflint
RUN ARCH=$(dpkg --print-architecture) \
    && curl -fsSL "https://github.com/terraform-linters/tflint/releases/download/${TFLINT_VERSION}/tflint_linux_${ARCH}.zip" -o /tmp/tflint.zip \
    && unzip -q /tmp/tflint.zip -d /usr/local/bin \
    && chmod +x /usr/local/bin/tflint \
    && rm /tmp/tflint.zip

# tfsec (Terraform static security scanner — advisory; not part of done-gate by default).
RUN ARCH=$(dpkg --print-architecture) \
    && curl -fsSL "https://github.com/aquasecurity/tfsec/releases/download/${TFSEC_VERSION}/tfsec-linux-${ARCH}" -o /usr/local/bin/tfsec \
    && chmod +x /usr/local/bin/tfsec

# trivy (broader config scanner — also advisory).
RUN ARCH=$(dpkg --print-architecture) \
    && curl -fsSL "https://github.com/aquasecurity/trivy/releases/download/v${TRIVY_VERSION}/trivy_${TRIVY_VERSION}_Linux-$([ "$ARCH" = "amd64" ] && echo 64bit || echo ARM64).tar.gz" -o /tmp/trivy.tgz \
    && tar -xzf /tmp/trivy.tgz -C /usr/local/bin trivy \
    && chmod +x /usr/local/bin/trivy \
    && rm /tmp/trivy.tgz

# terraform-docs (for generating module READMEs).
RUN ARCH=$(dpkg --print-architecture) \
    && curl -fsSL "https://github.com/terraform-docs/terraform-docs/releases/download/${TERRAFORM_DOCS_VERSION}/terraform-docs-${TERRAFORM_DOCS_VERSION}-linux-${ARCH}.tar.gz" -o /tmp/tfd.tgz \
    && tar -xzf /tmp/tfd.tgz -C /usr/local/bin terraform-docs \
    && chmod +x /usr/local/bin/terraform-docs \
    && rm /tmp/tfd.tgz

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
ENV PATH="/home/${USERNAME}/.local/bin:${PATH}"

# Per-user Python toolchain: ansible-core, ansible-lint, yamllint, pre-commit.
# pipx isolates each tool so they can't break each other.
RUN pipx ensurepath \
    && pipx install ansible-core \
    && pipx install ansible-lint \
    && pipx install yamllint \
    && pipx install pre-commit

# Persist agent CLI auth/config across rebuilds.
RUN mkdir -p /home/${USERNAME}/.config/claude /home/${USERNAME}/.config/codex
