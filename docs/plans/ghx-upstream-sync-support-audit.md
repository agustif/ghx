# Plan: ghx upstream sync and support audit

Status: active
Date: 2026-05-20
Owner issues: [#66](https://github.com/agustif/ghx/issues/66), [#71](https://github.com/agustif/ghx/issues/71)
Parent plan: [First-class ghx release migration](ghx-first-class-release-migration.md)

## Goal

Make upstream sync and non-release product identity safe before `ghx` becomes a first-class distribution.

This plan keeps inherited `gh` behavior compatible where possible, but it treats fork release identity, telemetry, support routing, and branch policy as owned `ghx` surfaces. An upstream merge is not complete until the fork-specific deltas are still intentional, documented, and smoke-tested.

## Task graph

```text
ghx governance and support readiness
|- #66 upstream sync policy
|  |- fetch origin and upstream
|  |- compare fork delta against upstream
|  |- route upstream sync through review branches
|  `- refresh shipped divergence docs
|- #71 telemetry support identity audit
|  |- classify telemetry endpoint and payload identity
|  |- classify support and security links
|  |- classify version and update links
|  `- define production release gates
`- release readiness gate
   |- branch protection and release branch policy
   |- source and package smoke matrix
   `- fork-owned publication targets
```

## Snapshot

Current evidence from the #66 and #71 audit pass:

- `origin/trunk`: `f83879f15`
- `upstream/trunk`: `81ed6d4e3`
- merge base: `f1c10e032a`
- `git rev-list --left-right --count upstream/trunk...origin/trunk`: `2 30`
- `agustif/ghx` latest release: none yet, `GET /repos/agustif/ghx/releases/latest` returned `404`
- upstream latest release: `cli/cli` `v2.92.0`, published 2026-04-28, with `gh_*` release assets
- live `agustif/ghx` branch metadata reported `trunk` as unprotected

The release policy must assume there is no proven `ghx` production artifact stream yet.

## Evidence anchors

| Surface | Current evidence | Classification |
| --- | --- | --- |
| Fork compatibility contract | `docs/ghx-vs-gh.md:28-41` says `ghx` is additive and inherits the normal command tree, flags, config files, auth storage, extension model, JSON conventions, and `api` escape hatch unless documented otherwise. | upstream-compatible baseline |
| Shipped fork deltas | `docs/ghx-vs-gh.md:16-26` lists binary identity, scoped account selection, auth status source evidence, helper recognition, subissues, parent issue creation, issue search matching, update links, and fork docs. | fork-owned |
| Upstream release config | `.goreleaser.yml:3-113` still builds `gh`, emits `gh_*` archives, and installs `gh` completions and manpages. | upstream-compatible `gh` lane |
| Fork release scaffold | `.goreleaser-ghx.yml:3-62` sets `project_name: ghx`, targets `agustif/ghx`, builds `bin/ghx`, injects `internal/build.Name=ghx`, and emits `ghx_*` archives. | fork-owned |
| Release trigger | `script/release:76-78` dispatches `deployment.yml` in `cli/cli`. | needs a `ghx` replacement before production |
| Release warning | `docs/releasing.md:5-7` says the page still describes upstream `gh` production release behavior and must not be used for production `ghx` until fork-owned automation is complete. | fork-owned guardrail |
| Release branch preflight | `docs/plans/ghx-release-operator-runbook.md:43-52` requires a clean `origin/trunk` snapshot and exact head checks. | fork-owned release gate |
| Release smoke matrix | `docs/plans/ghx-release-smoke-matrix.md:38-55` separates source smoke from package, signing, provenance, and registry readiness. | fork-owned release gate |
| Source install | `Makefile:124-152` installs and uninstalls `ghx` binary, manpages, and completions under `ghx` names. | fork-owned, side-by-side |
| Update checks | `internal/ghcmd/update_enabled.go:5-18` keeps the upstream default `cli/cli` for updateable `gh`, while `internal/ghcmd/cmd.go:323-355` remaps default `ghx` checks to `agustif/ghx`, uses `state-ghx.yml`, and suppresses the `brew upgrade gh` hint. | fork-owned for `ghx`, upstream-compatible for `gh` |
| Version links | `pkg/cmd/version/version.go:45-60` maps `ghx` changelog URLs to `agustif/ghx` and all other command names to `cli/cli`. | fork-owned |
| Telemetry opt-out | `internal/telemetry/telemetry.go:103-145` honors `GH_TELEMETRY`, `DO_NOT_TRACK`, and config values. `pkg/cmd/root/help_topic.go:120-138` documents the current upstream telemetry help link. | upstream-compatible but needs `ghx` policy |
| Telemetry sender | `pkg/cmd/send-telemetry/send_telemetry.go:21-109` defaults to `https://cafe.github.com`, sends app `github-cli`, and uses user agent `GitHub CLI <version>`. | needs a `ghx` replacement or release disable |
| Telemetry payload shape | `internal/telemetry/telemetry.go:228-331` includes `device_id`, `invocation_id`, OS, architecture, sample rate, command, flags, version, TTY, agent, CI, and accessibility/color dimensions. | needs a `ghx` privacy decision |
| API user agent | `api/http_client.go:54-63` uses `GitHub CLI <version>` for normal GitHub API requests. | upstream-compatible if accepted as additive fork identity |
| Support contact links | `.github/ISSUE_TEMPLATE/config.yml:3-8` points questions at `cli/cli` discussions and the GitHub API community forum. | needs fork-owned routing for `ghx` behavior |
| Bug report template | `.github/ISSUE_TEMPLATE/bug_report.md:14-17` asks reporters to run `gh version`. | needs fork-owned routing for `ghx` behavior |
| Security policy | `.github/SECURITY.md:1-18` names GitHub CLI and links private vulnerability reporting to `cli/cli`. | needs owner decision before production |
| Contributor routing | `.github/CONTRIBUTING.md:78-82` points bug, enhancement, and label links at `cli/cli`. | upstream-compatible for upstream contribution, fork-owned for `ghx` roadmap |

## Upstream Sync Policy

`origin/trunk` is the canonical `ghx` integration branch. `upstream/trunk` is an input, not a release base. Sync work must happen through reviewed fork branches.

Branch policy:

- Use `sync/upstream-YYYYMMDD-<upstream-short-sha>` for upstream import PRs.
- Use `release/ghx/vX.Y.Z` for fork release candidates after the release gates below pass.
- Use issue-scoped branches such as `af/ghx-<topic>` for normal implementation slices.
- Do not create production release tags or publish release assets from upstream sync branches.
- Protect `trunk` before the first production `ghx` release. The current live branch metadata says `trunk` is not protected, so this is a production blocker.

Sync command contract:

```bash
git status --short
git fetch --prune origin trunk
git fetch --prune upstream trunk
git rev-parse --short origin/trunk upstream/trunk
git rev-list --left-right --count upstream/trunk...origin/trunk
git switch -c sync/upstream-YYYYMMDD-<upstream-short-sha> origin/trunk
git merge --no-ff upstream/trunk
```

Use a merge commit for broad upstream imports so the upstream boundary is visible. Rebase an issue branch onto `origin/trunk` before pushing a focused PR, especially after unrelated docs-only planning PRs merge.

Required checks after each upstream merge:

```bash
git diff --name-only upstream/trunk...HEAD
git diff -- docs/ghx-vs-gh.md docs/plans/ghx-first-class-release-migration.md
make smoke-ghx-release
go test ./internal/ghcmd ./pkg/cmd/version ./pkg/cmd/auth/... ./pkg/cmd/issue/...
```

If the upstream merge touches release, install, update, telemetry, support, generated docs, issue commands, auth selection, or extension surfaces, expand the check set to the owning package or workflow. Do not treat a clean merge as a completed sync until the shipped divergence list in `docs/ghx-vs-gh.md` is still current.

Conflict watchlist:

- release scripts and configs: `.goreleaser.yml`, `.goreleaser-ghx.yml`, `script/release`, `.github/workflows/deployment.yml`
- update and version identity: `internal/ghcmd/update_enabled.go`, `internal/ghcmd/cmd.go`, `pkg/cmd/version/version.go`
- telemetry and support identity: `internal/telemetry`, `pkg/cmd/send-telemetry`, `pkg/cmd/root/help_topic.go`, `.github/ISSUE_TEMPLATE`, `.github/SECURITY.md`
- account selection and git helper behavior: `internal/config`, `internal/gh`, `pkg/cmd/auth`, `docs/multiple-accounts.md`
- issue command deltas: `api/queries_issue.go`, `pkg/cmd/issue`, `pkg/cmd/pr/shared/params.go`
- generated docs and source install: `cmd/gen-docs`, `Makefile`, `docs/install_*.md`, `share/`

## Release Branch Policy

Release candidates start from a clean `origin/trunk` snapshot, not from `upstream/trunk`, a local worktree head, or a branch that includes unreviewed sync work.

Preflight:

```bash
git status --short
git fetch --prune origin trunk
git rev-parse --short HEAD
git rev-parse --short origin/trunk
git diff --quiet
```

Stop if the worktree is dirty or `HEAD` is not the intended release base. A release branch may stage artifacts, smoke checks, and docs for a candidate, but it must not publish upstream-owned docs sites, formulas, package repositories, or release assets.

Production release gates:

- `trunk` branch protection is enabled or an explicit release exception is approved.
- `agustif/ghx` has a fork-owned release workflow or an explicit manual release path.
- `script/release` no longer dispatches production `ghx` releases to `cli/cli`.
- `make smoke-ghx-release` passes on the release candidate.
- `goreleaser check -f .goreleaser-ghx.yml` passes where GoReleaser is available.
- Packaged artifacts use `ghx_*` names and install `ghx` paths without overwriting stock `gh`.
- Update checks and version output point at `agustif/ghx`.
- Provenance, checksum, and attestation examples use `agustif/ghx` and `ghx_*` artifacts.
- Rollback leaves either the previous `ghx` release or the stock `gh` install usable.

## Telemetry And Support Audit

Telemetry is the highest-risk #71 surface because current `ghx` builds inherit upstream telemetry defaults while also adding fork-specific command names and behavior.

Classification:

| Surface | Current state | Policy |
| --- | --- | --- |
| GitHub API calls | Normal `gh` and `ghx` commands call GitHub APIs for user-requested work. | upstream-compatible |
| Update checks | Default update repository is remapped to `agustif/ghx` only when the command is built as `ghx`. | fork-owned and implemented |
| Version links | `ghx --version` links to `agustif/ghx` releases. | fork-owned and implemented |
| Command telemetry endpoint | `send-telemetry` defaults to `https://cafe.github.com`. | needs replacement or production disable |
| Telemetry app and user agent | Sender labels events as app `github-cli` and user agent `GitHub CLI <version>`. | needs replacement or production disable |
| GitHub API user agent | Normal API clients use user agent `GitHub CLI <version>`. | upstream-compatible unless GitHub starts product-specific routing from this value |
| Telemetry privacy docs | Help topic links to `https://cli.github.com/telemetry`. | needs `ghx` policy docs if telemetry remains enabled |
| Support questions | Issue contact links route usage questions to `cli/cli` discussions. | needs fork-owned route for `ghx`-specific behavior |
| Bug reports | Template asks for `gh version`. | needs fork-owned `ghx version` prompt |
| Security reports | Security policy points private vulnerability reports to `cli/cli`. | needs owner decision before production |

Production `ghx` must choose one of these telemetry outcomes before release:

1. Disable command telemetry for packaged `ghx` builds by default and document the opt-in path.
2. Keep telemetry enabled only after setting a fork-owned endpoint, app identity, user agent, privacy page, support owner, and data retention policy.

Until that decision is implemented, a production `ghx` release is blocked. A fork binary must not send fork-specific command usage, account-selection behavior, or issue-workflow behavior to an upstream-only telemetry channel by accident.

Telemetry proof commands for the eventual implementation:

```bash
GH_TELEMETRY=log GH_TELEMETRY_SAMPLE_RATE=100 bin/ghx version
GH_TELEMETRY=enabled GH_TELEMETRY_SAMPLE_RATE=100 GH_TELEMETRY_ENDPOINT_URL=http://127.0.0.1:<port> bin/ghx version
```

The first command should expose the payload shape without network send. The second command should be run only against a local capture endpoint and should prove the exact URL, app identity, user agent, and request body that a packaged `ghx` would send.

Minimum static audit before production:

```bash
rg -n "cafe.github.com|github-cli|GitHub CLI %s|github.com/cli/cli/discussions|formula-name: gh|cli.github.com|gh_|brew upgrade gh" .github docs internal pkg script .goreleaser*.yml
```

Support routing gates:

- `.github/ISSUE_TEMPLATE/config.yml` links fork-specific support to `agustif/ghx` discussions or issues.
- `.github/ISSUE_TEMPLATE/bug_report.md` asks for `ghx version` when the report is about fork behavior.
- `.github/SECURITY.md` says where `ghx` vulnerabilities go, and whether upstream GitHub security reporting still owns inherited `gh` vulnerabilities.
- `.github/CONTRIBUTING.md` separates upstream `cli/cli` contribution policy from fork roadmap contribution policy.
- User-facing docs keep upstream package-manager instructions labeled as stock `gh` until fork-owned `ghx` package channels exist.

## Post-merge Checklist

Use this checklist after every upstream sync and before every release branch.

- [ ] `origin/trunk`, `upstream/trunk`, and merge base are recorded.
- [ ] `docs/ghx-vs-gh.md` snapshot and shipped divergence table match the current fork head.
- [ ] Upstream added no overlapping feature that should replace a fork implementation.
- [ ] Release scripts, workflows, GoReleaser configs, update code, install docs, and generated docs were checked if touched.
- [ ] Telemetry endpoint, app identity, user agent, privacy docs, and support links were checked if touched.
- [ ] `make smoke-ghx-release` passed or the blocker is recorded.
- [ ] Release branch, package, signing, provenance, and rollback gates are still blocked or explicitly satisfied.

## Acceptance

This #66/#71 slice is complete when:

- upstream sync commands and branch naming are documented
- release branch naming and production blockers are documented
- the conflict watchlist distinguishes inherited upstream behavior from fork-owned `ghx` behavior
- telemetry, support, security, update, and version surfaces are classified
- production-ready gates block `ghx` release publication until branch protection, telemetry policy, support routing, package identity, updater identity, and smoke coverage are explicit
