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
