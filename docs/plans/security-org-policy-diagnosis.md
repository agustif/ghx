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
