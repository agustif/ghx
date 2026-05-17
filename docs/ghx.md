# ghx documentation index

Status: active
Date: 2026-05-17

This index collects the fork-specific plans, ADRs, and implementation notes for `ghx`.

## Product vision

`ghx` is the account-safe, agent-ready GitHub control plane. It should make it obvious which account, host, repo, scopes, rules, checks, and mutation target a command will use, then expose the deeper GitHub platform as composable terminal workflows.

The product direction is:

- control plane: `ghx ctx` explains account, repo, token, scope, and mutation intent before risky actions
- operational cockpit: `ghx pr ready`, `ghx ci doctor`, `ghx rules explain`, `ghx env approve`, `ghx deploy timeline`, `ghx mq status`, and `ghx sec inbox`
- API coverage: generated REST and GraphQL proxies with validated params plus raw escape hatches
- workflow engine: issue trees, Actions workflows, progress logs, attachments, and audit trails for human and agent collaboration
- toolchain compatibility: stable JSON and pipe-friendly output first, optional managed companions second
- identity safety: `.ghaccount`, `.ghx/`, session scopes, and cwd binding make wrong-account mutation hard

The near-term PM rule is to ship one useful read-only control-plane slice at a time, then add mutation only when identity, dry-run, JSON, and audit behavior are proven.

## Core docs

- [Multiple accounts](multiple-accounts.md): upstream multi-account behavior plus `ghx` scoped account selection.
- [Gap map](ghx-gap-map.md): feature gap map between current `gh`/`ghx` and the deeper GitHub platform surface.
- [Research index](research/README.md): broad investigation notes and product angle maps before decisions harden.
- [RFC index](rfcs/README.md): larger implementation proposals with rollout and validation plans.
- [ADR index](adr/README.md): accepted architecture decisions for the fork.
- [Plan index](plans/README.md): implementation plans and sequencing.
- [API coverage](ghx-api-coverage.md): placeholder for generated REST/GraphQL coverage reports.
- [Official surface report](ghx-official-surface-report.md): placeholder for generated product/docs gap reports.
- [Extension bundle](ghx-extension-bundle.md): curated extension catalog and wrapper posture.
- [Agent workflows](ghx-agent-workflows.md): contributor and agent guide for progress logs, JSON, screenshots, and collaboration-safe command patterns.

## First implementation lanes

1. Stabilize scoped account binding so `.ghaccount`, cwd scope, session scope, and explicit env overrides are predictable.
2. Generate validated API proxies from pinned GitHub REST OpenAPI and GraphQL schema inputs.
3. Mine official GitHub docs, schemas, changelog, and official repos for missing terminal workflows.
4. Adopt `gh aw` for repo-owned durable workflows while keeping `ghx` as the local interactive control plane.
5. Bundle and wrap proven extensions before rebuilding their behavior in-tree.
6. Keep agent progress and collaboration state machine-readable through `p` and `--progress-log`.
7. Deliver the read-only first slices: `ghx ctx`, `ghx pr ready`, `ghx pr threads`, `ghx ci doctor`, and `ghx rules explain`.
8. Add a layered `.ghx/` configuration directory for project-local profiles, extension bundles, workflow defaults, and trusted hooks.

## Operating rules

- Every mutating `ghx` command needs `--dry-run` or an equivalent preview path.
- Every implicit repo/account resolution needs `--explain` or visible status output.
- Agent-facing commands need `--json`.
- Long-running commands should support `--watch` and `--progress-log`.
- Raw escape hatches stay available through `ghx api`, generated operation metadata, and extension pass-through.
