# Plan: Workflows and extensions

Status: active
Date: 2026-05-17
Related ADRs: [ADR 0004](../adr/0004-gh-aw-workflow-companion.md), [ADR 0005](../adr/0005-extension-bundles-and-wrappers.md)

## Goal

Adopt proven GitHub workflow and extension tools before rebuilding them, while adding `ghx` account safety, provenance, JSON output, and progress behavior around them.

## `gh aw` adoption

Use `gh aw` for durable repo-owned agent workflows:

- workflows committed to the repo
- GitHub Actions execution
- approval gates
- safe outputs
- audit history
- issue and PR integration

Keep `ghx` responsible for:

- local account binding
- cwd/repo context
- generated API calls
- immediate diagnosis
- human-in-the-loop terminal UX

## Extension bundle commands

- `ghx ext bundle list`
- `ghx ext bundle install agent`
- `ghx ext bundle audit`
- `ghx ext wrap <extension>`

## Initial bundle candidates

| Extension | Posture | First action |
| --- | --- | --- |
| `github/gh-aw` | first-class companion | Add workflow docs and wrapper commands. |
| `github/gh-stack` | first-class companion | Document stacked PR flow and account wrapper needs. |
| `enthus-appdev/gh-attach` | wrap and watch | Add proof-artifact upload docs and basename collision guard. |
| `dlvhdr/gh-dash` | optional bundle | Keep as user-facing TUI, not machine-readable source of truth. |
| `agynio/gh-pr-review` | spike | Mine review-thread behavior before implementing `ghx pr threads`. |
| `actions/gh-actions-cache` | wrap or supersede | Compare against planned cache/storage reports. |
| `github/gh-actions-importer` | optional bundle | Include for migration-heavy repositories. |

## Provenance metadata

Bundle manifests should include:

- owner/repo
- version pin
- install URL
- license
- update age
- command mapping
- required scopes
- risk notes
- whether wrapper is required

## Acceptance checks

- Bundles install pinned versions.
- Audit output flags unpinned, stale, archived, missing-license, or scope-sensitive extensions.
- Wrappers preserve `.ghaccount` and `GH_ACCOUNT_SESSION`.
- `gh attach` docs mention unique basenames for screenshot sets.
- Extension provenance is visible in `--json`.
