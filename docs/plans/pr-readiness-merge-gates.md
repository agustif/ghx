# PR readiness, review threads, and merge gates

Status: draft implementation plan
Epic: https://github.com/agustif/ghx/issues/18
Branch: `af/plan-pr-readiness-merge-gates`

This plan turns the merge-readiness roadmap into a small stack of read-only commands before any merge mutation changes.

## Issue set

| Issue | Scope |
| --- | --- |
| https://github.com/agustif/ghx/issues/23 | `ghx pr ready` read-only merge cockpit |
| https://github.com/agustif/ghx/issues/24 | `ghx pr gate explain` blocker diagnosis |
| https://github.com/agustif/ghx/issues/25 | `ghx pr threads` unresolved review triage |
| https://github.com/agustif/ghx/issues/26 | `ghx mq status` and merge-queue watch |
| https://github.com/agustif/ghx/issues/27 | `ghx rules why-blocked` PR policy explanation |

## DAG

1. Define shared `PRGateReport` and `PRBlocker` models with stable JSON fields.
2. Add a resolver for PR number, URL, branch, or current branch using existing PR lookup patterns.
3. Implement unresolved review-thread GraphQL query and command output.
4. Add read-only ruleset and merge-queue adapters behind interfaces.
5. Compose `ghx pr ready` from review, check, deployment, ruleset, and queue signals.
6. Add `--watch` only after single-shot JSON output is stable.

## Acceptance

- `ghx pr ready <pr> --json` includes account, repository, PR number, head SHA, ready boolean, blockers, and next actions.
- `ghx pr threads <pr> --unresolved --json` avoids fetching large review bodies unless explicitly requested.
- Ruleset and merge-queue failures are explainable without opening the browser.
- No merge mutation lands in this stack.

## Validation

- `go test ./pkg/cmd/pr/...`
- `go test ./api/...`
- Manual smoke against `agustif/ghx` with a PR that has no checks and a PR with review or policy blockers.

## Implementation Note

- First CLI slice uses `gh pr gate explain` for the read-only merge cockpit because upstream `gh pr ready` already mutates draft state by marking a PR ready for review.
- The JSON report includes account, repository, PR summary, ready boolean, merge state, review decision, check counts, merge queue booleans, blockers, and next actions.
- Detailed unresolved threads, full ruleset matching, and merge queue watch remain separate follow-up slices.
