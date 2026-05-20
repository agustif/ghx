# API mining and generated coverage automation

Status: generated miner implemented, drift workflow pending
Epic: https://github.com/agustif/ghx/issues/20
Branch: `codex/ghx-api-coverage-miner`

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

1. Done for REST seed: add pinned source manifest for the curated REST OpenAPI subset.
2. Done for first miner: generate REST coverage from OpenAPI and local command inventory.
3. Partial: add GraphQL schema inventory from live introspection or schema JSON. Curated GraphQL operation validation remains.
4. Done: build `ghx api explain` around operation metadata before generated clients are used by commands.
5. Pilot generated REST adapters on a narrow Actions/Checks/Deployments subset.
6. Add CI drift detection once generated output is stable and reviewable.

## Implemented slice: `ghx api explain`

Worker C added the first concrete metadata-backed explain surface without adding
a giant generated client:

- `internal/ghapi/rest` contains a curated generated-operation registry seeded
  from `github/rest-api-description`.
- `pkg/cmd/api` registers `ghx api explain <operation-id>` and `ghx api explain
  --list`.
- Explain output includes method, path, params, scopes or permission notes,
  pagination, source, docs URL, coverage state, proposed command, and the raw
  `ghx api` escape hatch.
- JSON output is field-selected through the existing `--json`, `--jq`, and
  `--template` command contract.

## Implemented slice: `ghx mine github`

This branch adds the first full-surface miner:

- `pkg/cmd/mine` registers `ghx mine <command>` as a root command.
- `pkg/cmd/mine/github` implements `ghx mine github`.
- `internal/ghapi/coverage` builds command, REST, and GraphQL report models.
- REST mining reads the pinned OpenAPI URL by default, accepts a file or URL
  through `--rest-openapi`, and uses the plain HTTP client for public spec
  downloads so auth headers are not sent to `raw.githubusercontent.com`.
- GraphQL mining live-introspects the selected host or reads deterministic
  introspection JSON through `--graphql-schema`.
- Markdown output is summary-first and uses `--detail --limit N` for rows.
- JSON output is untruncated and intended for agents, `jq`, CI, and future drift
  checks.

Current command shape:

```sh
ghx mine github --source all --format md
ghx mine github --source rest --format json --rest-openapi /tmp/github-rest-openapi.json
ghx mine github --source graphql --format json --graphql-schema /tmp/github-graphql-schema.json
ghx mine github --source rest --tag actions --state missing --detail
```

Current quantitative gap:

| Metric | Value |
| --- | ---: |
| REST operations in current OpenAPI snapshot | 1186 |
| ghx operations with explicit generated metadata | 13 |
| Explicit REST metadata coverage | 1.1% |
| Remaining explicit REST metadata gap | 1173 |
| GraphQL schema validation coverage | 0% |

Reproduce the REST denominator and top tags with:

```sh
script/ghx-rest-coverage-summary /tmp/github-rest-openapi.json
```

Pinned REST source used for this slice:

| Field | Value |
| --- | --- |
| Source | `github/rest-api-description` |
| Ref | `133d385dfbee06825d4d4136a82dd2b4c79813ba` |
| URL | `https://raw.githubusercontent.com/github/rest-api-description/133d385dfbee06825d4d4136a82dd2b4c79813ba/descriptions/api.github.com/api.github.com.json` |
| Checksum | `sha256:93b14ec8053fde77ac78837e73db9346e3f8802fb4bf4b801ff4269dad89c4ca` |
| Runtime API version | `2022-11-28` |

Seeded operations for the first Actions, checks, and deployments subset:

| Operation id | Coverage state | Proposed command |
| --- | --- | --- |
| `actions/get-pending-deployments-for-run` | `raw-api` | `ghx env pending` |
| `actions/get-hosted-runners-limits-for-org` | `missing` | `ghx runners capacity` |
| `actions/list-hosted-runners-for-org` | `missing` | `ghx runners status` |
| `actions/list-jobs-for-workflow-run` | `thin` | `ghx checks inventory` |
| `actions/list-workflow-run-artifacts` | `thin` | `ghx actions storage report` |
| `actions/review-pending-deployments-for-run` | `raw-api` | `ghx env approve`, `ghx env reject` |
| `checks/list-for-ref` | `thin` | `ghx pr gate explain`, `ghx checks inventory` |
| `checks/rerequest-run` | `raw-api` | `ghx checks rerun --dry-run` |
| `checks/rerequest-suite` | `raw-api` | `ghx checks rerun --dry-run` |
| `repos/create-deployment-status` | `raw-api` | `ghx deploy status` |
| `repos/get-deployment` | `raw-api` | `ghx deploy timeline` |
| `repos/list-deployment-statuses` | `raw-api` | `ghx deploy timeline` |
| `repos/list-deployments` | `raw-api` | `ghx deploy timeline` |

Issue coverage:

| Issue | Coverage from this slice | Remaining work |
| --- | --- | --- |
| https://github.com/agustif/ghx/issues/20 | Adds reusable API metadata, explain surface, and generated miner. | Add drift workflow and generated proxy pilot. |
| https://github.com/agustif/ghx/issues/35 | Generates full REST coverage reports from a pinned or supplied OpenAPI snapshot. | Add reviewed coverage overlays beyond the 13-operation seed. |
| https://github.com/agustif/ghx/issues/36 | Adds GraphQL schema inventory from live introspection or schema JSON. | Add curated GraphQL operation validation and coverage registry. |
| https://github.com/agustif/ghx/issues/37 | Implements `ghx api explain <operation-id>` with JSON and list output. | Expand generated metadata as more operations are mined. |
| https://github.com/agustif/ghx/issues/38 | Source ref and checksum are present for future drift comparison. | Add scheduled drift workflow and reviewable snapshots. |
| https://github.com/agustif/ghx/issues/39 | Seeds Actions, checks, and deployments metadata without touching handwritten commands. | Add typed adapter pilot after metadata review. |

## Acceptance

- Generated reports reproduce `docs/ghx-api-coverage.md` shape with stable columns.
- Drift output links to official source URLs and the local command or proxy state.
- Generated adapters expose raw escape hatches and do not replace existing commands wholesale.
- CI opens reviewable drift PRs instead of failing on every upstream spec change.

## Validation

- `go test ./internal/ghapi/rest ./pkg/cmd/api`
- `go test ./internal/ghapi/coverage ./pkg/cmd/mine/...`
- `go test ./internal/...`
- `go test ./api/...`
- `ghx mine github --source rest --format json --rest-openapi /tmp/github-rest-openapi.json`
- `ghx mine github --source graphql --format json --graphql-schema /tmp/github-graphql-schema.json`
- Live GraphQL schema validation against GitHub.com with `GH_ACCOUNT=agustif`.
