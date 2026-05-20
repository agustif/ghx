package inventory

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

var inventoryFields = []string{
	"host",
	"repository",
	"pullRequest",
	"headSha",
	"name",
	"state",
	"status",
	"conclusion",
	"bucket",
	"detailsUrl",
	"app",
	"checkRunId",
	"checkSuiteId",
	"workflow",
	"workflowRunId",
	"required",
	"rerunSupported",
	"rerunKind",
	"rerunEndpoint",
	"rerunCommand",
	"rerunLimitation",
	"provider",
}

type InventoryOptions struct {
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
	Rerunnable  bool
	IncludePass bool
}

func NewCmdInventory(f *cmdutil.Factory, runF func(*InventoryOptions) error) *cobra.Command {
	opts := &InventoryOptions{
		HttpClient: f.HttpClient,
		IO:         f.IOStreams,
		Limit:      30,
		State:      "open",
	}

	cmd := &cobra.Command{
		Use:   "inventory",
		Short: "List checks across pull requests",
		Long: heredoc.Docf(`
			List check runs and status contexts across pull requests.

			By default, this command scans open pull requests and hides passing checks.
			Use %[1]s--include-pass%[1]s to include successful checks in the inventory.
		`, "`"),
		Example: heredoc.Doc(`
			# List failing checks on open pull requests
			$ gh checks inventory --bucket fail --json pullRequest,name,state,detailsUrl,checkRunId,rerunSupported

			# List all checks for a single pull request
			$ gh checks inventory --pr 42 --include-pass --json pullRequest,name,bucket,app,rerunCommand
		`),
		Args: cmdutil.NoArgsQuoteReminder,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.BaseRepo = f.BaseRepo
			if opts.Limit < 1 || opts.Limit > 100 {
				return cmdutil.FlagErrorf("invalid value for --limit: %v", opts.Limit)
			}
			if opts.PRNumber > 0 && cmd.Flags().Changed("state") {
				return cmdutil.FlagErrorf("cannot use `--state` with `--pr`")
			}
			if runF != nil {
				return runF(opts)
			}
			return inventoryRun(opts)
		},
	}

	cmdutil.StringEnumFlag(cmd, &opts.State, "state", "s", "open", []string{"open", "closed", "merged", "all"}, "Filter pull requests by state")
	cmd.Flags().IntVarP(&opts.Limit, "limit", "L", 30, "Maximum number of pull requests to scan")
	cmd.Flags().IntVar(&opts.PRNumber, "pr", 0, "Only scan one pull request number")
	cmd.Flags().StringVar(&opts.Name, "name", "", "Filter checks by case-insensitive name substring")
	cmd.Flags().StringVar(&opts.Conclusion, "conclusion", "", "Filter checks by conclusion, state, or status")
	cmdutil.StringEnumFlag(cmd, &opts.Bucket, "bucket", "", "", []string{"pass", "fail", "pending", "skipping", "cancel"}, "Filter checks by normalized bucket")
	cmd.Flags().BoolVar(&opts.Required, "required", false, "Only include required checks")
	cmd.Flags().BoolVar(&opts.Rerunnable, "rerunnable", false, "Only include checks with a known rerun target")
	cmd.Flags().BoolVar(&opts.IncludePass, "include-pass", false, "Include passing checks")
	cmdutil.AddJSONFlags(cmd, &opts.Exporter, inventoryFields)

	return cmd
}

func inventoryRun(opts *InventoryOptions) error {
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
		Rerunnable:  opts.Rerunnable,
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

	if opts.Exporter != nil {
		return opts.Exporter.Write(opts.IO, items)
	}
	if len(items) == 0 {
		return cmdutil.NewNoResultsError("no checks matched")
	}

	return printInventory(opts.IO, items)
}

func printInventory(io *iostreams.IOStreams, items []checksShared.CheckItem) error {
	if err := io.StartPager(); err == nil {
		defer io.StopPager()
	}

	tp := tableprinter.New(io, tableprinter.WithHeader("PR", "CHECK", "STATE", "APP", "RERUN", "URL"))
	for _, item := range items {
		tp.AddField(fmt.Sprintf("#%d", item.PullRequest.Number))
		tp.AddField(item.Name)
		tp.AddField(item.State)
		tp.AddField(item.App.Slug)
		if item.RerunSupported {
			tp.AddField(item.RerunKind)
		} else {
			tp.AddField("-")
		}
		tp.AddField(item.DetailsURL)
		tp.EndRow()
	}
	return tp.Render()
}
