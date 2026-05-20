# ghx git credential helper parity

This note covers the part of the `ghx` migration that crosses from `ghx` commands into plain `git` operations such as `git fetch`, `git pull`, and `git push`.

## Contract

`ghx` uses the same auth storage model as regular `gh`. Installing `ghx` side by side with `gh` does not create separate token storage, and account selectors such as `GH_ACCOUNT`, `GH_ACCOUNT_SESSION`, `.ghaccount`, or cwd bindings do not change which binary `git` invokes.

Git operations follow git credential helper configuration. When `git` invokes the `ghx` helper, the helper can use `ghx` account selection. When git config still points at stock `gh auth git-credential`, `git push` can use the host-global `gh` account even if `ghx api /user` resolves a scoped account.

## Install

Configure the helper for all authenticated hosts:

```sh
ghx auth setup-git
```

Configure one host:

```sh
ghx auth setup-git --hostname github.com
```

Successful `ghx auth switch` calls also sync the HTTPS credential helper for the selected host to the running `ghx`
binary. This keeps repo-scoped account choices and plain `git fetch` / `git push` on the same resolver without requiring
a separate setup step after every account switch.

Configure a host that is not currently in `ghx auth status`:

```sh
ghx auth setup-git --hostname github.example.com --force
```

The configured helper is the invoked executable path, for example:

```text
!/Users/you/.local/bin/ghx auth git-credential
```

`ghx` recognizes both upstream `gh` and forked `ghx` helper commands as GitHub CLI helpers, but only the binary named in git config handles a later git credential request.

## Validate

Check the helper that git will use for GitHub:

```sh
git config --global --get-all credential.https://github.com.helper
```

Check the helper for gists:

```sh
git config --global --get-all credential.https://gist.github.com.helper
```

Check whether a generic helper still exists earlier in the chain:

```sh
git config --show-origin --global --get-all credential.helper
```

If any host-specific output points at `gh auth git-credential`, git operations for that host are still going through stock `gh`. If output points at `ghx auth git-credential`, git operations are entering the fork helper.

## Side-by-side constraints

- `gh` and `ghx` share auth storage. Do not expect separate login state unless the future config directory work explicitly changes that contract.
- Explicit token environment variables such as `GH_TOKEN` and `GITHUB_TOKEN` still take precedence over stored credentials.
- `gh auth setup-git` and `ghx auth setup-git` both write git config. The last setup command for a host controls which helper git invokes for that host.
- `ghx auth switch` changes selected accounts inside `ghx` and best-effort syncs the host-specific HTTPS git credential
  helper to the running `ghx` binary.
- SSH remotes still use SSH keys, not OAuth tokens. Use HTTPS remotes when account selection must follow `ghx`.
- Local repository git config can override global config. Inspect local config too if a single repository behaves differently.

## Rollback

Record current values before changing them:

```sh
git config --global --get-all credential.https://github.com.helper
git config --global --get-all credential.https://gist.github.com.helper
```

Remove the `ghx` host helper values:

```sh
git config --global --unset-all credential.https://github.com.helper
git config --global --unset-all credential.https://gist.github.com.helper
```

Restore stock `gh` for GitHub:

```sh
gh auth setup-git --hostname github.com
```

Or restore a captured helper value explicitly:

```sh
git config --global --add credential.https://github.com.helper '!gh auth git-credential'
```
