# Plan: Generated API proxies

Status: active
Date: 2026-05-17
Related ADR: [ADR 0002](../adr/0002-generated-validated-api-proxies.md)

## Goal

Use pinned GitHub REST OpenAPI and GraphQL schema inputs to generate validated operation helpers that make platform coverage safer than raw `ghx api`, while preserving raw escape hatches.

## First slice

Target only a small Actions subset:

- workflow runs
- workflow jobs
- workflow run attempts
- pending deployments

Build only what is needed for `ghx api explain` and a read-only `ghx ci doctor` prototype.

## Package shape

- `internal/ghapi/specs/rest/`: pinned REST spec snapshot or trimmed subset.
- `internal/ghapi/specs/graphql/`: pinned schema or introspection snapshot.
- `internal/ghapi/restgen/`: generator driver.
- `internal/ghapi/rest/`: generated REST helpers.
- `internal/ghapi/graphqlgen/`: GraphQL operation validator.
- `internal/ghapi/graphql/`: generated GraphQL types for curated operations.

## Command shape

- `ghx api discover <keyword>`
- `ghx api explain <operation-id>`
- `ghx api coverage`

## Generator evaluation

Spike order:

1. `oapi-codegen`: pragmatic Go baseline with OpenAPI 3 support and templates.
2. `ogen`: evaluate generated validation and typed request structures.
3. ReadMe-style SDK generation: use as a product reference for docs, examples, request builders, and generated SDK ergonomics.
4. Speakeasy, Fern, Stainless, Kiota, and OpenAPI Generator: compare if the first two options are not enough.

## Validation and escape hatches

Generated operations should validate:

- required path/query/body params
- enum values
- scalar shapes such as date, URI, integer, boolean, and arrays
- precise one-of or mutually exclusive fields
- pagination style

Escape hatches:

- `Unchecked`
- custom headers
- preview media types
- unknown query params
- raw body
- raw `ghx api`

## Acceptance checks

- `go generate ./internal/ghapi/...` produces stable output.
- generated code is clean in CI.
- `ghx api explain <operation-id>` prints method, path, params, response, pagination, previews, and permission notes when known.
- operation metadata exposes a raw path and operation id.
- first generated wrappers are mockable in command tests.
