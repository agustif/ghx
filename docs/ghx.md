# ghx documentation index

Status: active
Date: 2026-05-17

This index collects the fork-specific plans, ADRs, and implementation notes for `ghx`.

## Start here: ghx vs gh

Read [ghx vs gh](ghx-vs-gh.md) first for the shipped behavior contract. The rest of this index includes roadmap and research material, so planned commands should not be treated as available unless they are listed in that comparison page.

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

- [ghx vs gh](ghx-vs-gh.md): shipped behavior that intentionally differs from regular upstream GitHub CLI.
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

## Manual and generated reference

The public manual at `https://cli.github.com/manual/` is upstream `gh` documentation. It remains owned by the upstream site and should not be used as a `ghx` publication target.

Generate fork-owned `ghx` reference pages into a local or fork-owned directory:

```sh
$ mkdir -p dist/ghx-manual
$ go run ./cmd/gen-docs --website --doc-path dist/ghx-manual --command-name ghx
```

Generate fork-owned `ghx` manpages with the source install target:

```sh
$ make manpages-ghx
```

The generated filenames, headings, links, and prompt examples use the configured command name, for example `ghx_issue_create.md`, `ghx-issue-create.1`, `## ghx issue create`, and `$ ghx issue create`.

Do not run the upstream `site-docs` target or production deployment workflow as a `ghx` manual publication path. Those paths check out or mutate the upstream `github/cli.github.com` site and are still part of the inherited `gh` release process.

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
