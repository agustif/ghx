# Research 0005: ghx CLI UX and human factors

Status: draft
Date: 2026-05-17
Related docs: [ghx index](../ghx.md), [Agent workflows](../ghx-agent-workflows.md), [Extension bundle](../ghx-extension-bundle.md), [First delivery slices](../plans/first-delivery-slices.md), [ADR 0006](../adr/0006-agent-progress-contract.md), [ADR 0007](../adr/0007-ghx-config-directory.md), [Research 0002](0002-ghx-identity-security-and-trust-boundaries.md), [Research 0003](0003-ghx-upstream-architecture-and-command-map.md)

## Why this note exists

`ghx` is adding account-safe, agent-ready GitHub workflows on top of the upstream `gh` command model. First-class issue subissue support is a useful pressure test because it is both simple and risky: a user may only want to list a tree, but the same command family can also add, remove, or reorder work items that coordinate humans, agents, and product plans.

Good CLI UX for `ghx` is not only about short flags or pretty output. It is about making the right target obvious, making dangerous actions previewable, giving humans readable status, giving agents stable contracts, and keeping the tool unsurprising inside shells, scripts, editors, CI, and companion pipelines.

This note records UX and human-factors guidance for subissues and future `ghx` workflows before the product surface hardens.

## Product posture

The core product promise is an account-safe GitHub control plane. Command ergonomics should serve that promise before optimizing for terseness.

The PM rule from the main docs is the right default: ship one useful read-only slice, prove identity, dry-run, JSON, and audit behavior, then add mutation. That sequencing should shape every command family:

- read paths come first and should be fast, obvious, and scriptable
- mutating paths should explain identity, host, repo, and resource targets before they perform remote changes
- raw API escape hatches remain available, but native commands should make common safe paths easier than raw calls
- issue trees should turn product intent into executable work, so the command surface must be ergonomic for both human planning and agent task routing

For subissues, this means `ghx issue subissue list` is more than a data fetch. It is a roadmap inspection primitive. `ghx issue create --parent`, `ghx issue subissue add`, `ghx issue subissue remove`, and `ghx issue subissue reprioritize` are coordination mutations and should carry stronger preflight expectations than ordinary text output suggests.

## Discoverability

Users should be able to find issue-tree behavior from the command names they already know.

Useful discovery paths:

- `ghx issue --help` should show `subissue` in the general command group near `list`, `create`, `view`, and `status`.
- `ghx issue create --help` should make `--parent` visible as the path for creating a new child under an existing issue.
- `ghx issue subissue --help` should show the lifecycle in task order: `list`, `add`, `remove`, `reprioritize`.
- Examples should include number inputs and URL inputs, since URLs are the safest way to disambiguate cross-repo cases.
- JSON examples should include the smallest useful field set for agents: `number,title,state,url,repository`.

Discoverability should not depend on users knowing GitHub's product vocabulary perfectly. Some users will search for "child issue", "issue tree", "task", "roadmap", "parent", or "dependency". The CLI can keep the canonical command as `subissue`, but help text and docs should include these alternate terms where they clarify intent.

Recommendation:

- Keep the canonical command `ghx issue subissue`.
- Keep the alias `subissues`.
- Consider `tree` or `children` only as help pathways or future aliases if usage data proves that `subissue` is blocking discovery.
- Use help text to mention "child issues" and "issue trees" without adding extra aliases unless usage proves they are needed.
- Prefer examples over prose for discovery because users copy commands from terminal help.

## Command naming

`subissue` is product-aligned with GitHub's current GraphQL vocabulary and keeps the command close to the upstream platform model. It is also slightly awkward for users who think in parent and child terms.

The current split is reasonable:

- `ghx issue create --parent <issue>` creates a new issue and attaches it under a parent.
- `ghx issue subissue list <issue>` reads the children of an existing issue.
- `ghx issue subissue add <parent> <child>` attaches an existing issue as a child.
- `ghx issue subissue remove <parent> <child>` detaches a child.
- `ghx issue subissue reprioritize <parent> <child> --before <sibling>` reorders a child.
- `ghx issue subissue reprioritize <parent> <child> --after <sibling>` reorders a child.

Human-factors risk:

- `add` can sound like "create a new issue" even though it attaches an existing issue.
- `remove` can sound destructive even though it detaches a relationship.
- `reprioritize` is precise, but long and less obvious than "move".

Recommendations:

- Keep `add` and `remove` if the help text says "Attach an existing issue as a subissue" and "Detach a subissue from its parent".
- Prefer `move` as the eventual human alias for `reprioritize`. Keep `reprioritize` as the precise API-aligned command, but let humans type `move` once the shared mutation result model and docs can support both names cleanly.
- Use argument names that encode relationships: `<parent>` and `<child>` are clearer than two generic issue numbers in help and errors.
- For mutation success messages, include the relationship and target repo: `Added subissue #12 to OWNER/REPO#34`.
- Avoid hidden cross-repo behavior. If parent and child are in different repos, text output should say so explicitly and JSON should include both repositories when supported.

## Explain and dry-run defaults

`--explain` and `--dry-run` are the main bridge between human confidence and agent safety.

`--explain` should answer:

- Which host will be used?
- Which account will be used?
- Which account source won the resolution chain?
- Which repo was inferred from cwd, flags, URL, or config?
- Which issue is the parent?
- Which issue is the child or sibling?
- Which GraphQL mutation or query will run?
- Whether the command is read-only or mutating?

`--dry-run` should answer:

- What remote mutation would be attempted?
- What resource ids or issue numbers were resolved?
- Whether the server supports the operation?
- What output would be emitted after success?
- Which parts could not be proven without performing the mutation?

Default posture:

- Read-only commands should not require `--dry-run`.
- Read-only commands should support `--explain` when identity, repo, or host resolution is implicit.
- Mutating commands should support `--dry-run` or an equivalent preview before they become first-class agent surfaces.
- Noninteractive mutating commands should be able to run without prompts only when the target is explicit enough and a confirmation flag or documented default makes the boundary clear.

For subissues, the strongest design is:

- `list`: normal execution by default, optional `--explain`.
- `add`, `remove`, `reprioritize`: support `--dry-run`, support `--explain`, and print the exact parent, child, sibling, host, account, and repo before mutation when a terminal is attached.
- `create --parent`: include the parent resolution in `--dry-run`, not only the issue title and body preview.

Open tradeoff:

- Making `--dry-run` the default for all new mutation could be safer for agents, but it may surprise humans who expect CLI verbs to execute. A better compromise is to require explicit preview support and make agent workflows default to preview in their own policy.

## Shared mutation result model

Mutation UX should not be rebuilt command by command. `add`, `remove`, `reprioritize`, future `move`, and `create --parent` should share one preflight and result model that can feed text, JSON, dry-run, explain output, and progress logs.

Suggested fields:

- `action`
- `parent`
- `child`
- `sibling`
- `repository`
- `host`
- `account`
- `accountSource`
- `operation`
- `dryRun`
- `state`
- `url`
- future: `position`

The text form can remain concise, but the JSON form should be the agent and companion-tool contract. `--dry-run --json` should return the same shape as a successful mutation with `dryRun: true`, resolved targets, and no remote state change.

This model should also carry enough information for better errors. If the child target is a pull request, the error should know it was resolving the child, not simply that a generic issue lookup failed.

## JSON and text duality

`ghx` needs two first-class output contracts:

- text for humans scanning a terminal
- JSON for agents, scripts, tests, and companion tools

Text output should optimize for immediate comprehension:

- use stable columns for lists
- keep stdout for data
- keep warnings, explanations, prompts, and progress on stderr
- show enough target context to catch wrong-account or wrong-repo mistakes
- avoid terminal-only decoration that becomes noisy in logs

JSON output should optimize for stability:

- field names must be explicit and documented through the standard `--json` help path
- flat fields are preferable until nesting buys a real contract
- resource identifiers should include numbers and URLs where useful
- repository identity should be present for issue-tree data because subissue work can cross repo boundaries
- mutation preview output should have a structured shape, not a human sentence hidden in a string

Recommended subissue JSON fields:

- `number`
- `title`
- `state`
- `url`
- `repository`
- future: `parent`
- future: `position`
- future: `viewerCanUpdate`

The current list fields are a good minimum. Future mutation previews should not reuse the list shape directly because they need operation metadata, resolved targets, and safety state.

Suggested first-class examples:

```sh
ghx issue create --parent 123 --title "Child task" --body "Details" --dry-run --json
ghx issue subissue list 123 --json number,title,state,url,repository
ghx issue subissue add 123 456 --dry-run --explain
ghx issue subissue add 123 456 --json action,parent,child,repository,host,account
ghx issue subissue move 123 456 --before 455 --dry-run --json
ghx issue subissue remove 123 456 --yes --json
```

## Interactivity and noninteractive behavior

Interactive behavior should help humans avoid mistakes without making scripts brittle.

Interactivity should be opt-in for subissue mutation. The explicit two-argument form is already scriptable, so the CLI should not add prompts as the primary safety mechanism. Prompts are useful when a terminal user asks for selection or confirmation, but dry-run and explain are the durable safety surface.

TTY behavior:

- prompts are acceptable for risky mutation when a required value is missing
- previews can be formatted for scanning
- pagers can be used for long read-only output when consistent with upstream `gh`
- progress can be animated if it does not obscure the final result

Noninteractive behavior:

- never prompt
- fail with a clear error when required intent is missing
- use exit codes consistently
- keep stdout parseable
- send warnings and progress to stderr
- provide flags for every value that an interactive prompt can collect

Agent behavior should be stricter than human terminal behavior:

- prefer explicit repo, issue URL, account, or explained cwd binding
- prefer `--json` for reads
- prefer `--dry-run` before mutation
- require exact next action in errors
- never require terminal-specific features to understand progress

For subissue commands, URL inputs are the best noninteractive affordance because they carry host, owner, repo, and number in one argument.

Possible opt-in interactive forms:

```sh
ghx issue subissue add 123 --interactive
ghx issue subissue move 123 456 --interactive
```

If an interactive selector uses `fzf`, `gum`, or another companion tool, that dependency should be explicit and optional. Noninteractive behavior must never depend on those tools being installed.

## Error messages

Errors should describe the failed intent, the resolved context, and the next action.

Weak error shape:

```text
not found
```

Better error shape:

```text
issue OWNER/REPO#123 was not found on github.com; pass an issue URL or use --repo OWNER/REPO
```

Subissue-specific errors should distinguish:

- parent issue not found
- child issue not found
- sibling issue not found
- an argument resolved to a pull request, not an issue
- issues are disabled for the repository
- account lacks permission to update the relationship
- GraphQL schema or GHES host does not support subissues
- `--before` and `--after` were both set
- neither `--before` nor `--after` was set

Recommendations:

- Use relationship words in errors: parent, child, sibling.
- Include the repo and host when known.
- Include the input source when helpful: cwd, `--repo`, URL, `.ghaccount`, or env.
- For flag errors, keep usage output aligned with upstream conventions.
- For permission errors, suggest `ghx ctx explain` or `ghx auth status` before asking the user to retry.

Example target-rich error:

```text
child target OWNER/REPO#44 is a pull request; subissues require issues. Inspect with: ghx pr view 44 -R OWNER/REPO
```

## Progress bars and long-running feedback

Most subissue operations are short. They should not need progress bars. But issue-tree workflows can become long-running when agents create, attach, reorder, or audit many subissues.

Guidance:

- single API operations should print one success line or one structured JSON result
- batch operations should support `--progress-log`
- watch-style commands should support `--watch` only when the state can change after the initial read
- progress should name the current resource and operation, not only show a spinner
- progress output belongs on stderr
- progress logs should be append-friendly and machine-readable

Progress bars are useful when they reveal bounded work. They are harmful when they hide which repo or issue is being mutated. Prefer phase lines for uncertain API work and percentage bars only when the total is known.

For subissue batch work, progress events should be phase events, not spinners:

- `start`
- `resolved-parent`
- `resolved-child`
- `resolved-sibling`
- `preflight`
- `mutated`
- `done`
- `blocked`

These can be represented inside the accepted progress contract as status and note events while preserving ASCII task trees and `0/100` through `100/100` progress.

## Icons, Unicode, and terminal decoration

`ghx` should be conservative with icons and Unicode.

Reasons:

- logs, CI, old terminals, screen readers, and pasted issue comments may not render glyphs reliably
- icons can hide meaning from users who do not know the visual convention
- agent transcripts and JSON pipelines should not depend on terminal decoration
- this repo currently prefers ASCII in docs and code

Recommendations:

- Default command output should be meaningful without icons.
- Use words such as `PASS`, `FAIL`, `PENDING`, `READ`, and `MUTATE` when a status needs to survive logs.
- If color or icons are later added, respect `NO_COLOR`, terminal capability checks, and upstream `gh` color behavior.
- JSON should never contain display-only icons.
- Docs and generated examples should stay ASCII unless a specific file has a reason to do otherwise.

Plain tree rendering should remain the safe baseline:

```text
#123 Parent issue
  |- #124 Child issue
  |- #125 Child issue
```

## Companion tools

The product docs already set the right compatibility-first stance. `ghx` should be easy to compose with existing tools before it manages installation.

Companion implications:

- `jq` needs stable JSON fields and predictable null behavior.
- `rg` needs output that can be searched without progress noise on stdout.
- `fzf` needs concise text lists plus enough hidden or adjacent data to select the right repo and issue.
- `delta` helps with diffs, but command output should not require it.
- `gum` can make scripts friendlier, but noninteractive paths must never require it.
- `yq` matters for `.ghx/` config, workflow files, and Actions manifests.

Future `ghx tools doctor` should explain:

- which companion tools are available
- which binary path will be used
- which version was detected
- whether a project pins a tool in `.ghx/tools.toml`
- whether the tool will affect output or only optional workflows

For subissues, useful examples should compose naturally:

```sh
ghx issue subissue list 123 --json number,title,state,url,repository --jq '.[] | select(.state == "OPEN")'
```

## Accessibility

Accessibility is a CLI concern.

Guidance:

- do not encode state only through color
- do not encode state only through icons
- keep text columns readable when titles are long
- provide JSON for users who rely on custom renderers or assistive tooling
- avoid progress animations that continuously rewrite important text
- keep prompts short and explicit
- make confirmation prompts include the resource that will change
- support `NO_COLOR` and upstream terminal-color behavior
- ensure examples are copyable without hidden formatting

Subissue list output should be readable in narrow terminals. If titles are too long, truncation should not remove the issue number, state, repo, or URL in contexts where those fields prevent wrong-target mistakes.

## Learning from zsh-style configuration

The user's shell environment shows a useful pattern: small files with clear responsibilities, fast startup, lazy loading, explicit PATH behavior, and discoverable tool state.

`ghx` config should learn from that rather than create one opaque global settings blob.

Relevant lessons:

- Keep identity binding separate from workflow defaults. `.ghaccount` should remain simple and auditable.
- Keep project config in `.ghx/` layered files with clear precedence.
- Make resolution explainable, like a shell can show which command path won.
- Lazy-load expensive checks. A prompt or help command should not trigger slow network work unless requested.
- Do not silently shadow system tools. Explain PATH and wrapper decisions.
- Keep local overrides local. Repo-level config should not mutate host-global auth state as a side effect.
- Provide doctor commands that show source, value, and precedence for each setting.

For subissues, `.ghx/config.yml` can safely express defaults such as preferred tree fields, progress-log path, text layout, templates, or snippets. It should not silently change account, host, repo, parent, child, or mutation target.

Command ergonomics should follow the same model. Users should be able to ask "why did ghx choose this account, repo, host, tool, or output mode?" and get a direct answer without reading implementation code.

## PM and product vision influence on ergonomics

PM decisions should shape command ergonomics early, not after implementation.

Questions product should answer before a command hardens:

- Is this command primarily for diagnosis, action, or coordination?
- Is the main user a human, an agent, or both?
- Is the command safe to run from cwd inference, or should URLs be preferred?
- Does success mean "remote state changed" or "a clear answer was produced"?
- What exact JSON fields need to be stable for scripts?
- What is the smallest read-only slice that proves the workflow?
- What mutation preview would make a wrong-account or wrong-repo mistake obvious?
- Which companion tools should examples encourage?
- Which errors should teach the user the next correct command?

For subissues, the PM framing should be "issue trees are the roadmap and work queue." That framing leads to different ergonomics than a generic graph mutation:

- list output should support planning and triage
- mutation output should support audit and collaboration
- errors should preserve the user's intended parent-child relationship
- examples should show roadmap maintenance, not only API capability
- JSON should be stable enough for agents to assign, verify, and report work

## Recommended command ergonomics for subissues

Short-term recommendations:

- Keep `ghx issue subissue list` as the read-only anchor.
- Keep `ghx issue create --parent` as the new-child creation path.
- Update help text to clarify that `add` attaches an existing issue and `remove` detaches the relationship.
- Add `--dry-run` and `--explain` before treating mutation commands as agent-ready.
- Include parent, child, sibling, repo, host, and account in mutation preview output.
- Add JSON support to mutation commands through a shared result model.
- Keep list JSON small and stable.
- Keep stdout parseable and send status messages to stderr.
- Prefer URL examples for cross-repo or agent workflows.

Medium-term recommendations:

- Add a structured preview schema shared by mutating commands.
- Add `ghx ctx explain --json` and reuse it in mutation preflight.
- Add `ghx tools doctor` before managed companion installs.
- Add issue-tree examples to agent workflow docs after the behavior is stable.
- Add a `move` alias for `reprioritize` once the command contract is stable.

Highest-value next slice:

- Implement a shared subissue preflight and result model first, then wire it into `add`, `remove`, `reprioritize`, `move`, and later `create --parent`. That model should power text output, JSON output, dry-run, explain, progress-log events, and better errors without each command inventing its own UX.

## Open questions

- Should agent mode require `--dry-run` before any subissue mutation, or should that remain policy outside the CLI?
- Should cross-repo subissue mutation require URL inputs for both parent and child?
- Should mutation success output include account and host by default, or only when resolution was implicit?
- Should `ghx issue subissue list` include position metadata once the API exposes it cheaply?
- Should `ghx issue create --parent` print a tree-friendly success object by default when `--json` is requested?
- Should `ghx` introduce a global `--explain` convention or keep explain commands scoped under `ghx ctx` and `ghx api`?
