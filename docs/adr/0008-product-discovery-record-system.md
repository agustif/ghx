# ADR 0008: Product discovery record system

Status: Accepted
Date: 2026-05-17
Roadmap: [#7](https://github.com/agustif/ghx/issues/7)

## Context

`ghx` is moving from a small fork patch set into an account-safe, agent-ready GitHub control plane. Ideas now arrive during implementation, issue triage, API exploration, and agent collaboration. If those ideas stay in chat, they are hard to resume, hard to review, and easy to lose.

The repo already has ADRs, plans, gap maps, and GitHub subissues. It now also needs a clear place for exploratory research and larger RFC proposals.

## Decision

`ghx` will use a four-layer product discovery record system:

- `docs/research/`: broad notes, competing options, risk maps, product angles, and source-backed investigation.
- `docs/rfcs/`: proposed larger changes with scope, design, rollout, validation, and acceptance criteria.
- `docs/adr/`: accepted decisions that constrain future implementation.
- GitHub issues and subissues: executable roadmap units with status and ownership.

Agents should capture new product ideas in this system before ending a session. Standing behavioral rules may be summarized in `AGENTS.md`, but implementation detail belongs in research notes, RFCs, ADRs, docs, and issues.

## Consequences

- Future agents can resume from files and issue trees rather than reconstructing product direction from chat.
- The fork can separate open exploration from accepted decisions.
- Roadmap issues can link to research and RFC sources.
- `AGENTS.md` stays concise and operational instead of becoming the full product backlog.

## Follow-ups

- Link research and RFC indexes from `docs/ghx.md`.
- Attach major roadmap issues to their research or RFC sources.
- Add a lightweight check later for broken links across `docs/research`, `docs/rfcs`, `docs/adr`, and roadmap docs.
