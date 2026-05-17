# ADR 0006: Agent progress contract

Status: Accepted
Date: 2026-05-17

## Context

`ghx` is being shaped for humans and agent swarms. Long-running work needs progress that is visible in chat, durable on disk, and machine-readable enough for other agents to resume. The current local convention uses `p` with `0/100` progress, task trees, lanes, agents, statuses, and a shared progress log.

## Decision

`ghx` commands that perform multi-step work should adopt a progress contract:

- Percent progress is reported as `0/100` through `100/100`.
- Work is represented as an ASCII task tree with stable lanes and subtasks.
- Long-running commands accept `--progress-log <path>`.
- Output can be emitted as human text, JSON, and progress-log events.
- Agents should update progress before major phases, after durable writes, and when blocked.

## Consequences

- Progress is resumable across terminals and agents.
- Human-facing output remains concise while structured logs retain detail.
- Commands need a shared progress writer rather than ad hoc print statements.
- Progress events must avoid secrets and token material.

## Follow-ups

- Add `internal/progress` or `internal/ghx/progress`.
- Add `--progress-log` helper plumbing to long-running ghx-only commands.
- Document status values: `start`, `progress`, `done`, `blocked`, `note`, `warn`.
- Add tests for JSON/progress output stability.
