# ghx Release Snapshot Workflow

Status: implemented for review
Owner issue: [#57](https://github.com/agustif/ghx/issues/57)
Workflow: `.github/workflows/ghx-release-snapshot.yml`

## Goal

Provide a manual CI lane that reproduces the local `ghx` GoReleaser snapshot validation without publishing anything.

This is a staging gate for the fork-owned release path. It is not the production deployment workflow, and it intentionally does not create GitHub Releases, update docs, update Homebrew, build package-manager repositories, sign binaries, emit attestations, or upload artifacts outside GitHub Actions run storage.

## Execution DAG

```text
ghx release snapshot workflow
|- checkout requested ref with full git history
|- install Go using go.mod
|- install GoReleaser v2.13.1
|- check .goreleaser-ghx.yml
|- guard release.disable and package/publisher sections
|- build GoReleaser snapshot with temporary release notes
|- run make smoke-ghx-release
|- verify dist contains ghx_* archives and ghx_*_checksums.txt only
`- upload ghx archives and checksums as workflow artifacts
```

## Manual Run

Run the workflow from GitHub Actions:

```text
Actions -> ghx Release Snapshot -> Run workflow
```

Inputs:

| Input | Required | Meaning |
| --- | --- | --- |
| `ref` | no | Branch, tag, or SHA to validate. Empty means the workflow ref. |

## No-Publish Contract

The workflow keeps the release lane non-publishing by design:

- trigger is `workflow_dispatch` only
- token permissions are `contents: read`
- checkout uses `persist-credentials: false`
- no secrets or environments are referenced
- no `id-token` permission is granted
- `.goreleaser-ghx.yml` must keep `release.disable: true`
- package and publisher sections are rejected before the snapshot build
- only `dist/ghx_*.tar.gz`, `dist/ghx_*.zip`, and `dist/ghx_*_checksums.txt` are uploaded as GitHub Actions artifacts

## Local Parity Checks

The workflow mirrors these local commands:

```bash
goreleaser check -f .goreleaser-ghx.yml
notes="$(mktemp)"
printf 'ghx release snapshot validation.\n' >"$notes"
goreleaser release -f .goreleaser-ghx.yml --snapshot --clean --release-notes="$notes"
make smoke-ghx-release
```

## Remaining Production Work

Production release automation remains blocked until separate release slices decide fork-owned publication, signing, package-manager channels, provenance, updater identity, and operator approval rules.
