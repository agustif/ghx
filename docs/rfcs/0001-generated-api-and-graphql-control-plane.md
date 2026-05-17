# RFC 0001: Generated GitHub REST and GraphQL control plane

Status: Draft
Date: 2026-05-17
Area: ghx API substrate, command integration, and drift control
Related ADRs:
- [ADR 0002](../adr/0002-generated-validated-api-proxies.md)
- [ADR 0003](../adr/0003-official-github-surface-mining.md)
- [ADR 0008](../adr/0008-product-discovery-record-system.md)

## Context

`ghx` already has a broad command surface, but many of the highest value workflows still depend on raw `gh api` calls or hand-written request logic. That makes the CLI brittle in the exact places where we want it to be safest: PR readiness, CI diagnosis, rules explanations, security inboxes, and future issue subissue automation.

GitHub already publishes enough source material to build a safer substrate:

- REST OpenAPI descriptions for most platform operations
- GraphQL schema and introspection output
- official docs and changelog content that describe permissions, previews, and product terminology
- existing local `ghx` command shapes that show which workflows matter first

This RFC proposes a generated API control plane that sits between raw API access and high-level commands. It should make the common cases safer, easier to explain, and easier to test, while still keeping raw escape hatches for endpoints that are new, unusual, or not yet modeled.

This is an implementation RFC, not a feature wish list. It should be specific enough to decompose into tracked issues and small PRs.

## Decision

`ghx` will introduce generated REST and GraphQL operation proxies backed by pinned upstream specs and curated operation registries.

The control plane must do four things:

1. Ingest pinned REST and GraphQL source snapshots into a stable local manifest.
2. Generate validated request and response shapes for a curated subset of operations.
3. Expose operation metadata for discovery, explain, and coverage reporting.
4. Preserve raw escape hatches for cases where the spec is incomplete or the user explicitly wants direct API access.

The first shipping slice should be narrow and read-only where possible. The initial goal is to prove the generated shape, the registry, the validation model, and command integration on a handful of workflows before broadening coverage.

## Scope

### In scope

- Pinned upstream spec ingestion for REST and GraphQL
- A manifest that records source versions, checksums, and generator inputs
- Generated operation proxies with typed params and validation
- Pagination helpers and metadata
- Raw request escape hatches
- GraphQL curated operation registry
- Drift detection and regeneration gates
- Command integration for the first targeted slices

### Out of scope

- Replacing `ghx api`
- Generating the entire GitHub REST or GraphQL surface up front
- Automatic product decisions from mining output
- Mutating workflows before the read-only substrate is proven
- Replacing hand-written command implementations everywhere at once

## Why this shape

The platform exposes too much surface area to model manually forever, and too much of the important surface is too dynamic to trust from memory alone.

The generator substrate should make the following true:

- A command author should not hand-write paths, body shapes, enum coercion, or pagination handling for a common official endpoint.
- A command should fail early on invalid params before it reaches the network.
- A user should be able to discover the exact operation id, raw path, and escape hatch for every generated call.
- A spec update should be visible as a deterministic diff, not a vague runtime surprise.

The result is safer than raw `ghx api`, but still less opinionated than a full high-level command rewrite.

## Design

### 1. Spec ingestion and pinned manifests

The repo should pin the upstream source inputs instead of generating from live network state during normal builds.

REST input:

- Pin a REST OpenAPI snapshot from `github/rest-api-description`.
- Store either the full snapshot or a trimmed, versioned subset in `internal/ghapi/specs/rest/`.
- Preserve the upstream operation ids, paths, methods, schema refs, parameter definitions, preview markers, and vendor extensions needed for validation and explain output.

GraphQL input:

- Pin a GraphQL schema snapshot or introspection result in `internal/ghapi/specs/graphql/`.
- Store curated operation documents separately from the schema snapshot so operation validation can be versioned and reviewed.

Manifest:

- Keep a pinned manifest at `internal/ghapi/specs/manifest.json`.
- Record source repository, source ref, snapshot timestamp, checksum, generator version, and included subsets.
- The manifest should be the single place that explains what spec inputs the generated code came from.

The manifest is not just bookkeeping. It is the reproducibility contract for codegen, CI drift detection, and release review.

### 2. Generated code shape

The generated packages should be small, explicit, and mockable.

Recommended package shape:

- `internal/ghapi/restgen/`: generator driver and input normalization
- `internal/ghapi/rest/`: generated REST operation metadata, request structs, response structs, validation, pagination helpers, and raw path helpers
- `internal/ghapi/graphqlgen/`: generator driver and schema validation for curated operations
- `internal/ghapi/graphql/`: generated GraphQL operation types and registry metadata

The codegen output should favor clear operation units over giant generic clients.

For REST operations:

- Each operation should have a stable operation id.
- Request params should be grouped by location: path, query, headers, body.
- Generated request types should encode requiredness and basic scalar constraints.
- Response types should mirror the smallest useful typed surface for the command layer.
- Each generated operation should expose a raw path builder for dropping to `ghx api`.

For GraphQL operations:

- Curated `.graphql` documents should define the actual operation boundary.
- Generated variables and result types should be specific to each curated operation.
- A registry should map operation name, purpose, schema version, and command consumers.

The output should be readable enough that a reviewer can inspect the generated shape without reverse engineering a framework abstraction.

### 3. Validation model

Generated proxies should validate before dispatch wherever the spec has enough information.

Validate:

- required path, query, and body params
- enum values
- scalar shapes such as integer, boolean, string, URI, date, and arrays
- mutually exclusive or one-of field combinations when the spec expresses them precisely
- pagination style and supported page controls
- known preview and media type requirements where the spec records them

Validation should fail locally and clearly before the request reaches the network.

Validation should not try to become a second hidden API parser. If the spec cannot express a constraint precisely, the generated proxy should prefer an explicit escape hatch over a fragile guess.

### 4. Pagination

Pagination should be first-class in the generated layer, not a command-specific afterthought.

The generated operation metadata should describe:

- whether the operation is paginated
- which pagination style it uses
- which request fields control paging
- whether the response is page-based, cursor-based, or link-header driven

The command layer should be able to consume the proxy through small helpers such as:

- `FirstPage`
- `EachPage`
- `AllPages`

Those helpers should hide transport details but preserve enough metadata for explain output and tests.

### 5. Raw escape hatches

The control plane should stay safe by default, but it must not trap the user inside the generated subset.

The generated request options should support explicit escape hatches:

- `Unchecked` to skip generated validation while keeping host selection, auth, telemetry, and error handling
- `Headers` to add custom headers and preview media types
- `Query` to add unknown query params when GitHub ships a new filter before the spec catches up
- `RawBody` to send raw JSON or streamed bodies when the body shape is not yet modeled
- `RawPath` and `OperationID` to let users drop directly to `ghx api`

The RFC should treat these as deliberate, documented options, not temporary hacks.

### 6. GraphQL operation registry

GraphQL needs a registry layer because the schema alone is too broad to map directly to command behavior.

The registry should store:

- operation name
- command consumer
- schema snapshot version
- required fields and variable types
- pagination behavior
- known permissions or product notes when they can be derived
- raw query document path
- explain metadata for `ghx api explain`

The registry is the bridge between the public schema and the command surface. It should make it easy to answer:

- which command uses this operation
- what the operation expects
- what changes when the schema drifts
- what raw document to inspect when a user wants to bypass the wrapper

### 7. Drift detection

Generated code must be tied to explicit drift checks.

The repo should enforce three drift gates:

1. `go generate ./internal/ghapi/...` is deterministic.
2. Generated code is clean in CI.
3. A scheduled update lane opens a PR when upstream specs change.

Drift detection should compare:

- pinned manifest entries
- generated file output
- operation ids and schema snapshots
- curated GraphQL operation validity against the pinned schema

When the generated output changes, the diff should make it obvious whether the cause is a new upstream field, a spec normalization change, or a local generator change.

## Command integration

Generated proxies should be introduced behind new `ghx` surfaces first, not by rewriting the entire command tree.

The first slices should consume the generated layer through small adapter interfaces so command tests can mock operations without depending on huge generated structs.

Initial target commands:

- `ghx ci doctor`
- `ghx pr ready`
- `ghx rules explain`
- `ghx sec inbox`

The command integration should favor a read-only first pass:

- `ghx ci doctor` should prove workflow-run, job, attempt, and pending-deployment wiring.
- `ghx pr ready` should prove PR state, review threads, merge queue, rules, and deployment blockers.
- `ghx rules explain` should prove ruleset and branch policy explanation.
- `ghx sec inbox` should prove alert aggregation across security products.

These commands do not need every final feature on day one. They do need enough real generated calls to prove the substrate is useful.

## First slices

The first implementation slice should be intentionally narrow.

### Slice 1: REST workflow diagnosis

Target REST operations that support `ghx ci doctor` and `ghx api explain`:

- workflow runs
- workflow jobs
- workflow run attempts
- pending deployments

This slice should prove:

- pinned REST spec ingestion
- generated request validation
- pagination helpers
- raw operation metadata
- read-only command integration

### Slice 2: PR readiness

Use a curated GraphQL operation set for `ghx pr ready`.

The first version should focus on:

- review decision
- unresolved review threads
- mergeability state
- merge queue state
- required checks
- required deployments or environment approvals when exposed

This slice should prove:

- GraphQL operation registry
- curated operation documents
- command-side adapters around generated output
- stable explain output for GraphQL-backed behavior

### Slice 3: Rules explanation

`ghx rules explain` should combine REST and GraphQL where needed to answer why a ref or PR is blocked.

The first version should focus on:

- matching rulesets
- branch protection
- required checks
- required review count
- review-thread resolution requirements
- deployment requirements

This slice should prove that the control plane can cross an API boundary without falling back to raw `gh api` for everything.

### Slice 4: Security inbox

`ghx sec inbox` should group alert products into one operational queue.

The first version should target read-only listing and grouping for:

- code scanning alerts
- secret scanning alerts
- Dependabot alerts

This slice should prove that the generated layer can support a higher-level queue shape without losing the underlying operation metadata.

## Testing

The test strategy should separate generator correctness from command behavior.

### Generator tests

- Validate pinned manifest parsing
- Verify deterministic generation from the same input snapshot
- Check that generated operation metadata includes operation id, raw path, and pagination style
- Verify validation failures for missing required params, invalid enums, and malformed scalars

### Registry tests

- Verify curated GraphQL operations register the correct schema snapshot and command consumer
- Verify explain metadata stays in sync with the registry
- Verify unknown or stale operations fail with a clear error

### Command tests

- Mock generated adapter interfaces at the command layer
- Prove `ghx ci doctor`, `ghx pr ready`, `ghx rules explain`, and `ghx sec inbox` can be exercised without live network calls
- Verify output includes the operation metadata needed for `--explain` and raw escape hatches

### Integration tests

- Add narrow HTTP mock coverage for the first REST slice
- Add schema-backed tests for curated GraphQL operations
- Keep tests focused on command behavior, not generated implementation details

The point of the tests is to make drift and contract breakage obvious, not to lock the generator into a particular templating library forever.

## Rollout

The rollout should follow a small-to-broad path:

1. Add the spec manifest and generator scaffolding.
2. Generate the first REST workflow subset.
3. Add `ghx api explain`.
4. Wire `ghx ci doctor` to the generated workflow calls.
5. Add curated GraphQL operation generation and registry metadata.
6. Wire `ghx pr ready`, then `ghx rules explain`, then `ghx sec inbox`.
7. Add CI drift checks and the scheduled spec update PR lane.
8. Expand coverage only after the first slices have stable tests and readable generated output.

## Risks and tradeoffs

- A generator can overfit to the first few operations if we optimize too early for convenience.
- A broad generated client can become harder to review than handwritten code if the codegen output is too abstract.
- Pinned specs can drift from reality if the update lane is not maintained.
- GraphQL validation can get too clever if the registry tries to infer more than the schema actually says.
- Escape hatches must stay easy to reach or the generated layer will become a new form of lock-in.

The design intentionally prefers small, explicit generated units over a monolithic SDK so these risks stay manageable.

## Acceptance criteria

- Pinned REST and GraphQL inputs are recorded in a manifest.
- The first generated REST subset covers workflow runs, jobs, run attempts, and pending deployments.
- The first GraphQL registry supports at least the initial `ghx pr ready` slice.
- Generated operations validate required params, enums, scalar shapes, and pagination style before dispatch.
- Raw escape hatches are present and documented.
- `ghx api explain <operation-id>` can report method, path, params, response, pagination, and metadata.
- `ghx ci doctor`, `ghx pr ready`, `ghx rules explain`, and `ghx sec inbox` can consume generated proxies through testable adapter layers.
- CI fails when generated output is stale or dirty.
- A scheduled update lane can surface spec drift as a PR.

## Open questions

- Whether the first REST snapshot should vendor the full upstream OpenAPI file or a trimmed subset plus extraction metadata.
- Whether GraphQL curated operations should live beside the registry or in a dedicated operations directory from day one.
- Whether `oapi-codegen` or `ogen` produces the clearest first slice for the workflow operations, given the balance between validation, reviewability, and template control.
- How much permission and scope metadata can be extracted reliably from the pinned specs versus from docs mining.
- Whether `ghx api explain` should be implemented entirely from the generated registry or should also inspect live auth and host context.

## Related work

- [docs/plans/generated-api-proxies.md](../plans/generated-api-proxies.md)
- [docs/ghx-gap-map.md](../ghx-gap-map.md)
- [docs/ghx-api-coverage.md](../ghx-api-coverage.md)
