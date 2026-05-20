# CI, checks, deployments, and runner operations

Status: draft implementation plan
Epic: https://github.com/agustif/ghx/issues/19
Branch: `af/plan-ci-checks-deployments-runners`

This plan makes CI operations inspectable before adding high-risk rerun, approval, or deletion mutations.

## Issue set

| Issue | Scope |
| --- | --- |
| https://github.com/agustif/ghx/issues/28 | `ghx checks inventory` across open PRs |
| https://github.com/agustif/ghx/issues/29 | `ghx checks rerun` dry-run and targeted mutation |
| https://github.com/agustif/ghx/issues/30 | `ghx env pending` and approval flow |
| https://github.com/agustif/ghx/issues/31 | `ghx deploy timeline` for PRs and refs |
| https://github.com/agustif/ghx/issues/32 | `ghx runners status`, images, and capacity |
| https://github.com/agustif/ghx/issues/33 | Actions storage report for artifacts and caches |
| https://github.com/agustif/ghx/issues/34 | External CI provider adapter contract |

## DAG

1. Define shared check inventory JSON with PR number, head SHA, check name, app, details URL, and rerun support.
2. Implement read-only check inventory for one PR, then broaden to open PRs.
3. Add dry-run rerun target selection without mutation.
4. Add environment pending deployment read path.
5. Add deployment timeline read path.
6. Add runner and storage reports as separate read-only commands.
7. Extract external provider adapter shape after GitHub check identity is stable.

## Acceptance

- Inventory skips closed PRs unless explicitly requested.
- Rerun dry-run lists exact check-run or check-suite IDs before any mutation.
- Deployment and environment commands clearly separate pending approval from failed checks.
- Runner and storage reports do not mutate remote state.

## Validation

- `go test ./pkg/cmd/run/...`
- `go test ./pkg/cmd/pr/checks/...`
- `go test ./api/...`
- Manual smoke with a repository that has a failed Actions run and an app-owned check.

## Implementation Note

- First CLI slice adds `gh checks inventory` for read-only check inventory across open PRs or one PR.
- `gh checks rerun --dry-run` plans exact check-run rerequest targets and emits commands without mutating remote state.
- Deployment environments, deployment timelines, runners, storage reports, and external provider-specific reruns remain separate follow-up slices after check identity is stable.
