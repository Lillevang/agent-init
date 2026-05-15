#!/usr/bin/env bash
# check.sh — the agent's "am I done" gate.
#
# Every step must succeed. If any step fails, the task is not complete.
# The agent must fix the failure, not bypass it.

set -euo pipefail

# Resolve repo root regardless of where this is called from
REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
cd "$REPO_ROOT"

# Colors (only if interactive)
if [[ -t 1 ]]; then
    RED=$'\033[31m'; GREEN=$'\033[32m'; YELLOW=$'\033[33m'; BOLD=$'\033[1m'; RESET=$'\033[0m'
else
    RED=""; GREEN=""; YELLOW=""; BOLD=""; RESET=""
fi

step() {
    local name="$1"; shift
    echo
    echo "${BOLD}→ $name${RESET}"
    if "$@"; then
        echo "${GREEN}✓ $name passed${RESET}"
    else
        echo "${RED}✗ $name failed${RESET}"
        exit 1
    fi
}

# Optional step: skip silently if recipe doesn't exist
maybe_step() {
    local recipe="$1"
    if just --list 2>/dev/null | grep -qE "^\s+${recipe}\b"; then
        step "$recipe" just "$recipe"
    else
        echo "${YELLOW}⊘ skipping '$recipe' (no Just recipe defined)${RESET}"
    fi
}

echo "${BOLD}Running done-gate checks for $(basename "$REPO_ROOT")${RESET}"

# Always regenerate the codemap so it reflects current state
if [[ -x .agent/scripts/gen-codemap.sh ]]; then
    step "codemap"  .agent/scripts/gen-codemap.sh
fi

# Generate clients from any OpenAPI specs present
maybe_step "generate-clients"

# Standard quality checks
maybe_step "fmt"
maybe_step "lint"
maybe_step "typecheck"
maybe_step "test"
maybe_step "playwright"

echo
echo "${GREEN}${BOLD}✓ all checks passed${RESET}"
