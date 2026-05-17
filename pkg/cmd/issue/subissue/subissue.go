package subissue

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/MakeNowJust/heredoc"
	"github.com/cli/cli/v2/api"
	"github.com/cli/cli/v2/internal/ghrepo"
	issueShared "github.com/cli/cli/v2/pkg/cmd/issue/shared"
	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/spf13/cobra"
)

var subIssueFields = []string{"number", "title", "state", "url", "repository"}

// SubIssue is a compact issue shape for subissue listings.
type SubIssue struct {
	Number     int                `json:"number"`
	Title      string             `json:"title"`
	State      string             `json:"state"`
	URL        string             `json:"url"`
	Repository subIssueRepository `json:"repository"`
	ID         string             `json:"-"`
}

// ExportData maps subissue fields for --json output.
func (s SubIssue) ExportData(fields []string) map[string]interface{} {
	return cmdutil.StructExportData(s, fields)
}

type subIssueRepository string

func (r *subIssueRepository) UnmarshalJSON(data []byte) error {
	var repo string
	if err := json.Unmarshal(data, &repo); err == nil {
		*r = subIssueRepository(repo)
		return nil
	}
	var response struct {
		NameWithOwner string `json:"nameWithOwner"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return err
	}
	*r = subIssueRepository(response.NameWithOwner)
	return nil
}

func (r subIssueRepository) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(r))
}

type options struct {
	HttpClient func() (*http.Client, error)
	IO         *iostreams.IOStreams
	BaseRepo   func() (ghrepo.Interface, error)

	Issue     string
	SubIssue  string
	Before    string
	After     string
	Exporter  cmdutil.Exporter
	Operation string
}

// NewCmdSubIssue builds the issue subissue command tree.
func NewCmdSubIssue(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "subissue <command>",
		Aliases: []string{"subissues"},
		Short:   "Manage subissues",
		Long: heredoc.Doc(`
			Manage GitHub subissues for an issue.
		`),
	}

	cmd.AddCommand(newCmdList(f, nil))
	cmd.AddCommand(newCmdAdd(f, nil))
	cmd.AddCommand(newCmdRemove(f, nil))
	cmd.AddCommand(newCmdReprioritize(f, nil))

	return cmd
}

func newOptions(f *cmdutil.Factory) *options {
	return &options{
		IO:         f.IOStreams,
		HttpClient: f.HttpClient,
	}
}

func newCmdList(f *cmdutil.Factory, runF func(*options) error) *cobra.Command {
	opts := newOptions(f)
	cmd := &cobra.Command{
		Use:   "list {<number> | <url>}",
		Short: "List subissues",
		Example: heredoc.Doc(`
			$ gh issue subissue list 123
			$ gh issue subissue list 123 --json number,title,state,url
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Operation = "list"
			opts.Issue = args[0]
			opts.BaseRepo = f.BaseRepo
			if runF != nil {
				return runF(opts)
			}
			return listRun(opts)
		},
	}
	cmdutil.AddJSONFlags(cmd, &opts.Exporter, subIssueFields)
	return cmd
}

func newCmdAdd(f *cmdutil.Factory, runF func(*options) error) *cobra.Command {
	opts := newOptions(f)
	cmd := &cobra.Command{
		Use:   "add {<number> | <url>} {<subissue-number> | <subissue-url>}",
		Short: "Add a subissue",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Operation = "add"
			opts.Issue = args[0]
			opts.SubIssue = args[1]
			opts.BaseRepo = f.BaseRepo
			if runF != nil {
				return runF(opts)
			}
			return mutateRun(opts)
		},
	}
	return cmd
}

func newCmdRemove(f *cmdutil.Factory, runF func(*options) error) *cobra.Command {
	opts := newOptions(f)
	cmd := &cobra.Command{
		Use:   "remove {<number> | <url>} {<subissue-number> | <subissue-url>}",
		Short: "Remove a subissue",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Operation = "remove"
			opts.Issue = args[0]
			opts.SubIssue = args[1]
			opts.BaseRepo = f.BaseRepo
			if runF != nil {
				return runF(opts)
			}
			return mutateRun(opts)
		},
	}
	return cmd
}

func newCmdReprioritize(f *cmdutil.Factory, runF func(*options) error) *cobra.Command {
	opts := newOptions(f)
	cmd := &cobra.Command{
		Use:   "reprioritize {<number> | <url>} {<subissue-number> | <subissue-url>}",
		Short: "Move a subissue before or after another subissue",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if (opts.Before == "") == (opts.After == "") {
				return cmdutil.FlagErrorf("must provide exactly one of `--before` or `--after`")
			}
			opts.Operation = "reprioritize"
			opts.Issue = args[0]
			opts.SubIssue = args[1]
			opts.BaseRepo = f.BaseRepo
			if runF != nil {
				return runF(opts)
			}
			return reprioritizeRun(opts)
		},
	}
	cmd.Flags().StringVar(&opts.Before, "before", "", "Move before another subissue by number or URL")
	cmd.Flags().StringVar(&opts.After, "after", "", "Move after another subissue by number or URL")
	return cmd
}

func listRun(opts *options) error {
	httpClient, err := opts.HttpClient()
	if err != nil {
		return err
	}
	baseRepo, err := opts.BaseRepo()
	if err != nil {
		return err
	}
	issue, repo, err := resolveIssue(httpClient, baseRepo, opts.Issue)
	if err != nil {
		return err
	}
	subIssues, err := listSubIssues(httpClient, repo, issue.Number)
	if err != nil {
		return err
	}
	if opts.Exporter != nil {
		return opts.Exporter.Write(opts.IO, subIssues)
	}
	if len(subIssues) == 0 {
		fmt.Fprintf(opts.IO.ErrOut, "Issue %s#%d has no subissues\n", ghrepo.FullName(repo), issue.Number)
		return nil
	}
	for _, subIssue := range subIssues {
		fmt.Fprintf(opts.IO.Out, "#%d\t%s\t%s\t%s\n", subIssue.Number, subIssue.State, subIssue.Title, subIssue.URL)
	}
	return nil
}

func mutateRun(opts *options) error {
	httpClient, err := opts.HttpClient()
	if err != nil {
		return err
	}
	baseRepo, err := opts.BaseRepo()
	if err != nil {
		return err
	}
	issue, repo, err := resolveIssue(httpClient, baseRepo, opts.Issue)
	if err != nil {
		return err
	}
	subIssue, _, err := resolveIssue(httpClient, repo, opts.SubIssue)
	if err != nil {
		return err
	}
	switch opts.Operation {
	case "add":
		err = addSubIssue(httpClient, repo, issue.ID, subIssue.ID)
	case "remove":
		err = removeSubIssue(httpClient, repo, issue.ID, subIssue.ID)
	default:
		err = fmt.Errorf("unsupported subissue operation %q", opts.Operation)
	}
	if err != nil {
		return err
	}
	verb := "Added"
	if opts.Operation == "remove" {
		verb = "Removed"
	}
	fmt.Fprintf(opts.IO.ErrOut, "%s subissue #%d for issue %s#%d\n", verb, subIssue.Number, ghrepo.FullName(repo), issue.Number)
	return nil
}

func reprioritizeRun(opts *options) error {
	httpClient, err := opts.HttpClient()
	if err != nil {
		return err
	}
	baseRepo, err := opts.BaseRepo()
	if err != nil {
		return err
	}
	issue, repo, err := resolveIssue(httpClient, baseRepo, opts.Issue)
	if err != nil {
		return err
	}
	subIssue, _, err := resolveIssue(httpClient, repo, opts.SubIssue)
	if err != nil {
		return err
	}

	var beforeID string
	var afterID string
	if opts.Before != "" {
		before, _, err := resolveIssue(httpClient, repo, opts.Before)
		if err != nil {
			return err
		}
		beforeID = before.ID
	}
	if opts.After != "" {
		after, _, err := resolveIssue(httpClient, repo, opts.After)
		if err != nil {
			return err
		}
		afterID = after.ID
	}
	if err := reprioritizeSubIssue(httpClient, repo, issue.ID, subIssue.ID, beforeID, afterID); err != nil {
		return err
	}
	fmt.Fprintf(opts.IO.ErrOut, "Reprioritized subissue #%d for issue %s#%d\n", subIssue.Number, ghrepo.FullName(repo), issue.Number)
	return nil
}

func resolveIssue(httpClient *http.Client, baseRepo ghrepo.Interface, arg string) (*api.Issue, ghrepo.Interface, error) {
	issueNumber, issueRepo, err := issueShared.ParseIssueFromArg(arg)
	if err != nil {
		return nil, nil, err
	}
	repo := baseRepo
	if parsedRepo, present := issueRepo.Value(); present {
		repo = parsedRepo
	}
	issue, err := issueShared.FindIssueOrPR(httpClient, repo, issueNumber, []string{"id", "number", "title", "state", "url"})
	if err != nil {
		return nil, nil, err
	}
	if issue.IsPullRequest() {
		return nil, nil, errors.New("subissues are only supported for issues")
	}
	return issue, repo, nil
}

func listSubIssues(httpClient *http.Client, repo ghrepo.Interface, number int) ([]SubIssue, error) {
	query := `
	query IssueSubIssues($owner: String!, $repo: String!, $number: Int!, $endCursor: String) {
		repository(owner: $owner, name: $repo) {
			hasIssuesEnabled
			issue(number: $number) {
				subIssues(first: 100, after: $endCursor) {
					nodes {
						id
						number
						title
						state
						url
						repository { nameWithOwner }
					}
					pageInfo { hasNextPage endCursor }
				}
			}
		}
	}`
	variables := map[string]interface{}{
		"owner":  repo.RepoOwner(),
		"repo":   repo.RepoName(),
		"number": number,
	}
	var all []SubIssue
	client := api.NewClientFromHTTP(httpClient)
	for {
		var response struct {
			Repository struct {
				HasIssuesEnabled bool
				Issue            *struct {
					SubIssues struct {
						Nodes    []SubIssue
						PageInfo struct {
							HasNextPage bool
							EndCursor   string
						}
					}
				}
			}
		}
		if err := client.GraphQL(repo.RepoHost(), query, variables, &response); err != nil {
			return nil, err
		}
		if !response.Repository.HasIssuesEnabled {
			return nil, fmt.Errorf("the '%s' repository has disabled issues", ghrepo.FullName(repo))
		}
		if response.Repository.Issue == nil {
			return nil, fmt.Errorf("issue %s#%d not found", ghrepo.FullName(repo), number)
		}
		all = append(all, response.Repository.Issue.SubIssues.Nodes...)
		if !response.Repository.Issue.SubIssues.PageInfo.HasNextPage {
			break
		}
		variables["endCursor"] = response.Repository.Issue.SubIssues.PageInfo.EndCursor
	}
	return all, nil
}

func addSubIssue(httpClient *http.Client, repo ghrepo.Interface, issueID, subIssueID string) error {
	query := `
	mutation AddSubIssue($input: AddSubIssueInput!) {
		addSubIssue(input: $input) {
			issue { id }
			subIssue { id }
		}
	}`
	return api.NewClientFromHTTP(httpClient).GraphQL(repo.RepoHost(), query, map[string]interface{}{
		"input": map[string]interface{}{
			"issueId":    issueID,
			"subIssueId": subIssueID,
		},
	}, &struct{}{})
}

func removeSubIssue(httpClient *http.Client, repo ghrepo.Interface, issueID, subIssueID string) error {
	query := `
	mutation RemoveSubIssue($input: RemoveSubIssueInput!) {
		removeSubIssue(input: $input) {
			issue { id }
			subIssue { id }
		}
	}`
	return api.NewClientFromHTTP(httpClient).GraphQL(repo.RepoHost(), query, map[string]interface{}{
		"input": map[string]interface{}{
			"issueId":    issueID,
			"subIssueId": subIssueID,
		},
	}, &struct{}{})
}

func reprioritizeSubIssue(httpClient *http.Client, repo ghrepo.Interface, issueID, subIssueID, beforeID, afterID string) error {
	query := `
	mutation ReprioritizeSubIssue($input: ReprioritizeSubIssueInput!) {
		reprioritizeSubIssue(input: $input) {
			issue { id }
		}
	}`
	var before interface{}
	if beforeID != "" {
		before = beforeID
	}
	var after interface{}
	if afterID != "" {
		after = afterID
	}
	variables := map[string]interface{}{"input": map[string]interface{}{
		"issueId":    issueID,
		"subIssueId": subIssueID,
		"beforeId":   before,
		"afterId":    after,
	}}
	return api.NewClientFromHTTP(httpClient).GraphQL(repo.RepoHost(), query, variables, &struct{}{})
}
