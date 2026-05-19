# ghx API coverage

Status: manual seed for generated reports
Date: 2026-05-19

This report compares the local `ghx` command surface against official GitHub REST and GraphQL surfaces. It is intentionally evidence-shaped so a later `ghx mine github` command can replace it with a generated report.

## Evidence snapshot

- Local command surface: `ghx help` and `pkg/cmd/*` in this checkout.
- REST surface: `github/rest-api-description` OpenAPI JSON downloaded on 2026-05-19.
- REST docs version: GitHub REST docs show API version `2026-03-10` as latest on 2026-05-19.
- GraphQL surface: live `ghx api graphql` introspection on 2026-05-19.
- Official CLI reference: `https://cli.github.com/manual/gh`.

High-volume REST tags from the OpenAPI snapshot:

| REST tag | Operations | Local command posture |
| --- | ---: | --- |
| `repos` | 201 | Broad `repo` coverage, but rules, rule suites, custom properties, webhooks, deployments, and environments remain mostly raw API workflows. |
| `actions` | 187 | `run`, `workflow`, `cache`, `secret`, and `variable` exist, but hosted runners, pending deployments, artifacts, jobs, annotations, and org storage diagnosis are thin. |
| `orgs` | 108 | `org list` exists, but roles, outside collaborators, fine-grained PATs, rules, webhooks, API insights, and custom properties are mostly missing. |
| `issues` | 55 | Good base issue CRUD, plus ghx issue search improvements. Issue dependencies, issue field values, issue types, timeline, and sub-issue workflows need richer operational commands. |
| `codespaces` | 48 | Existing command group is broad enough for now. |
| `apps` | 37 | Mostly raw for app installations, permissions, and webhook operations. |
| `activity` | 32 | Notifications and subscriptions are not a first-class agent queue. |
| `teams` | 32 | Org/team access diagnosis is raw or spread across API calls. |
| `copilot` | 31 | CLI has `copilot`, but metrics, user management, content exclusion, and coding-agent policy are not account-safe admin workflows. |
| `projects` | 26 | `project` exists, but Projects v2 status, field health, and PR sync workflows remain partial. |
| `dependabot` | 25 | No unified security inbox. |
| `code-scanning` | 21 | No first-class security inbox. |
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

Future generated reports should keep these columns stable:

- source category
- operation id or GraphQL type and field
- API version or schema snapshot
- local command coverage
- generated proxy package
- coverage state: `first-class`, `thin`, `raw-api`, `missing`
- permission and scope notes
- pagination style
- source URL
- proposed ghx command

## Generator target

Future command:

```sh
ghx mine github --source rest --format md > docs/ghx-api-coverage.md
```

The generator should compare official GitHub API operations against:

- high-level `ghx` command coverage
- generated proxy coverage
- raw-only `ghx api` coverage
- missing permission notes
- missing pagination helpers
- missing examples

It should also emit a machine-readable companion file so CI can detect drift when GitHub ships new operations in high-priority tags.
