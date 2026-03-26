# Shared Conventions

Conventions that apply across all Muga repositories. This file is synced to the
public CLI/SDK repository automatically.

## Language Policy

All code, comments, commit messages, documentation, identifiers, schemas, API
fields, PR titles, and review comments must be in **English**.

## CLI Conventions

**Command pattern:** `muga <noun> <verb>` — noun first, verb second:

```
muga monitor add
muga monitor list
muga alert silence
```

**Flags:**
- `--long-name` for all flags
- `-s` short aliases for frequently used flags
- Boolean flags do not require a value (`--verbose`, not `--verbose=true`)

**Output:**
- Default output is human-readable tables
- `--json` flag for machine-readable JSON output
- `--quiet` flag to suppress non-essential output

## API Conventions

**Error response format:**

```json
{"error": {"code": "not_found", "message": "Monitor not found"}}
```

Standard error codes: `bad_request`, `unauthorized`, `forbidden`, `not_found`,
`conflict`, `rate_limited`, `internal_error`.

**JSON fields:** always `snake_case`.

**Versioning:** `/v1/` prefix on all endpoints.

## Git Conventions

**Branches:** `{TICKET-ID}/{title-in-kebab-case}` (e.g., `MUG-123/add-auth`).

**Commits:** conventional commits with component scope:

```
feat(cli): add monitor list command
fix(sdk): handle timeout errors gracefully
```
