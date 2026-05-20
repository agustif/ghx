# Plan: Official GitHub surface mining

Status: active
Date: 2026-05-17
Related ADR: [ADR 0003](../adr/0003-official-github-surface-mining.md)

## Goal

Continuously discover useful `ghx` work from official GitHub sources instead of relying on memory or manual browsing.

## Inputs

- REST OpenAPI descriptions from `github/rest-api-description`.
- GraphQL public schema.
- Local Cobra command tree from the current checkout.
- `github/docs` article and navigation structure.
- GitHub changelog.
- Official GitHub extension repositories.
- Existing local command inventory under `pkg/cmd`.

## Commands

- `ghx mine github --source rest --format md`
- `ghx mine github --source rest --format json --rest-openapi /tmp/github-rest-openapi.json`
- `ghx mine github --source graphql --format json --graphql-schema /tmp/github-graphql-schema.json`
- `ghx mine github --source docs --area actions`
- `ghx mine github --source extensions --topic gh-extension`

## Outputs

- `docs/ghx-api-coverage.md`
- `docs/ghx-official-surface-report.md`
- `docs/ghx-extension-bundle.md`
- generated command, REST, and GraphQL coverage JSON for agents and drift checks
- `internal/ghapi/specs/manifest.json`

## Scoring

Each candidate should be scored by:

- user workflow value
- API availability
- current CLI coverage
- safety risk
- implementation size
- extension overlap
- need for account-aware wrapping
- whether generated proxies are enough

## Acceptance checks

- Reports include exact source links.
- Reports can be regenerated locally.
- Generated markdown diffs are reviewable.
- JSON output is available for agents.
- REST spec downloads do not send auth headers to public raw OpenAPI URLs.
- GraphQL report generation can use deterministic schema files in CI.
- The command never mutates remote state.
