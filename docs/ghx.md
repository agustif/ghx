# ghx documentation index

Status: active
Date: 2026-05-17

This index collects the fork-specific plans, ADRs, and implementation notes for `ghx`.

## Core docs

- [Multiple accounts](multiple-accounts.md): upstream multi-account behavior plus `ghx` scoped account selection.
- [Gap map](ghx-gap-map.md): feature gap map between current `gh`/`ghx` and the deeper GitHub platform surface.
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
