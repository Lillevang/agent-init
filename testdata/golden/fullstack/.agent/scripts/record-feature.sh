#!/usr/bin/env bash
# record-feature.sh — run a Playwright test and save the recording as mp4.
#
# Usage: record-feature.sh <feature-name> [test-file]
#
# - <feature-name> is used for the output filename
# - [test-file] defaults to tests/<feature-name>.spec.ts
#
# Requires: playwright, ffmpeg. Both are in the devcontainer Dockerfile.

set -euo pipefail

REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
cd "$REPO_ROOT"

if [[ $# -lt 1 ]]; then
    echo "Usage: $0 <feature-name> [test-file]" >&2
    exit 1
fi

FEATURE="$1"
TEST_FILE="${2:-tests/${FEATURE}.spec.ts}"
OUT_DIR=".agent/recordings/${FEATURE}"

if [[ ! -f "$TEST_FILE" ]]; then
    echo "ERROR: test file not found: $TEST_FILE" >&2
    echo "Hint: pass the path explicitly as the second argument." >&2
    exit 1
fi

mkdir -p "$OUT_DIR"

# Sanity-check tooling
for tool in npx ffmpeg; do
    if ! command -v "$tool" >/dev/null 2>&1; then
        echo "ERROR: '$tool' not installed. Inside the devcontainer it should be." >&2
        exit 1
    fi
done

echo "→ Running Playwright test with video capture..."

# Override the playwright config to force video recording
# This works whether or not the project's own config sets video:on
PLAYWRIGHT_OUTPUT_DIR="$OUT_DIR/raw" \
    npx playwright test "$TEST_FILE" \
        --output "$OUT_DIR/raw" \
        --reporter=list \
        --config=<(cat <<'EOF'
import { defineConfig } from '@playwright/test';
export default defineConfig({
  use: {
    video: 'on',
    trace: 'on',
  },
  retries: 0,
  reporter: 'list',
});
EOF
) || {
    echo "⚠ test failed; recording still saved" >&2
}

# Find the most recent .webm in the output dir and convert to mp4
WEBM=$(find "$OUT_DIR/raw" -name '*.webm' -type f -printf '%T@ %p\n' 2>/dev/null \
    | sort -rn | head -1 | cut -d' ' -f2-)

if [[ -z "$WEBM" ]]; then
    echo "ERROR: no .webm output found. Did the test run?" >&2
    exit 1
fi

MP4="$OUT_DIR/${FEATURE}-$(date +%Y%m%d-%H%M%S).mp4"
ffmpeg -y -loglevel error -i "$WEBM" -c:v libx264 -preset fast -crf 23 "$MP4"

echo
echo "✓ recording saved: $MP4"
echo
echo "Watch it before declaring the feature done."
