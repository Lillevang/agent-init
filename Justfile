# Justfile for agent-init, a Go CLI.

set shell := ["bash", "-uc"]

# Default: list recipes
default:
    @just --list

# Full done-gate (what the agent runs)
check:
    ./.agent/scripts/check.sh

# Regenerate codemap
codemap:
    ./.agent/scripts/gen-codemap.sh

# Format Go code
fmt:
    #!/usr/bin/env bash
    set -euo pipefail
    files=$(find cmd internal test -name '*.go' -print)
    gofmt -w $files
    if command -v goimports >/dev/null 2>&1; then
        goimports -w $files
    fi

# Lint Go code
lint:
    #!/usr/bin/env bash
    set -euo pipefail
    if ! command -v golangci-lint >/dev/null 2>&1; then
        echo "ERROR: golangci-lint not installed. Rebuild the devcontainer." >&2
        exit 1
    fi
    golangci-lint run

# Type check
typecheck:
    go vet ./...

# Run tests
test:
    go test ./...

# Build supported release targets
cross-build:
    GOOS=linux GOARCH=amd64 go build -o /tmp/agent-init-linux-amd64 ./cmd/agent-init
    GOOS=darwin GOARCH=arm64 go build -o /tmp/agent-init-darwin-arm64 ./cmd/agent-init

# Smoke-test scaffold output for the fullstack flavor
smoke-test:
    #!/usr/bin/env bash
    set -euo pipefail
    tmp=$(mktemp -d)
    trap 'rm -rf "$tmp"' EXIT
    target="$tmp/fullstack"
    go run ./cmd/agent-init init --no-git fullstack "$target"
    test -f "$target/.agent/AGENTS.md"
    test -L "$target/AGENTS.md"
    test -x "$target/.agent/scripts/check.sh"
    (cd "$target" && just check)
    diff -ruN testdata/golden/fullstack "$target"

# Regenerate the fullstack smoke-test golden snapshot
smoke-test-update:
    #!/usr/bin/env bash
    set -euo pipefail
    tmp=$(mktemp -d)
    trap 'rm -rf "$tmp"' EXIT
    target="$tmp/fullstack"
    go run ./cmd/agent-init init --no-git fullstack "$target"
    (cd "$target" && just check)
    rm -rf testdata/golden/fullstack
    mkdir -p testdata/golden
    cp -a "$target" testdata/golden/fullstack

# Invoke reviewer agent
review:
    ./.agent/scripts/review.sh
