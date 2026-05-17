# ghx API coverage

Status: placeholder
Date: 2026-05-17

This file is the target location for generated REST and GraphQL coverage reports.

## Intended generator

Future command:

```sh
ghx mine github --source rest --format md > docs/ghx-api-coverage.md
```

The generated report should compare official GitHub API operations against:

- high-level `ghx` command coverage
- generated proxy coverage
- raw-only `ghx api` coverage
- missing permission notes
- missing pagination helpers
- missing examples

## Initial tracked areas

| Area | Source | Expected first command |
| --- | --- | --- |
| Actions workflow runs | REST OpenAPI | `ghx ci doctor` |
| Actions workflow jobs | REST OpenAPI | `ghx ci doctor` |
| Pending deployments | REST OpenAPI | `ghx ci pending-deployments` |
| Pull request review threads | GraphQL schema | `ghx pr threads` |
| Merge queue | GraphQL schema | `ghx pr ready` |
| Rulesets | REST and GraphQL | `ghx rules explain` |
| Code scanning alerts | REST OpenAPI | `ghx sec inbox` |
| Secret scanning alerts | REST OpenAPI | `ghx sec inbox` |
| Dependabot alerts | REST OpenAPI | `ghx sec inbox` |

## Required report columns

- source category
- operation id or GraphQL type/field
- API version or schema snapshot
- local command
- generated proxy package
- coverage state
- permission and scope notes
- pagination style
- source URL
