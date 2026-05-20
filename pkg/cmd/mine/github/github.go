package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/MakeNowJust/heredoc"
	apiClient "github.com/cli/cli/v2/api"
	"github.com/cli/cli/v2/internal/gh"
	"github.com/cli/cli/v2/internal/ghapi/coverage"
	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/spf13/cobra"
)

// Options captures options for `gh mine github`.
type Options struct {
	Config          func() (gh.Config, error)
	HttpClient      func() (*http.Client, error)
	PlainHTTPClient func() (*http.Client, error)
	IO              *iostreams.IOStreams
	RootCommand     *cobra.Command

	Source        string
	Format        string
	RESTOpenAPI   string
	GraphQLSchema string
	Hostname      string
	Tag           string
	State         string
	Search        string
	Detail        bool
	Limit         int
}

// NewCmdGithub creates the `gh mine github` command.
func NewCmdGithub(f *cmdutil.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{
		Config:          f.Config,
		HttpClient:      f.HttpClient,
		PlainHTTPClient: f.PlainHttpClient,
		IO:              f.IOStreams,
		RESTOpenAPI:     coverage.DefaultRESTOpenAPIURL,
		Limit:           50,
	}

	cmd := &cobra.Command{
		Use:   "github",
		Short: "Generate GitHub REST and GraphQL coverage reports",
		Long: heredoc.Doc(`
			Generate a coverage map that compares GitHub REST and GraphQL source
			surfaces with the explicit ghx metadata and command surface.

			This is different from raw endpoint reachability. Regular gh and ghx can
			both reach broad API surfaces through the api command. The coverage map
			tracks explicit metadata, coverage state, proposed local commands, and
			stable machine-readable rows.
		`),
		Example: heredoc.Doc(`
			# Generate a Markdown summary from live REST OpenAPI and GraphQL introspection
			$ gh mine github --source all --format md

			# Generate a full REST JSON coverage map from a pinned OpenAPI file
			$ gh mine github --source rest --format json --rest-openapi /tmp/github-rest-openapi.json

			# Show missing Actions operations in Markdown detail mode
			$ gh mine github --source rest --tag actions --state missing --detail
		`),
		Args: cmdutil.NoArgsQuoteReminder,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !coverage.ValidCoverageState(opts.State) {
				return cmdutil.FlagErrorf("invalid value for --state: %s", opts.State)
			}
			if opts.Limit < 0 {
				return cmdutil.FlagErrorf("invalid value for --limit: %d", opts.Limit)
			}
			if opts.Source == "graphql" && cmd.Flags().Changed("rest-openapi") {
				return cmdutil.FlagErrorf("cannot use `--rest-openapi` with `--source graphql`")
			}
			if opts.Source == "rest" && cmd.Flags().Changed("graphql-schema") {
				return cmdutil.FlagErrorf("cannot use `--graphql-schema` with `--source rest`")
			}
			if opts.Source == "rest" && cmd.Flags().Changed("hostname") {
				return cmdutil.FlagErrorf("cannot use `--hostname` with `--source rest`")
			}
			opts.RootCommand = cmd.Root()
			if runF != nil {
				return runF(opts)
			}
			return githubRun(opts)
		},
	}

	cmdutil.StringEnumFlag(cmd, &opts.Source, "source", "", "all", []string{"all", "rest", "graphql"}, "API source to mine")
	cmdutil.StringEnumFlag(cmd, &opts.Format, "format", "", "md", []string{"md", "json"}, "Output format")
	cmd.Flags().StringVar(&opts.RESTOpenAPI, "rest-openapi", coverage.DefaultRESTOpenAPIURL, "REST OpenAPI file path or URL")
	cmd.Flags().StringVar(&opts.GraphQLSchema, "graphql-schema", "", "GraphQL introspection JSON file")
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "GitHub hostname for GraphQL introspection")
	cmd.Flags().StringVar(&opts.Tag, "tag", "", "Filter REST operations by tag")
	cmd.Flags().StringVar(&opts.State, "state", "", "Filter rows by coverage state")
	cmd.Flags().StringVar(&opts.Search, "search", "", "Filter rows by case-insensitive text")
	cmd.Flags().BoolVar(&opts.Detail, "detail", false, "Include operation or field detail rows in Markdown output")
	cmd.Flags().IntVar(&opts.Limit, "limit", 50, "Maximum Markdown detail rows to show, or 0 for all")

	return cmd
}

func githubRun(opts *Options) error {
	report := coverage.Report{}
	if opts.RootCommand != nil {
		report.Commands = coverage.BuildCommandReport(opts.RootCommand)
	}

	if opts.Source == "all" || opts.Source == "rest" {
		restReport, err := buildRESTReport(opts)
		if err != nil {
			return err
		}
		report.REST = restReport
	}

	if opts.Source == "all" || opts.Source == "graphql" {
		graphQLReport, err := buildGraphQLReport(opts)
		if err != nil {
			return err
		}
		report.GraphQL = graphQLReport
	}

	switch opts.Format {
	case "json":
		encoder := json.NewEncoder(opts.IO.Out)
		encoder.SetEscapeHTML(false)
		encoder.SetIndent("", "  ")
		return encoder.Encode(report)
	case "md":
		return coverage.WriteMarkdown(opts.IO.Out, report, coverage.MarkdownOptions{
			Detail: opts.Detail,
			Limit:  opts.Limit,
		})
	default:
		return fmt.Errorf("unsupported format %q", opts.Format)
	}
}

func buildRESTReport(opts *Options) (*coverage.RESTReport, error) {
	reader, source, closeFn, err := openRESTOpenAPI(opts)
	if err != nil {
		return nil, err
	}
	if closeFn != nil {
		defer closeFn()
	}
	return coverage.BuildRESTReport(reader, source, coverage.RESTFilter{
		Tag:   opts.Tag,
		State: opts.State,
		Query: opts.Search,
	})
}

func buildGraphQLReport(opts *Options) (*coverage.GraphQLReport, error) {
	if opts.GraphQLSchema != "" {
		schema, err := readGraphQLSchemaFile(opts.GraphQLSchema)
		if err != nil {
			return nil, err
		}
		host := opts.Hostname
		if host == "" {
			host = "schema-file"
		}
		return coverage.BuildGraphQLReport(host, schema, coverage.GraphQLFilter{
			State: opts.State,
			Query: opts.Search,
		})
	}

	cfg, err := opts.Config()
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration: %w", err)
	}
	host, _ := cfg.Authentication().DefaultHost()
	if opts.Hostname != "" {
		host = opts.Hostname
	}
	httpClient, err := opts.HttpClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create http client: %w", err)
	}

	var response struct {
		Schema coverage.GraphQLSchema `json:"__schema"`
	}
	client := apiClient.NewClientFromHTTP(httpClient)
	if err := client.GraphQL(host, coverage.GraphQLIntrospectionQuery, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to introspect GitHub GraphQL schema: %w", err)
	}
	return coverage.BuildGraphQLReport(host, response.Schema, coverage.GraphQLFilter{
		State: opts.State,
		Query: opts.Search,
	})
}

func openRESTOpenAPI(opts *Options) (io.Reader, string, func(), error) {
	if isHTTPURL(opts.RESTOpenAPI) {
		httpClientFunc := opts.PlainHTTPClient
		if httpClientFunc == nil {
			httpClientFunc = opts.HttpClient
		}
		httpClient, err := httpClientFunc()
		if err != nil {
			return nil, "", nil, fmt.Errorf("failed to create http client: %w", err)
		}
		resp, err := httpClient.Get(opts.RESTOpenAPI)
		if err != nil {
			return nil, "", nil, fmt.Errorf("failed to download REST OpenAPI: %w", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, "", nil, fmt.Errorf("failed to download REST OpenAPI: HTTP %d", resp.StatusCode)
		}
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, "", nil, fmt.Errorf("failed to read REST OpenAPI: %w", err)
		}
		return bytes.NewReader(data), opts.RESTOpenAPI, func() {}, nil
	}

	f, err := os.Open(opts.RESTOpenAPI)
	if err != nil {
		return nil, "", nil, fmt.Errorf("failed to open REST OpenAPI %q: %w", opts.RESTOpenAPI, err)
	}
	return f, opts.RESTOpenAPI, func() { f.Close() }, nil
}

func isHTTPURL(value string) bool {
	return strings.HasPrefix(value, "https://") || strings.HasPrefix(value, "http://")
}

func readGraphQLSchemaFile(path string) (coverage.GraphQLSchema, error) {
	f, err := os.Open(path)
	if err != nil {
		return coverage.GraphQLSchema{}, fmt.Errorf("failed to open GraphQL schema %q: %w", path, err)
	}
	defer f.Close()

	var envelope struct {
		Data struct {
			Schema coverage.GraphQLSchema `json:"__schema"`
		} `json:"data"`
		Schema coverage.GraphQLSchema `json:"__schema"`
		coverage.GraphQLSchema
	}
	if err := json.NewDecoder(f).Decode(&envelope); err != nil {
		return coverage.GraphQLSchema{}, fmt.Errorf("failed to parse GraphQL schema %q: %w", path, err)
	}
	switch {
	case len(envelope.Data.Schema.Types) > 0:
		return envelope.Data.Schema, nil
	case len(envelope.Schema.Types) > 0:
		return envelope.Schema, nil
	case len(envelope.GraphQLSchema.Types) > 0:
		return envelope.GraphQLSchema, nil
	default:
		return coverage.GraphQLSchema{}, fmt.Errorf("GraphQL schema %q did not contain __schema data", path)
	}
}
