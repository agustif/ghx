# ADR 0000: ghx fork scope and principles

Status: Accepted
Date: 2026-05-17

## Context

The upstream GitHub CLI is broad and stable, but its product direction leaves gaps for users and agents who work across many accounts, forks, repos, workflows, and GitHub platform surfaces at once. The `ghx` fork exists to move faster on these operational workflows without breaking the base CLI contract.

## Decision

`ghx` is a maintained fork of `gh`, not a clean-room replacement.

The fork should stay compatible with upstream command behavior unless a command is explicitly `ghx`-only or a safety fix is required. New behavior should be additive where possible.

`ghx`-only behavior is justified when it improves:

- scoped account and repo intent
- generated REST/GraphQL API access
- cross-surface diagnosis
- durable agent workflows
- extension bundling and account-aware wrappers
- progress logs and resumable collaboration
- safer mutation previews

## Consequences

- Upstream mergeability matters. Avoid broad rewrites that make routine upstream rebases painful.
- Fork-specific docs live under `docs/ghx.md`, `docs/adr/`, and `docs/plans/`.
- Fork-specific command groups should use names that make the added contract clear.
- Every new feature should preserve raw escape hatches.

## Follow-ups

- Keep `docs/ghx-gap-map.md` as the high-level roadmap.
- Keep ADRs short and update them when a decision changes.
- Split implementation detail into `docs/plans/`.
- Prefer wrappers and generated substrates before bespoke hand-written command sprawl.
