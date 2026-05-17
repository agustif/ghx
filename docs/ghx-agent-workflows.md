# ghx agent workflows

Status: draft
Date: 2026-05-17
Related ADR: [ADR 0006](adr/0006-agent-progress-contract.md)

This guide describes how agents should use `ghx` safely and collaboratively.

## Account safety

Before remote mutation, agents should verify:

- selected host
- selected account
- account source
- target owner/repo
- token scopes

Preferred command path:

```sh
ghx ctx explain --json
```

Until `ghx ctx` exists, use:

```sh
ghx auth status
ghx repo view --json nameWithOwner,url
```

## Progress logs

Long-running work should report:

- `0/100` at start
- task tree updates during major phases
- blockers with exact next action
- `100/100` only after verification and durable writes

Future `ghx` commands should accept:

```sh
--progress-log /path/to/progress.log
```

## Output modes

Agent-facing commands should support:

- `--json`
- `--jq`
- `--template` when aligned with upstream patterns
- `--explain`
- `--dry-run` for mutation

## Screenshot and proof artifacts

Use `gh attach` for visual evidence when a PR or issue needs screenshot-backed proof.

Rules:

- use unique basenames before attaching many files
- attach generated proof sheets and evidence JSON together
- avoid uploading secrets or private local paths in screenshots
- link the attached proof in the PR summary or comment

## Collaboration pattern

Agents should:

- work in scoped branches or worktrees when practical
- avoid reverting unrelated WIP
- log progress before and after durable changes
- keep generated reports reproducible
- prefer read-only diagnosis before mutation
- use wrappers for extensions until account behavior is verified
