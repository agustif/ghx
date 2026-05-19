# 0004 ghx release, upstream sync, and distribution

Status: draft
Date: 2026-05-17

## Purpose

This note records the current release, packaging, update, docs, and compatibility shape of the `ghx` fork so the repo can decide whether `ghx` should stay a local sidecar binary or become a first-class distributable product.

User-facing summary: [ghx vs gh](../ghx-vs-gh.md).
Execution map: [first-class ghx release and automation migration](../plans/ghx-first-class-release-migration.md).

The central question is not whether `ghx` can be built. It can. The real question is which distribution contract the fork wants to support:

1. local-only sidecar binary
2. first-class `ghx` side-by-side distribution
3. full replacement of stock `gh`

The current repo state clearly supports option 1, partially supports option 2, and is not yet safe for option 3.

## Local snapshot

- Local repo root inspected: `/Users/af/gh/cli`
- Current branch: `af/scoped-gh-accounts`
- Current `HEAD`: `aa7fd4fb0e39821bf782d07e16ab5de65853ddbd`
- Remotes:
  - `origin` -> `https://github.com/agustif/ghx.git`
  - `upstream` -> `https://github.com/cli/cli.git`
- Local branch state inspected from `git branch -vv`:
  - `af/scoped-gh-accounts` tracks `origin/trunk`
  - local `trunk` is behind `origin/trunk` by 7 commits

This is the relevant upstream sync frame: the fork is still organized as an additive delta on top of upstream `cli/cli`, not as a clean-room product with separate release infrastructure.

## Proven current state

### 1. Local build support exists and is intentionally narrow

`ghx` is a real build target today.

- `/Users/af/gh/cli/Makefile` defines `bin/ghx` and `install-ghx`.
- `/Users/af/gh/cli/script/build.go` builds both `bin/gh` and `bin/ghx` from `./cmd/gh`.
- `/Users/af/gh/cli/script/build.go` injects `internal/build.Name` from the output filename, so a `bin/ghx` build changes command self-identification without requiring a separate entrypoint.
- `/Users/af/gh/cli/internal/build/build.go`, `/Users/af/gh/cli/internal/ghcmd/cmd.go`, and `/Users/af/gh/cli/pkg/cmd/root/root.go` carry that name through root help and version output.

That is enough for local development and explicit side-by-side invocation.

It is not yet a full distribution story because:

- `/Users/af/gh/cli/Makefile` `install-ghx` installs only the binary.
- `install-ghx` does not install completions.
- `install-ghx` does not install manpages.
- there is no `uninstall-ghx` path
- source install docs still describe `gh`, not `ghx`

### 2. Release automation is still upstream `gh` automation

The release pipeline is still branded, named, and packaged as `gh`.

- `/Users/af/gh/cli/.goreleaser.yml` sets `project_name: gh`
- `/Users/af/gh/cli/.goreleaser.yml` builds `bin/gh`, not `bin/ghx`
- `/Users/af/gh/cli/.goreleaser.yml` emits `gh_*` archive names
- `/Users/af/gh/cli/.goreleaser.yml` packages `/usr` contents as `gh`
- `/Users/af/gh/cli/.github/workflows/deployment.yml` creates temporary tags and runs `script/release --local` for `gh` artifact names
- `/Users/af/gh/cli/.github/workflows/deployment.yml` uploads `dist/*.tar.gz`, `dist/*.deb`, `dist/*.rpm`, `dist/*.zip`, `dist/*.pkg`, all named from the `gh` GoReleaser config
- `/Users/af/gh/cli/docs/releasing.md` documents a "GitHub CLI" release, upstream release notes generation, marketing-site update, and Homebrew formula bump for `gh`
- `/Users/af/gh/cli/docs/release-process-deep-dive.md` describes the same upstream pipeline in detail
- `/Users/af/gh/cli/script/pkgmacos` packages `/bin/gh` and emits `gh_<version>_macOS_universal.pkg`
- `/Users/af/gh/cli/.github/workflows/homebrew-bump.yml` bumps formula `gh`

There is also a release-contract mismatch already present in upstream-shaped tooling:

- `/Users/af/gh/cli/.goreleaser.yml` still supports prerelease naming behavior
- `/Users/af/gh/cli/.github/workflows/deployment.yml` creates prereleases when a tag contains a hyphen
- `/Users/af/gh/cli/.github/workflows/deployment.yml` also rejects any tag that is not strict `vN.N.N`
- `/Users/af/gh/cli/docs/release-process-deep-dive.md` already calls out that contradiction

So even before introducing `ghx`, the current release lane is tuned for stable upstream tags, not for a fork that may want more explicit prerelease or sync-build tags.

Conclusion: the repo does not yet have a `ghx` release lane. It has a `gh` release lane plus a local `ghx` build target.

Issue #56 update on 2026-05-19:

- `.goreleaser-ghx.yml` now exists as a dry-run-only fork release scaffold.
- It leaves `.goreleaser.yml` unchanged for upstream-compatible `gh` release behavior.
- It sets `project_name: ghx`, targets disabled SCM release metadata at `agustif/ghx`, builds `bin/ghx`, injects `internal/build.Name=ghx`, and names archives with the `ghx_{{ .Version }}_*` prefix.
- It does not yet create Linux packages, macOS pkg output, Windows MSI output, generated `ghx` completions, generated `ghx` manpages, signing artifacts, or update-channel metadata.
- It should be validated with `goreleaser check -f .goreleaser-ghx.yml` and a local dry run:

```bash
goreleaser release -f .goreleaser-ghx.yml --snapshot --clean --skip publish,announce --release-notes="$(mktemp)"
```

### 3. Version, changelog, and updater identity are only partially fork-aware

The command surface is partially rebranded but the release identity is not.

What is fork-aware:

- `/Users/af/gh/cli/pkg/cmd/root/root.go` parameterizes the command name in help and examples
- `/Users/af/gh/cli/pkg/cmd/version/version.go` renders version output with the chosen command name

What is still upstream-bound:

- `/Users/af/gh/cli/pkg/cmd/version/version.go` always points changelog links at `https://github.com/cli/cli/releases/...`
- `/Users/af/gh/cli/internal/ghcmd/update_enabled.go` sets `updaterEnabled = "cli/cli"` when the `updateable` build tag is present
- `/Users/af/gh/cli/internal/ghcmd/cmd.go` prints `A new release of gh is available:` and suggests `brew upgrade gh`
- `/Users/af/gh/cli/internal/update/update.go` checks `https://api.github.com/repos/<repo>/releases/latest`, so the chosen repo slug matters directly

This means a packaged `ghx` release is not safe to ship until version links, update notifier repo selection, and upgrade instructions stop pointing at upstream `gh`.

### 4. Stock `gh` compatibility is real, but side-by-side behavior has sharp edges

The repo intentionally supports side-by-side installation.

- `/Users/af/gh/cli/README.md` says `ghx` can be installed without replacing upstream `gh`
- `/Users/af/gh/cli/README.md` also suggests an optional `gh` symlink to `ghx` earlier on `PATH`

That is a useful compatibility posture, but it creates two important behavior modes:

1. explicit sidecar mode
2. shadow or override mode

#### Explicit sidecar mode

In explicit sidecar mode the user invokes `ghx` directly and keeps stock `gh` available.

Benefits:

- lowest blast radius
- easiest rollback
- preserves upstream packaging and tooling
- avoids confusing stock `gh` users on the same machine

Risks:

- docs, completions, and manpages still skew toward `gh`
- git credential helper may still point at stock `gh`
- users may assume `ghx` scoped selection also covers `git push`

#### Shadow or override mode

In shadow mode the user symlinks or aliases `gh` to `ghx`.

Benefits:

- uniform command name
- better parity between GitHub CLI commands and git credential helper calls if the helper actually resolves to the forked binary

Risks:

- release/update messaging still says `gh`
- packaged docs/completions are not fully fork-specific
- any fork-only config or migration mistake would hit the user's primary `gh` workflow

### 5. Auth and config state are still shared with stock `gh`

Today `ghx` is not isolated from upstream config and secret storage.

- `/Users/af/gh/cli/internal/config/config.go` still uses the standard config surface
- `/Users/af/gh/cli/pkg/cmd/auth/status/status.go` reconstructs token file paths under `GH_CONFIG_DIR/hosts.yml`
- `/Users/af/gh/cli/internal/config/config.go` still uses keyring service name `gh:<hostname>`
- `/Users/af/gh/cli/internal/config/migration/multi_account.go` keeps old host-level config in place for forward compatibility
- `/Users/af/gh/cli/docs/multiple-accounts.md` documents the multi-account migration and its rollback caveat for insecure storage

This shared-state design is a strong rollback anchor because uninstalling `ghx` does not strand secrets in a fork-only store.

It is also a compatibility hazard because stock `gh` and `ghx` can mutate the same auth and config state in different ways.

### 6. Scoped account override behavior is real and already stronger than upstream

This is the main reason `ghx` exists.

- `/Users/af/gh/cli/internal/config/config.go` resolves active user in this order:
  - `GH_ACCOUNT`
  - `GH_ACCOUNT_SESSION`
  - nearest `.ghaccount`
  - cwd-scoped account
  - host-global active account
- `/Users/af/gh/cli/pkg/cmd/auth/switch/switch.go` adds `--scope cwd` and `--scope session`
- `/Users/af/gh/cli/pkg/cmd/auth/status/status.go` reports the winning source as `GH_ACCOUNT`, `GH_ACCOUNT_SESSION=<selector>`, `nearest .ghaccount`, or `cwd scope`
- `/Users/af/gh/cli/docs/multiple-accounts.md` documents the same precedence and notes that `GH_TOKEN` and `GITHUB_TOKEN` still outrank account selection

This is release-relevant because it is the user-visible fork value that distribution must not break.

### 7. Git credential helper compatibility is not yet safe by default

This is the main side-by-side footgun.

- `/Users/af/gh/cli/pkg/cmd/auth/shared/gitcredentials/helper_config.go` treats both `gh` and `ghx` helper commands as "ours"
- `/Users/af/gh/cli/pkg/cmd/auth/shared/git_credential.go` assumes either binary can serve the active token path
- the scoped resolver lives in `/Users/af/gh/cli/internal/config/config.go`, not in upstream `gh`

That means a machine can easily end up in this state:

1. `ghx api` honors `.ghaccount` or `GH_ACCOUNT_SESSION`
2. `git push` still invokes stock `gh auth git-credential`
3. git operations therefore use host-global stock behavior instead of the fork's scoped account behavior

A first-class `ghx` release should not encourage side-by-side install unless the helper path and account-selection behavior are explicit and testable.

### 8. Docs generation and command docs are still `gh`-first

The docs pipeline is generated from the live cobra tree, which is good for upstream sync, but it still defaults to `gh`.

- `/Users/af/gh/cli/cmd/gen-docs/main.go` builds the root command without a fork-specific command name override
- `/Users/af/gh/cli/Makefile` `manpages` calls `go run ./cmd/gen-docs --man-page`
- `/Users/af/gh/cli/Makefile` `site-docs` removes and regenerates `site/manual/gh*.md`
- `/Users/af/gh/cli/Makefile` completions are generated by executing `bin/gh`, not `bin/ghx`

So even though `bin/ghx` help output is branded at runtime, generated docs, manual pages, and completions still represent `gh`.

### 9. Extension and wrapper distribution is still policy, not product

The docs already describe a future distribution and compatibility posture for wrapped extensions, but the implementation is not here yet.

- `/Users/af/gh/cli/docs/adr/0005-extension-bundles-and-wrappers.md`
- `/Users/af/gh/cli/docs/ghx-extension-bundle.md`
- `/Users/af/gh/cli/docs/adr/0007-ghx-config-directory.md`
- `/Users/af/gh/cli/docs/ghx-agent-workflows.md`

Those docs define `ghx ext bundle ...`, `ghx ext wrap`, `.ghx/` config, and `ghx ctx`, but current root registration in `/Users/af/gh/cli/pkg/cmd/root/root.go` still follows the stock extension flow.

This matters for release because a side-by-side `ghx` package cannot yet promise wrapper-safe extension execution or repo-local `.ghx/` configuration behavior.

## Upstream sync implications

The repo's own ADRs already point in the right direction.

- `/Users/af/gh/cli/docs/adr/0000-ghx-fork-scope-and-principles.md` says `ghx` is a maintained fork of `gh`, not a clean-room replacement
- `/Users/af/gh/cli/docs/research/0003-ghx-upstream-architecture-and-command-map.md` argues for additive deltas and rebase-safe command wiring

That implies a release strategy:

1. keep the fork delta narrow
2. avoid forking the whole release pipeline until the product contract is clear
3. prefer side-by-side distribution before any full override story
4. keep auth storage rollback-compatible with upstream as long as possible

A separate `ghx` release lane should therefore be introduced only after it can stay mostly parallel to upstream, not by rewriting the upstream release flow in-place.

## Distribution options

### Option A: local-only fork

Definition:

- keep `make bin/ghx`
- keep `make install-ghx`
- do not publish official `ghx` artifacts
- document `ghx` as a development and power-user binary

Pros:

- matches current repo reality
- lowest maintenance
- minimal divergence from upstream release engineering
- rollback is trivial

Cons:

- no clear install path for wider users
- value stays trapped in local builds
- docs and release notes remain confusing

### Option B: first-class side-by-side `ghx` distribution

Definition:

- publish `ghx` artifacts under their own repo and version stream
- keep stock `gh` install valid on the same machine
- require explicit `ghx` invocation unless users opt into override

Pros:

- safest path to distributability
- lets the fork prove account-safe behavior without taking over stock `gh`
- rollback remains easy: remove `ghx`, keep `gh`

Cons:

- requires parallel packaging, updater, docs, and install work
- must solve the git credential helper mismatch
- must define release versioning policy relative to upstream

### Option C: full `gh` replacement

Definition:

- package the fork so users can transparently use it as `gh`
- optimize for shadowing or replacing stock `gh`

Pros:

- simplest user story once mature
- best chance of consistent CLI and git credential behavior

Cons:

- highest blast radius
- update notifier, packaging identity, docs generation, and migration safety all need to be correct first
- harder rollback if the fork writes incompatible state

## Recommended direction

The repo should explicitly target option B first.

That means:

1. keep the current local sidecar flow
2. add a real side-by-side `ghx` release lane
3. delay any recommended `gh -> ghx` override until helper behavior, updater behavior, and docs identity are correct

This matches the current product value and avoids overcommitting to a full replacement before the fork-specific safety guarantees are proven.

## Versioning and naming guidance

The version scheme needs to answer two separate questions:

1. how the fork tracks upstream compatibility
2. how users detect fork-specific deltas

The least confusing release scheme is:

- keep upstream semantic version as the compatibility base
- append a fork suffix or build metadata for fork-only releases
- ensure updater and changelog URLs resolve to the fork repo, not `cli/cli`

Examples of possible schemes:

- `v2.73.0-ghx.1`
- `v2.73.0+ghx.1`
- `v2.73.0-ghx-scoped1`

Research conclusion:

- a pure upstream tag like `v2.73.0` is too ambiguous for a public `ghx` artifact
- a totally separate version line would make upstream sync harder to reason about

## Release gates

If `ghx` becomes distributable, every release should pass these gates.

### Gate 1: upstream sync gate

- the fork release commit should be based on a known upstream `trunk` point
- the fork delta should be enumerated in the release note
- rebases or merges from `/Users/af/gh/cli` `upstream` should be clean before release cut

### Gate 2: identity gate

- `ghx version` must link to fork release notes
- update notifier must check the fork repo, not `cli/cli`
- upgrade instructions must say `ghx`, not `gh`
- asset labels, package names, and release titles must be fork-correct

### Gate 3: coexistence gate

- side-by-side `gh` and `ghx` install must be tested
- `ghx auth status`, `ghx api /user`, and git credential helper behavior must agree in explicit test cases
- `.ghaccount`, `GH_ACCOUNT_SESSION`, and `GH_ACCOUNT` precedence must be validated in a side-by-side machine setup

### Gate 4: docs gate

- install docs must describe `ghx`
- source install docs must cover `make install-ghx`
- generated docs, manpages, and completions must be available for `ghx`
- rollback instructions must be documented

### Gate 5: rollback gate

- no irreversible config migration outside upstream-compatible state
- release artifacts can be removed cleanly
- users can revert to stock `gh` without losing auth state or corrupting host config

### Gate 6: changelog gate

- changelog must distinguish upstream sync from fork-specific changes
- release notes should link to the upstream base commit or version
- fork-only behavior changes should be summarized separately from merged upstream PRs

## Changelog and docs generation strategy

The repo should not treat changelog generation and command docs generation as the same problem.

Command docs generation is already structurally sound because it comes from the live cobra tree:

- `/Users/af/gh/cli/cmd/gen-docs/main.go`
- `/Users/af/gh/cli/internal/docs/markdown.go`
- `/Users/af/gh/cli/internal/docs/man.go`

The missing piece is fork identity:

- add a fork-aware command-name option to docs generation
- generate `ghx` manpages and completions separately from `gh`
- decide whether `ghx` has its own site docs or only repo-local docs

Changelog generation should be two-part:

1. upstream sync summary
2. fork delta summary

If the release uses GitHub-generated notes alone, users will not be able to tell which changes come from upstream `gh` and which are fork-specific scoped-account or workflow features.

## CI and rehearsal gaps

The repo has normal code-quality CI, but not a fork-aware release rehearsal lane.

- `/Users/af/gh/cli/.github/workflows/go.yml` validates tests and builds `./cmd/gh`
- `/Users/af/gh/cli/.github/workflows/lint.yml` runs lint and license generation checks
- `/Users/af/gh/cli/.github/workflows/govulncheck.yml` and `/Users/af/gh/cli/.github/workflows/codeql.yml` cover security scanning
- acceptance coverage under `/Users/af/gh/cli/acceptance/testdata/release/` exercises end-user `gh release` commands, not the repo's own `script/release` or deployment workflow

What is missing for `ghx` release confidence:

- a dry-run or staging workflow that validates `bin/ghx` artifact names
- a docs-generation check that proves `ghx` manpages and completions are emitted correctly
- a side-by-side install test that exercises stock `gh`, `ghx`, and git credential helper behavior together
- a rollback rehearsal that proves a bad `ghx` artifact can be removed without damaging stock `gh`

## Rollback strategy

The safest rollback plan is still side-by-side rollback, not in-place replacement rollback.

### Short-term rollback

For a bad local or packaged `ghx` build:

1. stop recommending `gh -> ghx` symlinks
2. remove `ghx` from `PATH`
3. keep stock `gh` in place
4. preserve shared auth state in existing config and keyring locations

This works today because `ghx` still uses upstream config and keyring surfaces.

### Release rollback

For a bad published `ghx` release:

1. delete or mark the bad release as superseded
2. publish a corrected release with clear rollback notes
3. if package-manager metadata was changed, revert formula or package references
4. verify updater and changelog links no longer point users at the bad artifact

### Migration rollback constraint

Do not introduce a fork-only auth or config migration until:

- the rollback story is proven
- the `.ghx/` config directory is implemented
- the fork can explain the exact precedence and storage graph

The current shared-state design is imperfect, but it is easier to unwind than a partially isolated fork state model.

## Open questions

1. Is `ghx` meant to stay a developer-focused sidecar, or is the intent to publish a maintained public binary stream?
2. If public, should the canonical install name stay `ghx`, or should the project eventually recommend `gh` shadow mode?
3. What exact git credential helper contract should be guaranteed for side-by-side installs?
4. Should `ghx` keep sharing `GH_CONFIG_DIR` and keyring storage until `.ghx/` is implemented, or should release packaging isolate fork state earlier?
5. What fork version scheme will best communicate both upstream compatibility and fork-specific deltas?

## Evidence used

- `/Users/af/gh/cli/README.md`
- `/Users/af/gh/cli/Makefile`
- `/Users/af/gh/cli/.goreleaser.yml`
- `/Users/af/gh/cli/.github/workflows/deployment.yml`
- `/Users/af/gh/cli/.github/workflows/homebrew-bump.yml`
- `/Users/af/gh/cli/.github/workflows/go.yml`
- `/Users/af/gh/cli/.github/workflows/lint.yml`
- `/Users/af/gh/cli/.github/workflows/govulncheck.yml`
- `/Users/af/gh/cli/.github/workflows/codeql.yml`
- `/Users/af/gh/cli/script/build.go`
- `/Users/af/gh/cli/script/release`
- `/Users/af/gh/cli/script/pkgmacos`
- `/Users/af/gh/cli/cmd/gen-docs/main.go`
- `/Users/af/gh/cli/internal/build/build.go`
- `/Users/af/gh/cli/internal/config/config.go`
- `/Users/af/gh/cli/internal/config/migration/multi_account.go`
- `/Users/af/gh/cli/internal/ghcmd/cmd.go`
- `/Users/af/gh/cli/internal/ghcmd/update_enabled.go`
- `/Users/af/gh/cli/internal/update/update.go`
- `/Users/af/gh/cli/pkg/cmd/root/root.go`
- `/Users/af/gh/cli/pkg/cmd/version/version.go`
- `/Users/af/gh/cli/pkg/cmd/release/create/http.go`
- `/Users/af/gh/cli/pkg/cmd/auth/status/status.go`
- `/Users/af/gh/cli/pkg/cmd/auth/switch/switch.go`
- `/Users/af/gh/cli/pkg/cmd/auth/shared/gitcredentials/helper_config.go`
- `/Users/af/gh/cli/pkg/cmd/auth/shared/git_credential.go`
- `/Users/af/gh/cli/docs/releasing.md`
- `/Users/af/gh/cli/docs/release-process-deep-dive.md`
- `/Users/af/gh/cli/docs/install_source.md`
- `/Users/af/gh/cli/docs/multiple-accounts.md`
- `/Users/af/gh/cli/docs/ghx.md`
- `/Users/af/gh/cli/docs/ghx-agent-workflows.md`
- `/Users/af/gh/cli/docs/ghx-extension-bundle.md`
- `/Users/af/gh/cli/docs/adr/0000-ghx-fork-scope-and-principles.md`
- `/Users/af/gh/cli/docs/adr/0005-extension-bundles-and-wrappers.md`
- `/Users/af/gh/cli/docs/adr/0007-ghx-config-directory.md`
- `/Users/af/gh/cli/docs/plans/scoped-account-rollout.md`
- `/Users/af/gh/cli/docs/research/0003-ghx-upstream-architecture-and-command-map.md`
- `/Users/af/gh/cli/acceptance/testdata/release/release-create.txtar`
