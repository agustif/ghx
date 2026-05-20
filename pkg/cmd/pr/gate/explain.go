package gate

import (
	"fmt"
	"net/http"

	"github.com/MakeNowJust/heredoc"
	"github.com/cli/cli/v2/api"
	"github.com/cli/cli/v2/internal/gh"
	"github.com/cli/cli/v2/internal/ghrepo"
	checksShared "github.com/cli/cli/v2/pkg/cmd/checks/shared"
	prShared "github.com/cli/cli/v2/pkg/cmd/pr/shared"
	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/spf13/cobra"
)

var explainFields = []string{
	"host",
	"account",
	"repository",
	"pullRequest",
	"ready",
	"summary",
	"mergeStateStatus",
	"mergeable",
	"reviewDecision",
	"checks",
	"mergeQueue",
	"blockers",
	"nextActions",
}

type ExplainOptions struct {
	HttpClient func() (*http.Client, error)
	IO         *iostreams.IOStreams
	Config     func() (gh.Config, error)
	Finder     prShared.PRFinder
	Exporter   cmdutil.Exporter

	SelectorArg string
}

// Report is the JSON-first merge gate explanation for one pull request.
type Report struct {
	Host             string                          `json:"host"`
	Account          string                          `json:"account"`
	Repository       string                          `json:"repository"`
	PullRequest      checksShared.PullRequestSummary `json:"pullRequest"`
	Ready            bool                            `json:"ready"`
	Summary          string                          `json:"summary"`
	MergeStateStatus string                          `json:"mergeStateStatus"`
	Mergeable        string                          `json:"mergeable"`
	ReviewDecision   string                          `json:"reviewDecision"`
	Checks           CheckSummary                    `json:"checks"`
	MergeQueue       MergeQueueSummary               `json:"mergeQueue"`
	Blockers         []Blocker                       `json:"blockers"`
	NextActions      []string                        `json:"nextActions"`
}

// ExportData supports gh's --json field selection.
func (r *Report) ExportData(fields []string) map[string]interface{} {
	return cmdutil.StructExportData(r, fields)
}

type CheckSummary struct {
	Total    int `json:"total"`
	Passing  int `json:"passing"`
	Failing  int `json:"failing"`
	Pending  int `json:"pending"`
	Skipped  int `json:"skipped"`
	Canceled int `json:"canceled"`
}

type MergeQueueSummary struct {
	Enabled bool `json:"enabled"`
	InQueue bool `json:"inQueue"`
}

type Blocker struct {
	Type       string `json:"type"`
	Severity   string `json:"severity"`
	Message    string `json:"message"`
	DetailsURL string `json:"detailsUrl"`
	Command    string `json:"command"`
}

func NewCmdExplain(f *cmdutil.Factory, runF func(*ExplainOptions) error) *cobra.Command {
	opts := &ExplainOptions{
		HttpClient: f.HttpClient,
		IO:         f.IOStreams,
		Config:     f.Config,
	}

	cmd := &cobra.Command{
		Use:   "explain [<number> | <url> | <branch>]",
		Short: "Explain whether a pull request can merge",
		Long: heredoc.Doc(`
			Explain merge readiness for one pull request using structured blocker fields.

			This command is read-only. It reports PR state, review decision, check summary,
			merge queue state, and exact next actions where gh can infer them.
		`),
		Example: heredoc.Doc(`
			# Explain blockers for the current branch pull request
			$ gh pr gate explain --json pullRequest,ready,blockers,nextActions

			# Explain blockers for a specific pull request
			$ gh pr gate explain 42 --json checks,mergeStateStatus,reviewDecision,blockers
		`),
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Finder = prShared.NewFinder(f)
			if repoOverride, _ := cmd.Flags().GetString("repo"); repoOverride != "" && len(args) == 0 {
				return cmdutil.FlagErrorf("argument required when using the --repo flag")
			}
			if len(args) > 0 {
				opts.SelectorArg = args[0]
			}
			if runF != nil {
				return runF(opts)
			}
			return explainRun(opts)
		},
	}

	cmdutil.AddJSONFlags(cmd, &opts.Exporter, explainFields)
	return cmd
}

func explainRun(opts *ExplainOptions) error {
	findOptions := prShared.FindOptions{
		Selector: opts.SelectorArg,
		Fields: []string{
			"id",
			"number",
			"title",
			"state",
			"url",
			"isDraft",
			"headRefName",
			"headRefOid",
			"baseRefName",
			"mergeable",
			"mergeStateStatus",
			"reviewDecision",
			"requiresStrictStatusChecks",
			"isInMergeQueue",
			"isMergeQueueEnabled",
		},
	}
	pr, repo, err := opts.Finder.Find(findOptions)
	if err != nil {
		return err
	}
	httpClient, err := opts.HttpClient()
	if err != nil {
		return fmt.Errorf("failed to create http client: %w", err)
	}

	checkItems, err := checksShared.FetchChecksForPullRequest(httpClient, repo, pullRequestSummary(pr), checksShared.Filter{IncludePass: true})
	if err != nil {
		return err
	}
	report := buildReport(opts, repo, pr, checkItems)

	if opts.Exporter != nil {
		return opts.Exporter.Write(opts.IO, &report)
	}
	return printReport(opts.IO, report)
}

func buildReport(opts *ExplainOptions, repo ghrepo.Interface, pr *api.PullRequest, checks []checksShared.CheckItem) Report {
	report := Report{
		Host:             repo.RepoHost(),
		Repository:       ghrepo.FullName(repo),
		Account:          activeAccount(opts, repo),
		PullRequest:      pullRequestSummary(pr),
		MergeStateStatus: pr.MergeStateStatus,
		Mergeable:        string(pr.Mergeable),
		ReviewDecision:   pr.ReviewDecision,
		Checks:           summarizeChecks(checks),
		MergeQueue: MergeQueueSummary{
			Enabled: pr.IsMergeQueueEnabled,
			InQueue: pr.IsInMergeQueue,
		},
	}
	report.Blockers = explainBlockers(repo, pr, report.Checks)
	report.Ready = len(report.Blockers) == 0
	if report.Ready {
		report.Summary = "pull request is merge-ready"
		report.NextActions = append(report.NextActions, fmt.Sprintf("gh pr merge %d --repo %s", pr.Number, ghrepo.FullName(repo)))
		if pr.IsMergeQueueEnabled && !pr.IsInMergeQueue {
			report.NextActions = append(report.NextActions, fmt.Sprintf("gh pr merge %d --auto --repo %s", pr.Number, ghrepo.FullName(repo)))
		}
	} else {
		report.Summary = fmt.Sprintf("%d merge blocker(s) found", len(report.Blockers))
		for _, blocker := range report.Blockers {
			if blocker.Command != "" {
				report.NextActions = append(report.NextActions, blocker.Command)
			}
		}
	}
	return report
}

func explainBlockers(repo ghrepo.Interface, pr *api.PullRequest, checks CheckSummary) []Blocker {
	var blockers []Blocker
	prURL := pr.URL
	repoName := ghrepo.FullName(repo)

	if pr.State != "OPEN" {
		blockers = append(blockers, Blocker{Type: "state", Severity: "blocking", Message: "pull request is not open", DetailsURL: prURL})
		return blockers
	}
	if pr.IsDraft {
		blockers = append(blockers, Blocker{
			Type:       "draft",
			Severity:   "blocking",
			Message:    "pull request is still a draft",
			DetailsURL: prURL,
			Command:    fmt.Sprintf("gh pr ready %d --repo %s", pr.Number, repoName),
		})
	}
	switch pr.ReviewDecision {
	case "CHANGES_REQUESTED":
		blockers = append(blockers, Blocker{Type: "review", Severity: "blocking", Message: "changes were requested", DetailsURL: prURL})
	case "REVIEW_REQUIRED":
		blockers = append(blockers, Blocker{Type: "review", Severity: "blocking", Message: "review approval is required", DetailsURL: prURL})
	}
	if checks.Failing > 0 {
		blockers = append(blockers, Blocker{
			Type:       "checks",
			Severity:   "blocking",
			Message:    fmt.Sprintf("%d check(s) are failing", checks.Failing),
			DetailsURL: ghrepo.GenerateRepoURL(repo, "pull/%d/checks", pr.Number),
			Command:    fmt.Sprintf("gh checks inventory --pr %d --bucket fail --repo %s --json pullRequest,name,state,detailsUrl,rerunCommand", pr.Number, repoName),
		})
	}
	if checks.Pending > 0 {
		blockers = append(blockers, Blocker{
			Type:       "checks",
			Severity:   "pending",
			Message:    fmt.Sprintf("%d check(s) are pending", checks.Pending),
			DetailsURL: ghrepo.GenerateRepoURL(repo, "pull/%d/checks", pr.Number),
		})
	}
	if pr.Mergeable == api.PullRequestMergeableConflicting || pr.MergeStateStatus == "DIRTY" {
		blockers = append(blockers, Blocker{Type: "conflict", Severity: "blocking", Message: "pull request has merge conflicts", DetailsURL: prURL})
	}
	if pr.BaseRef.BranchProtectionRule.RequiresStrictStatusChecks && pr.MergeStateStatus == "BEHIND" {
		blockers = append(blockers, Blocker{
			Type:       "stale",
			Severity:   "blocking",
			Message:    "head branch is behind the base branch",
			DetailsURL: prURL,
			Command:    fmt.Sprintf("gh pr update-branch %d --repo %s", pr.Number, repoName),
		})
	}
	if len(blockers) == 0 {
		switch pr.MergeStateStatus {
		case "BLOCKED", "HAS_HOOKS":
			blockers = append(blockers, Blocker{Type: "policy", Severity: "blocking", Message: "GitHub reports the merge gate as blocked", DetailsURL: prURL})
		case "UNKNOWN":
			blockers = append(blockers, Blocker{Type: "unknown", Severity: "pending", Message: "GitHub has not computed mergeability yet", DetailsURL: prURL})
		}
	}
	return blockers
}

func summarizeChecks(items []checksShared.CheckItem) CheckSummary {
	var summary CheckSummary
	for _, item := range items {
		summary.Total++
		switch item.Bucket {
		case "pass":
			summary.Passing++
		case "fail":
			summary.Failing++
		case "pending":
			summary.Pending++
		case "skipping":
			summary.Skipped++
		case "cancel":
			summary.Canceled++
		}
	}
	return summary
}

func pullRequestSummary(pr *api.PullRequest) checksShared.PullRequestSummary {
	return checksShared.PullRequestSummary{
		ID:          pr.ID,
		Number:      pr.Number,
		Title:       pr.Title,
		URL:         pr.URL,
		State:       pr.State,
		HeadRefName: pr.HeadRefName,
		HeadSHA:     pr.HeadRefOid,
		IsDraft:     pr.IsDraft,
	}
}

func activeAccount(opts *ExplainOptions, repo ghrepo.Interface) string {
	if opts.Config == nil {
		return ""
	}
	cfg, err := opts.Config()
	if err != nil {
		return ""
	}
	account, err := cfg.Authentication().ActiveUser(repo.RepoHost())
	if err != nil {
		return ""
	}
	return account
}

func printReport(io *iostreams.IOStreams, report Report) error {
	out := io.Out
	fmt.Fprintf(out, "PR #%d: %s\n", report.PullRequest.Number, report.Summary)
	for _, blocker := range report.Blockers {
		fmt.Fprintf(out, "- %s: %s\n", blocker.Type, blocker.Message)
	}
	for _, action := range report.NextActions {
		fmt.Fprintf(out, "next: %s\n", action)
	}
	return nil
}
