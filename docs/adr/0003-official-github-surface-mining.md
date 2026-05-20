# ADR 0003: Official GitHub surface mining

Status: Accepted
Date: 2026-05-17

## Context

GitHub ships platform features faster than the core CLI exposes them as coherent terminal workflows. Feature discovery through manual docs browsing or ad hoc `gh api` commands is not enough. The fork needs a repeatable way to compare official GitHub surfaces against local command coverage.

## Decision

`ghx` will add an official-source mining lane that reads from:

- `github/rest-api-description`
- GitHub GraphQL public schema
- `github/docs`
- GitHub changelog
- official GitHub extension repos and CLI-adjacent repos

The mining lane should produce reports, not automatic product decisions.

## Consequences

- `ghx mine github` generates Markdown and JSON reports with source links for
  REST, GraphQL, and local command inventory.
- Report output should separate missing API coverage from missing high-level command workflows.
- Generated reports should be stable enough to diff in PRs.
- Human review still decides whether to build, wrap, bundle, defer, or ignore a candidate.

## Follow-ups

- Create `docs/ghx-official-surface-report.md`.
- Create `docs/ghx-api-coverage.md`.
- Add drift automation around `ghx mine github --source rest`.
- Add curated GraphQL operation validation on top of `ghx mine github --source graphql`.
- Build `ghx mine github --source docs`.
- Build `ghx mine github --source extensions`.
