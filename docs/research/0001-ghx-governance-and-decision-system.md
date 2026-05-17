# Research 0001: ghx governance and decision system

Status: draft
Date: 2026-05-17
Related docs: [ghx index](../ghx.md), [ADR index](../adr/README.md), [RFC index](../rfcs/README.md), [Research index](README.md)

## Why this note exists

`ghx` is no longer just a narrow fork patch set. It is being shaped as an account-safe, agent-ready GitHub control plane with its own docs, plans, ADRs, and issue-tree expectations. That creates a governance problem: ideas can now appear in implementation threads, doc passes, API mining, issue triage, and agent handoffs. If those ideas do not move through one durable system, the fork drifts, agents overfit to chat context, and roadmap work becomes hard to resume.

This note defines the governance layer above individual feature plans. It focuses on record taxonomy, promotion rules, roadmap issue discipline, negative signals, release gates, and the contract that keeps docs, issues, and `AGENTS.md` aligned.

## Canonical record system

The current repo already implies a five-layer system:

1. `docs/research/` captures broad investigation, tradeoffs, constraints, risks, and open questions before the fork commits to a direction.
2. `docs/rfcs/` captures larger proposed changes that need design scope, rollout, validation, and ownership before implementation spreads.
3. `docs/adr/` captures accepted decisions that constrain future implementation and should survive chat sessions.
4. GitHub issues and subissues capture executable roadmap work with status, ownership, and acceptance-sized decomposition.
5. `AGENTS.md` captures only standing operational behavior for contributors and agents. It is not the product backlog and it should not hold detailed acceptance criteria.

This hierarchy matches the accepted product discovery record system in [ADR 0008](../adr/0008-product-discovery-record-system.md) and the repo-wide rules in [AGENTS.md](../../AGENTS.md).

## Taxonomy and promotion rules

### Research note

Use a research note when the repo needs one or more of:

- source-backed constraints
- competing options
- product angle mapping
- risk inventory
- unclear boundaries between local control-plane work and repo-owned automation

Research notes may stay unresolved. Their job is to sharpen the next decision, not to freeze one.

### RFC

Promote research to an RFC when the direction is credible enough to implement, but still needs:

- scope boundaries
- rollout sequencing
- validation strategy
- ownership
- non-goals
- issue-tree links

RFCs should stay implementation-shaped. If a note cannot name rollout and validation, it is probably still research.

### ADR

Promote research or RFC conclusions to an ADR when the fork commits to a durable rule that future work should assume by default. ADRs should answer "what is the decision and what constraints follow from it?" rather than re-litigating all options.

### Issue and subissue

Promote an RFC or plan into issues once work becomes executable. Parent issues should represent a theme, capability, or rollout lane. Subissues should represent acceptance-sized slices that can be reviewed, verified, and resumed independently.

### AGENTS.md

Promote behavior into `AGENTS.md` only when it is a standing operating rule for contributors or agents. Examples include output-shape expectations, account-safety rules, or where detailed acceptance criteria must live. If a rule depends on one feature proposal, one branch, or one temporary rollout phase, it belongs in docs or issues, not in `AGENTS.md`.

## Decision lifecycle

The preferred lifecycle for `ghx` work is:

1. Capture the idea in `docs/research/` if the direction is still open.
2. Promote to `docs/rfcs/` when design, rollout, and validation are needed.
3. Promote to `docs/adr/` when the fork commits to a durable rule.
4. Open or attach a parent issue when the work becomes roadmap-worthy.
5. Break the parent into subissues when slices are acceptance-sized and parallelizable.
6. Keep plans, docs, and issue links updated as implementation changes the real shape of the work.
7. Move any truly standing operational rule into `AGENTS.md` only after the rule is stable.

This preserves a clean distinction:

- research explains uncertainty
- RFCs explain proposed change
- ADRs explain accepted constraints
- issues drive execution
- `AGENTS.md` governs contribution behavior

## Roadmap issue-tree governance

The roadmap should live in issue trees, not only in prose. The repo already points in this direction: `docs/plans/first-delivery-slices.md` calls for converting the `ghx` roadmap into parent issues and acceptance-sized subissues, and `docs/ghx-agent-workflows.md` already treats subissues as the preferred queue shape for human and multi-agent work.

Issue-tree rules:

- Use one parent issue per capability lane, rollout, or governance theme.
- Use subissues for reviewable slices, not vague reminders.
- Link each parent issue back to the governing research note, RFC, ADR, or plan.
- Link each subissue to the parent and to the acceptance source when the slice is constrained by a plan or RFC.
- Prefer same-repo issue trees unless a URL intentionally points at another repo.
- Do not treat PRs as issue-tree parents. The current subissue surface is issue-first.
- Reprioritize subissues instead of opening duplicate siblings when order changes.
- Close or supersede stale work explicitly so the tree remains a reliable queue for agents.

## Negative signals

The control plane should treat the following as governance failures or escalation triggers:

- `AGENTS.md` contains feature backlog detail, acceptance criteria, or branch-local policy that should live in docs or issues.
- A plan or ADR exists with no issue or subissue linkage once work is actively being implemented.
- An issue tree exists with no research, RFC, ADR, or plan source for why the work matters.
- A mutating command ships without identity/target explainability.
- Agent-facing commands lack stable `--json` output.
- Long-running multi-step commands lack progress semantics or durable logs.
- A fork-only behavior silently diverges from upstream `gh` without an ADR, RFC, or explicit compatibility rationale.
- Index docs reference records that do not exist on disk.
- Companion-tool compatibility regresses for JSON, pipes, stdin, or predictable args.

Current repo evidence already shows some of these signals:

- The RFC layer exists, but both RFCs are still `Draft`, so the promotion ladder is defined more clearly than it is yet exercised.
- The documented agent contract expects `--json`, `--explain`, `--dry-run`, and `--progress-log`, but the new mutating subissue commands do not yet expose the full contract.
- The docs call for a roadmap issue tree, but the repo evidence is still primarily doc-shaped rather than broadly issue-linked.
- The release workflow is heavily documented, but the deep-dive doc makes clear that parts of the current release system were inherited and had to be reverse engineered by current maintainers.

These are not reasons to stop the fork. They are reasons to keep governance explicit until the implementation catches up.

## Upstream compatibility constraints

`ghx` is a maintained fork of `gh`, not a clean-room replacement. Governance should assume the following constraints from the start:

- Upstream behavior stays the default unless `ghx` needs an additive safety or control-plane feature.
- Fork-only features should compose with existing `gh` command patterns rather than forcing broad rewrites.
- JSON, `--jq`, `--template`, stdin behavior, and pipe-friendly output must remain predictable for both humans and companion tools.
- Raw escape hatches must remain available even when higher-level `ghx` workflows are added.
- Repo-owned automation belongs with `gh aw` or Actions workflows when auditability, approvals, and durable workflow files matter more than local interactivity.

The governance implication is simple: every fork-only behavior should carry a compatibility story. If that story is weak, the behavior should stay in research or RFC form until it is stronger.

## Release gates

A `ghx` feature is not ready to ship merely because the code works locally. Governance needs two release gates: one for shipping a feature contract inside `ghx`, and one for shipping binaries from this repo.

### Feature-governance gates

Feature readiness should pass the following gates:

1. Identity safety: repo, host, account, and token-scope resolution are visible or explainable before mutation.
2. Upstream compatibility: behavior is additive, reviewable, and does not create unnecessary rebase drag.
3. Agent contract: agent-facing reads support stable `--json`; mutating flows have preview or `--dry-run`; long-running flows have progress semantics when applicable.
4. Record integrity: the governing research, RFC, ADR, plan, or issue tree exists and matches the shipped behavior.
5. Documentation integrity: `docs/ghx.md`, the relevant index, and the feature guide do not disagree about what exists.
6. Proof integrity: tests, command examples, and validation notes cover the promised contract, especially for account safety and output shape.

Governance-heavy features should not skip gate 4 or gate 5. That would ship behavior faster at the cost of making future agents less reliable.

### Repo release-workflow gates

Binary release readiness should also respect the repo's existing release controls:

1. Release entry is manual through `script/release` and `workflow_dispatch`.
2. Maintainer approval is required for the deployment workflow.
3. Tag validation must pass before any platform build jobs run.
4. Production-only steps such as signing, notarization, attestations, and publish actions stay guarded by `inputs.environment == 'production'`.
5. The release job depends on the platform build lanes and should not be treated as a lightweight afterthought.

For `ghx`, the governance point is that feature maturity and binary release maturity are related but not identical. A feature can be doc-ready before it is release-ready, and a release workflow can succeed while still shipping a weak control-plane contract if the docs and issue tree were not kept current.

## Agent sync contract

Agents working in this fork should keep the durable record system synchronized as part of the change, not as optional cleanup.

When to update each surface:

- Update `AGENTS.md` only when the standing contributor or agent behavior changes.
- Update a research note when new constraints, options, or negative signals appear.
- Update or add an RFC when the repo needs rollout, validation, and ownership for a larger change.
- Update or add an ADR when a decision becomes accepted and future work should assume it.
- Update plans when sequencing or slice boundaries change.
- Open or update parent issues and subissues when the work is actionable.
- Update `docs/ghx.md` and the relevant index when a new durable record is added.

Minimum sync behavior for agents:

- Do not leave product-shaping decisions only in chat.
- Do not put detailed acceptance criteria only in `AGENTS.md`.
- Do not open roadmap-scale implementation without linking it back to a durable source document.
- Do not claim a workflow exists if the record, issue tree, and command surface are out of sync.

## Recommended governance posture

`ghx` should keep a strict split between durable decision records and executable roadmap work:

- research for open exploration
- RFCs for shaped proposals
- ADRs for accepted rules
- issues and subissues for execution
- `AGENTS.md` for stable operating rules

That posture fits the current fork goals: account safety first, agent-readiness second, and upstream compatibility throughout. It also gives agents a reliable way to resume work from disk and issue trees instead of reconstructing product intent from chat history.
