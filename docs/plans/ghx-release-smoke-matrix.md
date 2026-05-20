# Plan: ghx release smoke matrix

Status: active
Date: 2026-05-20
Owner issue: [#69](https://github.com/agustif/ghx/issues/69)
Parent plan: [First-class ghx release and automation migration](ghx-first-class-release-migration.md)
Readiness gate: [ghx production release readiness](ghx-production-release-readiness.md)

## Goal

Keep a fast, secret-free release smoke check in the repo while documenting the fuller packaged artifact matrix that must pass before a production `ghx` release.

The local check proves that a source-built `bin/ghx` carries `ghx` command identity through build metadata, version output, and root help. It does not claim package-manager, signing, notarization, provenance, or registry readiness.

## Local source smoke command

Run:

```bash
make smoke-ghx-release
```

The target delegates to `script/smoke-ghx-release`.

Current local assertions:

- force a fresh source build of `bin/ghx`
- set deterministic smoke version metadata through `GH_VERSION`
- avoid user auth state by using a temporary `GH_CONFIG_DIR`
- assert `ghx version ...` instead of upstream `gh version ...`
- assert root help uses `ghx <command> <subcommand> [flags]`
- assert root help describes `Show ghx version`

Environment overrides:

- `GHX_SMOKE_VERSION`: smoke version, default `v0.0.0-smoke`
- `GHX_SMOKE_SOURCE_DATE_EPOCH`: reproducible build date input, default `1704067200`

## Local artifact smoke command

After a `.goreleaser-ghx.yml` snapshot build, run:

```bash
packaging/ghx/scripts/smoke-release-artifacts
```

The artifact smoke is secret-free. It inspects the local `dist/` directory and
does not install packages, sign artifacts, notarize anything, or publish
repository metadata.

Current local assertions:

- fail if any `dist/gh_*` upstream-shaped artifacts appear
- require `ghx_*` archives and `ghx_*_checksums.txt`
- verify Linux archives contain `bin/ghx`, not `bin/gh`
- verify macOS archives contain `bin/ghx`, not `bin/gh`
- verify Windows archives contain `bin/ghx.exe`, not `bin/gh.exe`
- verify checksum contents list `ghx_*` assets, not `gh_*` assets
- when `dpkg-deb` is present, verify deb package name and payload paths
- when `rpm` is present, verify rpm package name and payload paths

CI can require the package inspection tools:

```bash
GHX_SMOKE_REQUIRE_PACKAGE_TOOLS=1 packaging/ghx/scripts/smoke-release-artifacts
```

## Matrix

| Lane | Scope | Command or evidence | Secrets | Status |
| --- | --- | --- | --- | --- |
| Source build identity | local `bin/ghx` built from checkout | `make smoke-ghx-release` | none | implemented |
| Side-by-side source install | `install-ghx` into a temporary prefix while upstream `gh` remains untouched | future script lane | none | planned |
| Completions and manpages | fork-owned shell completion and manpage paths | future #59 lane | none | planned |
| Linux archives | `ghx_*_linux_*` archive names, binary name, license payload, checksum | `packaging/ghx/scripts/smoke-release-artifacts` | none for local artifact inspection | implemented for artifact inspection |
| Debian and RPM packages | package metadata, installed files, conflicts, uninstall behavior | `packaging/ghx/scripts/smoke-release-artifacts` plus future host install smoke | local package build only | implemented for package payload inspection |
| macOS archives and pkg | archive names, binary identity, pkg receipt, side-by-side install | `packaging/ghx/scripts/smoke-release-artifacts` plus future pkg lane | signing and notarization only for production | archive inspection implemented, pkg planned |
| Windows zip and MSI | archive names, binary identity, install path, uninstall behavior | `packaging/ghx/scripts/smoke-release-artifacts` plus future MSI lane | signing only for production | zip inspection implemented, MSI planned |
| Provenance and checksums | release checksum names, attestation repo identity, verification docs | future #65 lane | production release secrets only | planned |
| Update channel | update notifier points at fork release metadata or is disabled | future #58 lane | none for dry-run metadata checks | planned |
| Operator runbook | staging release, production release, rollback, and verification steps | future #70 lane | production-only | planned |

## Release gate rule

A `ghx` release can use the local source smoke as an early gate, but production readiness requires the packaged artifact lanes above to pass against artifacts produced by the fork release path. Any lane that depends on signing, notarization, registry publication, or GitHub release secrets must have a dry-run or inspection mode that remains runnable without those secrets.

## Follow-ups

- Expand `script/smoke-ghx-release` with a temporary-prefix `install-ghx` lane after #59 defines completion and manpage ownership.
- Add live install/uninstall package smoke lanes for disposable Debian/Ubuntu, Fedora/RHEL-like, macOS, and Windows hosts.
- Link this matrix from the release operator runbook in #70 when that runbook exists.
