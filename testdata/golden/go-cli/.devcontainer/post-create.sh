#!/usr/bin/env bash
# Runs once after the container is created.
set -euo pipefail

cd "$(git rev-parse --show-toplevel 2>/dev/null || pwd)"

# Install pre-commit hooks if config exists
if [[ -f .pre-commit-config.yaml ]]; then
    pre-commit install --install-hooks || echo "pre-commit install failed (non-fatal)"
fi

# Generate initial codemap if scripts present
if [[ -x .agent/scripts/gen-codemap.sh ]]; then
    .agent/scripts/gen-codemap.sh || echo "codemap generation failed (non-fatal)"
fi

# Warm the Go module cache.
if [[ -f go.mod ]]; then
    go mod download || echo "go mod download failed (non-fatal)"
fi

echo "✓ post-create complete. Run 'just' to see available commands."
