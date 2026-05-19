# Plan: ghx workflow automation audit

Status: active
Date: 2026-05-19
Related issues: [#57](https://github.com/agustif/ghx/issues/57), [#72](https://github.com/agustif/ghx/issues/72)
Parent plan: [First-class ghx release and automation migration](ghx-first-class-release-migration.md)

## Goal

Classify every checked-in GitHub Actions workflow before `ghx` becomes a first-class release. The audit is intentionally documentation-only: production deployment YAML stays unchanged until the fork-owned release path is designed and reviewed.

## Classification key

| Classification | Meaning |
| --- | --- |
| `unchanged upstream behavior` | Safe to keep as inherited `gh` behavior because it validates source, security, or build health without fork-specific publication or repo identity. |
| `fork-compatible` | Safe enough for the fork if labels, secrets, and repo settings are intentionally present in `agustif/ghx`. These workflows may still need owner review. |
| `fork-disabled` | Should stay off for the fork until a named downstream owner or channel exists. |
| `needs a ghx replacement` | Must be replaced, parameterized, or guarded before it can run as `ghx` automation. |

## Audit DAG

```text
workflow automation audit
|- done: inventory .github/workflows/*.yml
|- done: classify CI and security checks
|- done: classify release and package publication
|- done: classify issue and PR automation
|- next: implement fork-owned release workflow for #57
`- next: decide which #72 triage automations ghx actually wants enabled
```

## Workflow classifications

| Workflow | Classification | Evidence | Decision |
| --- | --- | --- | --- |
| `.github/workflows/go.yml` | `unchanged upstream behavior` | Pushes to `trunk` and PRs, uses `contents: read`, runs Go tests and builds `./cmd/gh` at `.github/workflows/go.yml:2`, `.github/workflows/go.yml:8`, `.github/workflows/go.yml:31`, `.github/workflows/go.yml:35`. | Keep as inherited CI. A later `ghx` binary rename can add a fork-specific smoke lane without changing this baseline. |
| `.github/workflows/lint.yml` | `unchanged upstream behavior` | Runs on Go and license file changes, uses `contents: read`, runs `go mod tidy`, `golangci-lint`, license generation, and source-mode vulnerability checks at `.github/workflows/lint.yml:2`, `.github/workflows/lint.yml:19`, `.github/workflows/lint.yml:33`, `.github/workflows/lint.yml:48`, `.github/workflows/lint.yml:59`, `.github/workflows/lint.yml:82`. | Keep as inherited source hygiene. |
| `.github/workflows/codeql.yml` | `unchanged upstream behavior` | Runs on push, PR, and weekly schedule, grants SARIF upload only through `security-events: write`, analyzes Go and Actions at `.github/workflows/codeql.yml:3`, `.github/workflows/codeql.yml:13`, `.github/workflows/codeql.yml:24`, `.github/workflows/codeql.yml:36`, `.github/workflows/codeql.yml:58`. | Keep as inherited security scanning. |
| `.github/workflows/govulncheck.yml` | `unchanged upstream behavior` | Runs weekly or manually, uses `contents: read` and `security-events: write`, writes SARIF from `govulncheck` at `.github/workflows/govulncheck.yml:2`, `.github/workflows/govulncheck.yml:10`, `.github/workflows/govulncheck.yml:24`, `.github/workflows/govulncheck.yml:28`. | Keep as inherited vulnerability scanning. |
| `.github/workflows/deployment.yml` | `needs a ghx replacement` | Manual production default, repository write, attestations, signing, upstream docs checkout, `gh_*` asset globs, upstream release title, upstream manual paths, and Homebrew formula mutation at `.github/workflows/deployment.yml:8`, `.github/workflows/deployment.yml:13`, `.github/workflows/deployment.yml:19`, `.github/workflows/deployment.yml:97`, `.github/workflows/deployment.yml:213`, `.github/workflows/deployment.yml:291`, `.github/workflows/deployment.yml:294`, `.github/workflows/deployment.yml:306`, `.github/workflows/deployment.yml:320`, `.github/workflows/deployment.yml:345`, `.github/workflows/deployment.yml:385`, `.github/workflows/deployment.yml:390`, `.github/workflows/deployment.yml:399`, `.github/workflows/deployment.yml:418`. | Do not use for a production `ghx` release. #57 should create a fork-owned staging path first, then production publication with `ghx_*` assets, fork-owned docs, fork-owned package channels, and fork-owned signing or explicit unsigned policy. |
| `.github/workflows/homebrew-bump.yml` | `fork-disabled` | Manual workflow uses `contents: write`, production environment default, `formula-name: gh`, and `HOMEBREW_PR_PAT` at `.github/workflows/homebrew-bump.yml:3`, `.github/workflows/homebrew-bump.yml:6`, `.github/workflows/homebrew-bump.yml:12`, `.github/workflows/homebrew-bump.yml:19`, `.github/workflows/homebrew-bump.yml:23`, `.github/workflows/homebrew-bump.yml:26`. | Keep disabled for `ghx` until #63 defines a fork-owned tap or explicitly keeps Homebrew unsupported. |
| `.github/workflows/bump-go.yml` | `needs a ghx replacement` | Scheduled write workflow calls `.github/workflows/scripts/bump-go.sh --apply go.mod` at `.github/workflows/bump-go.yml:2`, `.github/workflows/bump-go.yml:6`, `.github/workflows/bump-go.yml:21`, `.github/workflows/bump-go.yml:29`; that helper hard-codes `REPO="cli/cli"` at `.github/workflows/scripts/bump-go.sh:38` and searches PRs against that repo at `.github/workflows/scripts/bump-go.sh:107`. | Replace with a repo-context-aware helper before enabling scheduled Go bump PRs in `agustif/ghx`. |
| `.github/workflows/detect-spam.yml` | `needs a ghx replacement` | Issue-open automation writes issues, uses `models: read`, runs in `cli-automation`, and uses `AUTOMATION_TOKEN` at `.github/workflows/detect-spam.yml:2`, `.github/workflows/detect-spam.yml:6`, `.github/workflows/detect-spam.yml:14`, `.github/workflows/detect-spam.yml:18`, `.github/workflows/detect-spam.yml:20`; the prompt describes the GitHub CLI project at `.github/workflows/scripts/spam-detection/generate-sys-prompt.sh:38`. | Replace or parameterize the product prompt, model policy, labels, and token before using it for `ghx` issues. |
| `.github/workflows/triage-discussion-label.yml` | `needs a ghx replacement` | Uses `pull_request_target`, calls the shared desktop workflow, and hard-codes `target_repo: 'github/cli'`, `cc_team: '@github/cli'`, `cli-discuss-automation`, and `CLI_DISCUSSION_TRIAGE_TOKEN` at `.github/workflows/triage-discussion-label.yml:8`, `.github/workflows/triage-discussion-label.yml:17`, `.github/workflows/triage-discussion-label.yml:19`, `.github/workflows/triage-discussion-label.yml:20`, `.github/workflows/triage-discussion-label.yml:21`, `.github/workflows/triage-discussion-label.yml:23`. | Replace with fork-owned discussion routing or keep disabled. |
| `.github/workflows/triage-issues.yml` | `fork-compatible` | Issue events only, current-repo issue and PR permissions, and shared desktop triage workflows at `.github/workflows/triage-issues.yml:2`, `.github/workflows/triage-issues.yml:7`, `.github/workflows/triage-issues.yml:9`, `.github/workflows/triage-issues.yml:13`, `.github/workflows/triage-issues.yml:21`, `.github/workflows/triage-issues.yml:57`. | May be enabled if `ghx` intentionally keeps the upstream label taxonomy and accepts the shared workflow dependency. Otherwise fork it under #72. |
| `.github/workflows/triage-pull-requests.yml` | `fork-compatible` | Uses `pull_request_target` but delegates to shared workflows without checking out PR code, sets repo-local PR permissions, and uses `default_branch: trunk` at `.github/workflows/triage-pull-requests.yml:2`, `.github/workflows/triage-pull-requests.yml:13`, `.github/workflows/triage-pull-requests.yml:23`, `.github/workflows/triage-pull-requests.yml:25`, `.github/workflows/triage-pull-requests.yml:33`, `.github/workflows/triage-pull-requests.yml:44`. | May be enabled after owner review confirms the shared workflow policy matches `ghx`. |
| `.github/workflows/triage-scheduled-tasks.yml` | `fork-compatible` | Manual, comment, and schedule triggers call shared desktop issue workflows with issue-write permissions at `.github/workflows/triage-scheduled-tasks.yml:2`, `.github/workflows/triage-scheduled-tasks.yml:6`, `.github/workflows/triage-scheduled-tasks.yml:14`, `.github/workflows/triage-scheduled-tasks.yml:20`, `.github/workflows/triage-scheduled-tasks.yml:32`. | May be enabled if `ghx` wants inherited stale, no-response, and pitch-surfacing policy. Otherwise fork under #72. |

## Release workflow blockers for #57

The release workflow is the only checked-in workflow that can produce and publish production artifacts. It is not fork-ready because it still assumes upstream `gh` identity across every publication boundary:

- Artifact names and checksums use `gh_*` in `.github/workflows/deployment.yml:320`, `.github/workflows/deployment.yml:345`, `.github/workflows/deployment.yml:385`, and `.github/workflows/deployment.yml:399`.
- The release title is `GitHub CLI ${TAG_NAME#v}` at `.github/workflows/deployment.yml:390`.
- Manual pages are written to upstream `manual/gh*.md` paths at `.github/workflows/deployment.yml:306` and `.github/workflows/deployment.yml:308`.
- The workflow checks out `github/cli.github.com` with `SITE_DEPLOY_PAT` at `.github/workflows/deployment.yml:291` through `.github/workflows/deployment.yml:297`.
- Homebrew publication still targets formula `gh` and `williammartin/homebrew-core` at `.github/workflows/deployment.yml:418` through `.github/workflows/deployment.yml:427`.

Required replacement properties:

1. A staging workflow must build `ghx_*` artifacts without writing external repos.
2. Production publication must use fork-owned GitHub environments, secrets, and approval rules.
3. Docs publication must target a fork-owned manual site or stay disabled.
4. Package publication must target fork-owned package channels or stay disabled.
5. Attestation and checksum examples must name `agustif/ghx` and `ghx_*` artifacts.

## Automation blockers for #72

These are the smallest safe next changes before enabling fork automation:

1. Replace `.github/workflows/scripts/bump-go.sh` hard-coded `cli/cli` repo selection with current-repo discovery or a documented `GHX_TARGET_REPO` input.
2. Replace the spam-detection prompt and labels with `ghx` product context before running `.github/workflows/detect-spam.yml`.
3. Disable or replace `.github/workflows/triage-discussion-label.yml` because it routes to `github/cli` and `@github/cli`.
4. Review the shared desktop workflow dependency for `.github/workflows/triage-issues.yml`, `.github/workflows/triage-pull-requests.yml`, and `.github/workflows/triage-scheduled-tasks.yml`; pin, fork, or accept it explicitly.
