# Plan: First delivery slices

Status: active
Date: 2026-05-17
Related docs: [Gap map](../ghx-gap-map.md), [Generated API proxies](generated-api-proxies.md), [Scoped account rollout](scoped-account-rollout.md)

## Goal

Ship useful `ghx` workflows in small read-only slices before adding mutating commands.

## PM sequencing

First principles:

- users and agents need to know identity and target before they need more mutation power
- diagnostics should collapse GitHub's scattered pages into one terminal answer with exact blockers and next commands
- every new surface should be scriptable by default through stable JSON
- raw API escape hatches and generated metadata should keep the product current while native workflows mature
- issue trees should hold the roadmap so each idea becomes an executable, reviewable unit of work

Next after first-class subissues:

1. Convert the existing `ghx` roadmap into parent issues and acceptance-sized subissues.
2. Ship `ghx ctx explain --json` as the identity and intent primitive.
3. Use generated API metadata for `ghx api explain <operation-id>`.
4. Build read-only `ghx pr ready` on top of `ctx`, review threads, checks, rulesets, and merge queue state.
5. Build `ghx ci doctor` with Actions run/job/log triage and pending deployment visibility.
6. Add opt-in companion tooling discovery through `ghx tools doctor` before managed installation.

## Slice 1: `ghx ctx`

Purpose: prove account and repo intent before other workflows depend on it.

Commands:

- `ghx ctx`
- `ghx ctx explain`
- `ghx ctx doctor`

Acceptance:

- shows host, active login, source, repo, remotes, token scopes, and mutation warnings
- supports `--json`
- never mutates config unless called through future `ghx ctx bind`

## Slice 2: `ghx api explain`

Purpose: prove generated metadata before wiring it into higher-level commands.

Commands:

- `ghx api explain <operation-id>`
- `ghx api discover <keyword>`

Acceptance:

- uses pinned spec metadata
- prints method, path, params, pagination, response shape, previews, and permission notes when known
- shows raw `ghx api` equivalent

## Slice 3: `ghx pr ready`

Purpose: answer whether a PR can merge and what exact blocker remains.

Inputs:

- PR number
- PR URL
- current branch

Blockers:

- checks
- unresolved review threads
- required reviews
- merge queue state
- deployment gates
- rulesets

Acceptance:

- read-only first
- supports `--json`
- lists exact next commands

## Slice 4: `ghx pr threads`

Purpose: make unresolved review conversations terminal-native.

Commands:

- `ghx pr threads [pr]`
- `ghx pr threads [pr] --unresolved`

Acceptance:

- shows file, line, author, state, URL, and latest body snippet
- supports `--json`
- can feed `ghx pr ready`

## Slice 5: `ghx ci doctor`

Purpose: find the failed run, failed job, failed step, and actionable log excerpt without opening the browser.

Commands:

- `ghx ci doctor`
- `ghx ci fail`
- `ghx ci logs --failed`
- `ghx ci pending-deployments`

Acceptance:

- starts with generated Actions REST subset
- supports branch, SHA, and PR inputs
- supports `--json`
- redacts secrets in logs

## Slice 6: `ghx rules explain`

Purpose: explain which branch, ruleset, or required workflow policy is blocking a ref or PR.

Commands:

- `ghx rules explain [ref]`
- `ghx rules why-blocked [pr]`

Acceptance:

- combines rulesets, branch protection, required checks, required deployments, bypass actors, and rule suite evaluations where available
- supports `--json`
- no mutation in first slice
