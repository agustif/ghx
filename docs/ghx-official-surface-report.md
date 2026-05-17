# ghx official surface report

Status: placeholder
Date: 2026-05-17

This file is the target location for generated reports that mine official GitHub docs, schemas, changelog entries, and official repos for terminal workflow gaps.

## Intended generator

Future command:

```sh
ghx mine github --source docs --format md > docs/ghx-official-surface-report.md
```

## Sources

- `github/rest-api-description`
- GitHub GraphQL public schema
- `github/docs`
- GitHub changelog
- official GitHub CLI extensions and adjacent repos

## Initial high-value gaps

| Product surface | Current state | Candidate command |
| --- | --- | --- |
| PR readiness | spread across `pr`, checks, reviews, rules, deployments | `ghx pr ready` |
| Review threads | mostly GraphQL/raw browser workflow | `ghx pr threads` |
| Merge queue | no focused terminal workflow | `ghx mq status` |
| CI failure diagnosis | `run view` exists but no blocker explanation | `ghx ci doctor` |
| Pending deployments | mostly raw API | `ghx env pending` |
| Ruleset diagnosis | commands exist but diagnosis is thin | `ghx rules explain` |
| Security alerts | no unified terminal inbox | `ghx sec inbox` |
| Org access and PAT governance | mostly raw API | `ghx org access why` |
| Extension ecosystem | install/search exists but no curated bundle | `ghx ext bundle` |

## Required report fields

- product area
- official source URL
- local command coverage
- API coverage
- extension overlap
- safety risk
- proposed `ghx` action
