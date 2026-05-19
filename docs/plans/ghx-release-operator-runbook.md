# ghx release operator runbook

Status: active
Date: 2026-05-19
Issue: [#70](https://github.com/agustif/ghx/issues/70)
Related docs: [Releasing](../releasing.md), [Release process deep dive](../release-process-deep-dive.md), [ghx vs gh](../ghx-vs-gh.md), [First-class release migration](ghx-first-class-release-migration.md)

## Goal

Provide the operator path for staging, verifying, publishing, and rolling back a `ghx` release without depending on upstream GitHub CLI team infrastructure unless the release plan explicitly says that dependency is unavailable.

This runbook is a fork-specific companion to the existing release docs merged in PR #73. Those docs still describe the upstream-shaped `gh` pipeline, so this page records the `ghx` operator sequence and the fork-owned decision points.

## Task graph

```text
release operator flow
|- preflight
|  |- clean origin/trunk snapshot
|  |- confirm release inputs and version tag
|  `- confirm branch ownership
|- staging release
|  |- run script/release with --staging
|  |- capture run id
|  `- download staged artifacts
|- verification
|  |- source install check
|  |- package install checks
|  |- auth status and helper checks
|  `- smoke matrix link
|- publication
|  |- create the fork release
|  |- publish fork-owned packages or explicitly skip upstream-owned surfaces
|  `- verify release notes and checksums
`- rollback
   |- stop before publish or revert after publish
   |- remove bad release assets
   `- restore previous release pointer
```

## Preflight

Start only from a clean `origin/trunk` snapshot.

```sh
git status --short
git fetch origin
git rev-parse --short HEAD
git rev-parse --short origin/trunk
```

Do not proceed if the worktree is dirty or if the branch tip is not the expected release base.

## Staging release

The current release docs already expose the staging entrypoint. Use it before any production publish.

```sh
script/release --staging vX.Y.Z --branch patch-1 -p macos
```

If the staging lane needs to cover more than one platform, repeat the same entrypoint for each platform instead of jumping directly to production.

Capture the workflow or run identifier before moving on.

```sh
gh run list --limit 5
```

## Production release

Only move to production after staging and verification have passed.

```sh
script/release vX.Y.Z
```

The production command is the same release entrypoint documented in [Releasing](../releasing.md). The fork-specific difference is not the command name. It is the release identity, artifact naming, and publication target.

## Distribution matrix

Issue [#69](https://github.com/agustif/ghx/issues/69) owns the release smoke matrix. This operator checklist uses the same platform and install surface split so the release and smoke lanes stay aligned.

| Surface | Build or install command | Operator check |
| --- | --- | --- |
| Source install | `make bin/ghx` then `make install-ghx prefix=$HOME/.local` | `ghx version` and `ghx auth status` |
| Linux package | Release artifact produced by the deployment workflow | `ghx version` after package install |
| macOS archive or pkg | Release artifact produced by the deployment workflow | `ghx version` after install |
| Windows zip or MSI | Release artifact produced by the deployment workflow | `ghx.exe version` after install |
| Git helper | `ghx auth setup-git` | `git config --get credential.helper` |
| Side-by-side install | `gh` plus `ghx` on the same machine | both `gh version` and `ghx version` succeed |

## Verification

The operator gate is not complete until the fork binary, helper path, and install surface all match the intended release name.

Suggested checks:

```sh
make bin/ghx
make install-ghx prefix=$HOME/.local
$HOME/.local/bin/ghx version
$HOME/.local/bin/ghx auth status
ghx auth setup-git
git config --get credential.helper
```

When a package artifact is available, install it and repeat the binary-name check on the installed path.

Use the smoke matrix from issue [#69](https://github.com/agustif/ghx/issues/69) for the full platform sweep. This runbook only records the operator-facing sequence and the stop points.

## Publication

Publish only after staging and verification succeed.

Do not publish to upstream-owned repositories, package registries, or docs sites unless the owning issue or release plan explicitly says that target is still upstream-hosted for this slice.

Record the exact release tag, the run that produced the artifacts, and the package surfaces that were updated.

## Rollback

Split rollback by timing:

- before publication: cancel the staging run and fix the input or build issue
- after publication but before broad adoption: delete the bad release assets and retag only if the release plan allows it
- after package publication: revert the package pointer or formula change from the fork-owned lane

Rollback should always leave the previous `ghx` release or the stock `gh` install usable on the same machine.

## Acceptance checklist

- staging release, production release, artifact verification, package publication, and rollback are separated
- the checklist points at issue [#69](https://github.com/agustif/ghx/issues/69) for the smoke matrix
- the checklist points at this page for operator steps and at the distribution matrix above for install surfaces
- no step assumes upstream GitHub CLI team infrastructure without saying so

## Evidence anchors

- [docs/releasing.md](../releasing.md)
- [docs/release-process-deep-dive.md](../release-process-deep-dive.md)
- [docs/ghx-vs-gh.md](../ghx-vs-gh.md)
- [docs/plans/ghx-first-class-release-migration.md](ghx-first-class-release-migration.md)
- [issue #70](https://github.com/agustif/ghx/issues/70)
- [issue #69](https://github.com/agustif/ghx/issues/69)
