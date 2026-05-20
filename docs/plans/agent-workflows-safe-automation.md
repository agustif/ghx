# Agent workflow, webhooks, boards, and safe automation

Status: draft implementation plan
Epic: https://github.com/agustif/ghx/issues/22
Branch: `af/plan-agent-workflows-safe-automation`

This plan makes ghx safer for long-running agent workflows that need durable state, copy-safe text, webhook diagnosis, and project-board alignment.

## Issue set

| Issue | Scope |
| --- | --- |
| https://github.com/agustif/ghx/issues/45 | `ghx hooks deliveries` and redelivery dry-run |
| https://github.com/agustif/ghx/issues/46 | `ghx board status` and PR sync |
| https://github.com/agustif/ghx/issues/47 | `ghx agent handoff` and task-to-PR views |
| https://github.com/agustif/ghx/issues/48 | `ghx script run` and shell-safe heredoc helpers |
| https://github.com/agustif/ghx/issues/49 | Shell-safe body-file warnings and literal body UX |

## DAG

1. Add copy-safe remote text rules to command help and docs before new shell helpers.
2. Build webhook delivery read path and dry-run redelivery target output.
3. Add Projects v2 board status for issue/PR field drift.
4. Add agent handoff views that join tasks, issues, branches, PRs, checks, and progress logs.
5. Add `ghx script run` only after the command recording and redaction contract is explicit.
6. Add body-literal or warning behavior for high-risk markdown inputs after compatibility review.

## Acceptance

- Webhook redelivery requires explicit delivery IDs and dry-run output.
- Board sync shows proposed field changes before mutation.
- Agent handoff output is stable JSON and includes exact next commands.
- Shell-safe helpers avoid login-shell defaults and redact token-looking values.

## Validation

- Unit tests for command output contracts and redaction.
- Manual smoke for body-file flows using stdin.
- Manual smoke for webhook delivery listing against a repo with hooks.

## Worker D implementation pass

Status: partial implementation landed in scoped packages only.

Implemented command surfaces:

| Issue | Surface | Status |
| --- | --- | --- |
| https://github.com/agustif/ghx/issues/49 | `gh issue create --body-literal` | Literal body flag added alongside existing `--body-file -` support. |
| https://github.com/agustif/ghx/issues/49 | `gh issue comment --body-literal` | Literal issue comment body flag added; keeps editor/web/delete mutual exclusion. |
| https://github.com/agustif/ghx/issues/49 | `gh issue edit --body-literal` | Literal issue edit body flag added; mutually exclusive with `--body` and `--body-file`. |
| https://github.com/agustif/ghx/issues/49 | Inline body warning | TTY warning for inline body text containing newlines, backticks, or `$(`; docs and examples prefer `--body-file -`. |
| https://github.com/agustif/ghx/issues/22 | `gh issue create --dry-run` | Safe create preview that prints the issue payload and sends no mutation. |

Existing coverage verified:

| Issue | Surface | Status |
| --- | --- | --- |
| https://github.com/agustif/ghx/issues/49 | `gh issue list --search "..." --match body,comments` | Already present. Focused tests kept this path covered. |
| https://github.com/agustif/ghx/issues/47 | `gh agent-task list --json ...` and `gh agent-task view --json ...` | Existing JSON-friendly agent task surfaces remain the current handoff base. |

Follow-up nodes:

1. Mirror `--body-literal` onto PR comment/create paths in a PR-owned lane.
2. Add `gh agent-task handoff` only after agreeing whether it belongs under `agent-task` or a new top-level `agent` control-plane command.
3. Implement webhook delivery diagnostics for https://github.com/agustif/ghx/issues/45 in a hooks-owned lane because a top-level `hooks` command crosses this worker boundary.
4. Implement Projects v2 board status and sync dry-run for https://github.com/agustif/ghx/issues/46 in a board/project-owned lane.
5. Implement `ghx script run` for https://github.com/agustif/ghx/issues/48 only after the command-recording artifact and redaction contract are finalized.

Validation run:

```bash
go test ./pkg/cmd/issue/shared ./pkg/cmd/issue/create ./pkg/cmd/issue/comment ./pkg/cmd/issue/edit ./pkg/cmd/issue/list
```
