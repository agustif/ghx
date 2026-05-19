# ghx rugged automation and CI operations

Status: draft
Date: 2026-05-18
Related docs: [Agent workflows](../ghx-agent-workflows.md), [Gap map](../ghx-gap-map.md), [RFC 0002](../rfcs/0002-agent-workflows-issue-trees-and-toolchain.md)

This note captures product ideas from an ArcusCI P0 incident where agents had to inspect private PR state, address review comments, deploy a CI runner fix, and rerun failed external checks across many open pull requests. The goal is not Arcus-specific behavior in `ghx`. The goal is a more rugged GitHub control plane for humans and agents doing high-pressure operational work.

## Incident lessons

The workflow exposed several classes of avoidable fragility:

- Shell interpolation made a PR comment command unsafe. Markdown backticks in a `gh pr comment --body "..."` command were executed by the shell before the API call.
- Long `zsh` one-liners failed on reserved variables and quoting edge cases. Multi-line automation needs a stable execution contract, not a user-login shell contract.
- `gh pr view --json latestReviews` produced review bodies that contained control characters and broke downstream `jq` parsing. Agents should not need to parse untrusted review bodies just to poll merge readiness.
- App-owned checks and external CI checks need a first-class inspect and rerun workflow. Operators should not need to reconstruct check-run IDs from PR pages and ad hoc API calls during an outage.
- Branch protection and review gates need explainable blocker output. The important fact was "one non-author approving review with write access is required", not just a generic blocked merge state.
- Failed-check inventory across open PRs is a real operational primitive. After fixing a shared runner bug, operators need to rerun only still-open PRs with relevant failed checks.
- Progress and handoff state needs to outlive chat context. The useful artifacts were branch, head SHA, proof commands, live version, deploy gate, failed PR list, and next action.

## Product principles

1. Prefer argument vectors, stdin, and files over shell-composed strings.
2. Every mutating command should have a dry-run or explain mode that shows the target repo, account, resolved IDs, and exact mutation.
3. Polling commands should support narrow JSON fields that exclude large or unsafe bodies by default.
4. CI and PR readiness commands should model external checks, GitHub Actions, branch policy, and review gates in one result.
5. Rerun workflows should be targetable by check name, PR, SHA, app, conclusion, and open/closed state.
6. Agent-facing output should include exact next commands, not only prose diagnostics.
7. The default posture should be "safe to resume": write progress, candidates, and proofs to durable files or issues when the task spans more than one command.

## API-backed additions beyond the incident

The 2026-05-19 API scan found that the most useful `ghx` additions are not one-command wrappers around single endpoints. They are terminal workflows that join related REST and GraphQL surfaces into one explainable state.

| Need | Official API surface | Why current `gh` is not enough | Candidate |
| --- | --- | --- | --- |
| Can this PR merge now? | PR review decision, `PullRequestReviewThread`, checks, rulesets, deployments, merge queue | Data is split across `pr`, `run`, `ruleset`, raw GraphQL, and the browser | `ghx pr ready`, `ghx pr gate explain` |
| What review threads block us? | GraphQL review threads with `isResolved`, `isOutdated`, path, line, and comments | `pr view` is not an actionable unresolved-thread queue | `ghx pr threads --unresolved` |
| Which failed checks need rerun after a shared fix? | REST checks, check suites, workflow jobs, open PRs, app-owned details URLs | `run rerun` is workflow-centric and too broad | `ghx checks inventory`, `ghx checks rerun` |
| Is a deploy gate blocking the merge? | Pending deployments, environments, deployment reviews, statuses | Mostly raw API or browser state | `ghx env pending`, `ghx deploy timeline` |
| Is policy blocking the merge? | REST/GraphQL rulesets, rule suites, branch protection, required deployments | `ruleset check` does not explain PR-specific blockers | `ghx rules why-blocked` |
| Are runners or images the real failure source? | Hosted runners, runner limits, machine sizes, images, self-hosted runner permissions | No first-class runner status command | `ghx runners status`, `ghx runners images` |
| What is wasting Actions storage? | Artifact, cache usage, retention, and storage-limit endpoints | Existing commands are item-level | `ghx actions storage report` |
| What security work is open? | Code scanning, secret scanning, Dependabot, security advisories | No unified CLI queue | `ghx sec inbox` |
| Why can or cannot this actor mutate the repo? | Org roles, teams, outside collaborators, SAML, fine-grained PAT requests, app installations | `org list` is not access diagnosis | `ghx org access why` |
| Did webhooks deliver to our automation? | Org/repo webhook delivery and redelivery endpoints | Raw API or browser-only | `ghx hooks deliveries` |
| Are agent tasks and Copilot policy configured safely? | REST `agent-tasks`, Copilot coding-agent policy, Copilot metrics and user management | Preview user commands do not cover admin posture | `ghx agent policy`, `ghx agent handoff` |
| Does the project board reflect reality? | Projects v2 fields, items, views, workflows, PR/issue links | `project` is CRUD, not health | `ghx board status`, `ghx board sync-pr` |

Deep exploration conclusion: build explainers before mutators. A read-only command that returns stable JSON, exact blockers, resolved account/repo context, source URLs, and the next command is immediately useful to agents. Mutation should come only after the read-only shape has proven itself in real incidents.

## Candidate command surfaces

### Shell-safe remote text

```sh
ghx pr comment <pr> --body-file -
ghx pr comment <pr> --body-literal 'short text without shell expansion'
ghx pr comment <pr> --from-template review-proof --field proof=proof.json
```

Acceptance:

- documentation must prefer `--body-file -` for multi-line agent comments
- `--body-literal` must never invoke a shell
- error messages should warn when a command likely lost markdown backticks to shell interpolation

### Merge gate explanation

```sh
ghx pr gate explain <pr> --json
ghx pr gate wait <pr> --watch --progress-log progress.log
```

Fields:

- `mergeStateStatus`
- `reviewDecision`
- `requiredApprovingReviewCount`
- `hasAuthorOnlyApproval`
- `missingReviewers`
- `statusChecks`
- `adminMergeAllowed`
- `nextActions`

The command should distinguish these states:

- checks still running
- checks failed
- branch is dirty
- branch protection requires a non-author review
- token cannot see or mutate the repository
- merge would be possible after approval without further code changes

### Check inventory and rerun

```sh
ghx checks inventory --repo FlatFilers/obvious --state open --name 'arcus/*' --conclusion failure --json
ghx checks rerun --repo FlatFilers/obvious --pr 20891 --name arcus/api-e2e-tests --dry-run
ghx checks rerun --repo FlatFilers/obvious --state open --name arcus/api-e2e-tests --conclusion failure --since 24h --confirm
```

Fields:

- `prNumber`
- `prUrl`
- `headSha`
- `headRef`
- `checkName`
- `conclusion`
- `detailsUrl`
- `checkRunUrl`
- `checkRunId`
- `checkSuiteId`
- `isOpen`
- `rerequestEndpoint`
- `rerunSupported`

Acceptance:

- inventory must skip closed PRs unless `--include-closed` is passed
- rerun must support dry-run output before mutation
- rerun must report app-owned check limitations distinctly from auth failures
- rerun must be able to target only failed checks matching a pattern
- output must be stable enough to pipe into a later rerun command

### External CI provider adapter

```sh
ghx ci provider add arcus --details-url-prefix https://arcus-turin.ci.obvi.dev
ghx ci provider inspect arcus --details-url <url> --json
ghx ci provider rerun arcus --pr 20891 --check arcus/api-e2e-tests --sha <sha> --dry-run
```

The provider contract should be generic:

- parse details URLs into run IDs and check names
- fetch provider run status
- link GitHub check-run identity to provider run identity
- express whether rerun should use GitHub check rerequest or provider API trigger
- support host-specific deploy gate checks when the provider is self-hosted

Arcus-specific behavior can live in config or an extension, not core `ghx`.

### Rugged script runner

```sh
ghx script run --shell bash --noprofile --norc --progress-log progress.log -- <script-file>
ghx script heredoc --body-file - -- gh pr comment 154 --repo FlatFilers/ArcusCI
```

This should not replace normal shells. It should provide safe defaults for agent automation:

- use `bash --noprofile --norc` unless a different shell is explicit
- fail on unset variables and pipeline failures when requested
- redact tokens in echoed commands
- record command, cwd, account, repo, duration, exit code, and stdout/stderr paths
- recommend `--body-file` for multi-line remote text

## Anti-patterns to catch

- embedding markdown backticks inside double-quoted shell strings for remote comments
- using a login shell for long polling loops
- broad `jq` over untrusted review bodies when only merge/check fields are needed
- rerunning every workflow when only one failed external check needs rerun
- deploying a self-hosted runner without a strict drain gate
- claiming CI is fixed before the live runner version proves the fix SHA

## First implementation slices

1. Add `ghx pr gate explain` as a read-only command that returns branch protection, review, and check blockers with next actions.
2. Add `ghx checks inventory` for open PR failed-check discovery with stable JSON and no rerun mutation.
3. Add `ghx checks rerun --dry-run` using GitHub check-run and check-suite rerequest endpoints.
4. Add `ghx pr comment --body-file -` examples and warnings to the agent workflow guide.
5. Add provider config for external CI details URLs, starting with Arcus as a local config example.

## Open questions

- Should `ghx checks rerun` require explicit `--confirm` for more than one PR?
- Should provider adapters live in core config, extensions, or `gh aw` workflows?
- Can `ghx` infer required approving review count across rulesets reliably on GitHub.com and GHES?
- Should long-running commands write a resumable state file in addition to progress logs?
