# ghx extension bundle

Status: draft
Date: 2026-05-17
Related ADR: [ADR 0005](adr/0005-extension-bundles-and-wrappers.md)

This file is the human-reviewed catalog for extensions that `ghx` should bundle, wrap, mine, or ignore.

## Bundle commands

- `ghx ext bundle list`
- `ghx ext bundle install <bundle>`
- `ghx ext bundle audit`
- `ghx ext wrap <extension>`

## Initial catalog

| Extension | Owner | Posture | Notes |
| --- | --- | --- | --- |
| `gh aw` | `github/gh-aw` | first-class companion | Use for durable repo-owned agent workflows in GitHub Actions. |
| `gh stack` | `github/gh-stack` | first-class companion | Use for stacked PR workflows before rebuilding stack mechanics. |
| `gh attach` | `enthus-appdev/gh-attach` | wrap and watch | Useful for screenshot-backed PR/issue evidence. Requires unique basenames for large screenshot sets. |
| `gh dash` | `dlvhdr/gh-dash` | optional bundle | Mature TUI for humans; do not treat as machine-readable source of truth. |
| `gh pr-review` | `agynio/gh-pr-review` | spike | Mine unresolved-thread behavior before building `ghx pr threads`. |
| `gh actions-cache` | `actions/gh-actions-cache` | wrap or supersede | Compare against `ghx cache top` and storage reports. |
| `gh actions-importer` | `github/gh-actions-importer` | optional bundle | Useful for migration-heavy repos. |

## Audit metadata

Each catalog entry should eventually include:

- pinned version
- install command
- license
- update age
- required scopes
- supported output formats
- noninteractive support
- wrapper required
- risk notes

## Wrapper requirements

Wrappers should preserve:

- `.ghaccount`
- `GH_ACCOUNT`
- `GH_ACCOUNT_SESSION`
- `GH_HOST`
- `--repo`
- `--json` when the extension supports it
- `--progress-log` when `ghx` adds the wrapper behavior
