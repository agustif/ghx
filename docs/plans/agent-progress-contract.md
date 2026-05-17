# Plan: Agent progress contract

Status: active
Date: 2026-05-17
Related ADR: [ADR 0006](../adr/0006-agent-progress-contract.md)

## Goal

Make long-running `ghx` work understandable and resumable for humans and multiple agents.

## Format

Progress entries use:

- agent
- lane
- status
- progress percent from 0 to 100
- title
- message
- task tree

Statuses:

- `start`
- `progress`
- `done`
- `blocked`
- `note`
- `warn`

## Command contract

Long-running ghx-only commands should support:

- `--progress-log <path>`
- `--json`
- `--watch`
- `--explain`

Commands should emit progress when:

- work starts
- a durable artifact is written
- a remote mutation is about to happen
- a blocker is found
- a verification step passes or fails
- work finishes

## Example task tree

```text
ghx ci doctor
|- done: resolved scoped account and repo
|- done: fetched workflow runs
|- progress: reading failed job logs
|- pending: summarize blockers
|- pending: write JSON output
```

## Acceptance checks

- Progress logs never contain tokens or secret values.
- JSON output is stable enough for another agent to resume.
- Human progress text is concise.
- `100/100` is only used when the command has completed its intended scope.
