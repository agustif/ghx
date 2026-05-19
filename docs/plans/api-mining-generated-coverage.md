# API mining and generated coverage automation

Status: draft implementation plan
Epic: https://github.com/agustif/ghx/issues/20
Branch: `af/plan-api-mining-generated-coverage`

This plan turns the manual REST/OpenAPI and GraphQL mining into repeatable local and CI workflows.

## Issue set

| Issue | Scope |
| --- | --- |
| https://github.com/agustif/ghx/issues/35 | `ghx mine github` REST coverage report |
| https://github.com/agustif/ghx/issues/36 | GraphQL schema validation |
| https://github.com/agustif/ghx/issues/37 | `ghx api explain` generated operation metadata |
| https://github.com/agustif/ghx/issues/38 | Spec drift workflow and coverage snapshots |
| https://github.com/agustif/ghx/issues/39 | Generated REST proxy pilot for Actions, checks, and deployments |

## DAG

1. Add pinned source manifest for REST OpenAPI and GraphQL schema snapshots.
2. Generate a coverage report from REST tags and local command inventory.
3. Add GraphQL operation validation for curated ghx queries.
4. Build `ghx api explain` around operation metadata before generated clients are used by commands.
5. Pilot generated REST adapters on a narrow Actions/Checks/Deployments subset.
6. Add CI drift detection once generated output is stable and reviewable.

## Acceptance

- Generated reports reproduce `docs/ghx-api-coverage.md` shape with stable columns.
- Drift output links to official source URLs and the local command or proxy state.
- Generated adapters expose raw escape hatches and do not replace existing commands wholesale.
- CI opens reviewable drift PRs instead of failing on every upstream spec change.

## Validation

- `go test ./internal/...`
- `go test ./api/...`
- Generated coverage command against the pinned REST snapshot.
- GraphQL schema validation against GitHub.com with `GH_ACCOUNT=agustif`.
