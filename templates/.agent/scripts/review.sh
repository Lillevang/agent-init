#!/usr/bin/env bash
# review.sh — invoke a separate agent to review the current diff.
#
# The reviewer reads AGENTS.md, CORRECTIONS.md, and CODEBASE.md, then comments
# on the diff against $BASE_REF (default: main). Output goes to .agent/REVIEW.md.
#
# Set REVIEWER=claude (default) or REVIEWER=codex to choose.

set -euo pipefail

REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
cd "$REPO_ROOT"

BASE_REF="${BASE_REF:-main}"
REVIEWER="${REVIEWER:-claude}"
OUTPUT=".agent/REVIEW.md"

# Ensure base ref exists
if ! git rev-parse --verify "$BASE_REF" >/dev/null 2>&1; then
    echo "ERROR: base ref '$BASE_REF' does not exist" >&2
    exit 1
fi

DIFF=$(git diff "$BASE_REF"...HEAD)
if [[ -z "$DIFF" ]]; then
    echo "No diff against $BASE_REF. Nothing to review."
    exit 0
fi

PROMPT=$(cat <<EOF
You are a code reviewer. Review the diff below against:
- .agent/AGENTS.md (project conventions)
- .agent/CORRECTIONS.md (known anti-patterns)
- .agent/CODEBASE.md (architecture)

Read those three files first, then review the diff.

For each issue, output:
- **File:line** — the location
- **Severity** — blocker / major / minor / nit
- **Issue** — what's wrong
- **Suggested fix** — concrete change

Then a one-paragraph summary at the end.

Only flag real issues. Don't pad. If the diff is fine, say so in one sentence.

You are read-only. Do not modify files.

Diff:
\`\`\`diff
$DIFF
\`\`\`
EOF
)

mkdir -p .agent
echo "→ Running reviewer ($REVIEWER) against $BASE_REF..."

case "$REVIEWER" in
    claude)
        if ! command -v claude >/dev/null 2>&1; then
            echo "ERROR: claude CLI not found" >&2
            exit 1
        fi
        # Read-only tool set. Adjust to current Claude Code flags if needed.
        claude -p "$PROMPT" \
            --allowedTools "Read,Grep,Glob,Bash(git diff:*),Bash(git log:*),Bash(git show:*)" \
            > "$OUTPUT"
        ;;
    codex)
        if ! command -v codex >/dev/null 2>&1; then
            echo "ERROR: codex CLI not found" >&2
            exit 1
        fi
        codex exec --sandbox read-only "$PROMPT" > "$OUTPUT"
        ;;
    *)
        echo "ERROR: unknown REVIEWER='$REVIEWER' (use claude or codex)" >&2
        exit 1
        ;;
esac

echo
echo "✓ review written to $OUTPUT"
echo
echo "--- Summary ---"
tail -20 "$OUTPUT"
