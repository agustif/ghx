# ghx release provenance and verification runbook

Status: active
Date: 2026-05-19
Issue: [#65](https://github.com/agustif/ghx/issues/65)
Parent plan: [First-class ghx release and automation migration](ghx-first-class-release-migration.md)
Related docs: [Releasing](../releasing.md), [Release process deep dive](../release-process-deep-dive.md), [Release operator runbook](ghx-release-operator-runbook.md), [Release smoke matrix](ghx-release-smoke-matrix.md)

## Goal

Give `ghx` users and release operators a fork-owned way to verify release artifacts before they install them.

The trusted release identity for `ghx` artifacts is:

- release repository: `agustif/ghx`
- release workflow: `.github/workflows/deployment.yml` or its fork-owned replacement
- production release branch or ref: `trunk` unless a release plan records a different approved ref
- archive prefix: `ghx_`
- checksum file: `ghx_<version>_checksums.txt`

The upstream `cli/cli` examples in the inherited README and release docs are not sufficient for `ghx` artifacts. They prove upstream `gh` provenance for `gh_*` files. They do not prove that a downloaded `ghx_*` file came from `agustif/ghx`.

## Artifact names

Fork-owned archive names must use the `ghx_` prefix and the version without the leading `v` in the filename:

```text
ghx_<version>_linux_386.tar.gz
ghx_<version>_linux_amd64.tar.gz
ghx_<version>_linux_arm64.tar.gz
ghx_<version>_macOS_amd64.zip
ghx_<version>_macOS_arm64.zip
ghx_<version>_windows_386.zip
ghx_<version>_windows_amd64.zip
ghx_<version>_windows_arm64.zip
ghx_<version>_checksums.txt
```

If an artifact or checksum file starts with `gh_`, treat it as an upstream-shaped `gh` artifact until a release plan proves otherwise. Do not rename `gh_*` artifacts after the build and call them `ghx_*`; provenance must match the build that produced the file.

## User verification

Use these checks against downloaded release assets from `https://github.com/agustif/ghx/releases`.

Set the version and artifact you downloaded:

```sh
version=X.Y.Z
artifact="ghx_${version}_macOS_arm64.zip"
checksums="ghx_${version}_checksums.txt"
repo="agustif/ghx"
```

Confirm the files are fork artifacts before verifying:

```sh
case "$artifact" in
  ghx_* ) ;;
  * )
    printf 'refusing non-ghx artifact: %s\n' "$artifact" >&2
    exit 1
    ;;
esac

case "$checksums" in
  ghx_*_checksums.txt ) ;;
  * )
    printf 'refusing non-ghx checksum file: %s\n' "$checksums" >&2
    exit 1
    ;;
esac
```

Verify the checksum with standard shell tools:

```sh
expected="$(awk -v file="$artifact" '$2 == file { print $1 }' "$checksums")"
actual="$(shasum -a 256 "$artifact" | awk '{ print $1 }')"

if [ -z "$expected" ]; then
  printf 'missing %s in %s\n' "$artifact" "$checksums" >&2
  exit 1
fi

if [ "$actual" != "$expected" ]; then
  printf 'checksum mismatch for %s\nexpected %s\nactual   %s\n' "$artifact" "$expected" "$actual" >&2
  exit 1
fi
```

If `ghx` is already installed and the production release published a GitHub attestation storage record, verify the artifact attestation through the GitHub API:

```sh
ghx at verify -R "$repo" "$artifact"
```

The successful output must identify `agustif/ghx` as the attesting repository. It must not identify `cli/cli`.

If only upstream `gh` is installed, it can still verify the fork artifact as long as the repository is explicit and the same attestation storage record exists:

```sh
gh at verify -R agustif/ghx "$artifact"
```

Do not run `gh at verify -R cli/cli` for `ghx_*` files.

To verify without installing either `ghx` or upstream `gh`, download the Sigstore bundle recorded by the release operator and verify the blob with `cosign`:

```sh
cosign verify-blob-attestation \
  --bundle agustif-ghx-attestation.sigstore.json \
  --new-bundle-format \
  --certificate-oidc-issuer="https://token.actions.githubusercontent.com" \
  --certificate-identity="https://github.com/agustif/ghx/.github/workflows/deployment.yml@refs/heads/trunk" \
  "$artifact"
```

If the release plan uses a dedicated release branch or a replacement workflow, replace the certificate identity with the exact approved workflow ref recorded by that plan. Do not loosen the identity to `cli/cli` or to a wildcard while verifying production artifacts.

## Operator checklist

Before publishing a `ghx` release, record these values in the release issue or release notes:

```text
tag:
target commit:
workflow run:
workflow ref:
release repository: agustif/ghx
artifact prefix: ghx_
checksum file:
attestation storage:
signing status:
package channels:
```

Then check:

- release assets are attached to `https://github.com/agustif/ghx/releases/tag/<tag>`
- every archive intended for users starts with `ghx_`
- the checksum file is named `ghx_<version>_checksums.txt`
- checksum contents list `ghx_*` assets, not `gh_*` assets
- `ghx at verify -R agustif/ghx <artifact>` succeeds for at least one artifact from each built platform
- the attestation output names `agustif/ghx`
- the signer workflow identity matches the approved workflow ref
- no release note, README snippet, or handoff tells users to verify `ghx_*` artifacts against `cli/cli`

Stop the release if any of those checks fail.

## Snapshot versus production

Not every `ghx` build has the same provenance contract.

| Build type | Expected location | Expected proof | Not expected |
| --- | --- | --- | --- |
| Local snapshot from `.goreleaser-ghx.yml` | local `dist/` directory | local checksum generated by the operator | GitHub Release, GitHub attestation, package signing, notarization |
| Staging workflow artifact | GitHub Actions workflow artifacts in `agustif/ghx` | workflow run id, branch or commit, `ghx_*` archive names, optional local checksum | durable release URL, package-manager publication, production signing |
| Production release | GitHub Release in `agustif/ghx` | `ghx_*` checksums, GitHub artifact attestations, approved signing or explicit unsigned policy | upstream `cli/cli` release identity |

For staging, it is acceptable for artifacts to be unsigned or unattested while the fork-owned workflow is still being proven. The staging handoff must say that clearly and must include the workflow run id and target commit.

For production, the release is not ready until users can verify the downloaded `ghx_*` artifact against `agustif/ghx` without trusting an upstream `cli/cli` example.

## Future gates

Issue #65 stays open until the fork-owned release path has these gates:

- workflow permissions include `contents: write`, `id-token: write`, and `attestations: write` only in the release workflow that owns publication
- attestation subject paths use `dist/ghx_*`
- the release plan states whether attestations are stored in the GitHub attestation API, attached as Sigstore bundles, or both
- checksum generation uses `shasum -a 256 ghx_* > ghx_<version>_checksums.txt`
- production release creation targets `agustif/ghx`
- macOS, Windows, RPM, and Debian signing identities are fork-owned or explicitly documented as unavailable for the release
- release notes link to this runbook or an equivalent user-facing verification section
- CI or release smoke checks fail if production docs combine `ghx_*` artifacts with `cli/cli` verification commands

## Minimal local helper

This snippet is intentionally documentation-only. It needs `awk` and `shasum`, which are already present on macOS and common Unix environments.

```sh
verify_ghx_checksum() {
  artifact="$1"
  checksums="$2"

  case "$artifact" in
    ghx_* ) ;;
    * )
      printf 'not a ghx artifact: %s\n' "$artifact" >&2
      return 2
      ;;
  esac

  expected="$(awk -v file="$artifact" '$2 == file { print $1 }' "$checksums")"
  actual="$(shasum -a 256 "$artifact" | awk '{ print $1 }')"

  if [ -z "$expected" ]; then
    printf 'missing %s in %s\n' "$artifact" "$checksums" >&2
    return 3
  fi

  if [ "$actual" != "$expected" ]; then
    printf 'checksum mismatch for %s\n' "$artifact" >&2
    return 4
  fi

  printf 'checksum ok: %s\n' "$artifact"
}
```

## Evidence anchors

- [issue #65](https://github.com/agustif/ghx/issues/65)
- [first-class release migration](ghx-first-class-release-migration.md)
- [workflow automation audit](ghx-workflow-automation-audit.md)
- [release smoke matrix](ghx-release-smoke-matrix.md)
- [release operator runbook](ghx-release-operator-runbook.md)
- [release process deep dive](../release-process-deep-dive.md)
