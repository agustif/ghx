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
| Official surface mining | Manual docs/API review only | `github/rest-api-description`, public GraphQL schema, `github/docs`, GitHub changelog | `none` | `ghx mine github` generates gap reports and candidate command specs from official sources |
| Workflow automation | `workflow`, `run`, and `agent-task` are primitives | `github/gh-aw`, Actions, issues, PRs, safe outputs, approval gates | `partial` | Adopt `gh aw` as a first-class companion for durable repo workflows, not as a replacement for interactive `ghx` |
| Extension ecosystem | `extension install/search/list/upgrade` exists | `gh-extension` topic and installed extensions | `thin` | `ghx ext bundle`, curated extension manifests, provenance checks, account-aware wrappers |

## Highest-value additions

### 0. Generated validated API proxies

Goal: make GitHub's REST and GraphQL surface usable without hand-writing fragile `ghx api` calls, while preserving raw escape hatches for new or unusual endpoints.

This should be an internal substrate for the rest of the roadmap, not a replacement for `ghx api`.

Commands and packages:

- `internal/ghapi/restgen`: generator driver for REST OpenAPI input.
- `internal/ghapi/rest`: generated typed REST operations, request structs, response structs, enums, pagination helpers, and endpoint metadata.
- `internal/ghapi/graphqlgen`: generator driver for GraphQL schema and curated operation files.
- `internal/ghapi/graphql`: generated GraphQL variable/input/output structs for curated operations.
- `ghx api discover <keyword>`: endpoint discovery over generated metadata.
- `ghx api explain <operation-id>`: method, path, required params, scopes/permissions when known, previews, response type, pagination style.

REST source:

- Use GitHub's public `github/rest-api-description` OpenAPI descriptions as the canonical REST input.
- Vendor a pinned bundled spec snapshot under `internal/ghapi/specs/rest/`.
- Keep the API version explicit. The current client hardcodes `X-GitHub-Api-Version: 2022-11-28`; generated clients should make this configurable while defaulting to the repo's chosen supported version.
- Generate per-operation metadata from `operationId`, method, path, parameters, request body schema, response schemas, pagination shape, and vendor extensions where present.

GraphQL source:

- Use the public GraphQL schema or schema introspection output as the canonical type input.
- Do not try to generate every possible GraphQL query as a command. GraphQL is a typed graph, not an endpoint list.
- Generate schema types and validate curated `.graphql` operation files for the workflows we own, such as PR readiness, review threads, merge queue, projects, rulesets, and inbox.
- Keep the current raw `Client.GraphQL` path as the fallback for exploratory queries.

Generator options:

- `oapi-codegen` is a pragmatic first candidate for Go because it generates clients/types from OpenAPI 3 and allows custom templates.
- `ogen` is worth a spike because it advertises generated validation, no reflection, typed request structures, and stronger OpenAPI-driven request parsing.
- ReadMe-style SDK generation is useful as a product reference: ReadMe treats OpenAPI as a source for docs, request builders, code examples, and SDK-generated code. That is the same direction we want for CLI help and examples.
- Speakeasy, Fern, Stainless, Kiota, and OpenAPI Generator are worth comparing, but the first shipping slice should avoid adding a paid/cloud generator to the normal build unless it clearly outperforms Go-native generation.

Proxy design:

```go
type RequestOptions struct {
	Headers    map[string]string
	Query      map[string][]string
	RawBody    io.Reader
	APIVersion string
	Preview    []string
	Unchecked  bool
}

type Operation[TParams any, TResponse any] interface {
	ID() string
	Method() string
	Path(params TParams) (string, error)
	Validate(params TParams) error
	Do(ctx context.Context, client *api.Client, params TParams, opts ...RequestOption) (TResponse, error)
}
```

Validation rules:

- Required path/query/body params are checked before the request.
- Enums become typed constants.
- Date, URI, integer, boolean, and array shapes are parsed before dispatch.
- Mutually exclusive query params and one-of request bodies should be enforced where the spec is precise enough.
- Unknown fields are rejected by default in generated calls.
- Pagination helpers should expose `AllPages`, `EachPage`, and `FirstPage`.

Escape hatches:

- `RequestOptions.Unchecked`: skip generated param validation but still use auth, host, telemetry, cache, and error handling.
- `RequestOptions.Headers` and `Preview`: opt into custom media types, previews, and new API headers.
- `RequestOptions.Query`: add unknown query params for newly shipped API filters.
- `RequestOptions.RawBody`: send raw JSON or stream bodies when the spec lags.
- `ghx api` remains the fully raw CLI path for REST and GraphQL.
- Generated operations should expose `RawPath(params)` and `OperationID` so a user can drop to `ghx api` with the exact endpoint.

Command integration:

- First use generated proxies behind new `ghx` surfaces, not by rewriting all old commands.
- Good first targets are `ghx pr ready`, `ghx ci doctor`, `ghx rules explain`, and `ghx sec inbox`.
- Existing handwritten `api/queries_*.go` remains valid until a command is touched.
- Generated code lives behind small adapter interfaces so command tests can mock operations without depending on huge generated structs.

CI and drift control:

- Pin the upstream OpenAPI and GraphQL schema snapshot.
- `go generate ./internal/ghapi/...` regenerates all proxy code.
- CI checks generated code is clean.
- A scheduled workflow opens a PR when specs change.
- A generated `docs/ghx-api-coverage.md` reports which operations have high-level CLI commands, which only have generated proxies, and which remain raw-only.

First implementation slice:

1. Vendor one small REST spec subset for Actions workflow runs, workflow jobs, and pending deployments.
2. Generate metadata plus typed request structs, not the full GitHub API.
3. Build `ghx api explain <operation-id>`.
4. Build `internal/ghapi/rest/workflow_runs` wrappers used by a read-only `ghx ci doctor` prototype.
5. Add `--unchecked` and `--raw-field` escape hatches to the prototype command.
6. Expand once the generated code shape is proven reviewable.

### 0.1 Official GitHub surface mining

Goal: keep the roadmap fed from the full depth of GitHub's official docs, public schemas, and maintained repos instead of only from the current `gh` command tree.

Inputs to mine:

- `github/rest-api-description`: operation ids, categories, parameters, response schemas, pagination, previews, and vendor extensions.
- GitHub GraphQL public schema: objects, mutations, deprecations, connections, and query cost surfaces.
- `github/docs`: REST and GraphQL article structure, product terminology, permission notes, examples, deprecation notes, and "in this article" navigation.
- GitHub changelog and release notes: newly shipped endpoints, deprecations, preview exits, and product feature flags.
- Official GitHub extension repos such as `github/gh-aw`, `github/gh-stack`, `github/gh-actions-importer`, `github/gh-gei`, `github/gh-copilot`, and `github/gh-models`.

Proposed commands:

- `ghx mine github --source rest --format md`: compare official REST operations against first-class `ghx` commands and generated proxies.
- `ghx mine github --source graphql --operation-dir internal/ghapi/graphql/operations`: validate curated GraphQL operations against the latest public schema.
- `ghx mine github --source docs --area actions`: scan docs navigation for product surfaces that have no local command group.
- `ghx mine github --source extensions --topic gh-extension`: score extension candidates by owner, update recency, license, stars, install shape, and overlap with our roadmap.

Output artifacts:

- `docs/ghx-api-coverage.md`: operation-level coverage.
- `docs/ghx-official-surface-report.md`: docs/product areas with no usable terminal surface.
- `docs/ghx-extension-bundle.md`: extension adoption candidates and wrapper decisions.
- `internal/ghapi/specs/manifest.json`: pinned spec/schema versions and checksums.

This mining should be automated as a scheduled workflow and as a local command. It should never make a feature decision by itself; it should produce evidence, proposed slices, and exact source links.

### 0.2 Workflow and extension adoption

Goal: use proven GitHub CLI extensions and workflow tools before rebuilding a whole product area in-tree.

`gh aw` decision:

- Adopt `github/gh-aw` as a first-class companion for durable repo workflows.
- Do not replace `ghx` with `gh aw`. `ghx` remains the interactive local control plane for identity, account binding, typed API access, diagnostics, and human-in-the-loop commands.
- Use `gh aw` when the task should live in the repo and run through GitHub Actions with guardrails, approval gates, safe outputs, and audit history.
- Use `ghx` when the task needs local repo/account context, immediate terminal feedback, cross-repo diagnosis, or generated REST/GraphQL calls.

Adoption commands:

- `ghx workflows init`: bootstrap our preferred `gh aw` workflow templates into a repo.
- `ghx workflows doctor`: validate workflow files, permissions, safe-output usage, and pinned dependencies.
- `ghx workflows run <name>`: call through to `gh aw` or `gh workflow run` with scoped-account confirmation.
- `ghx workflows watch`: combine `gh aw` run state, Actions logs, progress log output, and PR/issue links.

Extension bundle model:

- `ghx ext bundle list`: show curated extension bundles.
- `ghx ext bundle install agent`: install pinned extensions used by our agent workflows.
- `ghx ext bundle audit`: report unpinned, stale, unsigned, archived, or scope-sensitive extensions.
- `ghx ext wrap <extension>`: expose an account-aware wrapper that injects `.ghaccount`/`GH_ACCOUNT_SESSION` context and consistent JSON/progress behavior when possible.

Initial bundle candidates:

| Extension | Owner | Adopt posture | Why |
| --- | --- | --- | --- |
| `gh aw` | `github/gh-aw` | first-class companion | Agentic workflows in repo-owned Actions with guardrails and approval gates. |
| `gh stack` | `github/gh-stack` | first-class companion | Official stacked PR workflow; do not reinvent stack mechanics unless ghx needs account-aware wrappers. |
| `gh attach` | `enthus-appdev/gh-attach` | wrap and watch | Already installed locally; uploads images to GitHub PRs/issues while preserving repo visibility, useful for screenshots and visual QA evidence. |
| `gh dash` | `dlvhdr/gh-dash` | optional bundle | Mature terminal dashboard; useful as an interactive review/status UI while ghx owns machine-readable diagnostics. |
| `gh pr-review` | `agynio/gh-pr-review` | spike | Direct overlap with `ghx pr threads`; mine behavior before building our own unresolved-thread UI. |
| `gh actions-cache` | `actions/gh-actions-cache` | wrap or supersede | Existing Actions cache management; ghx can add org/repo storage diagnosis and account safety around it. |
| `gh actions-importer` | `github/gh-actions-importer` | optional bundle | Official migration workflow; useful for repos moving into GitHub Actions. |

Bundle rules:

- Prefer install-by-pin over "latest" for anything agents call automatically.
- Prefer official GitHub-owned extensions for mutating workflows when feature coverage is close.
- Keep third-party extensions behind wrappers until license, maintenance, and behavior are reviewed.
- Never hide extension provenance. `ghx ext bundle list --json` should show owner, repo, version, pin, license, update age, and command mapping.
- Do not fork an extension unless we need `.ghaccount` awareness, stable JSON, noninteractive mode, progress logging, or security changes that upstream will not take.

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

1. Generated validated API proxy spike: small Actions REST subset plus `ghx api explain`.
2. `ghx ctx`: mostly local config plus existing auth APIs. It makes every later command safer.
3. `ghx pr ready --json`: read-only GraphQL/REST aggregation. High daily value.
4. `ghx pr threads`: focused read-only GraphQL surface that feeds `pr ready`.
5. `ghx ci doctor`: read-only REST aggregation over runs/jobs/logs.
6. `ghx rules explain`: mostly read-only REST/GraphQL rulesets.
7. `ghx deploy status` and `ghx env pending`: read-only deployment/environment gates.
8. `ghx sec inbox`: read-only security alert aggregation.
9. Mutating variants: approvals, deployment statuses, alert dismissal, ruleset import, cache/artifact pruning.

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
- GitHub REST OpenAPI description: https://github.com/github/rest-api-description
- GitHub GraphQL public schema: https://docs.github.com/en/graphql/overview/public-schema
- GitHub REST API versioning: https://docs.github.com/en/rest/about-the-rest-api/api-versions
- ReadMe OpenAPI support and SDK-generated code: https://docs.readme.com/main/docs/openapi
- oapi-codegen Go OpenAPI generator: https://github.com/oapi-codegen/oapi-codegen
- ogen Go OpenAPI generator: https://github.com/ogen-go/ogen
- GitHub Docs repo: https://github.com/github/docs
- GitHub changelog: https://github.blog/changelog
- GitHub CLI extension install docs: https://cli.github.com/manual/gh_extension_install
- GitHub CLI extension topic: https://github.com/topics/gh-extension
- GitHub Agentic Workflows: https://github.com/github/gh-aw
- GitHub Stacked PRs extension: https://github.com/github/gh-stack
- gh attach extension: https://github.com/enthus-appdev/gh-attach
