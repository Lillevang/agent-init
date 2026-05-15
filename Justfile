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

# Lint shell scripts. shellcheck is installed via the devcontainer Dockerfile.
lint-shell:
    #!/usr/bin/env bash
    set -euo pipefail
    if ! command -v shellcheck >/dev/null 2>&1; then
        echo "ERROR: shellcheck not installed. Rebuild the devcontainer." >&2
        exit 1
    fi
    files=$(find . -name '*.sh' -not -path './.git/*' -not -path './testdata/*' -print)
    shellcheck $files

# Run govulncheck against the module. Soft signal — surfaces stdlib + dep CVEs.
vulncheck:
    #!/usr/bin/env bash
    set -euo pipefail
    if ! command -v govulncheck >/dev/null 2>&1; then
        echo "ERROR: govulncheck not installed. Rebuild the devcontainer." >&2
        exit 1
    fi
    govulncheck ./...

# Type check
typecheck:
    go vet ./...

# Verify go.mod / go.sum are tidy. Fails if `go mod tidy` would change them.
mod-tidy:
    #!/usr/bin/env bash
    set -euo pipefail
    if ! diff=$(go mod tidy -diff 2>&1); then
        echo "ERROR: go.mod or go.sum is not tidy. Run 'go mod tidy' and commit the result." >&2
        echo "$diff" >&2
        exit 1
    fi

# Run tests with the race detector enabled.
test:
    go test -race ./...

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
        # Fresh mode
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
        # Agents-only mode (only flavors that support it). Probed with --dry-run
        # so flavors without SupportsAgentsOnly exit cleanly here.
        # `just check` is intentionally NOT run inside the agents-only output:
        # the scaffold expects to be added to an existing project, so it
        # ships no go.mod / cmd/ and golangci-lint has nothing to check
        # against in a bare tempdir. We still regenerate the codemap because
        # the golden snapshot includes its output.
        if go run ./cmd/agent-init init --dry-run --no-git --agents-only "$flavor" "$tmp/_probe" >/dev/null 2>&1; then
            echo "→ smoke-testing flavor: $flavor (agents-only)"
            target="$tmp/$flavor-agents-only"
            go run ./cmd/agent-init init --no-git --agents-only "$flavor" "$target"
            test -f "$target/AGENTS.md" || test -f "$target/.agent/AGENTS.md"
            if [[ -x "$target/.agent/scripts/gen-codemap.sh" ]]; then
                (cd "$target" && ./.agent/scripts/gen-codemap.sh) >/dev/null
            fi
            diff -ruN "testdata/golden/$flavor-agents-only" "$target"
        fi
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
        # Fresh mode
        echo "→ regenerating golden for flavor: $flavor"
        target="$tmp/$flavor"
        go run ./cmd/agent-init init --no-git "$flavor" "$target"
        if [[ -f "$target/Justfile" ]]; then
            (cd "$target" && just check)
        fi
        rm -rf "testdata/golden/$flavor"
        cp -a "$target" "testdata/golden/$flavor"
        # Agents-only mode (only flavors that support it). See `smoke-test`
        # comment for why `just check` is skipped here.
        if go run ./cmd/agent-init init --dry-run --no-git --agents-only "$flavor" "$tmp/_probe" >/dev/null 2>&1; then
            echo "→ regenerating golden for flavor: $flavor (agents-only)"
            target="$tmp/$flavor-agents-only"
            go run ./cmd/agent-init init --no-git --agents-only "$flavor" "$target"
            if [[ -x "$target/.agent/scripts/gen-codemap.sh" ]]; then
                (cd "$target" && ./.agent/scripts/gen-codemap.sh) >/dev/null
            fi
            rm -rf "testdata/golden/$flavor-agents-only"
            cp -a "$target" "testdata/golden/$flavor-agents-only"
        fi
    done

# Invoke reviewer agent
review:
    ./.agent/scripts/review.sh
