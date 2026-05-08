# Vendored / mounted-in repos

This directory is the mount point for sibling repos that the agent should be able to read for context — simulating a monorepo without merging the histories.

Configure mounts in `.devcontainer/devcontainer.json`. Example:

```json
"mounts": [
  "source=${localEnv:HOME}/code/other-service,target=/workspaces/__PROJECT_NAME__/vendor/other-service,type=bind,readonly"
]
```

Use `readonly` unless the agent needs to write. Even with write access, modifications to mounted repos won't show in this repo's git history — they go into the source repo. That's usually what you want for tightly-coupled but independently-versioned code.

The directory itself is empty in the scaffold; it gets populated by the bind mount when the container starts.
