# ghx research index

Status: active
Date: 2026-05-17

Research notes capture broad investigation, product angles, constraints, risks, and open questions before a decision is ready for an ADR or a proposal is ready for an RFC.

## Records

| Note | Roadmap issue | Area |
| --- | --- | --- |
| [0001](0001-ghx-governance-and-decision-system.md) | [#10](https://github.com/agustif/ghx/issues/10) | Governance, ADR/RFC taxonomy, issue tree discipline, and release gates. |
| [0002](0002-ghx-identity-security-and-trust-boundaries.md) | [#11](https://github.com/agustif/ghx/issues/11) | Account safety, token scopes, extension trust, generated API escape hatches, and audit boundaries. |
| [0003](0003-ghx-upstream-architecture-and-command-map.md) | [#12](https://github.com/agustif/ghx/issues/12) | Upstream command architecture, JSON conventions, testing patterns, and rebase-safe evolution. |
| [0004](0004-ghx-release-upstream-and-distribution.md) | [#13](https://github.com/agustif/ghx/issues/13) | Release lanes, upstream sync, side-by-side distribution, versioning, and rollback. |
| [0005](0005-ghx-cli-ux-and-human-factors.md) | [#14](https://github.com/agustif/ghx/issues/14) | Command naming, explain/dry-run ergonomics, output contracts, companion tools, and accessibility. |

## Rules

- Research notes may include unresolved tradeoffs and competing options.
- Promote a research conclusion to an ADR when the fork commits to a direction.
- Promote an implementation-shaped proposal to an RFC when it needs rollout, acceptance criteria, and ownership.
- Link roadmap issues or subissues when the note creates executable work.
