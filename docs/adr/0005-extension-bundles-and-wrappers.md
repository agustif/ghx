# ADR 0005: Extension bundles and wrappers

Status: Accepted
Date: 2026-05-17

## Context

The GitHub CLI extension ecosystem already contains tools that solve parts of the `ghx` roadmap. Examples include `gh attach`, `gh dash`, `gh stack`, `gh pr-review`, `gh actions-cache`, and official migration or agent workflow extensions. Rebuilding each tool in-tree would slow the fork and increase maintenance load.

At the same time, arbitrary extension execution can be risky for agents because extensions may not honor account scoping, JSON output, noninteractive behavior, or progress logging.

## Decision

`ghx` will support curated extension bundles and account-aware wrappers.

Bundle commands:

- `ghx ext bundle list`
- `ghx ext bundle install <bundle>`
- `ghx ext bundle audit`
- `ghx ext wrap <extension>`

Initial bundle candidates:

- `github/gh-aw`
- `github/gh-stack`
- `enthus-appdev/gh-attach`
- `dlvhdr/gh-dash`
- `agynio/gh-pr-review`
- `actions/gh-actions-cache`
- `github/gh-actions-importer`

## Consequences

- Agents can install a known toolchain without rediscovering extensions every session.
- Wrappers can inject `.ghaccount` and `GH_ACCOUNT_SESSION` context before invoking extensions.
- Third-party extensions remain visible and auditable instead of hidden behind `ghx`.
- Proven extension behavior can inform native `ghx` commands before the fork commits to reimplementation.

## Follow-ups

- Create `docs/ghx-extension-bundle.md`.
- Add a machine-readable bundle manifest under `internal/ghx/extensions/`.
- Add extension audit metadata: owner, repo, pin, license, update age, command mapping, required scopes, and risk notes.
- Wrap `gh attach` first because it is already installed locally and useful for screenshot-backed PR evidence.
