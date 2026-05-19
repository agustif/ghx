# ghx vs gh

Status: active
Date: 2026-05-19

This document tracks behavior that is intentionally different in `ghx` compared with regular upstream GitHub CLI (`gh`). It only covers behavior that is shipped on this fork's `trunk`; roadmap ideas live in the gap map, RFCs, plans, and research notes.

Snapshot used for this page:

- `ghx`: `origin/trunk` at `6ae6652d0`
- regular `gh`: `upstream/trunk` at `81ed6d4e3`
- comparison command: `git diff upstream/trunk...origin/trunk`

## Quick summary

| Area | Regular `gh` | `ghx` divergence | User-facing command or behavior |
| --- | --- | --- | --- |
| Binary identity | Builds and self-references `gh`. | Also builds a rebranded `ghx` binary without requiring replacement of upstream `gh`. | `make bin/ghx`, `make install-ghx prefix=$HOME/.local`, `ghx --version`, help text that says `ghx`. |
| Scoped account selection | Active account is host-global after `gh auth switch`. | Account can be selected by process, named session, nearest `.ghaccount`, or cwd binding before falling back to host-global active account. | `GH_ACCOUNT=<login> ghx ...`, `GH_ACCOUNT_SESSION=<name> ghx ...`, `.ghaccount`, `ghx auth switch --scope cwd`. |
| Auth status evidence | Shows active account, token source, scopes, and git protocol. | Also reports the active account source so automation can explain why an account was selected. | `ghx auth status` includes `Active account source: ...`. |
| Git credential helper | Regular helper self-reference assumes `gh`. | `ghx` recognizes its own helper name as well as upstream `gh`. | Git credential setup and discovery work when the binary is installed as `ghx`. |
| Issue subissues | No first-class command group in the compared upstream snapshot. | Adds a native subissue command group with JSON output for listings. | `ghx issue subissue list/add/remove/reprioritize`, alias `ghx issue subissues`. |
| Issue creation parent link | Regular issue creation has no parent issue flag in the compared upstream snapshot. | Can create a new issue as a subissue of an existing issue in the same repo. | `ghx issue create --parent 123 --title "Child" --body-file body.md`. |
| Issue search fields | `gh issue list --search` does not expose issue-field scoping in the compared upstream snapshot. | Search can be constrained to issue title, body, and comments. | `ghx issue list --search "runner failed" --match body,comments`. |
| Runtime update and version links | Update checks and changelog links point at `cli/cli`, and Homebrew users can be told to run `brew upgrade gh`. | `ghx` update checks and version links point at `agustif/ghx`; `ghx` does not print the upstream `brew upgrade gh` hint. | `ghx --version`, update notifier output. |
| Fork documentation | Upstream docs describe regular `gh` development and usage. | Adds fork-specific ADRs, RFCs, plans, research notes, API coverage reports, and operational gap maps. | Start at `docs/ghx.md`. |
| Generated reference | The public manual at `cli.github.com/manual` is generated for upstream `gh`. | `ghx` generated reference can be produced locally with `--command-name ghx`, but is not published to the upstream site. | `go run ./cmd/gen-docs --website --doc-path dist/ghx-manual --command-name ghx`. |

## Compatibility contract

`ghx` is an additive fork of upstream `gh`, not a clean-room replacement. The baseline expectation is that normal `gh` commands, configuration, auth storage, and extension conventions keep working unless this page calls out a deliberate difference.

The fork currently supports two operating modes:

- Side-by-side mode: invoke `ghx` directly while leaving upstream `gh` installed.
- Shadow mode: place a `gh` symlink to `ghx` earlier on `PATH` after accepting that the fork now handles commands typed as `gh`.

Side-by-side mode is the safer default while the fork's release lane, completions, manpages, and update messaging are still maturing.

## What stays the same as gh

Unless noted in this page, `ghx` inherits the upstream command tree, flags, config files, keyring-backed auth storage, extension model, JSON conventions, and raw `api` escape hatch from `gh`.

Important shared behavior:

- `GH_TOKEN` and `GITHUB_TOKEN` still take priority over stored auth.
- Existing `gh auth login`, `gh auth token`, `gh auth logout`, and host-global `gh auth switch` behavior remains available.
- `ghx api` remains the raw REST and GraphQL escape hatch.
- Git credential behavior still depends on the helper configured in git config.

## What ghx adds today

The shipped fork delta is:

- a rebranded `ghx` binary and source install target
- scoped account selection through env, session, nearest `.ghaccount`, and cwd bindings
- visible account-source evidence in `auth status`
- `ghx` git credential helper recognition
- issue subissue commands
- parent issue creation through `issue create --parent`
- field-scoped issue text search through `issue list --match`
- fork-local docs, ADRs, RFCs, plans, research notes, and gap maps

## What is roadmap, not shipped

Docs such as [the gap map](ghx-gap-map.md), [first delivery slices](plans/first-delivery-slices.md), and research notes describe intended future work. They are not user-facing command guarantees until the behavior appears in this page.

Examples of planned surfaces that are not currently shipped as first-class commands:

- `ghx ctx`
- `ghx api explain`
- `ghx pr ready`
- `ghx ci doctor`
- `ghx rules explain`
- `ghx tools doctor`

## Binary identity and installation

`ghx` keeps the upstream command tree compatible while allowing the fork to be installed side by side with regular `gh`.

Implemented divergence:

- `Makefile` adds `bin/ghx` and `install-ghx` targets.
- `script/build.go` builds both `bin/gh` and `bin/ghx` through the same entry point while setting `internal/build.Name` from the output binary name.
- `internal/ghcmd/cmd.go`, `pkg/cmd/root/root.go`, `pkg/cmd/root/help.go`, and `pkg/cmd/version/version.go` use the configured command name in help, auth guidance, executable lookup, and version output.
- `README.md` documents side-by-side build and install commands.

Expected behavior:

```sh
make bin/ghx
make install-ghx prefix=$HOME/.local
ghx --version
ghx issue list
```

Installing `ghx` does not require replacing upstream `gh`. A machine can still opt into `ghx` as default `gh` by placing a symlink earlier on `PATH`.

Current install caveat: `make install-ghx` installs the fork binary, fork-owned completions, and fork-owned manpages, but packaged release artifacts are not yet first-class. Package manager instructions in the upstream install docs install regular `gh`, not `ghx`. The execution map for closing these gaps is [first-class ghx release and automation migration](plans/ghx-first-class-release-migration.md).

## Generated reference and manual publishing

The public manual at `https://cli.github.com/manual/` remains upstream `gh` documentation. `ghx` does not publish generated command reference pages to that site.

Implemented divergence:

- `cmd/gen-docs` accepts `--command-name ghx`.
- `make manpages-ghx` writes `ghx*.1` manpages under `share/man/man1`.
- A local website reference can be generated into a fork-owned directory:

```sh
mkdir -p dist/ghx-manual
go run ./cmd/gen-docs --website --doc-path dist/ghx-manual --command-name ghx
```

Expected generated files include:

- `dist/ghx-manual/ghx.md`
- `dist/ghx-manual/ghx_help_environment.md`
- `dist/ghx-manual/ghx_issue.md`
- `dist/ghx-manual/ghx_issue_create.md`

Generated headings, links, and shell prompt examples use the configured `ghx` command name.

The inherited `site-docs` Make target and `.github/workflows/deployment.yml` release workflow remain upstream `gh` publication paths. They check out or update `github/cli.github.com`, operate on `manual/gh*.md`, and must not be treated as the `ghx` manual publication path unless a future release issue replaces them with fork-owned targets and credentials.

## Runtime update and version identity

Regular `gh` builds keep the upstream update channel and changelog links:

```sh
gh --version
```

When a binary is built as `ghx`, runtime release identity follows the fork:

- updateable `ghx` builds check `agustif/ghx` release metadata when the upstream default update repository is still configured
- updateable `ghx` builds use their own update state file so recent `gh` checks do not suppress fork release checks
- `ghx --version` links to `https://github.com/agustif/ghx/releases/...`
- update notifications say `A new release of ghx is available`
- `ghx` does not print the upstream Homebrew hint `brew upgrade gh`

Packagers can still override or disable the update repository explicitly. The fork remap only protects the default upstream `cli/cli` setting from leaking into `ghx` builds.

## Scoped account selection

Regular `gh` supports multiple logged-in accounts, but the active account is primarily host-global. `ghx` adds scoped account selection for local multi-account workflows.

Selection order in `ghx`:

1. Explicit token environment variables such as `GH_TOKEN` and `GITHUB_TOKEN` remain the strongest authentication signal.
2. `GH_ACCOUNT=<login>` selects one logged-in account for the current process.
3. `GH_ACCOUNT_SESSION=<name>` selects a named session binding from gh config.
4. The nearest `.ghaccount` file, searched upward from the invocation directory, selects a local account.
5. Cwd-scoped bindings from `ghx auth switch --scope cwd` select the closest matching configured path.
6. Host-global active account remains the fallback.

Implemented divergence:

- `internal/config/config.go` adds scoped account selectors, `.ghaccount` discovery, cwd and session bindings, and scoped token lookup.
- `internal/gh/gh.go` adds active-user source metadata.
- `pkg/cmd/auth/switch/switch.go` adds scoped switching.
- `pkg/cmd/auth/status/status.go` prints source information.
- `docs/multiple-accounts.md` documents the behavior.

Common forms:

```sh
GH_ACCOUNT=agustif ghx pr list --repo agustif/ghx
GH_ACCOUNT_SESSION=work ghx issue list
printf "agustif\n" > .ghaccount
ghx auth switch --user agustif --scope cwd
```

`.ghaccount` is intended as local machine state. It should normally be globally git-ignored and contain only a login.

## Auth status source visibility

`ghx auth status` is designed to answer not only "who am I logged in as?" but also "why did this account win?".

Examples of source labels:

- `GH_ACCOUNT`
- `GH_ACCOUNT_SESSION`
- `.ghaccount`
- cwd scoped binding
- host-global active account

This is important for agents and automation because remote mutations need a short audit trail that includes account, repository, token source, and selector source.

## Git credential helper compatibility

Regular `gh` recognizes the upstream helper shape. `ghx` also recognizes the `ghx` helper so git credential setup remains coherent when the fork is installed without replacing `gh`.

Side-by-side caveat: `ghx api` can honor scoped account selection while `git push` may still invoke stock `gh auth git-credential` if git config points at `gh`. Check the configured helper before assuming scoped `ghx` account selection also applies to git operations.

Implementation touches:

- `pkg/cmd/auth/shared/git_credential.go`
- `pkg/cmd/auth/shared/gitcredentials/helper_config.go`
- related tests under `pkg/cmd/auth/shared/gitcredentials`

## Issue subissue operations

`ghx` adds a first-class issue subissue command group:

```sh
ghx issue subissue list 123
ghx issue subissue add 123 456
ghx issue subissue remove 123 456
ghx issue subissue reprioritize 123 456 --before 455
ghx issue subissues list 123 --json number,title,state,url,repository
```

Implemented divergence:

- `pkg/cmd/issue/subissue/subissue.go` implements `list`, `add`, `remove`, and `reprioritize`.
- `pkg/cmd/issue/issue.go` registers the command group.
- `pkg/cmd/issue/subissue/subissue_test.go` covers command behavior.

Subissue listing supports `--json`, `--jq`, and `--template` with these fields:

- `number`
- `repository`
- `state`
- `title`
- `url`

## Parent issue creation

`ghx issue create` accepts `--parent` to create a new issue as a subissue of an existing issue.

```sh
ghx issue create --parent 123 --title "Add check inventory" --body-file issue.md
```

Implemented divergence:

- `pkg/cmd/issue/create/create.go` adds `--parent`, resolves a parent issue number or URL, rejects pull requests, and requires same-repository parent issues.
- `api/queries_issue.go` allows `parentIssueId` in the create mutation input.
- tests in `pkg/cmd/issue/create/create_test.go` cover parent creation behavior.

The parent must be an issue, not a pull request, and must live in the selected repository.

## Issue search field matching

`ghx issue list` adds `--match` to constrain `--search` to issue title, body, comments, or a comma-separated combination.

```sh
ghx issue list --search "timeout" --match title
ghx issue list --search "runner failed" --match body,comments
ghx issue list --search "auth" --match title,body,comments
```

Implemented divergence:

- `pkg/cmd/issue/list/list.go` adds `SearchFields` and validates that `--match` is only used with `--search`.
- `pkg/cmd/pr/shared/params.go` maps the selected fields to search qualifiers.
- tests in `pkg/cmd/issue/list/list_test.go` and `pkg/cmd/pr/shared/params_test.go` cover the behavior.

Invalid usage:

```sh
ghx issue list --match comments
```

returns:

```text
specify `--search` when using `--match`
```

## Fork-specific documentation

The fork now has a repo-local documentation system that regular upstream `gh` does not carry:

- `docs/ghx.md`: fork-specific docs index
- `docs/ghx-gap-map.md`: platform and CLI gap map
- `docs/ghx-api-coverage.md`: REST and GraphQL coverage seed report
- `docs/ghx-official-surface-report.md`: official surface report
- `docs/ghx-agent-workflows.md`: agent workflow and automation conventions
- `docs/ghx-extension-bundle.md`: extension bundle and wrapper posture
- `docs/adr/`: accepted architecture decisions
- `docs/rfcs/`: implementation-shaped proposals
- `docs/plans/`: implementation plans
- `docs/research/`: research notes and product discovery records

These docs are divergence too, but mostly as governance and planning. They should not be read as shipped command behavior unless the relevant command is also listed above.

## Upstream drift note

This fork currently has local divergence from upstream and is also behind the newest upstream `trunk` by at least one upstream documentation commit in the snapshot above. Rebase or merge work should keep behavior divergence explicit:

- If upstream adds the same feature, prefer upstream behavior unless the `ghx` behavior is intentionally account-safe or agent-oriented.
- If upstream adds a related feature with different flags or output, document the compatibility decision before changing `ghx`.
- Keep this file current whenever a merged PR changes user-facing behavior relative to regular `gh`.
