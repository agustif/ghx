# ghx rpm Repository Template

Status: template
Owner issue: https://github.com/agustif/ghx/issues/60

Use this file as the required record before publishing rpm repository metadata
for `ghx`. Do not publish dnf, yum, or zypper install docs until every required
field is filled in a release issue, release PR, or fork-owned repository.

## Required identity

| Field | Value |
| --- | --- |
| Package name | `ghx` |
| Binary path | `/usr/bin/ghx` |
| Source repository | `https://github.com/agustif/ghx` |
| Release asset prefix | `ghx_` |
| Repository owner | TODO |
| Repository URL | TODO |
| Signing key owner | TODO |
| Signing key fingerprint | TODO |
| Key rotation policy | TODO |
| Publishing workflow | TODO |
| Rollback owner | TODO |

## Side-by-side policy

The package must not declare `Conflicts`, `Obsoletes`, or `Provides` for upstream
`gh` by default. A user who already has upstream GitHub CLI installed must keep
both commands available:

```sh
gh version
ghx version
```

## Required smoke evidence

Record the command output in the release issue before publication:

```sh
sudo dnf install ./ghx_<version>_linux_amd64.rpm
command -v gh
command -v ghx
gh version
ghx version
rpm -ql ghx
sudo dnf remove ghx
gh version
```

## Documentation gate

Only after the repository and signing key exist may `docs/install_linux.md`
include rpm repository instructions for `ghx`. Until then, package-manager
instructions in this repo must continue to say that inherited upstream commands
install regular `gh`, not `ghx`.
