# CLAUDE.md — muga

## Contract-first development

The OpenAPI specification at `docs/contracts/openapi.yaml` is the single source of truth for the Muga API surface. This file is automatically synced from the upstream server repository.

Before implementing or modifying any CLI command or SDK method:

1. Read `docs/contracts/openapi.yaml` to understand the available endpoints, request/response schemas, and error codes.
2. Match CLI commands and SDK methods 1:1 with API endpoints — do not invent endpoints that are not in the spec.
3. Use the exact field names, types, and enum values defined in the spec.
4. When the spec changes, update the affected CLI commands and SDK methods to stay in sync.

## Repository layout

```
cli/              # Go CLI — muga <noun> <verb>
sdks/
├── python/       # Python SDK (PyPI: muga)
└── node/         # Node.js SDK (planned)
docs/
├── contracts/    # OpenAPI spec (synced from upstream)
└── upstream/     # Shared conventions (synced from upstream)
```

## CLI

- Command pattern: `muga <noun> <verb>` (e.g., `muga monitor add`).
- Flags: `--long-name` always, `-s` short aliases for common flags.
- Output: default to human-readable tables; support `--output json` for scripting.
- Every command must map to an OpenAPI endpoint.

## SDKs

- Method names follow the `<noun>_<verb>` pattern (e.g., `client.monitors.list()`).
- Every public method must map to an OpenAPI endpoint.
- Use the error codes from the spec (`bad_request`, `not_found`, etc.).
