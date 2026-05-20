# ghx API coverage

Status: generated miner plus curated metadata seed
Date: 2026-05-20

This report compares the local `ghx` command surface against official GitHub REST and GraphQL surfaces. `ghx mine github` now generates Markdown or JSON reports from REST OpenAPI, GraphQL introspection, the local Cobra command tree, and the curated `internal/ghapi/rest` metadata seed.

## Evidence snapshot

- Local command surface: `ghx mine github` walks the live Cobra tree from this checkout.
- Generated REST metadata seed: `internal/ghapi/rest/registry.go`.
- Generated metadata command: `ghx api explain <operation-id>` and `ghx api explain --list`.
- REST surface: `github/rest-api-description` OpenAPI JSON. The miner defaults to the pinned PR #86 ref and accepts `--rest-openapi` for a file or URL.
- REST docs version: GitHub REST docs show API version `2026-03-10` as latest on 2026-05-19.
- GraphQL surface: `ghx mine github --source graphql` can live-introspect a host or read deterministic schema JSON through `--graphql-schema`.
- Official CLI reference: `https://cli.github.com/manual/gh`.
- Repro commands:
  - `ghx mine github --source rest --format md --rest-openapi /tmp/github-rest-openapi.json`
  - `ghx mine github --source graphql --format json --graphql-schema /tmp/github-graphql-schema.json`
  - `script/ghx-rest-coverage-summary /tmp/github-rest-openapi.json`

## Current quantitative gap

Regular `gh` has broad raw API reachability through `gh api`, including the
GraphQL endpoint. The explicit gap is different: regular `gh` does not expose
a generated REST operation registry, source checksums, coverage states,
parameter metadata, pagination hints, GraphQL schema coverage rows, or
operation-to-command mapping. `ghx` keeps the raw API escape hatch and now
adds generated coverage reports around it.

| Coverage question | Regular `gh` explicit metadata | Current `ghx` explicit metadata | Gap to full explicit metadata |
| --- | ---: | ---: | ---: |
| Raw REST reachability through `api` | broad escape hatch | broad escape hatch | not the target |
| Raw GraphQL reachability through `api graphql` | broad escape hatch | broad escape hatch | not the target |
| REST operations inventoried by generated report | 0 tracked | 1186 of 1186 (100%) | 0 operation inventory gap |
| REST operations with explicit generated metadata | 0 tracked | 13 of 1186 (1.1%) | 1173 operations (98.9%) |
| REST operations with coverage state and proposed command | 0 tracked | 13 of 1186 (1.1%) | 1173 operations (98.9%) |
| GraphQL root query and mutation inventory | 0 tracked | generated from schema | explicit coverage remains 0% until a GraphQL registry lands |

The first `ghx` metadata slice is intentionally small: Actions pending
deployments, hosted runners, workflow jobs and artifacts, checks list/rerun,
and deployments/statuses. The next useful milestone is not more hand-written
rows, but a generated coverage report that marks every REST operation as
`first-class`, `thin`, `raw-api`, or `missing`.

Coverage state means explicit local workflow coverage, not raw endpoint
reachability:

- `first-class`: a purpose-built `ghx` command covers the workflow.
- `thin`: an existing command partially covers the operation but does not expose
  the full agent or operator workflow.
- `raw-api`: `ghx api explain` can show the operation and raw command, but no
  first-class workflow command exists.
- `missing`: the operation is known to the generated report, but no local
  workflow command or reviewed raw-only decision exists yet.

High-volume REST tags from the OpenAPI snapshot:

| REST tag | Operations | Local command posture |
| --- | ---: | --- |
| `repos` | 201 | Broad `repo` coverage, but rules, rule suites, custom properties, webhooks, deployments, and environments remain mostly raw API workflows. |
| `actions` | 187 | `run`, `workflow`, `cache`, `secret`, and `variable` exist, but hosted runners, pending deployments, artifacts, jobs, annotations, and org storage diagnosis are thin. |
| `orgs` | 108 | `org list` exists, but roles, outside collaborators, fine-grained PATs, rules, webhooks, API insights, and custom properties are mostly missing. |
| `issues` | 55 | Good base issue CRUD, plus ghx issue search improvements. Issue dependencies, issue field values, issue types, timeline, and sub-issue workflows need richer operational commands. |
| `codespaces` | 48 | Existing command group is broad enough for now. |
| `users` | 47 | User lookup exists indirectly, but global user administration and audit workflows remain raw. |
| `apps` | 37 | Mostly raw for app installations, permissions, and webhook operations. |
| `activity` | 32 | Notifications and subscriptions are not a first-class agent queue. |
| `teams` | 32 | Org/team access diagnosis is raw or spread across API calls. |
| `copilot` | 31 | CLI has `copilot`, but metrics, user management, content exclusion, and coding-agent policy are not account-safe admin workflows. |
| `agents` | 30 | Coding-agent task and policy flows need account-safe handoff and status commands. |
| `copilot-spaces` | 28 | Spaces operations are raw API candidates until a concrete workflow emerges. |
| `packages` | 27 | Package inventory and cleanup are raw or dashboard-oriented. |
| `pulls` | 27 | PR CRUD is strong, but merge queue, thread, and policy diagnosis remain thin. |
| `projects` | 26 | `project` exists, but Projects v2 status, field health, and PR sync workflows remain partial. |
| `dependabot` | 25 | No unified security inbox. |
| `migrations` | 22 | Organization migration status and recovery remain raw API workflows. |
| `code-scanning` | 21 | No first-class security inbox. |
| `code-security` | 20 | Security configuration rollout and policy diagnosis remain raw. |
| `checks` | 12 | `pr checks` exists, and ghx now has a first checks inventory/rerun dry-run slice. |
| `secret-scanning` | 9 | No first-class security inbox. |
| `agent-tasks` | 5 | `agent-task` exists, but handoff, progress, and task-to-PR workflows are still immature. |

## REST gaps to promote first

| Surface | Example operation IDs | Current CLI coverage | First `ghx` command | Why this helps agents and us |
| --- | --- | --- | --- | --- |
| PR gate diagnosis | `repos/get-repo-rulesets`, `repos/get-repo-rule-suites`, `checks/list-for-ref` | Spread across `pr`, `ruleset`, `run`, and raw API | `ghx pr ready`, `ghx pr gate explain` | One answer for "can this PR merge, and what exact blocker remains?" |
| Review thread triage | GraphQL `PullRequestReviewThread` | No focused command | `ghx pr threads --unresolved` | Avoid browser-only review cleanup and brittle parsing of large review bodies. |
| Failed check inventory | `checks/list-for-ref`, `checks/rerequest-run`, `checks/rerequest-suite`, workflow job/run endpoints | `pr checks` and `run rerun` do not inventory open PR failures | `ghx checks inventory`, `ghx checks rerun --dry-run` | After shared CI fixes, rerun only relevant failed checks on open PRs. |
| Pending deployments | `actions/get-pending-deployments-for-run`, environment protection endpoints | Mostly raw API | `ghx env pending`, `ghx env approve`, `ghx env reject` | Merge blockers often sit behind environment approvals, not only checks. |
| Deployment timeline | Deployment and deployment-status endpoints, GraphQL `Deployment` | Mostly raw API | `ghx deploy timeline` | Trace preview, staging, and production state from a branch or PR. |
| Hosted runners | `actions/list-hosted-runners-for-org`, `actions/get-hosted-runners-limits-for-org`, image and machine-size endpoints | No first-class operations | `ghx runners status`, `ghx runners capacity`, `ghx runners images` | Debug capacity, label drift, image changes, and expensive CI behavior without dashboard spelunking. |
| Artifact and cache hygiene | Artifact, cache usage, retention, and storage-limit endpoints | `run download` and `cache list/delete` are narrow | `ghx actions storage report`, `ghx artifact sweep`, `ghx cache top` | Find storage/cost regressions and stale branch waste across repos and orgs. |
| Security alerts | Code scanning, secret scanning, Dependabot, security advisory endpoints | No unified queue | `ghx sec inbox`, `ghx sec report` | One cross-product remediation queue with stable JSON. |
| Org access and tokens | Org roles, members, outside collaborators, fine-grained PAT request endpoints | `org list` only | `ghx org access why`, `ghx org tokens`, `ghx org apps` | Explain why a user, app, or token can or cannot mutate a repo. |
| Webhook delivery debugging | Org/repo webhook delivery and redelivery endpoints | Raw API | `ghx hooks deliveries`, `ghx hooks redeliver --dry-run` | Diagnose automation outages without copying delivery IDs from the browser. |
| API insights | Org API Insights endpoints | Raw API | `ghx api insights` | Track which tokens/apps are consuming API budget or failing routes. |
| Copilot and coding agent admin | Copilot metrics, user management, content exclusion, coding-agent policy endpoints | `copilot` is user-facing | `ghx copilot admin`, `ghx agent policy` | Keep agent permissions and policy visible from the same account-safe CLI. |
| Project v2 status | Project item, field, view, and workflow endpoints plus GraphQL `ProjectV2` | `project` exists but is not a status cockpit | `ghx board status`, `ghx board sync-pr` | Keep issue trees, PR state, and project fields synchronized. |

## GraphQL gaps to promote first

| GraphQL object | Fields confirmed in live schema | Current CLI coverage | First `ghx` command |
| --- | --- | --- | --- |
| `PullRequestReviewThread` | `isResolved`, `isOutdated`, `path`, `line`, `comments`, `viewerCanResolve`, `viewerCanUnresolve` | No focused thread triage | `ghx pr threads --unresolved` |
| `MergeQueue` and `MergeQueueEntry` | `entries`, `position`, `state`, `estimatedTimeToMerge`, `pullRequest` | Hidden behind merge behavior | `ghx mq status`, `ghx pr ready` |
| `RepositoryRuleset` and `RepositoryRule` | `conditions`, `enforcement`, `rules`, `bypassActors`, `source`, `target` | `ruleset` can list/view/check, but not explain PR blockers | `ghx rules why-blocked` |
| `RuleSuite` | Present in REST as repo/org rule-suite endpoints | Not surfaced as diagnosis | `ghx rules explain --suite` |
| `Environment` | `protectionRules`, `latestCompletedDeployment` | Raw API | `ghx env pending`, `ghx deploy status` |
| `Deployment` and `DeploymentStatus` | `latestStatus`, `environmentUrl`, `logUrl`, `task`, `state` | Raw API | `ghx deploy timeline` |
| `WorkflowRun`, `CheckRun`, `CheckSuite` | `pendingDeploymentRequests`, `annotations`, `isRequired`, `steps`, `detailsUrl`, `matchingPullRequests` | Split between `run`, `workflow`, `pr checks`, and raw API | `ghx ci doctor`, `ghx checks inventory` |
| `SecurityAdvisory` and `RepositoryVulnerabilityAlert` | `severity`, `state`, `vulnerableManifestPath`, `dependabotUpdate` | Raw API or no command | `ghx sec inbox` |
| `ProjectV2` and `ProjectV2Item` | `fields`, `items`, `views`, `workflow`, `fieldValueByName`, `content` | Partial `project` CRUD | `ghx board status`, `ghx board sync-pr` |
| `Organization` and `Team` | `rulesets`, `repositoryCustomProperties`, `samlIdentityProvider`, `membersWithRole`, `reviewRequestDelegation*` | `org list` plus raw API | `ghx org access why`, `ghx org rules`, `ghx team review-load` |

## Output contract for generated coverage

`ghx mine github` emits a summary report with optional detail rows. JSON output
keeps these top-level objects stable:

- `commands`: local Cobra command inventory, including command path, runnable
  status, hidden/deprecated flags, aliases, and JSON fields.
- `rest`: REST OpenAPI inventory merged with `internal/ghapi/rest` coverage
  metadata. Rows include `operationId`, `method`, `path`, `tag`, `summary`,
  `docsUrl`, `coverageState`, `registered`, `localCommand`, `proposedCommand`,
  `rawCommand`, `pagination`, and notes.
- `graphql`: GraphQL schema inventory for root query and mutation fields. Rows
  include `coordinate`, `parentType`, `name`, `kind`, `description`,
  `returnType`, `args`, `deprecated`, `coverageState`, `proposedCommand`, and
  `rawCommand`.

The nested `rest` summary also reports total operation count, matching row
count, explicit metadata count, remaining explicit metadata gap, state counts,
and tag counts. The nested `graphql` summary reports host, schema hash, root
type names, type counts, field counts, deprecated field count, state counts,
and explicit coverage percentage.

## Generator commands

```sh
ghx mine github --source rest --format md > docs/ghx-api-coverage.md
ghx mine github --source rest --format json --rest-openapi /tmp/github-rest-openapi.json
ghx mine github --source graphql --format json --graphql-schema /tmp/github-graphql-schema.json
ghx mine github --source all --format md --detail --limit 100
```

The generator should compare official GitHub API operations against:

- high-level `ghx` command coverage
- generated proxy coverage
- raw-only `ghx api` coverage
- missing permission notes
- missing pagination helpers
- missing examples

It should also emit a machine-readable companion file so CI can detect drift when GitHub ships new operations in high-priority tags.
