package tokens

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/cli/cli/v2/api"
	"github.com/cli/cli/v2/internal/gh"
	"github.com/cli/cli/v2/internal/tableprinter"
	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/spf13/cobra"
)

var tokenRequestFields = []string{
	"id",
	"owner",
	"tokenName",
	"repositorySelection",
	"reason",
	"expiresAt",
	"lastUsedAt",
}

type ListOptions struct {
	IO         *iostreams.IOStreams
	Config     func() (gh.Config, error)
	HttpClient func() (*http.Client, error)

	Org      string
	Limit    int
	Exporter cmdutil.Exporter
}

type ReviewOptions struct {
	IO *iostreams.IOStreams

	Org       string
	RequestID int
	Decision  string
	DryRun    bool
}

type TokenRequest struct {
	ID                  int
	Owner               string
	TokenName           string
	RepositorySelection string
	Reason              string
	ExpiresAt           string
	LastUsedAt          string
}

func (r TokenRequest) ExportData(fields []string) map[string]interface{} {
	return cmdutil.StructExportData(r, fields)
}

type tokenRequestResponse struct {
	ID                  int    `json:"id"`
	Reason              string `json:"reason"`
	RepositorySelection string `json:"repository_selection"`
	TokenName           string `json:"token_name"`
	TokenExpiresAt      string `json:"token_expires_at"`
	TokenLastUsedAt     string `json:"token_last_used_at"`
	Owner               struct {
		Login string `json:"login"`
	} `json:"owner"`
}

func NewCmdTokens(f *cmdutil.Factory) *cobra.Command {
	return newCmdTokens(f, nil)
}

func newCmdTokens(f *cmdutil.Factory, runF func(*ListOptions) error) *cobra.Command {
	listOpts := &ListOptions{
		IO:         f.IOStreams,
		Config:     f.Config,
		HttpClient: f.HttpClient,
		Limit:      30,
	}

	cmd := &cobra.Command{
		Use:   "tokens <org>",
		Short: "Inspect organization fine-grained PAT requests",
		Long: heredoc.Doc(`
			List organization fine-grained personal access token requests.

			The list command is read-only. Review actions are currently dry-run only
			and require explicit request IDs.
		`),
		Example: heredoc.Doc(`
			$ gh org tokens github --limit 20
			$ gh org tokens github --json id,owner,tokenName,reason
			$ gh org tokens review github --request-id 123 --decision deny --dry-run
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			listOpts.Org = args[0]
			if listOpts.Limit < 1 {
				return cmdutil.FlagErrorf("invalid limit: %v", listOpts.Limit)
			}
			if runF != nil {
				return runF(listOpts)
			}
			return tokensRun(listOpts)
		},
	}

	cmd.Flags().IntVarP(&listOpts.Limit, "limit", "L", listOpts.Limit, "Maximum number of token requests to fetch")
	cmdutil.AddJSONFlags(cmd, &listOpts.Exporter, tokenRequestFields)
	cmd.AddCommand(newCmdReview(f, nil))

	return cmd
}

func newCmdReview(f *cmdutil.Factory, runF func(*ReviewOptions) error) *cobra.Command {
	opts := &ReviewOptions{
		IO:       f.IOStreams,
		Decision: "deny",
	}
	cmd := &cobra.Command{
		Use:   "review <org>",
		Short: "Preview a fine-grained PAT request review",
		Long: heredoc.Doc(`
			Preview a fine-grained personal access token request review.

			This command is dry-run only. It prints the intended mutation target and
			exits without sending a review request.
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Org = args[0]
			if opts.RequestID < 1 {
				return cmdutil.FlagErrorf("required flag `--request-id` not specified")
			}
			if !opts.DryRun {
				return cmdutil.FlagErrorf("`--dry-run` is required")
			}
			if runF != nil {
				return runF(opts)
			}
			return reviewRun(opts)
		},
	}
	cmd.Flags().IntVar(&opts.RequestID, "request-id", 0, "Fine-grained PAT request ID to review")
	cmdutil.StringEnumFlag(cmd, &opts.Decision, "decision", "", opts.Decision, []string{"approve", "deny"}, "Review decision to preview")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "Print the review action without sending it")
	return cmd
}

func tokensRun(opts *ListOptions) error {
	httpClient, err := opts.HttpClient()
	if err != nil {
		return err
	}
	cfg, err := opts.Config()
	if err != nil {
		return err
	}
	host, _ := cfg.Authentication().DefaultHost()

	query := url.Values{}
	query.Set("per_page", fmt.Sprintf("%d", opts.Limit))
	path := fmt.Sprintf("orgs/%s/personal-access-token-requests?%s", url.PathEscape(opts.Org), query.Encode())

	var response []tokenRequestResponse
	client := api.NewClientFromHTTP(httpClient)
	if err := client.REST(host, "GET", path, nil, &response); err != nil {
		return err
	}

	requests := make([]TokenRequest, 0, len(response))
	for _, r := range response {
		requests = append(requests, TokenRequest{
			ID:                  r.ID,
			Owner:               r.Owner.Login,
			TokenName:           r.TokenName,
			RepositorySelection: r.RepositorySelection,
			Reason:              r.Reason,
			ExpiresAt:           r.TokenExpiresAt,
			LastUsedAt:          r.TokenLastUsedAt,
		})
	}

	if opts.Exporter != nil {
		return opts.Exporter.Write(opts.IO, requests)
	}
	printTokenRequests(opts.IO, requests)
	return nil
}

func reviewRun(opts *ReviewOptions) error {
	fmt.Fprintf(opts.IO.Out, "Would %s fine-grained PAT request %d in %s\n", opts.Decision, opts.RequestID, opts.Org)
	fmt.Fprintf(opts.IO.Out, "endpoint: POST /orgs/%s/personal-access-token-requests\n", opts.Org)
	fmt.Fprintf(opts.IO.Out, "body: {\"pat_request_ids\":[%d],\"action\":\"%s\"}\n", opts.RequestID, opts.Decision)
	fmt.Fprintln(opts.IO.Out, "dry-run: true")
	fmt.Fprintln(opts.IO.Out, "No review was submitted.")
	return nil
}

func printTokenRequests(io *iostreams.IOStreams, requests []TokenRequest) {
	if len(requests) == 0 {
		fmt.Fprintln(io.Out, "no fine-grained PAT requests found")
		return
	}
	tp := tableprinter.New(io, tableprinter.WithHeader("ID", "Owner", "Token", "Repositories", "Reason", "Expires"))
	for _, request := range requests {
		tp.AddField(fmt.Sprintf("%d", request.ID))
		tp.AddField(request.Owner)
		tp.AddField(request.TokenName)
		tp.AddField(request.RepositorySelection)
		tp.AddField(strings.TrimSpace(request.Reason))
		tp.AddField(request.ExpiresAt)
		tp.EndRow()
	}
	_ = tp.Render()
}
