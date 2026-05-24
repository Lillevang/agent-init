# Releases (tag-driven)

`agent-init` cuts public GitHub Releases from semver tags. Pushing a tag that
matches `v*` (for example `v1.2.3`) is the only thing that publishes a release.
Pushes and merges to `main` run CI (check + build) but never publish.

This is the behavior decided in
[#28](https://github.com/Lillevang/agent-init/issues/28) and tracked in
[#37](https://github.com/Lillevang/agent-init/issues/37).

## How it works

The release workflow ([.github/workflows/release.yml](../../.github/workflows/release.yml))
triggers on two events:

- `push` to `main` — runs the `check` and `build` jobs only.
- `push` of a tag matching `v*` — runs `check`, `build`, and `release`.

The `release` job is gated with `if: startsWith(github.ref, 'refs/tags/v')`, so
it is skipped entirely on main pushes. The release tag, name, and body all come
from the pushed tag via `github.ref_name`, not from `github.run_number`.

## Cutting a release

```bash
git tag v1.2.3
git push origin v1.2.3
```

This builds binaries for Linux `amd64`, Linux `arm64`, macOS `arm64`, and
Windows `amd64`, packages them as `.tar.gz` (Linux and macOS) and `.zip`
(Windows) with a `checksums.txt`, and attaches them to a release named
`agent-init v1.2.3`.

Use a real semver tag. The `version` subcommand embeds the commit and build
date via `-ldflags`; the tag itself is the human-facing version.

## Build provenance

Each published archive carries a signed SLSA build provenance attestation,
produced by the `release` job via `actions/attest-build-provenance`. The
attestation binds the archive to the workflow run and the commit it was built
from, so a `checksums.txt` substitution by anyone who can publish a release no
longer goes undetected.

Verify a downloaded asset:

```bash
gh attestation verify agent-init-linux-amd64.tar.gz --repo Lillevang/agent-init
```

The attestation subject is the archive set (`dist/*.tar.gz`, `dist/*.zip`); the
step runs after checksums and before the release is cut, so the exact files that
get uploaded are the ones attested. The `checksums.txt` file is not attested —
it only proves integrity against itself.

## Permissions

The workflow runs with `contents: read` by default. Only the `release` job
opts up to `contents: write`, since it is the sole job that publishes. It also
holds `id-token: write` and `attestations: write`, the minimal scopes
`actions/attest-build-provenance` needs to mint a signed provenance statement;
these stay confined to the tag-gated `release` job. The `check` and `build`
jobs — which also run on main pushes — stay least-privilege.

## Source

- Trigger, permissions, and the tag gate: [.github/workflows/release.yml](../../.github/workflows/release.yml)
- CI on pull requests: [.github/workflows/pr.yml](../../.github/workflows/pr.yml)
- Shared setup for the check job: [.github/actions/setup/action.yml](../../.github/actions/setup/action.yml)
