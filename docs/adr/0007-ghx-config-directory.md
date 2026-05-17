# ADR 0007: `.ghx/` config directory

Status: Accepted
Date: 2026-05-17

## Context

The zsh ecosystem works well because users can compose local config, plugins, aliases, themes, completions, hooks, and project-specific behavior without patching the shell itself. `ghx` needs a similar user-owned configuration surface for GitHub workflows, but with stricter safety because it can mutate remote repositories.

The existing `.ghaccount` file should stay as the smallest possible account binding. A richer project configuration needs namespacing and room to grow.

## Decision

`ghx` will support a layered `.ghx/` directory for repo-local or folder-local configuration.

Initial layout:

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

`config.yml` is the main project config. `local.yml` is for uncommitted personal overrides and should be ignored by default. `.ghaccount` remains supported as a simple shortcut that can be represented inside `.ghx/accounts.yml` later.

Precedence:

1. command flags
2. explicit env such as `GH_ACCOUNT` and `GH_ACCOUNT_SESSION`
3. nearest `.ghx/local.yml`
4. nearest `.ghx/config.yml`
5. nearest `.ghaccount`
6. cwd-scoped config
7. global `ghx` config
8. upstream `gh` host-global config

## Consequences

- Config loading must be explainable through `ghx ctx explain`.
- Config files must never contain tokens or secrets.
- Hooks must not run by default from untrusted repos.
- Extension bundles can be declared per repo without hiding provenance.
- Agents can read project intent from `.ghx/` before running workflows.

## Follow-ups

- Add `ghx config doctor`.
- Add `ghx config trust` for hook-enabled repos.
- Add `ghx config explain --json`.
- Add schema validation for `.ghx/*.yml`.
- Add starter templates for common repo profiles.
