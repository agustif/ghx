# Plan: First-class ghx release and automation migration

Status: active
Date: 2026-05-19
Roadmap epic: [#55](https://github.com/agustif/ghx/issues/55)
Parent epic: [#2](https://github.com/agustif/ghx/issues/2)
Expanded slice: [#5](https://github.com/agustif/ghx/issues/5)
Related research: [0004](../research/0004-ghx-release-upstream-and-distribution.md)
Related audit: [ghx workflow automation audit](ghx-workflow-automation-audit.md)
Shipped behavior contract: [ghx vs gh](../ghx-vs-gh.md)

## Goal

Map every upstream `gh` release, packaging, update, automation, and docs assumption that must be replaced, forked, disabled, or explicitly preserved before `ghx` becomes a first-class side-by-side distribution.

The plan keeps `ghx` compatible with upstream `gh` where possible, but it treats release identity as a product boundary. Anything that publishes artifacts, points users at an update channel, installs binaries, modifies git credential behavior, signs packages, writes external docs, or verifies provenance must be audited before a `ghx` release is called production-ready.

## Issue tree

| Issue | Ownership |
| --- | --- |
| [#55](https://github.com/agustif/ghx/issues/55) | Epic: first-class `ghx` release and automation migration. |
| [#56](https://github.com/agustif/ghx/issues/56) | GoReleaser identity and artifact names. |
| [#57](https://github.com/agustif/ghx/issues/57) | GitHub Actions deployment workflow. |
| [#58](https://github.com/agustif/ghx/issues/58) | Update notifier and version links. |
| [#59](https://github.com/agustif/ghx/issues/59) | Source install, completions, manpages, and uninstall. |
| [#60](https://github.com/agustif/ghx/issues/60) | Linux package metadata and repository path. |
| [#61](https://github.com/agustif/ghx/issues/61) | macOS archives, pkg, signing, and notarization. |
| [#62](https://github.com/agustif/ghx/issues/62) | Windows zip, MSI, WiX, and signing. |
| [#63](https://github.com/agustif/ghx/issues/63) | Homebrew tap and package-manager handoff. |
| [#64](https://github.com/agustif/ghx/issues/64) | Manual site, install docs, and generated reference. |
| [#65](https://github.com/agustif/ghx/issues/65) | Release attestations, provenance, and verification docs. |
| [#66](https://github.com/agustif/ghx/issues/66) | Upstream sync and release branch policy. |
| [#67](https://github.com/agustif/ghx/issues/67) | Side-by-side migration and rollback runbook. |
| [#68](https://github.com/agustif/ghx/issues/68) | Credential helper install and git-operation parity. |
| [#69](https://github.com/agustif/ghx/issues/69) | Release smoke test matrix. |
| [#70](https://github.com/agustif/ghx/issues/70) | Release operator runbook. |
| [#71](https://github.com/agustif/ghx/issues/71) | Telemetry, support, and product identity audit. |
| [#72](https://github.com/agustif/ghx/issues/72) | Repository automation and CI workflow audit. |

## Migration DAG

```text
first-class ghx release
|- foundation
|  |- #66 upstream sync and release branch policy
|  |- #72 repository automation and CI workflow audit
|  `- #71 telemetry support and product identity audit
|- build identity
|  |- #56 GoReleaser identity and artifact names
|  |- #59 source install completions manpages and uninstall
|  `- #58 update notifier and version links
|- platform packages
|  |- #60 Linux package metadata and repository path
|  |- #61 macOS archives pkg signing and notarization
|  `- #62 Windows zip MSI WiX and signing
|- publication
|  |- #57 GitHub Actions deployment workflow
|  |- #63 Homebrew tap and package-manager handoff
|  |- #64 manual site docs and generated reference
|  `- #65 release attestations provenance and verification docs
|- user migration
|  |- #67 side-by-side migration and rollback runbook
|  `- #68 credential helper install and git-operation parity
`- release gate
   |- #69 release smoke test matrix
   `- #70 release operator runbook
```

## Surface map

### Release config and artifact identity

Owned by [#56](https://github.com/agustif/ghx/issues/56).

Primary files:

- `.goreleaser.yml`
- `script/release`
- `Makefile`
- generated `dist/` layout

Known upstream assumptions:

- `project_name: gh`
- release name template says `GitHub CLI`
- build outputs use `binary: bin/gh`
- archives use `gh_{{ .Version }}_*`
- nfpm metadata names upstream GitHub CLI and installs `gh` completions and manpages

Migration decision:

- Either fork a dedicated `ghx` release config or parameterize the current config with an explicit release product name.
- Keep any upstream-compatible `gh` build lane separate from the fork release lane.

Issue #56 scaffold:

- `.goreleaser-ghx.yml` is the fork-specific GoReleaser scaffold for local validation.
- `.goreleaser.yml` remains the upstream-compatible `gh` release config.
- The scaffold sets `project_name: ghx`, points the disabled SCM release target at `agustif/ghx`, builds `bin/ghx`, injects `internal/build.Name=ghx`, and emits `ghx_{{ .Version }}_*` archive names.
- The scaffold intentionally does not define nFPM packages, macOS pkg output, Windows MSI output, signing hooks, generated completions, or generated manpages.
- The local dry-run command is:

```bash
goreleaser release -f .goreleaser-ghx.yml --snapshot --clean --skip publish,announce --release-notes="$(mktemp)"
```

Acceptance for #56:

- `goreleaser check -f .goreleaser-ghx.yml` passes.
- The dry-run command above produces only fork-owned `ghx_*` archives and `bin/ghx` binaries under `dist/`.
- No command in this issue publishes a release or uploads artifacts.
- Production publication stays blocked until #57, #58, #60, #61, #62, #65, #69, and #70 decide workflow publication, updater identity, packages, signing, provenance, smoke tests, and the operator runbook.

### GitHub Actions deployment automation

Owned by [#57](https://github.com/agustif/ghx/issues/57).

Primary files:

- `.github/workflows/deployment.yml`
- release artifact upload and download steps
- release job site update steps
- platform signing steps

Known upstream assumptions:

- workflow labels and checkout steps reference `cli/cli`
- artifact globs collect `gh_*`
- release asset preparation moves `gh_*`
- manual publication checks out `github/cli.github.com`
- production signing and package publication secrets are upstream-team shaped

Migration decision:

- Create a fork-owned staging release path before production release publication.
- Any workflow that mutates external repos must use fork-owned credentials and dry-run gates.
- Current classification: `needs a ghx replacement`. See the [workflow automation audit](ghx-workflow-automation-audit.md#workflow-classifications) for exact `.github/workflows/deployment.yml` blockers.

### Runtime update and version identity

Owned by [#58](https://github.com/agustif/ghx/issues/58).

Primary files:

- `internal/ghcmd/update_enabled.go`
- `internal/ghcmd/cmd.go`
- `internal/update/update.go`
- `pkg/cmd/version/version.go`

Known upstream assumptions:

- updateable builds check `cli/cli`
- update messaging can suggest `brew upgrade gh`
- changelog links point at upstream release pages

Migration decision:

- `ghx` builds must point at fork release metadata or disable update checks explicitly.
- A binary built as upstream `gh` must not be silently repointed unless the fork deliberately retires upstream-compatible builds.

Implementation status:

- Runtime `ghx` update checks remap the upstream default `cli/cli` release repository to `agustif/ghx`.
- Runtime `ghx` update checks use a fork-specific update state file.
- Runtime `ghx` version links point at `agustif/ghx` release pages.
- Runtime `ghx` update messages use the `ghx` command name and do not print the upstream `brew upgrade gh` hint.
- Upstream-compatible `gh` builds keep the upstream release repository, changelog links, and Homebrew hint.

### Install surfaces, completions, and manpages

Owned by [#59](https://github.com/agustif/ghx/issues/59).

Primary files:

- `Makefile`
- `cmd/gen-docs`
- `docs/install_source.md`
- generated `share/` contents

Known upstream assumptions:

- `make completions` builds bash, fish, and zsh files for `gh`
- `make manpages` generates `gh*.1`
- `make completions-ghx` builds bash, fish, and zsh files for `ghx`
- `make manpages-ghx` generates `ghx*.1`
- `install-ghx` installs the `ghx` binary, completions, and manpages
- `uninstall-ghx` removes only `ghx`-owned source install files

Migration decision:

- `install-ghx` should install only fork-owned paths.
- generated `ghx` docs and shell integration should not replace upstream `gh` files unless the user selects shadow mode.

### Platform packages

Owned by [#60](https://github.com/agustif/ghx/issues/60), [#61](https://github.com/agustif/ghx/issues/61), and [#62](https://github.com/agustif/ghx/issues/62).

Primary files:

- `.goreleaser.yml`
- `script/pkgmacos`
- `build/windows/gh.wixproj`
- `.github/workflows/deployment.yml`
- `docs/install_linux.md`
- `docs/install_macos.md`
- `docs/install_windows.md`

Known upstream assumptions:

- deb and rpm packages describe upstream `gh`
- macOS archives and pkg names are `gh_*`
- Windows MSI build loops over `dist/gh_*_windows_*.zip`
- WiX project identity is upstream `gh`

Migration decision:

- Each platform must prove side-by-side installation with upstream `gh`.
- Upgrade and uninstall behavior is a release gate, not a packaging afterthought.

### Distribution channels and docs

Owned by [#63](https://github.com/agustif/ghx/issues/63) and [#64](https://github.com/agustif/ghx/issues/64).

Primary files:

- `.github/workflows/homebrew-bump.yml`
- `docs/install_*.md`
- `README.md`
- `docs/ghx.md`
- `docs/ghx-vs-gh.md`
- generated manual output

Known upstream assumptions:

- Homebrew bump automation targets formula `gh`
- install docs point at package-manager instructions for upstream `gh`
- generated site output targets `manual/gh*.md`
- official release docs assume upstream GitHub CLI ownership

Migration decision:

- Until a channel is fork-owned, docs must say it installs upstream `gh`, not `ghx`.
- `ghx` manual publication must not write to upstream-owned docs sites by default.

### Provenance, signing, and verification

Owned by [#65](https://github.com/agustif/ghx/issues/65).

Primary files:

- `.github/workflows/deployment.yml`
- `README.md`
- release docs and verification examples
- signing scripts

Known upstream assumptions:

- attestation examples reference `cli/cli`
- verification examples use `gh_*` artifact names
- production signing identities are upstream-team shaped

Migration decision:

- Users need verification examples for `ghx_*` artifacts and fork repo identity.
- Staging and production provenance differences must be explicit.

### Upstream sync and repo automation

Owned by [#66](https://github.com/agustif/ghx/issues/66) and [#72](https://github.com/agustif/ghx/issues/72).

Primary files:

- `.github/workflows/*.yml`
- `.github/actions/*`
- release and docs scripts
- `docs/ghx-vs-gh.md`

Known upstream assumptions:

- upstream drift can change docs, release scripts, or command behavior without updating fork docs
- repo automation may mutate upstream-owned systems or use upstream naming

Migration decision:

- Every workflow is classified as unchanged upstream behavior, fork-compatible, fork-disabled, or needing a `ghx` replacement.
- Divergence docs are required when shipped behavior changes relative to upstream `gh`.
- Current classification source: [ghx workflow automation audit](ghx-workflow-automation-audit.md). The high-risk automation is release publication, Homebrew publication, Go bump PR creation, spam detection, and discussion routing.

### Side-by-side account and git behavior

Owned by [#67](https://github.com/agustif/ghx/issues/67) and [#68](https://github.com/agustif/ghx/issues/68).

Primary files:

- `docs/multiple-accounts.md`
- `docs/ghx-vs-gh.md`
- `pkg/cmd/auth/shared/git_credential.go`
- `pkg/cmd/auth/shared/gitcredentials/helper_config.go`
- install docs

Known upstream assumptions:

- `ghx` commands can use scoped account selection while `git push` still invokes the helper configured in git config
- shadow mode can make `gh` resolve to `ghx`
- shared auth storage can surprise users if docs imply isolated account state

Migration decision:

- Side-by-side mode stays the safer default.
- Any automated helper installation must explain what it changes and how to roll back.

### Release gates and operator runbook

Owned by [#69](https://github.com/agustif/ghx/issues/69) and [#70](https://github.com/agustif/ghx/issues/70).

Primary files:

- [ghx release operator runbook](ghx-release-operator-runbook.md)
- [ghx side-by-side migration and rollback runbook](ghx-side-by-side-migration-runbook.md)
- `script/smoke-ghx-release`
- `docs/plans/ghx-release-smoke-matrix.md`
- `docs/releasing.md`
- `docs/release-process-deep-dive.md`
- future fork release checklist

Known upstream assumptions:

- release docs assume upstream team infrastructure
- the current release process is not a fork-owned checklist
- there is no end-to-end side-by-side smoke matrix for `ghx`

Migration decision:

- A release is not production-ready until the smoke matrix passes on source install and packaged artifacts.
- The operator runbook must separate staging release, production release, verification, package publication, and rollback.

Runbook links:

- [#67 side-by-side migration and rollback runbook](https://github.com/agustif/ghx/issues/67)
- [#69 release smoke test matrix](https://github.com/agustif/ghx/issues/69)
- [#70 release operator runbook](https://github.com/agustif/ghx/issues/70)
- Current release docs: [docs/releasing.md](../releasing.md), [docs/release-process-deep-dive.md](../release-process-deep-dive.md)

## Acceptance

- [ ] #55 remains a child of #2.
- [ ] #56 through #72 remain children of #55 or are replaced by narrower children with the same coverage.
- [ ] Every release-facing PR links the owning issue slice.
- [ ] `docs/ghx-vs-gh.md` is updated whenever a migration changes shipped behavior relative to upstream `gh`.
- [ ] No release automation publishes to upstream-owned repos, package registries, docs sites, or formulas without an explicit fork decision.
