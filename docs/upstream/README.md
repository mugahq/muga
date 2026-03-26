# Public Documentation

Files in this directory are **automatically synced** to the public repository
[mugahq/muga](https://github.com/mugahq/muga) via the `sync-public-docs` GitHub
Actions workflow.

## Rules

1. **Never put sensitive information here.** No infrastructure details, pricing,
   internal architecture, deployment configs, or business strategy.
2. **Only conventions and contracts.** This directory should contain coding
   conventions, API contracts (OpenAPI), CLI behavior specs, and agent knowledge
   that the public CLI/SDK needs.
3. **Review before merging.** Changes here trigger a PR in the public repo.
   Always review the PR to verify nothing sensitive leaked.

## What belongs here

- CLI conventions and command patterns
- API error format and standard codes
- OpenAPI spec for the public API (sanitized)
- Agent instructions for the CLI/SDK repos (language-agnostic)

## What does NOT belong here

- Infrastructure details (hosting, providers, IPs)
- Pricing tiers and billing logic
- Internal ADRs and architectural decisions
- Database schemas (ClickHouse internals)
- Deployment playbooks
- Research and competitive analysis
- Internal project roadmaps
