# ADR 0001: Scoped account binding

Status: Accepted
Date: 2026-05-17

## Context

Upstream `gh` supports multiple accounts per host, but the active account remains host-global. That is not enough for users and agents working across many repositories, forks, employers, and terminals at the same time. A global `gh auth switch` can make the next command in another working tree target the wrong account.

The `ghx` fork already adds scoped account selection through:

- `GH_ACCOUNT`
- `GH_ACCOUNT_SESSION`
- cwd-scoped config
- `.ghaccount`
- visible account source in `auth status`

## Decision

`ghx` will keep upstream multi-account behavior, but account selection for API and git credential flows will resolve through a scoped precedence chain before falling back to the host-global active account.

Precedence:

1. Explicit command flags when available.
2. `GH_ACCOUNT`.
3. `GH_ACCOUNT_SESSION`.
4. nearest `.ghaccount`.
5. cwd-scoped config.
6. host-global active account.

`.ghaccount` is a local account binding file. It does not have to be committed. Users can put it in `.git/info/exclude`, global git ignore, or any parent folder that should govern a subtree.

## Consequences

- `ghx auth status` must keep showing the selected account and the source that selected it.
- Mutating commands should expose account/repo intent through `--explain` or preflight output.
- Existing upstream flows still work for users who only use host-global accounts.
- Tests must cover precedence, nearest-file lookup, missing accounts, and git credential helper behavior.

## Follow-ups

- Add `ghx ctx`, `ghx ctx explain`, and `ghx ctx doctor`.
- Add `ghx ctx bind --account <login> [--cwd <path>]`.
- Add `.ghaccount` docs with ignore examples.
- Add JSON output for account resolution so agents can verify target identity before mutation.
