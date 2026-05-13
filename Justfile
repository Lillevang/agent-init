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

# Smoke-test scaffold output for every registered flavor
smoke-test:
    #!/usr/bin/env bash
    set -euo pipefail
    tmp=$(mktemp -d)
    trap 'rm -rf "$tmp"' EXIT
    flavors=$(go run ./cmd/agent-init list-flavors | awk '{print $1}')
    for flavor in $flavors; do
        echo "→ smoke-testing flavor: $flavor"
        target="$tmp/$flavor"
        go run ./cmd/agent-init init --no-git "$flavor" "$target"
        # Every flavor must write AGENTS.md somewhere — either at the root
        # (doc-collab flavors) or inside .agent/ (code flavors).
        test -f "$target/AGENTS.md" || test -f "$target/.agent/AGENTS.md"
        # Run the done-gate only when the flavor ships a Justfile.
        if [[ -f "$target/Justfile" ]]; then
            (cd "$target" && just check)
        fi
        diff -ruN "testdata/golden/$flavor" "$target"
    done

# Regenerate every flavor's smoke-test golden snapshot
smoke-test-update:
    #!/usr/bin/env bash
    set -euo pipefail
    tmp=$(mktemp -d)
    trap 'rm -rf "$tmp"' EXIT
    flavors=$(go run ./cmd/agent-init list-flavors | awk '{print $1}')
    mkdir -p testdata/golden
    for flavor in $flavors; do
        echo "→ regenerating golden for flavor: $flavor"
        target="$tmp/$flavor"
        go run ./cmd/agent-init init --no-git "$flavor" "$target"
        if [[ -f "$target/Justfile" ]]; then
            (cd "$target" && just check)
        fi
        rm -rf "testdata/golden/$flavor"
        cp -a "$target" "testdata/golden/$flavor"
    done

# Invoke reviewer agent
review:
    ./.agent/scripts/review.sh
