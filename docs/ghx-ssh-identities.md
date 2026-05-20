# ghx SSH identities

`ghx` can manage SSH remotes as a first-class account switching path.

GitHub SSH auth is key based. It does not use OAuth tokens and it does not call `ghx auth git-credential`. This means multi-account SSH support has to bind a GitHub account to an SSH private key and then make the current repository use that key.

## Contract

- SSH identities are stored per GitHub host and login in `ghx` config.
- `ghx auth ssh link` records the private key path for one account.
- `ghx auth ssh sync` writes repo-local `core.sshCommand`.
- `ghx auth switch` best-effort syncs repo-local `core.sshCommand` when the current repository has an SSH remote for the selected host.
- No global `~/.ssh/config` mutation is performed.
- HTTPS remotes still use the git credential helper path.

## Link an identity

```sh
ghx auth ssh link --hostname github.com --user agustiobvious --identity ~/.ssh/id_ed25519_github_agustiobvious
```

The identity file must exist. `~` is expanded before the path is stored.

List configured identities:

```sh
ghx auth ssh list --hostname github.com
```

Unlink an identity:

```sh
ghx auth ssh unlink --hostname github.com --user agustiobvious
```

Example output:

```text
github.com
  agustif: /Users/af/.ssh/id_ed25519_github_agustif
  agustiobvious: /Users/af/.ssh/id_ed25519_github_agustiobvious
```

## Sync a repository

For a checkout with an SSH remote such as:

```text
git@github.com:FlatFilers/ArcusCI.git
```

run:

```sh
ghx auth switch --hostname github.com --user agustiobvious --scope cwd
```

or:

```sh
ghx auth ssh sync --hostname github.com --user agustiobvious
```

The repository receives local config like:

```sh
git config --local core.sshCommand "ssh -i /Users/af/.ssh/id_ed25519_github_agustiobvious -o IdentitiesOnly=yes"
```

This keeps the SSH key selection scoped to the current repository. Other repositories and global SSH behavior are not changed.

## Validate

Check the remote protocol:

```sh
git remote -v
```

Check the repo-local SSH command:

```sh
git config --local --get core.sshCommand
```

Ask `ghx` to explain the current context:

```sh
ghx ctx doctor --json host,login,sshRemote,sshIdentity,sshCommand,warnings
```

Expected healthy state for an SSH checkout:

```json
{
  "host": "github.com",
  "login": "agustiobvious",
  "sshRemote": true,
  "sshIdentity": "/Users/af/.ssh/id_ed25519_github_agustiobvious",
  "sshCommand": "ssh -i /Users/af/.ssh/id_ed25519_github_agustiobvious -o IdentitiesOnly=yes"
}
```

## Failure modes

If an SSH remote is present but no identity is linked, `ghx auth switch` prints:

```text
SSH remote detected for github.com; link an identity with: ghx auth ssh link --hostname github.com --user <login> --identity ~/.ssh/<key>
```

If an identity is linked but repo-local config is not synced, run:

```sh
ghx auth ssh sync
```

If GitHub still resolves the wrong SSH account, confirm the key is attached to the expected GitHub account:

```sh
ssh -T git@github.com
```

GitHub decides the account from the key that was accepted by SSH. `ghx` can select the key for a repository, but it cannot make GitHub treat one key as a different account.

## Rollback

Remove the repo-local SSH override:

```sh
git config --local --unset core.sshCommand
```

Switch the remote to HTTPS if you want git operations to use OAuth token based account selection instead:

```sh
git remote set-url origin https://github.com/OWNER/REPO.git
```
