# ghx production release readiness

Status: active
Date: 2026-05-20
Owner issues: [#55](https://github.com/agustif/ghx/issues/55), [#60](https://github.com/agustif/ghx/issues/60), [#61](https://github.com/agustif/ghx/issues/61), [#62](https://github.com/agustif/ghx/issues/62), [#63](https://github.com/agustif/ghx/issues/63), [#65](https://github.com/agustif/ghx/issues/65), [#69](https://github.com/agustif/ghx/issues/69)
Related docs: [First-class release migration](ghx-first-class-release-migration.md), [Platform package channels](ghx-platform-package-channels.md), [Release smoke matrix](ghx-release-smoke-matrix.md), [Release operator runbook](ghx-release-operator-runbook.md), [Release provenance runbook](ghx-release-provenance-runbook.md)

## Goal

Give release operators a single secret-free checklist that separates what can be
validated in this repository today from what still requires fork-owned signing,
package repository, notarization, or release credentials.

This page is not permission to publish. It is the readiness gate before a
production release issue can be marked ready.

## Task graph

```text
production release readiness
|- no-publish config gate
|  |- done: validate .goreleaser-ghx.yml identity
|  `- done: reject publisher and signing sections in snapshot config
|- artifact boundary gate
|  |- done: inspect ghx archives and checksum names
|  |- done: inspect deb and rpm payload metadata when tools are present
|  `- next: install/uninstall package artifacts on disposable hosts
|- package channel gate
|  |- done: add apt, rpm, macOS pkg, Windows MSI, and Homebrew templates
|  `- blocked: fork-owned repository, signing, notarization, and tap ownership
|- workflow gate
|  |- done: add no-publish ghx release readiness workflow
|  `- blocked: fork-owned production deployment workflow
`- provenance gate
   |- done: document agustif/ghx verification identity
   `- blocked: production attestations for published release assets
```

## Secret-free local gate

Run these commands from the repository root:

```bash
packaging/ghx/scripts/check-release-readiness
notes="$(mktemp)"
printf 'ghx release readiness validation.\n' >"$notes"
goreleaser release -f .goreleaser-ghx.yml --snapshot --clean --release-notes="$notes"
make smoke-ghx-release
packaging/ghx/scripts/smoke-release-artifacts
```

For CI or release-operator checks where package inspection tools are expected:

```bash
GHX_SMOKE_REQUIRE_PACKAGE_TOOLS=1 packaging/ghx/scripts/smoke-release-artifacts
```

To prove that production is not ready while known blockers remain:

```bash
packaging/ghx/scripts/check-release-readiness --strict-production
```

The strict command should fail until the production blockers below are closed.

## Manual workflow gate

Run the no-publish workflow before staging or production release work:

```text
Actions -> ghx Release Readiness -> Run workflow
```

The workflow:

- checks out the requested ref without persisted credentials
- installs Go and GoReleaser
- validates `.goreleaser-ghx.yml`
- builds a local snapshot with release publication disabled
- runs the source smoke and packaged artifact smoke checks
- uploads only `dist/ghx_*` artifacts to the workflow run

The workflow must not request environments, signing secrets, id-token
permissions, or write permissions.

## Issue checklist

| Issue | Readiness state | Remaining blocker |
| --- | --- | --- |
| [#55](https://github.com/agustif/ghx/issues/55) | Release readiness, package templates, and smoke workflow are represented. | Epic remains open until all child release slices are complete. |
| [#60](https://github.com/agustif/ghx/issues/60) | `.goreleaser-ghx.yml` owns local deb/rpm metadata and artifact smoke inspects package payloads. | apt/rpm repositories, signing keys, and install/uninstall host proof. |
| [#61](https://github.com/agustif/ghx/issues/61) | macOS archive names are validated and pkg identity inputs are templated. | fork-owned pkg implementation, Developer ID signing, notarization, and uninstall proof. |
| [#62](https://github.com/agustif/ghx/issues/62) | Windows zip names are validated and WiX identity inputs are templated. | fork-owned MSI implementation, stable UpgradeCode values, signing, and install/uninstall proof. |
| [#63](https://github.com/agustif/ghx/issues/63) | Homebrew formula shape is drafted as a non-publishing template. | tap owner, formula publication policy, token ownership, and smoke proof. |
| [#65](https://github.com/agustif/ghx/issues/65) | Verification docs name `agustif/ghx`, `ghx_*` checksums, and staging versus production proof. | production release artifacts with attestations or recorded Sigstore bundles. |
| [#69](https://github.com/agustif/ghx/issues/69) | Source smoke and artifact inspection smoke are runnable without credentials. | real package install/uninstall matrix across Linux, macOS, and Windows. |

## Package channel templates

Template files live under [packaging/ghx](../../packaging/ghx/README.md).
They are not package-manager source of truth yet. They are the minimum records
that must be filled before each channel can become user-facing documentation or
publishing automation.

## Production blockers

The following blockers are expected until dedicated release work removes them:

- `.goreleaser-ghx.yml` keeps `release.disable: true`, so it cannot publish production releases.
- `.github/workflows/deployment.yml` is still upstream-shaped and builds `gh_*` production artifacts.
- `script/pkgmacos` still installs `/usr/local/bin/gh` with `com.github.cli` receipts.
- `build/windows/gh.wxs` still installs `gh.exe` with upstream product and registry identity.
- `.github/workflows/homebrew-bump.yml` still targets formula `gh`.
- no fork-owned signing, notarization, package repository, tap, or attestation credentials are available in this secret-free lane.

## Ready-to-publish rule

A release is not production-ready until the no-publish workflow passes and every
production blocker above is either removed by a fork-owned implementation or
explicitly accepted in a release issue as unsupported for that release.
