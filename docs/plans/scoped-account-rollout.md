# Plan: Scoped account rollout

Status: active
Date: 2026-05-17
Related ADR: [ADR 0001](../adr/0001-scoped-account-binding.md)

## Goal

Make `ghx` reliably select the intended GitHub account per command without changing the host-global active account unless the user explicitly asks for that.

## Scope

- `.ghaccount` nearest-file lookup.
- cwd-scoped account selection.
- session-scoped account selection.
- explicit env override with `GH_ACCOUNT`.
- visible account source in `auth status`.
- future `ghx ctx` commands.

## Not in scope

- Replacing upstream `gh auth switch`.
- Storing tokens outside existing `gh` auth storage.
- Persisting secrets in `.ghaccount`.

## Implementation slices

1. Finish tests for account source precedence.
2. Add docs for `.ghaccount` and ignore patterns.
3. Add `ghx ctx` read-only command that prints host, account, source, repo, remotes, and token scopes.
4. Add `ghx ctx explain --json`.
5. Add `ghx ctx bind --account <login> [--cwd <path>]`.
6. Add mutation preflight helpers that can show account/repo intent.

## Acceptance checks

- `ghx auth status` shows the active account source.
- `GH_ACCOUNT` beats `.ghaccount`.
- `GH_ACCOUNT_SESSION` beats `.ghaccount` but not `GH_ACCOUNT`.
- nearest `.ghaccount` beats host-global active account.
- missing account names produce a clear error and suggested `ghx auth status`.
- JSON output contains enough fields for agents to verify target identity.
