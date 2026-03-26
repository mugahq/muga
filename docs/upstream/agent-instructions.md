# Agent Instructions

Rules for AI agents working on Muga CLI and SDK repositories.

## Rules

1. **Read before implementing.** Check existing code and docs before writing.
2. **Follow existing patterns.** Read an existing command/handler before creating
   a new one.
3. **No LLM calls in CI.** Pipelines must be deterministic.
4. **One package per ticket.** A ticket touches files in exactly one scope.
5. **Binary acceptance criteria.** Every criterion must be yes/no testable.
6. **Explicit "Out of scope".** Every ticket states what it does NOT cover.

## CLI Patterns

- Commands follow `muga <noun> <verb>` pattern — noun first, verb second
- Each command is in its own file named after the command
- Output formatting must be consistent: human-readable tables by default, `--json` for machine output
- No business logic in command handlers — delegate to a service/client layer

## SDK Patterns

- SDK surface should be minimal: `init()`, structured logging, and tracing helpers
- No business logic in the SDK — it is a thin client over the Muga API
- Each SDK wraps OpenTelemetry for log/trace export
- Language-specific conventions (naming, error handling, packaging) are defined in each SDK's own `AGENTS.md`

## Validation

Each repo defines its own validation commands in its `AGENTS.md`. All repos must pass lint, type-check, and tests before opening a PR.

## Issue Tracker

Linear — workspace: `mugahq`.

Status flow: Backlog → Todo → In Progress → Human Review → Merging → Done.

Agents pick tickets when status is `Todo` and all `blockedBy` tickets are `Done`.
