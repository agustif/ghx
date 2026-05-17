# ADR 0004: `gh aw` workflow companion

Status: Accepted
Date: 2026-05-17

## Context

Some work should happen as durable, repo-owned workflows with reviewable files, Actions logs, approval gates, and audit history. Other work should happen as fast local terminal commands with account resolution, repo context, generated API calls, and immediate feedback.

`github/gh-aw` already exists as a GitHub Agentic Workflows extension. Rebuilding that entire workflow substrate inside `ghx` would duplicate an official tool before proving the gap.

## Decision

`ghx` will adopt `gh aw` as a first-class workflow companion, not as a replacement for `ghx`.

Use `gh aw` when work should live in the repo and run through GitHub Actions.

Use `ghx` when work needs local account binding, cwd context, immediate diagnosis, generated API calls, or interactive human control.

## Consequences

- `ghx workflows init` can bootstrap preferred `gh aw` templates.
- `ghx workflows doctor` can validate workflow permissions, safe outputs, pinned dependencies, and repository readiness.
- `ghx workflows run` can call through to `gh aw` or `gh workflow run` with scoped-account confirmation.
- `ghx workflows watch` can combine Actions state, logs, progress log output, and PR/issue links.
- `ghx` should not fork or vendor `gh aw` unless account binding, JSON output, safety, or maintenance gaps require it.

## Follow-ups

- Spike a repo workflow template for generated GitHub surface mining.
- Spike a repo workflow template for spec drift updates.
- Add account-aware wrappers for workflow launch and watch.
- Document when a task belongs in `gh aw` versus local `ghx`.
