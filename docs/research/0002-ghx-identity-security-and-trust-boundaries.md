# Research 0002: ghx identity, security, and trust boundaries

Status: draft
Date: 2026-05-17
Related docs: [ghx index](../ghx.md), [Multiple accounts](../multiple-accounts.md), [ADR 0001](../adr/0001-scoped-account-binding.md), [ADR 0002](../adr/0002-generated-validated-api-proxies.md), [ADR 0004](../adr/0004-gh-aw-workflow-companion.md), [ADR 0005](../adr/0005-extension-bundles-and-wrappers.md), [ADR 0007](../adr/0007-ghx-config-directory.md)

## Why this note exists

`ghx` is trying to make GitHub mutation safe for users and agents who operate across many accounts, hosts, repositories, shells, and helper tools. That makes identity resolution the core security boundary, not a convenience feature.

Upstream `gh` already has strong primitives such as keychain-backed auth, explicit token environment variables, `--repo`, and raw `api` access. The `ghx` fork adds more power by introducing scoped account binding, repo-local config, wrappers, generated API proxies, and workflow companions. Each new layer can reduce accidental mutation, but each new layer can also hide which account, token, host, or repo will be used if the resolution chain is not visible.

This note records the main trust boundaries, failure modes, and guardrails that should shape `ghx` command design before more mutation-heavy surfaces ship.

## Boundary inventory

| Boundary | Assets at risk | Main failure mode |
| --- | --- | --- |
| explicit env such as `GH_TOKEN`, `GITHUB_TOKEN`, `GH_HOST`, `GH_ACCOUNT`, `GH_ACCOUNT_SESSION` | remote mutations, token authority, host targeting | a stronger env override silently defeats local account intent |
| host-global `gh` auth state and keyring entries | default login, token scopes, git credential behavior | one shell changes the active account and another shell mutates with it |
| local scope selectors such as `.ghaccount` and cwd binding | repo-local account intent | a stale or misplaced selector applies to the wrong subtree |
| layered `.ghx/` config and future hooks | per-repo behavior, wrappers, defaults, templates | untrusted repo config changes mutation behavior or runs code |
| repo and host resolution inputs such as `--repo`, git remotes, cwd, hostname flags | owner/repo target, API host, web host | commands resolve the wrong repository or wrong GitHub host |
| generated REST and GraphQL proxy layers | request validation, discoverability, safety defaults | users assume generated coverage means a command is safe when it still falls through to raw mutation |
| raw escape hatches such as `ghx api`, custom headers, preview media types, unchecked bodies | full platform surface, unsupported operations | validation and explain layers are bypassed without a durable signal |
| extension wrappers and companion tools | execution provenance, stdout/stderr contracts, account context | a wrapped tool drops account context or leaks secrets in output |
| workflow companions such as `gh aw` | durable repo-owned automation, Actions permissions, logs | local identity safety is lost when the action runs remotely with broader credentials |
| proof artifacts such as logs, progress logs, JSON, screenshots, and attachments | tokens, repo names, private paths, issue data | operators upload or paste secrets while proving what happened |
| autonomous agents | remote state, issue trees, workflow runs, comments, attachments | the agent mutates the wrong account, host, repo, or issue without a clear preflight |

## Core threat model

The main attacker is often not an external intruder. It is an operator or agent with valid credentials acting on the wrong target because context was ambiguous, stale, or hidden.

Important classes:

- wrong-account mutation: the token is valid, but it belongs to the wrong person, bot, or employer
- wrong-host mutation: `github.com` and GHES or staging hosts are confused
- wrong-repo mutation: local cwd or remote inference points at the wrong owner or fork
- over-broad authority: a token has `repo`, `workflow`, `admin:org`, or similar scopes that exceed the intended action
- silent fallback: the safe path fails and the command silently falls back to host-global auth or raw API behavior
- untrusted local policy: repo-local config or hooks alter command behavior without trust establishment
- evidence leakage: progress output, debug logs, screenshots, or attachments disclose secrets or private paths
- autonomous escalation: agent mode converts a diagnosis flow into a mutating flow without an explicit boundary crossing

## Multi-account safety

Scoped account binding is the primary safety feature because upstream host-global switching is not sufficient for multi-repo, multi-employer work.

Observed resolution chain:

1. explicit command flags when available
2. `GH_ACCOUNT`
3. `GH_ACCOUNT_SESSION`
4. nearest `.ghaccount`
5. cwd-scoped config
6. host-global active account

Research conclusions:

- The chain is reasonable only if every mutating command can explain which step won.
- The docs do not yet describe a single final precedence contract. ADR 0001 documents the current scoped-account order, while ADR 0007 adds `.ghx/local.yml` and `.ghx/config.yml` above `.ghaccount`. Until there is one canonical chain, precedence drift is itself a trust-boundary risk.
- Explicit token environment variables remain stronger than account selectors. That means a shell can present one account in `.ghaccount` and still authenticate as another account if `GH_TOKEN` or `GITHUB_TOKEN` is set.
- The danger is not only a wrong login. The token may map to the right login on the wrong host, or to a login that lacks org SSO authorization for the target repo.
- Session scopes help reduce global churn, but they also create invisible state if the session selector is not shown in status output and preflight output.

Guardrails worth adopting:

- `ghx ctx explain --json` should become the mandatory machine-readable preflight for any mutating automation.
- Mutating commands should print or emit the selected login, source, host, owner/repo, and token scope summary before dispatch.
- When multiple accounts exist for a host and the command is about to mutate without an explicit account source, `ghx` should prefer refusal over guessing.
- A mismatch between account selectors and explicit token env should be surfaced as a warning or hard error in `--explain` output.

## Risks in `.ghaccount` and `.ghx/`

`.ghaccount` is intentionally small and local, but that does not make it harmless.

`.ghaccount` risks:

- nearest-parent search can accidentally pick up a file from a broader parent directory than the operator expects
- committed `.ghaccount` files can turn local machine state into repo policy by accident
- stale values can outlive a branch or worktree move and silently keep targeting the old account
- text-only account names do not prove host or token scope

`.ghx/` risks are broader because the directory can grow into a local policy engine:

- `config.yml` and `local.yml` can alter defaults for host, repo, wrappers, workflows, and future aliases
- hook support creates a code execution boundary inside arbitrary repositories
- shared templates or snippets can encode mutating defaults that look read-only at first glance
- personal overrides in `.ghx/local.yml` can diverge from the repo-visible config and make debugging hard

Guardrails worth adopting:

- `.ghaccount` should stay a local binding, not a collaboration artifact; docs should keep steering users toward ignore rules
- `ghx` should report the exact file path that supplied account or config state
- config files must never store tokens or secret material
- hooks must stay off by default for untrusted repos and require an explicit trust ceremony
- `ghx config explain --json` should show the merged config graph, not only the final values

## Token scopes and SSO

Tokens are the real authority boundary. Account names are only a hint until the token is known and the host accepts it.

Scope risks:

- the right login may still have the wrong scopes for the intended command
- `auth refresh` can widen scopes over time and surprise automation that assumed a smaller authority set
- `auth refresh` can also mint a token for the wrong browser-selected account even when `gh` refuses to store it, which is still a platform-side authority leak worth acknowledging
- a broad PAT or OAuth token can make accidental mutation much worse than a narrow token would
- different accounts on the same host can hold materially different scopes and org memberships

SSO risks:

- a token can authenticate successfully and still fail against an org that requires SSO authorization
- errors can look like generic authorization failures, which encourages operators to retry with broader credentials
- agents may interpret SSO errors as transient unless the CLI distinguishes "login present" from "org authorization missing"
- GHES device-flow login lacks the github.com account-switcher interstitial, so browser identity must already be correct before login or refresh flows begin

Guardrails worth adopting:

- preflight output should include a token scope summary and source, but never the token value
- mutating commands should fail early when required scopes are clearly absent
- `ghx ctx doctor` should detect likely SSO mismatches for the target owner when possible and explain the next action
- login and refresh flows should warn when the browser-selected identity can diverge from the selected local account, especially on GHES
- docs and command help should favor least-privilege tokens, especially for agent and workflow flows

## Repo and host resolution

Account safety is incomplete without target safety. The same token may have authority over many repositories and hosts.

Primary confusion points:

- `--repo` versus cwd remote inference
- API host versus web host
- forks versus upstream repos
- enterprise hosts that mirror `github.com` naming conventions
- extension or workflow commands that accept repo-like strings but resolve them differently than core `gh`

Guardrails worth adopting:

- `--explain` output should include both the selected host and the selected `owner/repo`
- commands should report the source of repo resolution, such as explicit flag, branch remote, or cwd default
- mutating commands should refuse ambiguous remote state instead of picking an arbitrary remote
- wrapper layers should preserve `GH_HOST` and `--repo` exactly and avoid rewriting them implicitly

## Extension wrappers and managed companion tools

Extensions and companion tools are part of the trust boundary because they execute code outside the core `ghx` tree.

Wrapper-specific risks:

- an extension may ignore `GH_ACCOUNT`, `.ghaccount`, or `GH_ACCOUNT_SESSION`
- stdout may not be stable or machine-readable, which makes agents misread outcomes
- extensions may prompt interactively or mutate without a dry-run mode
- unpinned versions can change behavior underneath an audited wrapper
- third-party extensions can require broader scopes than the narrow command they appear to expose

Managed companion tool risks:

- auto-installing `jq`, `rg`, `fzf`, `delta`, `gum`, or `yq` can silently shadow user binaries
- PATH order and wrapper shims can make it hard to know which binary actually ran
- a companion tool can receive sensitive stdin or env and leak it into shell history, temp files, or crash output

Guardrails worth adopting:

- `ghx` should stay compatibility-first: stable JSON and pipe-friendly output before managed installs
- wrappers should preserve resolved account, host, repo, and dry-run intent, then record that the call crossed an extension boundary
- mutating wrappers should require pinned versions and explicit audit metadata
- `ghx tools doctor` and `ghx tools path` should expose resolved paths, versions, and provenance
- automatic install or update should never happen on the hot path of a mutation command

## Generated API proxies and raw escape hatches

Generated proxies can reduce errors, but they do not erase the risk of raw API access.

Benefits:

- operation ids, typed params, enum validation, and pagination metadata make dangerous calls more inspectable
- generated coverage can expose which operations still have no safe command surface

Risks:

- users may mistake "generated" for "fully safe" even when an operation is inherently destructive
- custom headers, preview media types, unchecked validation, unknown query fields, and raw bodies can bypass the safe path
- generated code may drift from live behavior, especially around previews or newly added fields

Guardrails worth adopting:

- `ghx api explain <operation-id>` should show operation id, path, method, known scopes, and whether validation was bypassed
- raw or partially validated calls should be marked as such in explain output and audit logs
- escape hatches must remain available, but the CLI should make the downgrade from validated to raw obvious
- dry-run support for generated operations should prefer request previews over silent dispatch when possible

## Audit logs, dry-runs, and proof artifacts

The product direction already points toward `--dry-run`, `--progress-log`, and machine-readable output. Those are security controls as much as UX features.

Audit expectations:

- every mutating command should be able to emit a structured preflight record with account source, host, repo, command path, dry-run state, and wrapper or raw-api status
- durable logs should record operation ids, resource identifiers, and exit status, but never raw tokens, secret bodies, or auth headers
- wrapper and workflow calls should record provenance so operators know whether core `ghx`, an extension, or a remote workflow performed the mutation

Dry-run expectations:

- dry-run should show exactly what repo, host, account, and API operation would be used
- when a downstream tool has no real dry-run, the wrapper should say so instead of implying safety
- agents should default to dry-run or explain-first behavior for new command classes

Proof artifact risks:

- screenshots can capture tokens, private repo names, branch names, org URLs, filesystem paths, or pending review content
- progress logs can leak command lines, request bodies, or local paths
- JSON intended for machines can still land in chat, PR comments, or issues

Guardrails worth adopting:

- redact token values and auth headers everywhere by default
- keep stdout for data and stderr for human warnings so logs are easier to sanitize
- document screenshot hygiene, attachment naming, and secret scanning before upload
- treat debug logging as opt-in and clearly label it as potentially sensitive

## Agent mutation guardrails

Agent safety should be stricter than interactive human safety because the agent can move faster and at larger scale.

Recommended posture:

- default to read-only diagnosis until account, host, repo, and token scope are explicit
- require a machine-readable preflight before remote mutation
- require a clear mutation boundary crossing such as `--confirm`, `--yes`, or a durable workflow dispatch choice for noninteractive runs
- refuse to run mutating repo hooks from untrusted repositories
- prefer repo-owned workflows for high-risk or long-running automation so there is reviewable code and remote audit history
- make wrong-target risk visible in issue-tree and progress-log output, not only in terminal prompts

Suggested agent rules:

- do not infer account or repo when multiple plausible targets exist
- do not paste `auth status` output, screenshots, or debug output into issues without redaction
- do not upgrade from generated safe paths to raw `api` calls without an explicit "validation bypass" signal
- do not auto-install extensions or companion tools during a mutating workflow unless the user requested it

## Suggested follow-up decisions

This research note points toward a few likely next decisions:

1. Define the exact JSON contract for `ghx ctx explain` and `ghx config explain`.
2. Define a shared preflight and audit event schema for all mutating commands and wrappers.
3. Define trust and execution rules for `.ghx/hooks/*`, including how a repo becomes trusted.
4. Define wrapper requirements for extensions that do not support native dry-run or stable JSON.
5. Define how generated API operations report required scopes, validation state, and raw escape hatch usage.

## Open questions

- Should noninteractive mutating commands require an explicit account source when more than one account exists on the host?
- Should `ghx` hard-error when explicit token env conflicts with `.ghaccount` or only warn in `--explain` output?
- How much SSO status can `ghx ctx doctor` determine cheaply without turning every preflight into extra API traffic?
- Which precedence chain is canonical once `.ghx/local.yml` and `.ghx/config.yml` are live everywhere?
- Should wrapper audit events be local-only, repo-local under `.ghx/`, or exportable as issue or workflow artifacts?
- Which command classes should be forbidden entirely in agent mode unless they route through a durable workflow companion?
