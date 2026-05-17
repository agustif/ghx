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

## Issue trees

Use native subissues for roadmaps, task decomposition, and multi-agent work queues.

Supported commands:

```sh
ghx issue create --parent <issue-number-or-url> --title "Child task" --body "Details"
ghx issue subissue list <issue-number-or-url> --json number,title,state,url,repository
ghx issue subissue add <parent-number-or-url> <child-number-or-url>
ghx issue subissue remove <parent-number-or-url> <child-number-or-url>
ghx issue subissue reprioritize <parent-number-or-url> <child-number-or-url> --before <sibling-number-or-url>
ghx issue subissue reprioritize <parent-number-or-url> <child-number-or-url> --after <sibling-number-or-url>
```

Agent rules:

- prefer `--json number,title,state,url,repository` for machine-readable tree reads
- keep parent and child mutations in the intended repo unless a URL explicitly names another repo
- record broad ideas as roadmap issues or subissues before leaving them in chat-only form
- use subissues for acceptance-sized work, not only large epics

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
