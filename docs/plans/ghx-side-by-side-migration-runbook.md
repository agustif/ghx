# ghx side-by-side migration and rollback runbook

Status: active
Date: 2026-05-19
Issue: [#67](https://github.com/agustif/ghx/issues/67)
Related docs: [ghx vs gh](../ghx-vs-gh.md), [Multiple accounts](../multiple-accounts.md), [Installation from source](../install_source.md), [Release upstream and distribution research](../research/0004-ghx-release-upstream-and-distribution.md)

## Goal

Document the exact path for installing `ghx` next to stock `gh`, exercising scoped account selection, and rolling back without disturbing the upstream binary.

This runbook is a companion to the current release docs in [Releasing](../releasing.md) and [Release process deep dive](../release-process-deep-dive.md), but it only covers the fork-specific migration path that keeps `gh` usable.

## Trigger conditions

Use this runbook when:

- you want to install `ghx` without replacing `gh`
- you need to test the `ghx` account precedence rules on a real machine
- you need a repeatable rollback path for shadow mode or sidecar installs
- you need to explain the git credential helper caveat before a release or migration

## Task graph

```text
side-by-side migration
|- preflight
|  |- confirm stock gh still works
|  `- confirm target repo and branch
|- sidecar install
|  |- build ghx
|  |- install ghx outside the stock gh path
|  `- verify ghx version and help text
|- account selection
|  |- verify auth status source
|  |- test GH_ACCOUNT and GH_ACCOUNT_SESSION
|  `- test cwd binding or .ghaccount
|- git helper check
|  |- inspect credential.helper
|  `- switch to ghx auth git-credential when desired
`- rollback
   |- remove ghx from PATH
   |- restore the recorded helper string
   `- verify gh remains usable
```

## Preflight

Start from a clean tree and verify that the upstream binary still works before touching the fork install.

```sh
git status --short
git remote -v
gh version
gh auth status
```

If `gh` already depends on a custom helper or alias, record that first. Do not overwrite it without a rollback note.

## Sidecar install

The repo already documents the source install path for a fork-specific binary. Use that path first.

```sh
make bin/ghx
make install-ghx prefix=$HOME/.local
$HOME/.local/bin/ghx version
$HOME/.local/bin/ghx auth status
```

If you want the fork binary on your interactive `PATH`, add the install prefix without removing the stock `gh` binary.

## Scoped account check

The current fork uses the following precedence, with explicit token environment variables still winning:

1. `GH_TOKEN` or `GITHUB_TOKEN`
2. `GH_ACCOUNT`
3. `GH_ACCOUNT_SESSION`
4. nearest `.ghaccount`
5. cwd binding
6. host-global active account

The release docs in [ghx vs gh](../ghx-vs-gh.md) and [Multiple accounts](../multiple-accounts.md) already describe the user-facing behavior. This runbook only records the migration verification sequence.

Verify the account source and then exercise each scoped input separately:

```sh
ghx auth status
GH_ACCOUNT=<login> ghx auth status
GH_ACCOUNT_SESSION=<name> ghx auth status
ghx auth switch --user <login> --scope cwd
ghx auth status
```

If you need a local project default, create or update `.ghaccount` in the project root and re-run `ghx auth status` from that directory.

## Git helper check

Side-by-side installs are only safe when the git credential helper is intentional.

Inspect the current helper before changing it:

```sh
git config --get credential.helper
```

If git should use the fork binary, run the fork helper setup from `ghx` and re-check the helper string:

```sh
ghx auth setup-git
git config --get credential.helper
```

The helper should resolve to `ghx auth git-credential` when the fork is meant to serve git operations.

## Shadow mode rehearsal

Shadow mode means placing `ghx` earlier on `PATH` or symlinking `gh` to `ghx` on purpose. Do this only after the sidecar path is proven.

Suggested rehearsal:

```sh
command -v gh
command -v ghx
ln -sf "$HOME/.local/bin/ghx" "$HOME/.local/bin/gh"
gh version
ghx version
```

If the shell or scripts begin resolving the wrong binary, stop and remove the symlink immediately.

## Rollback

Rollback should leave stock `gh` available and should not assume any fork-only auth store.

Sidecar rollback:

```sh
rm -f "$HOME/.local/bin/ghx"
hash -r
command -v gh
gh version
```

Shadow mode rollback:

```sh
rm -f "$HOME/.local/bin/gh"
hash -r
command -v gh
gh version
```

If you changed the git credential helper, restore the helper string you recorded during preflight. Then re-run `git config --get credential.helper` and `gh auth status`.

If you created `.ghaccount` for the test, remove it or move it back into the project-specific working directory before you declare the rollback complete.

## Validation

Minimum checks for a successful migration rehearsal:

- `gh version` still works after the fork install
- `ghx version` reports the fork binary name
- `ghx auth status` reports the winning source for each scoped account input
- `git config --get credential.helper` matches the intended helper
- removing `ghx` or the `gh` symlink restores the stock binary path

## Evidence anchors

- [docs/ghx-vs-gh.md](../ghx-vs-gh.md)
- [docs/multiple-accounts.md](../multiple-accounts.md)
- [docs/install_source.md](../install_source.md)
- [docs/research/0004-ghx-release-upstream-and-distribution.md](../research/0004-ghx-release-upstream-and-distribution.md)
- [issue #67](https://github.com/agustif/ghx/issues/67)
