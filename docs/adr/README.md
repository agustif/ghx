# ghx ADR index

Status: active
Date: 2026-05-17

Architecture Decision Records in this directory capture fork-level choices that should survive individual chat sessions.

## Records

| ADR | Status | Decision |
| --- | --- | --- |
| [0000](0000-ghx-fork-scope-and-principles.md) | Accepted | Maintain `ghx` as an upstream-compatible fork with sharper account, API, workflow, and agent ergonomics. |
| [0001](0001-scoped-account-binding.md) | Accepted | Bind GitHub account context per session, cwd, and `.ghaccount` without changing the global active account. |
| [0002](0002-generated-validated-api-proxies.md) | Accepted | Build generated validated REST and GraphQL proxy layers behind `ghx` workflows. |
| [0003](0003-official-github-surface-mining.md) | Accepted | Mine official GitHub docs, schemas, changelog, and repos into repeatable gap reports. |
| [0004](0004-gh-aw-workflow-companion.md) | Accepted | Adopt `gh aw` as a companion for durable repo-owned workflows, not a replacement for `ghx`. |
| [0005](0005-extension-bundles-and-wrappers.md) | Accepted | Pre-bundle and wrap proven extensions before rebuilding their behavior in-tree. |
| [0006](0006-agent-progress-contract.md) | Accepted | Use a shared progress contract with 0/100 reporting, task trees, and machine-readable logs. |
| [0007](0007-ghx-config-directory.md) | Accepted | Support layered `.ghx/` config files for project-local profiles, extension bundles, workflows, and trusted hooks. |
| [0008](0008-product-discovery-record-system.md) | Accepted | Use research notes, RFCs, ADRs, and subissues as the durable product discovery record system. |

## ADR template

Each ADR should include:

- Status
- Date
- Context
- Decision
- Consequences
- Follow-ups
