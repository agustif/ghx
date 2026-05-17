# ghx gap map

Status: draft
Date: 2026-05-17
Scope: gaps between the current `ghx` command surface and the GitHub API/product surfaces that are exposed by REST or GraphQL but not made operationally first-class by the CLI.

## Current local surface

The local fork currently exposes the same broad groups as upstream `gh`, plus fork-specific account-scoping work:

- Account and auth: `auth login`, `auth logout`, `auth refresh`, `auth setup-git`, `auth status`, `auth switch`, `auth token`.
- Core collaboration: `issue`, `pr`, `repo`, `release`, `project`, `gist`, `org`, `label`, `search`.
- Actions: `cache`, `run`, `workflow`, `secret`, `variable`.
- Policy and supply chain: `ruleset`, `attestation`, `release verify`, `release verify-asset`.
- Utilities: `api`, `browse`, `codespace`, `extension`, `copilot`, `agent-task`, `skill`.
- ghx fork delta so far: scoped account selection via `GH_ACCOUNT`, `GH_ACCOUNT_SESSION`, cwd scope, `.ghaccount`, and `auth status` source visibility.

The CLI has breadth. The gap is that it mostly exposes individual nouns, while real GitHub work needs cross-surface answers like "can this PR merge now?", "why is this run blocked?", "which policy caused this?", and "which account/repo will this mutation hit?"

## Gap scoring

- `none`: no focused command surface beyond `gh api`.
- `thin`: a command exists, but important API data is missing.
- `partial`: useful command exists, but no operational diagnosis or cross-surface workflow.
- `good`: acceptable base, still improvable for `ghx` conventions.

## Priority gap matrix

| Area | Current CLI coverage | GitHub surface available | Gap | ghx target |
| --- | --- | --- | --- | --- |
| Identity and intent | `auth status`, `auth token`, `auth switch`; ghx adds scoped selection | Token scopes, host config, repo remotes, account identity, SSO failures | `partial` | `ghx ctx`, `ghx ctx explain`, `ghx ctx doctor`, every mutating command shows account/repo intent under `--explain` |
| PR readiness | `pr status`, `pr checks`, `pr view`, `pr merge` | GraphQL PR review state, review threads, merge queue, rulesets, deployments, checks | `thin` | `ghx pr ready` as a merge cockpit with blockers, unresolved threads, checks, deployments, rulesets, queue state |
| Review threads | `pr review` can create reviews; `pr view` is not a thread triage tool | GraphQL `PullRequestReviewThread` and review comments | `thin` | `ghx pr threads`, `ghx pr comments --unresolved`, exact file/line/action output |
| Merge queue | Some merge behavior is hidden under `pr merge` | GraphQL `MergeQueue`, `MergeQueueEntry`, positions, state, estimated time | `none` | `ghx mq status`, `ghx mq watch`, `ghx mq why-stuck` |
| Rulesets and branch policy | `ruleset list`, `ruleset view`, `ruleset check` | REST and GraphQL repo/org rulesets, rule suites, branch protection, required workflows | `partial` | `ghx rules explain <ref>`, `ghx rules why-blocked <pr>`, `ghx rules diff/export/import` |
| Actions run triage | `run list`, `run view`, `run watch`, `run rerun`, `run download` | Workflow runs, jobs, attempts, logs, annotations, pending deployments | `partial` | `ghx ci doctor`, `ghx ci fail`, `ghx ci logs --failed`, `ghx ci rerun --failed --reason` |
| Environment approvals | Mostly raw `api` | Workflow pending deployments and deployment protection review endpoints | `none` | `ghx env pending`, `ghx env approve`, `ghx env reject`, always dry-run friendly |
| Deployments | Mostly raw `api`; release commands are separate | Deployments, deployment statuses, environments, protection rules | `none` | `ghx deploy status`, `ghx deploy timeline`, `ghx deploy mark`, `ghx deploy open` |
| Actions artifacts | `run download`; no inventory or hygiene workflow | Artifact list, get, download, delete, digests, expiry | `thin` | `ghx artifact list`, `ghx artifact sweep`, `ghx artifact verify-digest`, storage reports |
| Actions cache | `cache list`, `cache delete` | Cache list, delete, usage at repo/org level | `partial` | `ghx cache top`, `ghx cache prune --stale-branches`, `ghx actions storage report` |
| Runners | No substantial hosted-runner operations | Hosted runners, self-hosted runners, images, limits, machine sizes | `none` | `ghx runners status`, `ghx runners images`, `ghx runners capacity`, `ghx runners doctor` |
| Security alerts | No first-class code/secret/dependabot inbox | Code scanning alerts, secret scanning alerts, Dependabot alerts, security advisories | `none` | `ghx sec inbox`, `ghx sec code`, `ghx sec secrets`, `ghx sec dismiss --dry-run`, `ghx sec report` |
| Supply chain | `attestation verify`, `release verify` | Artifact attestations, release assets, repo attestations, org/user attestations | `partial` | `ghx release doctor`, `ghx attest list`, `ghx attest policy`, provenance gates |
| Projects v2 operations | Many project commands exist | GraphQL Projects v2 fields, items, views, status updates | `partial` | `ghx board status`, `ghx board stale`, `ghx board sync-pr`, `ghx board plan` |
| Notifications and review queue | `status` is broad but shallow | Notifications, review requests, subscriptions, issue/PR state | `thin` | `ghx inbox`, `ghx review queue`, `ghx inbox triage` |
| Org access and governance | `org list`; repo/team details mostly missing | Org roles, teams, members, outside collaborators, fine-grained PAT inventory | `thin` | `ghx org access why`, `ghx org tokens`, `ghx org apps`, `ghx org audit tail` |
| Agent workflows | `agent-task` exists, but no general progress/control-plane contract | GitHub agent-task API, issues/PRs/actions/projects | `partial` | `ghx agent handoff`, `ghx agent progress`, `--progress-log` on long-running commands |
| API ergonomics | `api` is powerful but low-level | REST and GraphQL OpenAPI/schema | `good but raw` | `ghx api discover`, generated typed helpers, endpoint permission hints, examples from live auth scope |

## Highest-value additions

### 1. `ghx ctx`

Goal: remove account and repo ambiguity before any mutation.

Commands:

- `ghx ctx`
- `ghx ctx explain`
- `ghx ctx doctor`
- `ghx ctx bind --account <login> [--cwd <path>]`
- `ghx ctx unbind`

Output should include:

- host
- active login
- active user source, such as `GH_ACCOUNT`, `GH_ACCOUNT_SESSION`, `.ghaccount`, cwd scope, host-global
- token source and scopes
- cwd repo
- git remote owner/name
- selected API host and web host
- mutation safety warnings

### 2. `ghx pr ready`

Goal: one command answers whether a PR can merge and what exact blocker remains.

Inputs:

- PR number, URL, branch, or current branch
- `--json`
- `--watch`
- `--progress-log`
- `--explain`

Data to combine:

- PR review decision
- unresolved review threads
- requested reviewers
- required checks and latest job failures
- merge state and mergeable state
- auto-merge and merge queue state
- rulesets/branch protection blockers
- required deployment environments
- pending environment approvals
- stale branch state
- account/repo context from `ghx ctx`

Suggested JSON shape:

```json
{
  "ready": false,
  "account": "agustif",
  "repository": "OWNER/REPO",
  "pullRequest": 123,
  "headSha": "...",
  "blockers": [
    {
      "kind": "unresolved_review_thread",
      "summary": "1 unresolved thread",
      "url": "https://github.com/OWNER/REPO/pull/123#discussion_r..."
    }
  ],
  "nextActions": [
    "ghx pr threads 123 --unresolved"
  ]
}
```

### 3. `ghx ci doctor`

Goal: avoid manually opening Actions pages to find the failed job and real failing step.

Commands:

- `ghx ci doctor [--pr <n>|--branch <name>|--sha <sha>]`
- `ghx ci fail`
- `ghx ci logs --failed [--grep <pattern>]`
- `ghx ci rerun --failed --reason <text>`
- `ghx ci pending-deployments`

Data to combine:

- workflow runs
- run attempts
- workflow jobs and steps
- job logs
- annotations
- pending deployments
- runner labels and runner group where exposed

### 4. `ghx rules explain`

Goal: make branch/ruleset policy understandable from the terminal.

Commands:

- `ghx rules explain [ref]`
- `ghx rules why-blocked [pr]`
- `ghx rules diff --org <org> --repo <repo>`
- `ghx rules export`

Output should explain:

- matching rulesets
- inherited org rules
- required checks
- required review count
- required review-thread resolution
- required deployments
- bypass actors
- rule suite evaluations if available

### 5. `ghx sec inbox`

Goal: one security queue across GitHub alert products.

Commands:

- `ghx sec inbox [--org <org>|--repo <repo>]`
- `ghx sec code`
- `ghx sec secrets`
- `ghx sec dependabot`
- `ghx sec report --format md`

Initial read-only implementation should group by:

- severity
- repository
- alert age
- alert kind
- assigned user
- fixed/dismissed/open state

Mutating commands should require `--dry-run` default previews and explicit confirmation.

### 6. `ghx deploy status`

Goal: expose GitHub deployments as a first-class operational state, not a hidden API object.

Commands:

- `ghx deploy status [--env <name>]`
- `ghx deploy timeline`
- `ghx deploy mark --state success --env preview --url <url>`
- `ghx env pending`
- `ghx env approve`

This should connect deployments, environment URLs, logs, workflow pending deployment gates, and current branch/PR.

## Implementation order

The lowest-risk order is:

1. `ghx ctx`: mostly local config plus existing auth APIs. It makes every later command safer.
2. `ghx pr ready --json`: read-only GraphQL/REST aggregation. High daily value.
3. `ghx pr threads`: focused read-only GraphQL surface that feeds `pr ready`.
4. `ghx ci doctor`: read-only REST aggregation over runs/jobs/logs.
5. `ghx rules explain`: mostly read-only REST/GraphQL rulesets.
6. `ghx deploy status` and `ghx env pending`: read-only deployment/environment gates.
7. `ghx sec inbox`: read-only security alert aggregation.
8. Mutating variants: approvals, deployment statuses, alert dismissal, ruleset import, cache/artifact pruning.

## Design rules for ghx-only surfaces

- Every command that can mutate remote state must support `--dry-run`.
- Every command that resolves repo/account implicitly must support `--explain`.
- Every command used by agents must support `--json`.
- Long-running commands should support `--watch` and `--progress-log`.
- Secrets, tokens, and credential material are redacted by default.
- Scoped account context must be visible in status and blocker output.
- Errors should include the next exact command when possible.
- Raw API access remains available through `ghx api`, but common operational paths should not require hand-written GraphQL.

## Source map

- Local command inventory: `pkg/cmd/*` in this checkout, plus `ghx help`.
- Official CLI manual command list: https://cli.github.com/manual/gh
- GitHub REST API overview and current categories: https://docs.github.com/en/rest
- GitHub Actions artifacts REST API: https://docs.github.com/en/rest/actions/artifacts
- GitHub workflow runs REST API: https://docs.github.com/en/rest/actions/workflow-runs
- GitHub deployment statuses REST API: https://docs.github.com/en/rest/deployments/statuses
- GitHub rulesets REST API: https://docs.github.com/en/rest/repos/rules
- GitHub code scanning REST API: https://docs.github.com/en/rest/code-scanning/code-scanning
- GitHub secret scanning REST API: https://docs.github.com/en/rest/secret-scanning/secret-scanning
- GitHub hosted runners REST API: https://docs.github.com/en/rest/actions/hosted-runners
- GitHub organization fine-grained PAT REST API: https://docs.github.com/en/rest/orgs/personal-access-tokens
- GitHub GraphQL object schema: https://docs.github.com/en/graphql/reference/objects
