# Security, org access, and policy diagnosis

Status: draft implementation plan
Epic: https://github.com/agustif/ghx/issues/21
Branch: `af/plan-security-org-policy`

This plan keeps security and access work read-only first, then adds dry-run mutations only after diagnosis output is stable.

## Issue set

| Issue | Scope |
| --- | --- |
| https://github.com/agustif/ghx/issues/40 | `ghx sec inbox` read-only alert queue |
| https://github.com/agustif/ghx/issues/41 | `ghx org access why` permission diagnosis |
| https://github.com/agustif/ghx/issues/42 | `ghx org tokens` and PAT request review dry-run |
| https://github.com/agustif/ghx/issues/43 | Copilot admin and coding-agent policy status |
| https://github.com/agustif/ghx/issues/44 | Org API insights and audit diagnostics |

## DAG

1. Define shared principal and permission evidence models.
2. Build `ghx sec inbox` read-only aggregation for code scanning, secret scanning, Dependabot, and advisories.
3. Build `ghx org access why` for user/team/app/token access evidence.
4. Add fine-grained PAT request inventory before review mutations.
5. Add Copilot and coding-agent policy status as read-only admin posture.
6. Add API insights summaries for actors, routes, failures, and time windows.

## Acceptance

- Security inbox output is grouped by repo, severity, alert kind, age, and state.
- Access diagnosis names the actor, permission source, and missing blocker without exposing secrets.
- PAT review and alert dismissal commands are dry-run only until read paths are validated.
- Copilot and API insight commands surface account and org context before any admin mutation.

## Validation

- Unit tests around response mapping and redaction.
- Smoke tests against an org/repo where the token has read-only security visibility.
- Negative tests for missing scopes and SAML/SSO authorization gaps.

## Worker D implementation pass

Status: partial implementation landed in scoped packages only.

Implemented command surfaces:

| Issue | Surface | Status |
| --- | --- | --- |
| https://github.com/agustif/ghx/issues/40 | `gh org security inbox <org>` | Read-only list of org code scanning, secret scanning, and Dependabot alerts with `--json`, `--kind`, `--state`, and `--limit`. |
| https://github.com/agustif/ghx/issues/41 | `gh org access why <user> --repo OWNER/REPO` | Read-only repository permission diagnosis with required permission ranking and JSON output. |
| https://github.com/agustif/ghx/issues/42 | `gh org tokens <org>` | Read-only fine-grained PAT request inventory with JSON output. |
| https://github.com/agustif/ghx/issues/42 | `gh org tokens review <org> --request-id ID --decision approve|deny --dry-run` | Dry-run only preview. No mutation request is sent. |

Follow-up nodes:

1. Extend `gh org access why` beyond the repository collaborator permission endpoint by adding team, outside collaborator, SAML, app installation, and token-scope evidence.
2. Add partial-success handling to `gh org security inbox` so missing security scopes report per-alert-kind blockers instead of failing the whole inbox.
3. Add a Copilot/admin read-only surface for https://github.com/agustif/ghx/issues/43 once the command ownership boundary for top-level `copilot` versus `org copilot` is agreed.
4. Add org audit/API insight summaries for https://github.com/agustif/ghx/issues/44 after selecting the audit-log endpoint shape and retention window.

Validation run:

```bash
go test ./pkg/cmd/org ./pkg/cmd/org/access ./pkg/cmd/org/security ./pkg/cmd/org/tokens
```
