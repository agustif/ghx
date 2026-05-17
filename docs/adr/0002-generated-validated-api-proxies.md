# ADR 0002: Generated validated API proxies

Status: Accepted
Date: 2026-05-17

## Context

`gh api` is powerful, but it forces users and agents to hand-write REST paths, GraphQL payloads, request fields, pagination, previews, and error handling. The GitHub REST API already has OpenAPI descriptions, and the GraphQL API has a public schema. These can drive safer generated code and richer command help.

The CLI should not wait for every platform surface to get a hand-written command before it becomes usable.

## Decision

`ghx` will add generated validated API proxy layers:

- `internal/ghapi/restgen`: generator driver for REST OpenAPI inputs.
- `internal/ghapi/rest`: generated REST operation metadata, request structs, response structs, enums, validation, pagination helpers, and path builders.
- `internal/ghapi/graphqlgen`: generator driver for schema and curated operation validation.
- `internal/ghapi/graphql`: generated GraphQL variable/input/output types for curated operations.

The REST source is a pinned snapshot from `github/rest-api-description`. The GraphQL source is the public schema or introspection output plus curated `.graphql` operations for workflows owned by `ghx`.

Generated proxies do not replace `ghx api`. They sit between raw API access and high-level commands.

## Consequences

- The first generator spike should be narrow: Actions workflow runs, jobs, and pending deployments.
- Generated calls must validate required params, enum values, basic scalar shapes, one-of bodies where precise, and pagination shape before dispatch.
- Generated operations must expose operation id and raw path so users can drop to `ghx api`.
- Escape hatches remain first-class: unchecked validation, custom headers, preview media types, unknown query fields, raw bodies, and raw `ghx api`.
- SDK generators such as ReadMe-style tooling, Speakeasy, Fern, Stainless, Kiota, and OpenAPI Generator remain comparison inputs. The first shipping slice should prefer a Go-native generator such as `oapi-codegen` or `ogen` unless a stronger reason appears.

## Follow-ups

- Create `docs/plans/generated-api-proxies.md`.
- Add `ghx api discover <keyword>`.
- Add `ghx api explain <operation-id>`.
- Add `docs/ghx-api-coverage.md` generated from the pinned spec.
- Add CI that fails when generated code is dirty.
- Add a scheduled spec-update workflow that opens a PR with coverage diff.
