# ghx Packaging Readiness

This tree contains fork-owned release productionization artifacts for `ghx`.
It is intentionally separate from the inherited upstream `gh` packaging paths
until each package channel has a side-by-side install contract, signing policy,
and smoke evidence.

## No-publish rule

Files under this tree may be used to:

- validate `.goreleaser-ghx.yml`
- inspect local snapshot artifacts
- draft package-manager metadata
- document signing and repository ownership
- run secret-free CI checks

They must not publish GitHub Releases, package repositories, Homebrew formulae,
Windows installers, macOS packages, attestations, or signed artifacts.

## Scripts

| Script | Purpose |
| --- | --- |
| `scripts/check-release-readiness` | Validates the secret-free release config and reports production blockers. |
| `scripts/smoke-release-artifacts` | Inspects `dist/ghx_*` archives, checksums, deb packages, and rpm packages without installing them. |

## Templates

| Template | Owner issue | Purpose |
| --- | --- | --- |
| `linux/apt-repository.template.md` | #60 | Records the apt repository, signing key, and publication contract before docs advertise apt install commands. |
| `linux/rpm-repository.template.md` | #60 | Records the rpm repository, signing key, and publication contract before docs advertise rpm install commands. |
| `macos/pkg-identity.env.example` | #61 | Names the non-secret macOS pkg identity inputs that must replace upstream `com.github.cli`. |
| `windows/msi-identity.wxi.template` | #62 | Names the fork-owned WiX identity values that must replace upstream GitHub CLI metadata. |
| `homebrew/Formula/ghx.rb.template` | #63 | Drafts the Homebrew formula shape without publishing to any tap. |

## Integration points

- `.github/workflows/ghx-release-readiness.yml` runs these scripts in a
  no-publish workflow.
- `docs/plans/ghx-production-release-readiness.md` is the checklist that maps
  the scripts and templates back to the open release issues.
- `docs/plans/ghx-platform-package-channels.md` remains the package-channel
  decision record.
