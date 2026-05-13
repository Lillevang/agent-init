#!/usr/bin/env bash
# check.sh — done gate for the agent-init Go CLI.

set -euo pipefail

REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
cd "$REPO_ROOT"

if [[ -t 1 ]]; then
    RED=$'\033[31m'; GREEN=$'\033[32m'; YELLOW=$'\033[33m'; BOLD=$'\033[1m'; RESET=$'\033[0m'
else
    RED=""; GREEN=""; YELLOW=""; BOLD=""; RESET=""
fi

step() {
    local name="$1"; shift
    echo
    echo "${BOLD}-> $name${RESET}"
    if "$@"; then
        echo "${GREEN}✓ $name passed${RESET}"
    else
        echo "${RED}✗ $name failed${RESET}"
        exit 1
    fi
}

# soft_step runs a check but only warns on failure; useful for advisory
# signals (vuln scans, etc.) where action may not be immediate.
soft_step() {
    local name="$1"; shift
    echo
    echo "${BOLD}-> $name${RESET}"
    if "$@"; then
        echo "${GREEN}✓ $name passed${RESET}"
    else
        echo "${YELLOW}⚠ $name had findings (advisory — not failing the gate)${RESET}"
    fi
}

require() {
    local tool="$1"
    if ! command -v "$tool" >/dev/null 2>&1; then
        echo "${RED}ERROR: required tool '$tool' is not installed or not on PATH.${RESET}" >&2
        echo "Rebuild the devcontainer if this is a fresh checkout." >&2
        exit 1
    fi
}

echo "${BOLD}Running done-gate checks for agent-init${RESET}"

require go
require just

step "codemap" .agent/scripts/gen-codemap.sh
step "fmt" just fmt
step "lint" just lint
step "lint-shell" just lint-shell
step "mod-tidy" just mod-tidy
step "typecheck" just typecheck
step "test" just test
step "cross-build" just cross-build
step "smoke-test" just smoke-test
soft_step "vulncheck" just vulncheck

echo
echo "${GREEN}${BOLD}✓ all checks passed${RESET}"
