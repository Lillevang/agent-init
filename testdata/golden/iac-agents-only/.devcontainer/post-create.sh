#!/usr/bin/env bash
# Runs once after the container is created.
set -euo pipefail

cd "$(git rev-parse --show-toplevel 2>/dev/null || pwd)"

# Install pre-commit hooks if config exists.
if [[ -f .pre-commit-config.yaml ]]; then
    pre-commit install --install-hooks || echo "pre-commit install failed (non-fatal)"
fi

# Generate initial codemap.
if [[ -x .agent/scripts/gen-codemap.sh ]]; then
    .agent/scripts/gen-codemap.sh || echo "codemap generation failed (non-fatal)"
fi

# Terraform: init each module with -backend=false so `terraform validate`
# (used by `just typecheck`) doesn't require backend credentials. Real
# applies still need a proper init against the configured backend.
if command -v terraform >/dev/null 2>&1; then
    while IFS= read -r -d '' tfdir; do
        echo "→ terraform init -backend=false in $tfdir"
        (cd "$tfdir" && terraform init -backend=false -input=false -upgrade) \
            || echo "  (init failed in $tfdir, non-fatal)"
    done < <(find . -type d \
                  \( -path './.git' -o -path './.terraform' -o -path '*/.terraform' \) -prune -o \
                  -type f -name '*.tf' -printf '%h\0' 2>/dev/null \
              | sort -zu)
fi

# Ansible: install collection/role dependencies if declared.
if [[ -f requirements.yml ]] && command -v ansible-galaxy >/dev/null 2>&1; then
    ansible-galaxy install -r requirements.yml || echo "ansible-galaxy install failed (non-fatal)"
fi
if [[ -f ansible/requirements.yml ]] && command -v ansible-galaxy >/dev/null 2>&1; then
    ansible-galaxy install -r ansible/requirements.yml || echo "ansible-galaxy install failed (non-fatal)"
fi

echo "✓ post-create complete. Run 'just' to see available commands."
