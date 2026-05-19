# ghx plan index

Status: active
Date: 2026-05-17

This directory holds implementation plans for the fork-specific `ghx` roadmap.

## Plans

| Plan | Purpose |
| --- | --- |
| [Scoped account rollout](scoped-account-rollout.md) | Finish `.ghaccount`, cwd/session scope, status visibility, and context doctor work. |
| [Generated API proxies](generated-api-proxies.md) | Build the typed REST/GraphQL proxy substrate and first Actions subset. |
| [Official surface mining](official-surface-mining.md) | Generate coverage and gap reports from official GitHub sources. |
| [Workflows and extensions](workflows-and-extensions.md) | Adopt `gh aw`, bundle proven extensions, and wrap `gh attach`. |
| [Agent progress contract](agent-progress-contract.md) | Standardize `0/100`, task trees, and progress logs for agent-friendly commands. |
| [First delivery slices](first-delivery-slices.md) | Read-only-first plan for `ghx ctx`, PR readiness, review threads, CI doctor, and rules explanation. |
| [`.ghx/` config directory](ghx-config-directory.md) | Layered project-local config inspired by zsh-style user composition, with safety and trust gates. |
| [First-class release migration](ghx-first-class-release-migration.md) | Map every release, packaging, update, automation, docs, and verification surface that must be forked or explicitly preserved for `ghx`. |
| [ghx side-by-side migration and rollback runbook](ghx-side-by-side-migration-runbook.md) | Operator-safe sidecar and shadow-mode install, account selection, helper wiring, and rollback steps. |
| [ghx release operator runbook](ghx-release-operator-runbook.md) | Staging, verification, publication, distribution matrix, and rollback sequence for `ghx` releases. |

## Current order

1. Generated validated API proxy spike.
2. Scoped account context commands.
3. PR readiness and review threads.
4. CI doctor and pending deployments.
5. Rules, deployments, security inbox.
6. First-class release migration gates for side-by-side distribution.
7. Side-by-side migration and release operator runbooks.
8. Extension bundle and workflow companion polish in parallel where it reduces custom work.
