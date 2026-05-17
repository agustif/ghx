# ghx extension bundle

Status: draft
Date: 2026-05-17
Related ADR: [ADR 0005](adr/0005-extension-bundles-and-wrappers.md)

This file is the human-reviewed catalog for extensions that `ghx` should bundle, wrap, mine, or ignore.

## Bundle commands

- `ghx ext bundle list`
- `ghx ext bundle install <bundle>`
- `ghx ext bundle audit`
- `ghx ext wrap <extension>`

## Companion tool compatibility

`ghx` should treat Unix-style companion tools as a first-class compatibility surface before it tries to manage installs. The core rule is that every automation-oriented command should have stable JSON fields, predictable stdout/stderr, and pipe-friendly behavior so agents and humans can compose it with existing tools.

Initial companion set:

- `jq`: JSON filtering and scripted assertions. `ghx` already embeds `--jq`, but command JSON contracts must stay stable.
- `rg`: fast source, docs, logs, and downloaded artifact search. Prefer examples that compose `ghx ... --json` with `rg` only when text search is actually the right primitive.
- `fzf`: human selection over issues, PRs, runs, checks, and generated operation ids.
- `delta`: readable diffs for PR patches, generated code reviews, and ruleset/config diffs.
- `gum`: optional interactive prompts for local scripts, never required for noninteractive command paths.
- `yq`: YAML workflows, Actions manifests, and repo config inspection.

Future native surface:

- `ghx tools doctor`: show missing companion tools, detected versions, PATH source, and project requirements.
- `ghx tools install <name>`: install a curated companion without shadowing existing tools silently.
- `ghx tools path`: print the resolved binary path for each companion tool.
- `.ghx/tools.toml`: optional project-pinned companion requirements, separate from account binding.

Do not hard-bundle or shadow user tools by default. The safe path is discovery and opt-in installation, with clear output about what will be used.

## Initial catalog

| Extension | Owner | Posture | Notes |
| --- | --- | --- | --- |
| `gh aw` | `github/gh-aw` | first-class companion | Use for durable repo-owned agent workflows in GitHub Actions. |
| `gh stack` | `github/gh-stack` | first-class companion | Use for stacked PR workflows before rebuilding stack mechanics. |
| `gh attach` | `enthus-appdev/gh-attach` | wrap and watch | Useful for screenshot-backed PR/issue evidence. Requires unique basenames for large screenshot sets. |
| `gh dash` | `dlvhdr/gh-dash` | optional bundle | Mature TUI for humans; do not treat as machine-readable source of truth. |
| `gh pr-review` | `agynio/gh-pr-review` | spike | Mine unresolved-thread behavior before building `ghx pr threads`. |
| `gh actions-cache` | `actions/gh-actions-cache` | wrap or supersede | Compare against `ghx cache top` and storage reports. |
| `gh actions-importer` | `github/gh-actions-importer` | optional bundle | Useful for migration-heavy repos. |

## Audit metadata

Each catalog entry should eventually include:

- pinned version
- install command
- license
- update age
- required scopes
- supported output formats
- noninteractive support
- wrapper required
- risk notes

## Wrapper requirements

Wrappers should preserve:

- `.ghaccount`
- `GH_ACCOUNT`
- `GH_ACCOUNT_SESSION`
- `GH_HOST`
- `--repo`
- `--json` when the extension supports it
- `--progress-log` when `ghx` adds the wrapper behavior
