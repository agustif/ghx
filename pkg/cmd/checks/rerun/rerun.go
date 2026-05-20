package rerun

import (
	"fmt"
	"net/http"

	"github.com/MakeNowJust/heredoc"
	"github.com/cli/cli/v2/internal/ghrepo"
	"github.com/cli/cli/v2/internal/tableprinter"
	checksShared "github.com/cli/cli/v2/pkg/cmd/checks/shared"
	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/spf13/cobra"
)

var rerunFields = []string{
	"dryRun",
	"repository",
	"targetCount",
	"runnableCount",
	"targets",
	"nextActions",
}

type RerunOptions struct {
	HttpClient func() (*http.Client, error)
	IO         *iostreams.IOStreams
	BaseRepo   func() (ghrepo.Interface, error)
	Exporter   cmdutil.Exporter

	State       string
	Limit       int
	PRNumber    int
	Name        string
	Conclusion  string
	Bucket      string
	Required    bool
	DryRun      bool
	IncludePass bool
}

// Plan is the JSON-first dry-run output for checks rerun.
type Plan struct {
	DryRun        bool                     `json:"dryRun"`
	Repository    string                   `json:"repository"`
	TargetCount   int                      `json:"targetCount"`
	RunnableCount int                      `json:"runnableCount"`
	Targets       []checksShared.CheckItem `json:"targets"`
	NextActions   []string                 `json:"nextActions"`
}

// ExportData supports gh's --json field selection.
func (p *Plan) ExportData(fields []string) map[string]interface{} {
	return cmdutil.StructExportData(p, fields)
}

func NewCmdRerun(f *cmdutil.Factory, runF func(*RerunOptions) error) *cobra.Command {
	opts := &RerunOptions{
		HttpClient: f.HttpClient,
		IO:         f.IOStreams,
		Limit:      30,
		State:      "open",
	}

	cmd := &cobra.Command{
		Use:   "rerun",
		Short: "Plan targeted check reruns",
		Long: heredoc.Doc(`
			Plan reruns for checks that match explicit filters.

			This first slice is dry-run only. It prints exact check-run targets and
			the command that would rerequest each supported target.
		`),
		Example: heredoc.Doc(`
			# Plan rerunning matching checks without mutating remote state
			$ gh checks rerun --name lint --bucket fail --dry-run --json targets,targetCount,runnableCount
		`),
		Args: cmdutil.NoArgsQuoteReminder,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.BaseRepo = f.BaseRepo
			if !opts.DryRun {
				return cmdutil.FlagErrorf("only `--dry-run` is supported for `checks rerun` in this release")
			}
			if opts.Limit < 1 || opts.Limit > 100 {
				return cmdutil.FlagErrorf("invalid value for --limit: %v", opts.Limit)
			}
			if opts.PRNumber > 0 && cmd.Flags().Changed("state") {
				return cmdutil.FlagErrorf("cannot use `--state` with `--pr`")
			}
			if opts.Name == "" && opts.Conclusion == "" && opts.Bucket == "" && opts.PRNumber == 0 && !opts.Required {
				return cmdutil.FlagErrorf("at least one of `--name`, `--conclusion`, `--bucket`, `--required`, or `--pr` is required")
			}
			if runF != nil {
				return runF(opts)
			}
			return rerunRun(opts)
		},
	}

	cmdutil.StringEnumFlag(cmd, &opts.State, "state", "s", "open", []string{"open", "closed", "merged", "all"}, "Filter pull requests by state")
	cmd.Flags().IntVarP(&opts.Limit, "limit", "L", 30, "Maximum number of pull requests to scan")
	cmd.Flags().IntVar(&opts.PRNumber, "pr", 0, "Only scan one pull request number")
	cmd.Flags().StringVar(&opts.Name, "name", "", "Filter checks by case-insensitive name substring")
	cmd.Flags().StringVar(&opts.Conclusion, "conclusion", "", "Filter checks by conclusion, state, or status")
	cmdutil.StringEnumFlag(cmd, &opts.Bucket, "bucket", "", "", []string{"pass", "fail", "pending", "skipping", "cancel"}, "Filter checks by normalized bucket")
	cmd.Flags().BoolVar(&opts.Required, "required", false, "Only include required checks")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "Print exact rerun targets without mutating remote state")
	cmd.Flags().BoolVar(&opts.IncludePass, "include-pass", false, "Include passing checks")
	cmdutil.AddJSONFlags(cmd, &opts.Exporter, rerunFields)

	return cmd
}

func rerunRun(opts *RerunOptions) error {
	repo, err := opts.BaseRepo()
	if err != nil {
		return fmt.Errorf("failed to determine base repo: %w", err)
	}
	httpClient, err := opts.HttpClient()
	if err != nil {
		return fmt.Errorf("failed to create http client: %w", err)
	}

	filter := checksShared.Filter{
		Name:        opts.Name,
		Conclusion:  opts.Conclusion,
		Bucket:      opts.Bucket,
		Required:    opts.Required,
		IncludePass: opts.IncludePass,
	}

	var prs []checksShared.PullRequestSummary
	if opts.PRNumber > 0 {
		pr, err := checksShared.GetPullRequest(httpClient, repo, opts.PRNumber)
		if err != nil {
			return err
		}
		prs = []checksShared.PullRequestSummary{pr}
	} else {
		prs, err = checksShared.ListPullRequests(httpClient, repo, opts.State, opts.Limit)
		if err != nil {
			return err
		}
	}

	items, err := checksShared.FetchChecksForPullRequests(httpClient, repo, prs, filter)
	if err != nil {
		return err
	}
	plan := buildPlan(repo, items)

	if opts.Exporter != nil {
		return opts.Exporter.Write(opts.IO, &plan)
	}
	if plan.TargetCount == 0 {
		return cmdutil.NewNoResultsError("no checks matched")
	}
	return printPlan(opts.IO, plan)
}

func buildPlan(repo ghrepo.Interface, items []checksShared.CheckItem) Plan {
	plan := Plan{
		DryRun:      true,
		Repository:  ghrepo.FullName(repo),
		TargetCount: len(items),
		Targets:     items,
	}
	for _, item := range items {
		if item.RerunSupported {
			plan.RunnableCount++
		}
	}
	if plan.RunnableCount > 0 {
		plan.NextActions = append(plan.NextActions, "review the target list, then run the emitted rerun commands for the intended checks")
	}
	if plan.RunnableCount < plan.TargetCount {
		plan.NextActions = append(plan.NextActions, "unsupported targets are status contexts without check-run rerequest targets")
	}
	return plan
}

func printPlan(io *iostreams.IOStreams, plan Plan) error {
	if err := io.StartPager(); err == nil {
		defer io.StopPager()
	}

	tp := tableprinter.New(io, tableprinter.WithHeader("PR", "CHECK", "RERUN", "COMMAND"))
	for _, item := range plan.Targets {
		tp.AddField(fmt.Sprintf("#%d", item.PullRequest.Number))
		tp.AddField(item.Name)
		if item.RerunSupported {
			tp.AddField(item.RerunKind)
			tp.AddField(item.RerunCommand)
		} else {
			tp.AddField("unsupported")
			tp.AddField(item.RerunLimitation)
		}
		tp.EndRow()
	}
	return tp.Render()
}
