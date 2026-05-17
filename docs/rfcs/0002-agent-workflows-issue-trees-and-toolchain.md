# RFC 0002: Human and agent workflows in ghx

Status: Draft
Date: 2026-05-17
Related ADRs: [ADR 0004](../adr/0004-gh-aw-workflow-companion.md), [ADR 0005](../adr/0005-extension-bundles-and-wrappers.md), [ADR 0006](../adr/0006-agent-progress-contract.md), [ADR 0007](../adr/0007-ghx-config-directory.md)
Related plans: [Workflows and extensions](../plans/workflows-and-extensions.md), [Agent progress contract](../plans/agent-progress-contract.md), [`.ghx/` config directory](../plans/ghx-config-directory.md), [First delivery slices](../plans/first-delivery-slices.md)

## Summary

`ghx` should treat human and agent collaboration as one workflow system. The system needs a shared work graph, durable progress, proof artifacts, repo-owned workflows, project-local defaults, and compatibility with the Unix tools people already use to inspect and automate GitHub work.

This RFC proposes:

- subissue trees as the execution graph for roadmap work
- `p` progress logs as the resumable status layer for long-running commands
- `gh attach` backed proof artifacts for screenshots, logs, and evidence bundles
- compatibility-first support for `jq`, `rg`, `fzf`, `delta`, `gum`, and `yq`
- `gh aw` as the durable workflow companion for repo-owned automation
- `.ghx/` as the project-local configuration source for workflow defaults and tool posture
- noninteractive command behavior as the default contract

## Context

`ghx` already has accepted decisions for account binding, generated API proxies, progress reporting, extension bundles, `gh aw`, and the `.ghx/` config directory. What is still missing is one proposal that explains how those pieces fit together for everyday human and agent work.

The gap is not the existence of commands. The gap is the operating model:

- how a roadmap becomes a tree of executable subissues
- how agents report progress without flooding chat or leaking secrets
- how proof is collected and attached when work needs auditability
- how companion tools stay optional instead of becoming hidden runtime dependencies
- how repo-owned workflows and project-local config interact without undermining account safety
- how commands behave when there is no TTY and no human to answer prompts

This RFC defines the expected behavior so implementation can proceed slice by slice without each command inventing its own workflow rules.

## Goals

- Make issue trees the source of truth for executable roadmap work.
- Make progress logs durable, readable, and resumable by another agent or human.
- Make proof artifacts easy to collect, attach, and reference.
- Keep the existing Unix toolchain useful without forcing managed installs.
- Keep `gh aw` as the repo-owned workflow surface, not a replacement for `ghx`.
- Keep project-local defaults in `.ghx/` explainable and safe.
- Make noninteractive runs deterministic and scriptable by default.

## Non-goals

- Rebuilding `gh aw`, `gh attach`, `jq`, `rg`, `fzf`, `delta`, `gum`, or `yq` in-tree.
- Making interactive prompts mandatory for any core workflow.
- Turning `.ghx/` into a secret store or an opaque account binding mechanism.
- Replacing the existing progress contract with a new format.
- Introducing a separate task tracker UI outside the issue tree model.

## Proposal

### 1. Issue trees are the execution graph

Roadmap work should move from chat into issues and subissues as early as possible. Parent issues capture the broad objective. Subissues capture reviewable slices with clear acceptance criteria.

The RFC assumes a workflow like this:

```sh
ghx issue create --parent <issue-number-or-url> --title "Child task" --body "Details"
ghx issue subissue list <issue-number-or-url> --json number,title,state,url,repository
ghx issue subissue add <parent-number-or-url> <child-number-or-url>
ghx issue subissue remove <parent-number-or-url> <child-number-or-url>
ghx issue subissue reprioritize <parent-number-or-url> <child-number-or-url> --before <sibling-number-or-url>
ghx issue subissue reprioritize <parent-number-or-url> <child-number-or-url> --after <sibling-number-or-url>
```

Behavioral rules:

- issue tree reads should prefer stable JSON output for automation
- issue tree mutations should always identify the parent, child, and repository in the result
- broad ideas should be captured as roadmap issues or subissues before they stay in chat
- parent and child mutations should stay in the intended repository unless a URL explicitly names another repository
- subissues should be used for acceptance-sized work, not only large epics
- tree traversal should be deterministic so agents can compare state across runs

Agent workflow rules:

- treat the issue tree as the source of truth, not a chat summary
- add new follow-up work to the tree before leaving a thread
- keep one issue or subissue per independently reviewable slice
- avoid hidden backlog in free-form text when the work has a clear owner and acceptance shape

### 2. `p` progress logs define reporting cadence

The shared progress contract from ADR 0006 should be the reporting layer for long-running `ghx` commands. The important part here is not the symbol `p` itself, but the behavior it represents: a visible task tree with resumable checkpoints.

Required progress events:

- `start` when the command has resolved enough context to know its scope
- `progress` after each major phase
- `done` only when the intended scope is complete
- `blocked` when the command cannot continue without user or external action
- `note` for durable, low-noise checkpoints
- `warn` for recoverable but important concerns

Required reporting cadence:

- after account and repository resolution
- after config discovery and validation
- before any remote mutation
- after any durable local write
- after attachment or proof artifact creation
- after verification steps that confirm success or failure
- when the command hits a blocker
- at completion, with a final summary and stable machine-readable output

Cadence rules:

- no silent span should cross a major phase boundary
- `100/100` should only appear when the command has actually completed its intended scope
- progress output must stay secret-free
- progress logs must be resumable by another agent or a human operator

### 3. Proof artifacts are first-class

When a PR, issue, or investigation needs evidence, `ghx` should treat proof artifacts as a normal workflow output, not as an afterthought.

`gh attach` should be the primary proof-artifact transport for:

- screenshots
- proof sheets
- generated logs
- evidence JSON
- diff captures or review snapshots when they help explain a change

Expected behavior:

- attach related artifacts together when they form one proof bundle
- use unique basenames for large screenshot sets so attachments do not collide
- include the final attachment references in command output
- expose attachment metadata in JSON for automation
- keep proof generation noninteractive unless the user explicitly requests a TTY-driven step

Proof artifact rules:

- proof should be attached with the work that produced it
- attachments should be stable enough for a reviewer or another agent to fetch later
- proof output should not require a second manual copy step when the command already knows the artifact paths

### 4. Companion tools are compatibility surface, not core dependencies

`ghx` should stay compatible with the Unix tools that humans and agents already use to inspect and automate GitHub work. The compatibility rule is simple: native command output must be stable enough that the user can compose it with their own tools.

Initial companion set:

- `jq` for JSON filtering and scripted assertions
- `rg` for search across source, docs, logs, and fetched artifacts
- `fzf` for optional selection in TTY-driven flows
- `delta` for readable diffs and review output
- `gum` for optional prompts in local scripts
- `yq` for YAML workflows, Actions manifests, and config inspection

Compatibility rules:

- stable JSON fields come before managed companion installs
- text output should be predictable and pipe-friendly
- `fzf` and `gum` should remain optional and never gate noninteractive paths
- commands should continue to work if the companion tools are absent
- install or wrapper flows should never shadow an existing user binary silently

Expected native surfaces:

- `ghx tools doctor` to report missing tools, detected versions, and PATH source
- `ghx tools path` to print the resolved binary path for a companion
- `ghx tools install <name>` to install a curated companion with explicit user intent
- `.ghx/tools.toml` to declare optional project-pinned companion requirements

### 5. `gh aw` remains the durable workflow companion

`gh aw` should remain the repo-owned workflow execution surface. `ghx` should bind account safety, repo context, progress, and config to that surface rather than recreating it.

Workflow rules:

- `ghx` should locate repo workflow defaults from `.ghx/workflows.yml`
- `ghx` should explain account, repository, runner, and scope before executing a workflow that can mutate remote state
- `ghx` should prefer repo-owned workflow definitions over ad hoc local scripts when the repo has a declared workflow
- workflow runs should produce audit-friendly output and stable references for later review
- workflow orchestration should work in noninteractive contexts first

Relationship between `ghx` and `gh aw`:

- `gh aw` provides the durable workflow engine and repository-owned behavior
- `ghx` provides account binding, cwd and repo intent, config discovery, and reporting
- `ghx` should not replace `gh aw`
- `ghx` should make `gh aw` safer and easier to inspect

### 6. `.ghx/` config drives project-local defaults

The accepted `.ghx/` config directory should be the place where repo-local workflow posture is declared. This RFC uses that directory for defaults, not for hidden authority.

Relevant config responsibilities:

- workflow presets
- extension bundle posture
- progress log defaults
- proof artifact locations
- aliases and reusable snippets

Config rules:

- `.ghx/config.yml` and `.ghx/local.yml` may influence workflow defaults
- `.ghx/local.yml` should remain the personal override layer
- config should be explainable in output before mutation
- config should not be a secret store
- trusted hooks should remain disabled until the repo is explicitly trusted

### 7. Noninteractive behavior is the default contract

Human convenience is allowed, but the core behavior must work without a TTY.

Noninteractive rules:

- do not prompt unless the user explicitly asked for an interactive flow and a TTY is available
- return machine-readable results on success whenever possible
- return a clear error with next steps when a selection or prompt would otherwise be required
- keep `--json`, `--explain`, and `--progress-log` usable in CI and in agent runs
- avoid depending on `fzf` or `gum` for the only available path through a command

The default expectation is that a command can be run by a human, a bot, or an agent without changing the command shape.

## Acceptance Criteria

- `ghx issue subissue` commands can list, add, remove, and reprioritize subissues with stable JSON output.
- Long-running commands emit progress at the start, after durable writes, before remote mutation, on blockers, and on completion.
- Proof artifacts can be attached with unique basenames and the resulting references are exposed in text and JSON output.
- `jq`, `rg`, `fzf`, `delta`, `gum`, and `yq` remain optional compatibility tools, not mandatory runtime dependencies.
- `gh aw` workflows can be explained and executed with repo-local defaults from `.ghx/`.
- `.ghx/config.yml` and `.ghx/local.yml` can influence workflow defaults without changing account binding.
- Noninteractive invocations never hang on prompts and always produce either a complete machine-readable result or an actionable error.
- Every new workflow surface has docs, tests, and at least one issue tree entry linking the rollout work.

## Rollout Plan

1. Add or refine the issue tree commands and JSON shapes first.
2. Wire the shared progress contract into the issue tree and workflow commands.
3. Make proof artifact handling and `gh attach` integration deterministic and noninteractive.
4. Add companion-tool discovery and compatibility docs without hard dependencies.
5. Wire `.ghx/workflows.yml` and `.ghx/config.yml` into `gh aw` backed workflows.
6. Convert the RFC into issue tree epics and subissues, then ship slice by slice.
7. Track rollout with progress logs and attach proof artifacts for user-visible steps.

## Open Questions

- Should progress cadence include a timer-based heartbeat in addition to phase-based events?
- Should `ghx tools doctor` report repo-local tool requirements from `.ghx/tools.toml` or also from workflow manifests?
- Should proof artifact attachment be handled by a dedicated `ghx attach` wrapper or only by pass-through integration with the extension?
