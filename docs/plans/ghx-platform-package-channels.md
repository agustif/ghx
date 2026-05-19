# Plan: ghx platform package channels

Status: active
Date: 2026-05-20
Owner issues: [#60](https://github.com/agustif/ghx/issues/60), [#61](https://github.com/agustif/ghx/issues/61), [#62](https://github.com/agustif/ghx/issues/62), [#63](https://github.com/agustif/ghx/issues/63)
Parent plan: [First-class ghx release and automation migration](ghx-first-class-release-migration.md)

## Goal

Define the side-by-side package channel contract for `ghx` without publishing artifacts or changing the upstream-compatible `gh` release lane.

This page is the handoff record for the platform packaging slice. It separates local package metadata that can be validated now from production publication decisions that still require signing, repository ownership, registry ownership, and install or uninstall smoke tests.

## Task graph

```text
ghx platform package channels
|- #60 Linux package metadata
|  |- done: add local nFPM metadata to .goreleaser-ghx.yml
|  |- next: inspect deb and rpm metadata from a snapshot build
|  `- blocked: fork-owned apt and rpm repository publication
|- #61 macOS archive and pkg
|  |- done: keep ghx zip archives separate from upstream gh archives
|  |- next: fork or parameterize pkgmacos for ghx payloads
|  `- blocked: fork-owned installer signing and notarization policy
|- #62 Windows zip and MSI
|  |- done: keep ghx zip archive names separate from upstream gh archives
|  |- next: fork or parameterize WiX metadata for ghx.exe
|  `- blocked: fork-owned signing and MSI upgrade-code policy
`- #63 package-manager handoff
   |- done: keep upstream Homebrew automation disabled for ghx
   |- next: decide fork tap versus unsupported Homebrew channel
   `- blocked: fork-owned package repository and secret ownership
```

## Cross-channel invariant

The default package contract is side-by-side install, not replacement install.

Required behavior:

- package name is `ghx`, not `gh`
- command installed on `PATH` is `ghx`, not `gh`
- package payloads do not write upstream `gh` completions, manpages, registry keys, receipts, or formula names
- package metadata does not declare `conflicts`, `replaces`, or `provides` against `gh` unless a separate shadow-mode decision explicitly accepts that behavior
- uninstalling `ghx` leaves an existing upstream `gh` install usable
- package manager docs keep saying upstream package-manager commands install regular `gh` until a fork-owned `ghx` channel exists

## Current implementation state

`.goreleaser-ghx.yml` is the fork-owned local validation config.

It now owns:

- `project_name: ghx`
- disabled release target metadata for `agustif/ghx`
- `bin/ghx` binaries across macOS, Linux, and Windows
- `ghx_*` archive names
- generated `ghx` manpages and shell completions for local artifact packaging
- local nFPM deb and rpm metadata for package name `ghx`

It does not publish anything. `release.disable: true` stays in place, and the validation commands for this slice use local checks and optional snapshot artifact inspection only.

`.goreleaser.yml` remains the upstream-compatible `gh` release config.

## Linux package channel

Issue: [#60](https://github.com/agustif/ghx/issues/60)

Local metadata decision:

- Use nFPM only in `.goreleaser-ghx.yml`.
- Package name is `ghx`.
- Maintainer and vendor are `agustif`.
- Homepage is `https://github.com/agustif/ghx`.
- Runtime dependency is `git`, matching upstream `gh`.
- Installed command path is `/usr/bin/ghx`.
- Installed manpages and shell completions use only `ghx` paths.
- Do not declare conflicts or replacement metadata for upstream `gh`.

Local validation gates:

```bash
goreleaser check -f .goreleaser-ghx.yml
goreleaser release -f .goreleaser-ghx.yml --snapshot --clean --skip publish,announce --release-notes="$(mktemp)"
dpkg-deb -I dist/ghx_*_linux_*.deb
dpkg-deb -c dist/ghx_*_linux_*.deb
rpm -qip dist/ghx_*_linux_*.rpm
rpm -qlp dist/ghx_*_linux_*.rpm
```

Production gates:

- install the deb on a disposable Debian or Ubuntu host where upstream `gh` is already installed
- install the rpm on a disposable Fedora, RHEL-like, or openSUSE host where upstream `gh` is already installed
- prove `gh version` and `ghx version` both work after install
- prove removing `ghx` does not remove upstream `gh`
- define fork-owned apt and rpm repository paths before documenting package-manager install commands
- define fork-owned package signing keys before publishing repository metadata

## macOS package channel

Issue: [#61](https://github.com/agustif/ghx/issues/61)

Current state:

- `.goreleaser-ghx.yml` produces `ghx_*_macOS_*.zip` archives.
- `script/pkgmacos` is still upstream `gh` packaging. It reads `dist/macos_darwin_*/bin/gh`, writes `/usr/local/bin/gh`, uses package identifier `com.github.cli`, and emits `gh_<version>_macOS_universal.pkg`.

Decision:

- Do not use `script/pkgmacos` for production `ghx` packages until it is forked or parameterized with `ghx` inputs.
- The `ghx` pkg payload must install `/usr/local/bin/ghx`.
- The pkg identifier and receipt identity must be fork-owned and must not reuse `com.github.cli`.
- The pkg must include only `ghx` manpages and `ghx` completions.
- Zip signing, pkg signing, and notarization must be explicit gates, not implicit side effects of the upstream release script.

Local validation gates:

```bash
goreleaser check -f .goreleaser-ghx.yml
goreleaser release -f .goreleaser-ghx.yml --snapshot --clean --skip publish,announce --release-notes="$(mktemp)"
zipinfo dist/ghx_*_macOS_*.zip
```

Production gates:

- build a fork-owned universal pkg from `bin/ghx`
- verify the package payload with `pkgutil --payload-files`
- install on a clean macOS host with upstream `gh` already installed
- prove `gh version` and `ghx version` both work after install
- prove `pkgutil --forget` or uninstall guidance removes only `ghx` receipt-owned files
- sign the zip or embedded binary with the fork-approved Developer ID Application identity
- sign the pkg with the fork-approved Developer ID Installer identity
- notarize and staple production artifacts before telling users they are trusted installers

## Windows package channel

Issue: [#62](https://github.com/agustif/ghx/issues/62)

Current state:

- `.goreleaser-ghx.yml` produces `ghx_*_windows_*.zip` archives.
- `build/windows/gh.wixproj` and `build/windows/gh.wxs` are still upstream `gh` packaging. The WiX product name is `GitHub CLI`, the installed file is `gh.exe`, the install directory is `GitHub CLI`, and the registry key is `SOFTWARE\GitHub\CLI`.
- `script/sign.ps1` still signs with description `GitHub CLI`.

Decision:

- Do not use the current WiX project for production `ghx` MSI output.
- Fork or parameterize the WiX metadata before enabling MSI generation for `ghx`.
- The MSI must install `ghx.exe`, not `gh.exe`.
- Product name, manufacturer, install directory, registry key, UpgradeCode values, and ARP display metadata must be fork-owned.
- The MSI must not upgrade, repair, or uninstall upstream GitHub CLI.
- PATH mutation must add only the `ghx` install directory and must be removed on uninstall without touching upstream `gh`.
- Signing metadata must describe `ghx`, not `GitHub CLI`.

Local validation gates:

```bash
goreleaser check -f .goreleaser-ghx.yml
goreleaser release -f .goreleaser-ghx.yml --snapshot --clean --skip publish,announce --release-notes="$(mktemp)"
powershell -NoProfile -Command "Get-ChildItem dist -Filter 'ghx_*_windows_*.zip'"
```

Production gates:

- build MSI output from fork-owned WiX metadata
- inspect MSI tables for product name, component paths, registry keys, UpgradeCode values, and environment changes
- install on Windows with upstream GitHub CLI already installed
- prove `gh.exe version` and `ghx.exe version` both work after install
- prove uninstall removes only the `ghx` MSI product
- sign `ghx.exe` and the MSI with fork-owned signing material

## Homebrew and package-manager handoff

Issue: [#63](https://github.com/agustif/ghx/issues/63)

Current state:

- `.github/workflows/homebrew-bump.yml` still targets formula `gh`.
- `.github/workflows/deployment.yml` still includes upstream Homebrew publication assumptions for `gh`.
- Install docs correctly say current package-manager instructions install upstream `gh`, not `ghx`.

Decision:

- Keep Homebrew publication disabled for `ghx` until there is a fork-owned tap or an explicit decision to leave Homebrew unsupported.
- A future formula must be named `ghx` and install `bin/ghx`.
- Do not publish a `gh` alias, symlink, cask, or replacement formula in the default side-by-side channel.
- Do not mutate `homebrew/core` automation from this fork unless a future issue explicitly decides to hand off to Homebrew core.
- Package-manager install docs should change only after the fork-owned channel has passed local audit and install smoke tests.

Future tap gates:

```bash
brew audit --strict --online ghx
brew install --build-from-source ./Formula/ghx.rb
brew test ghx
brew uninstall ghx
```

Production gates:

- decide tap owner and repository
- decide token or GitHub App ownership for formula bumps
- decide prerelease exclusion rules
- prove formula installation does not conflict with `brew install gh`
- document upgrade and uninstall commands only after the tap is live

## No-publish rule for this slice

This slice may:

- validate GoReleaser config
- run local snapshot builds
- inspect local archives and package metadata
- update docs and package metadata in the repository

This slice must not:

- create releases
- upload artifacts
- push package repository metadata
- open Homebrew formula PRs
- sign with production material
- notarize production artifacts

## Acceptance checklist

- `.goreleaser-ghx.yml` validates with local `goreleaser check`
- any nFPM metadata uses `ghx` package, binary, manpage, and completion names
- platform docs name macOS pkg, Windows MSI, signing, notarization, and Homebrew gates explicitly
- side-by-side install and upstream `gh` non-interference stay visible
- package-manager docs keep pointing users to source install until package channels are fork-owned

## Evidence anchors

- [.goreleaser-ghx.yml](../../.goreleaser-ghx.yml)
- [.goreleaser.yml](../../.goreleaser.yml)
- [script/pkgmacos](../../script/pkgmacos)
- [build/windows/gh.wixproj](../../build/windows/gh.wixproj)
- [build/windows/gh.wxs](../../build/windows/gh.wxs)
- [.github/workflows/homebrew-bump.yml](../../.github/workflows/homebrew-bump.yml)
- [ghx vs gh](../ghx-vs-gh.md)
- [source install docs](../install_source.md)
