# 0003 ghx upstream architecture and command map

Status: draft
Date: 2026-05-17

## Purpose

This note records the current `gh` command, API, docs, and extension shape that `ghx` should keep aligned with while adding first-class issue subissue support.

The core rule is simple: keep `ghx` additive and upstream-compatible. New issue-tree behavior should live in the smallest possible set of files, reuse shared lookup and export helpers, and let the existing help and doc generation pipeline continue to work without fork-specific scaffolding.

## Proven repository shape

- The issue command root already owns the new subissue tree. `pkg/cmd/issue/issue.go:24` imports `pkg/cmd/issue/subissue`, and `pkg/cmd/issue/issue.go:46` registers it in the "General commands" group next to list, create, and status.
- `gh issue create` already has a parent-issue path. `pkg/cmd/issue/create/create.go:25` adds `Parent` to `CreateOptions`, `pkg/cmd/issue/create/create.go:136` adds `--parent`, and `pkg/cmd/issue/create/create.go:376` passes `parentIssueId` into `api.IssueCreate`.
- The new subissue package is a sibling command tree, not a rewrite of `issue`. `pkg/cmd/issue/subissue/subissue.go:18` defines the exported JSON fields, `pkg/cmd/issue/subissue/subissue.go:70` builds the command tree, and `pkg/cmd/issue/subissue/subissue.go:96`, `pkg/cmd/issue/subissue/subissue.go:120`, `pkg/cmd/issue/subissue/subissue.go:140`, and `pkg/cmd/issue/subissue/subissue.go:160` split list, add, remove, and reprioritize into distinct commands.
- Cobra group wiring is centralized. `pkg/cmdutil/cmdgroup.go:5` adds grouped commands while preserving a single registration path, which is why `issue` and other top-level command trees can stay consistent across help, manpages, and generated docs.
- Shared issue lookup stays central. `pkg/cmd/issue/shared/lookup.go:56` parses issue arguments, `pkg/cmd/issue/shared/lookup.go:136` resolves issue or pull request records, and `pkg/cmd/issue/shared/display.go:16` handles standard issue table output.
- The API layer is still the main boundary for GitHub operations. `api/client.go:57` and `api/client.go:104` expose GraphQL and REST through a single client, and `api/queries_issue.go:273` owns the `IssueCreate` mutation contract while `api/queries_issue.go:344` owns the issue status query.
- `IssueCreate` now whitelists `parentIssueId`. `api/queries_issue.go:289` is the relevant acceptance point for the new parent relationship.
- The JSON export pattern remains command-local but annotation-driven. `pkg/cmdutil/json_flags.go:26` wires `--json`, `--jq`, and `--template`, and `pkg/cmdutil/json_flags.go:118` writes `help:json-fields` into the command annotations.
- Help and manpage generation read those same annotations. `internal/docs/markdown.go:178` and `internal/docs/man.go:178` both recurse the live cobra tree and render JSON sections from `help:json-fields`.
- The doc generation entry point uses the root command, not a hand-built docs tree. `cmd/gen-docs/main.go:56` creates the command root with stubbed browser and extension manager implementations, then `cmd/gen-docs/main.go:65` and `cmd/gen-docs/main.go:71` generate markdown and man pages from that same tree.
- Extension registration is owned at the root command layer. `pkg/cmd/root/root.go:196` registers installed extensions, `pkg/cmd/root/root.go:200` avoids conflicts with core command names, and `pkg/cmd/root/root.go:244` adds official extension stubs after real extensions and aliases.
- The extension manager itself is wired from the factory. `pkg/cmd/factory/default.go:42` creates `ExtensionManager`, and `pkg/cmd/factory/default.go:261` shows that it is built from the configured HTTP client and cached transport.

## What this means for ghx

1. Keep issue-tree behavior in `pkg/cmd/issue/subissue/` and only add narrow glue to `pkg/cmd/issue/issue.go` and `pkg/cmd/issue/create/create.go`.
2. Keep all issue references going through `pkg/cmd/issue/shared/lookup.go` so number or URL parsing stays consistent across create, status, view, and subissue flows.
3. Keep JSON outputs flat and explicit. `subIssueFields` should remain the source of truth for `--json`, and the exported shape should stay small unless a new field is proven to need nesting.
4. Keep API changes in `api/queries_issue.go` and `api/client.go` rather than inventing a fork-only transport layer.
5. Let docs and help continue to derive from command annotations and live cobra registration. That is what keeps `ghx` rebases boring.
6. Keep extension compatibility intact by preserving the current root registration order and conflict checks. `ghx` should not consume names that extensions or official extension stubs can reasonably own.

## JSON and export conventions

The current codebase already shows two useful patterns for issue-shaped exports.

- Plain issue exports use the shared `api.Issue` export path. `pkg/cmd/issue/view/view.go:87`, `pkg/cmd/issue/list/list.go:116`, and `pkg/cmd/issue/status/status.go:49` all call `cmdutil.AddJSONFlags` with `api.IssueFields`.
- Nested or special-case fields use custom export shaping. `api/export_pr.go:8` shows how issue and pull request exports can special-case nested structures when the plain struct shape is not enough.
- Subissues currently take the minimal approach. `pkg/cmd/issue/subissue/subissue.go:20` defines a compact `SubIssue` type, `pkg/cmd/issue/subissue/subissue.go:30` uses `cmdutil.StructExportData`, and `pkg/cmd/issue/subissue/subissue.go:35` wraps `repository` in a custom string type so the export stays readable while GraphQL can unmarshal either a raw string or `nameWithOwner`.

That is the right default for ghx: prefer the smallest export shape that satisfies scripting and docs, and only add custom export code when a nested shape truly buys something.

## Tests and mocks

- Subissue coverage already exercises the command tree and the main GraphQL paths. `pkg/cmd/issue/subissue/subissue_test.go:18` validates command parsing, `pkg/cmd/issue/subissue/subissue_test.go:66` covers list output, `pkg/cmd/issue/subissue/subissue_test.go:99` covers JSON output, `pkg/cmd/issue/subissue/subissue_test.go:134` covers add, and `pkg/cmd/issue/subissue/subissue_test.go:158` covers reprioritize.
- Parent-issue creation has an explicit regression test. `pkg/cmd/issue/create/create_test.go:773` proves that `--parent` resolves to `parentIssueId`, and `pkg/cmd/issue/create/create_test.go:793` checks the actual mutation inputs.
- The command tests rely on the standard HTTP mock and command injection style rather than bespoke harnesses. That keeps the new subissue package easy to compare with existing issue, PR, and repo commands.
- The extension layer already has generated mocks. `pkg/extensions/extension.go:17` and `pkg/extensions/extension.go:31` define the `moq` generation points, and `pkg/extensions/manager_mock.go:16` provides the manager mock used by the root and extension command tests.

## Rebase-safe guidance

The easiest way to keep rebases manageable is to keep the ghx delta narrow and local.

- Treat `pkg/cmd/issue/issue.go`, `pkg/cmd/issue/create/create.go`, `pkg/cmd/issue/subissue/*`, and `api/queries_issue.go` as the only high-value product files for subissue support.
- Avoid changing the command framework unless the upstream tree forces it. `cmdutil.AddGroup`, JSON annotations, and docs generation already provide the hook points ghx needs.
- Avoid fork-only output formatting. If a new command needs JSON, wire it into the standard `cmdutil.AddJSONFlags` path so help, manpages, and scripts stay consistent.
- Avoid extension name churn. Root command registration intentionally keeps real extensions, aliases, and official stubs in a specific order, so new core commands should fit that order instead of bypassing it.
- If upstream `gh` later lands first-class subissues, the safest cleanup path is to shrink or delete `pkg/cmd/issue/subissue/` rather than unwind a wider architectural fork.

## Likely implications

- `ghx issue create --parent` should stay the single creation path for attaching a new child issue to a parent.
- `ghx issue subissue` should stay the tree-management path for list, add, remove, and reprioritize.
- Shared lookup and API code should remain the place where repo identity, host selection, and issue resolution logic live.
- Documentation should stay generated from the command tree so that `ghx` and upstream `gh` do not drift in help layout unless the command tree itself changes.

## Open questions

- Will upstream keep the same GraphQL names for subissues and reprioritization, or will `ghx` need a tiny adapter if the schema shifts?
- Will `parentIssueId` remain the stable input name for issue creation, or should ghx be ready to map a future rename?
- Should any future tree operations live only in `gh issue subissue`, or should some of them also surface as shortcuts in `gh issue create` and `gh issue view` help text?

## Evidence used

- `pkg/cmd/issue/issue.go:24`
- `pkg/cmd/issue/issue.go:46`
- `pkg/cmd/issue/create/create.go:25`
- `pkg/cmd/issue/create/create.go:136`
- `pkg/cmd/issue/create/create.go:376`
- `pkg/cmdutil/cmdgroup.go:5`
- `pkg/cmd/issue/subissue/subissue.go:18`
- `pkg/cmd/issue/subissue/subissue.go:20`
- `pkg/cmd/issue/subissue/subissue.go:30`
- `pkg/cmd/issue/subissue/subissue.go:35`
- `pkg/cmd/issue/subissue/subissue.go:70`
- `pkg/cmd/issue/shared/lookup.go:56`
- `pkg/cmd/issue/shared/lookup.go:136`
- `pkg/cmd/issue/shared/display.go:16`
- `pkg/cmd/issue/view/view.go:87`
- `pkg/cmd/issue/list/list.go:116`
- `pkg/cmd/issue/status/status.go:49`
- `api/client.go:57`
- `api/client.go:104`
- `api/queries_issue.go:273`
- `api/queries_issue.go:289`
- `api/queries_issue.go:344`
- `api/export_pr.go:8`
- `pkg/cmdutil/json_flags.go:26`
- `pkg/cmdutil/json_flags.go:118`
- `internal/docs/markdown.go:178`
- `internal/docs/man.go:178`
- `cmd/gen-docs/main.go:56`
- `cmd/gen-docs/main.go:65`
- `cmd/gen-docs/main.go:71`
- `pkg/cmd/root/root.go:196`
- `pkg/cmd/root/root.go:200`
- `pkg/cmd/root/root.go:244`
- `pkg/cmd/factory/default.go:42`
- `pkg/cmd/factory/default.go:261`
- `pkg/cmd/issue/subissue/subissue_test.go:18`
- `pkg/cmd/issue/subissue/subissue_test.go:66`
- `pkg/cmd/issue/subissue/subissue_test.go:99`
- `pkg/cmd/issue/subissue/subissue_test.go:134`
- `pkg/cmd/issue/subissue/subissue_test.go:158`
- `pkg/cmd/issue/create/create_test.go:773`
- `pkg/cmd/issue/create/create_test.go:793`
- `pkg/extensions/extension.go:17`
- `pkg/extensions/extension.go:31`
- `pkg/extensions/manager_mock.go:16`
