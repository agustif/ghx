# ghx official surface report

Status: manual seed for generated reports
Date: 2026-05-19

This report captures official GitHub product and API surfaces that are available through docs, REST, GraphQL, or official repos but are not yet first-class `ghx` workflows.

## Sources

- GitHub CLI manual: https://cli.github.com/manual/gh
- GitHub REST API docs: https://docs.github.com/en/rest
- GitHub REST OpenAPI description: https://github.com/github/rest-api-description
- GitHub GraphQL object reference: https://docs.github.com/en/graphql/reference/objects
- GitHub GraphQL public schema docs: https://docs.github.com/en/graphql/overview/public-schema
- GitHub Agentic Workflows: https://github.com/github/gh-aw
- GitHub Stacked PRs extension: https://github.com/github/gh-stack

## Main conclusion

The official CLI already covers the most common nouns: issues, PRs, repos, releases, projects, Actions runs, workflows, caches, rulesets, attestations, secrets, variables, and extensions. The gap for `ghx` is not breadth for its own sake. The leverage is cross-surface diagnosis, account-safe mutation, and generated coverage for official APIs that currently require hand-written `gh api` calls.

## High-value official surfaces

| Product surface | Official source | Current CLI coverage | API coverage | Extension overlap | Safety risk | Proposed `ghx` action |
| --- | --- | --- | --- | --- | --- | --- |
| PR merge readiness | CLI `pr`, REST checks/rules/deployments, GraphQL PR/review/merge queue objects | Spread across commands | Strong | `gh-stack` for stacked PR workflow | Merging with wrong account or stale blocker state | Build `ghx pr ready` and `ghx pr gate explain`. |
| Review thread resolution | GraphQL `PullRequestReviewThread` | Browser/raw GraphQL workflow | Strong | Some third-party review extensions | Missing unresolved comments before merge | Build `ghx pr threads --unresolved --json`. |
| Merge queue operations | GraphQL `MergeQueue`, `MergeQueueEntry` | Hidden under `pr merge` | Strong | None required | Misreading queue position or stuck state | Build `ghx mq status` and feed it into `ghx pr ready`. |
| Rules and rule suites | REST repo/org rulesets and rule suites, GraphQL ruleset objects | `ruleset` exists but diagnosis is thin | Strong | None required | Policy bypass or unexplained blocked merges | Build `ghx rules why-blocked` and `ghx rules diff`. |
| Environment approvals | REST pending deployments and environment endpoints, GraphQL environment/deployment objects | Mostly raw API | Strong | `gh-aw` can wrap durable approvals | Approving wrong environment or run | Build `ghx env pending`, `approve`, and `reject` with dry-run. |
| CI failure diagnosis | REST workflow runs/jobs/logs, checks, annotations | `run` and `pr checks` exist | Strong | `gh-aw` for repo-owned remediation | Rerunning wrong workflow or masking provider outage | Build `ghx ci doctor`, `ghx checks inventory`, and targeted rerun. |
| Hosted runner administration | REST hosted runner, image, platform, machine size, and limits endpoints | No first-class surface | Strong | None required | Capacity and cost changes are hard to audit | Build `ghx runners status`, `images`, `capacity`. |
| Artifact and cache hygiene | REST artifacts, cache usage, retention, storage-limit endpoints | `run download`, `cache list/delete` | Strong | `actions/gh-actions-cache` can inform behavior | Deleting useful evidence or wasting storage | Build `ghx actions storage report`, `artifact sweep`, `cache top`. |
| Security alert inbox | REST code scanning, secret scanning, Dependabot, advisories | No unified surface | Strong | None required | Dismissal and exposure mistakes | Build read-only `ghx sec inbox` first, mutations later. |
| Org access governance | REST org roles, teams, outside collaborators, PAT requests, webhooks, API insights | `org list` only | Strong | None required | Wrong principal or token can mutate repos | Build `ghx org access why`, `org tokens`, `org apps`, `api insights`. |
| Agent tasks and coding-agent policy | REST `agent-tasks`, Copilot coding-agent policy, GraphQL project/PR data | `agent-task` preview exists | Growing | `gh-aw` is the durable workflow companion | Agent actions become opaque or overprivileged | Build `ghx agent handoff`, `agent policy`, and task-to-PR views. |
| Projects v2 operational status | REST Projects, GraphQL `ProjectV2` | `project` CRUD exists | Strong | `gh-aw` can write repo-owned updates | Planning state drifts from PR/issue truth | Build `ghx board status` and `board sync-pr`. |
| Webhook delivery debugging | REST org/repo hook delivery endpoints | Raw API | Strong | None required | Redelivering to wrong integration | Build `ghx hooks deliveries` and dry-run redelivery. |

## What not to build first

- Do not duplicate every REST endpoint as a hand-written command. Generate typed helpers and promote only operational workflows.
- Do not replace `gh aw`. Use it for durable repo-owned Actions workflows and keep `ghx` as the local account-safe control plane.
- Do not build mutating admin commands before read-only explain output and `--dry-run` are stable.
- Do not hide extension provenance. Curated bundles should show owner, version, license, pin, age, and wrapper behavior.

## Intended generator

Future command:

```sh
ghx mine github --source docs --format md > docs/ghx-official-surface-report.md
```

## Required report fields

- product area
- official source URL
- local command coverage
- API coverage
- extension overlap
- safety risk
- proposed `ghx` action
- recommended first read-only command
- recommended first mutating command, if any
