# OpenAPI specs

Drop OpenAPI 3.x specs here as `*.yaml`, `*.yml`, or `*.json`. The `just generate-clients` recipe (run by `check.sh`) will pick them up and generate clients into `clients/`.

By default the Justfile uses `openapi-typescript` (TypeScript types only). Swap to `openapi-generator-cli` for full clients in other languages — see the comments in the `Justfile`.
