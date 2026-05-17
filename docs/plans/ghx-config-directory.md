# Plan: `.ghx/` config directory

Status: active
Date: 2026-05-17
Related ADR: [ADR 0007](../adr/0007-ghx-config-directory.md)

## Goal

Give `ghx` a user-owned, project-local configuration surface inspired by the zsh ecosystem, without making remote mutation unsafe or surprising.

## Proposed layout

```text
.ghx/
|- config.yml
|- local.yml
|- accounts.yml
|- extensions.yml
|- workflows.yml
|- aliases.yml
|- progress.yml
|- hooks/
|  |- pre-mutate
|  |- post-progress
|- templates/
|- snippets/
```

## File responsibilities

- `config.yml`: main project config and includes.
- `local.yml`: personal overrides, ignored by default.
- `accounts.yml`: account binding rules that can supersede `.ghaccount` once stable.
- `extensions.yml`: extension bundles, pins, and wrapper posture.
- `workflows.yml`: `gh aw` and `gh workflow` defaults.
- `aliases.yml`: ghx-specific aliases and command presets.
- `progress.yml`: default progress log path and output preferences.
- `hooks/`: optional trusted executable hooks.
- `templates/`: reusable PR, issue, release, and workflow templates.
- `snippets/`: reusable GraphQL, jq, and command snippets.

## First slice

1. Add config discovery with nearest `.ghx/config.yml` and `.ghx/local.yml`.
2. Add schema structs and validation errors, but no hook execution.
3. Add `ghx config explain --json`.
4. Add `ghx config doctor`.
5. Teach `ghx ctx explain` to include config file sources.
6. Document `.gitignore` patterns for `.ghx/local.yml`.

## Safety model

- No secrets in `.ghx/`.
- Hooks require explicit trust.
- Mutating commands must show config-derived account/repo intent under `--explain`.
- Extension commands must show owner, repo, version, and wrapper before install or execution.
- Local overrides should be visible in explain output without printing secret-like values.

## Example

```yaml
profile: agent
account:
  login: agustif
extensions:
  bundles:
    - agent
    - review
progress:
  log: /Users/af/apple-re/src/kernel/progress.log
workflows:
  default_runner: github-hosted
```

## Acceptance checks

- `ghx config explain --json` lists loaded files in precedence order.
- invalid keys produce exact file and line diagnostics where available.
- `ghx ctx explain --json` includes config-derived account source.
- `.ghx/local.yml` is documented as the default uncommitted override.
- hook files are ignored until the repo is trusted.
