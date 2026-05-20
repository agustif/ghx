package access

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/cli/cli/v2/api"
	"github.com/cli/cli/v2/internal/gh"
	"github.com/cli/cli/v2/internal/ghrepo"
	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/spf13/cobra"
)

var accessWhyFields = []string{
	"login",
	"repository",
	"host",
	"permission",
	"roleName",
	"requiredPermission",
	"decision",
	"blocker",
	"sources",
	"nextAction",
}

type WhyOptions struct {
	IO         *iostreams.IOStreams
	Config     func() (gh.Config, error)
	HttpClient func() (*http.Client, error)

	Username           string
	Repository         string
	RequiredPermission string
	Exporter           cmdutil.Exporter
}

type accessWhyResponse struct {
	Permission string `json:"permission"`
	RoleName   string `json:"role_name"`
	User       struct {
		Login string `json:"login"`
	} `json:"user"`
}

type AccessDiagnosis struct {
	Login              string
	Repository         string
	Host               string
	Permission         string
	RoleName           string
	RequiredPermission string
	Decision           string
	Blocker            string
	Sources            []string
	NextAction         string
}

func (d AccessDiagnosis) ExportData(fields []string) map[string]interface{} {
	return cmdutil.StructExportData(d, fields)
}

func NewCmdAccess(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "access <command>",
		Short: "Diagnose organization and repository access",
		Long: heredoc.Doc(`
			Diagnose organization and repository access using read-only GitHub APIs.
		`),
	}
	cmd.AddCommand(NewCmdWhy(f, nil))
	return cmd
}

func NewCmdWhy(f *cmdutil.Factory, runF func(*WhyOptions) error) *cobra.Command {
	opts := &WhyOptions{
		IO:                 f.IOStreams,
		Config:             f.Config,
		HttpClient:         f.HttpClient,
		RequiredPermission: "write",
	}

	cmd := &cobra.Command{
		Use:   "why <user> --repo OWNER/REPO",
		Short: "Explain a user's repository permission",
		Long: heredoc.Doc(`
			Explain the permission a user currently has on a repository.

			This command is read-only. It uses GitHub's repository collaborator
			permission endpoint and reports whether the observed permission satisfies
			the requested access level.
		`),
		Example: heredoc.Doc(`
			$ gh org access why monalisa --repo cli/cli
			$ gh org access why monalisa --repo cli/cli --permission maintain
			$ gh org access why monalisa --repo cli/cli --json login,permission,decision,blocker
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Username = args[0]
			if opts.Repository == "" {
				return cmdutil.FlagErrorf("required flag `--repo` not specified")
			}
			if runF != nil {
				return runF(opts)
			}
			return whyRun(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Repository, "repo", "R", "", "Repository to diagnose in OWNER/REPO format")
	cmdutil.StringEnumFlag(cmd, &opts.RequiredPermission, "permission", "", opts.RequiredPermission, []string{"read", "triage", "write", "maintain", "admin"}, "Permission required for the operation")
	cmdutil.AddJSONFlags(cmd, &opts.Exporter, accessWhyFields)

	return cmd
}

func whyRun(opts *WhyOptions) error {
	httpClient, err := opts.HttpClient()
	if err != nil {
		return err
	}
	cfg, err := opts.Config()
	if err != nil {
		return err
	}
	host, _ := cfg.Authentication().DefaultHost()

	repo, err := ghrepo.FromFullNameWithHost(opts.Repository, host)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("repos/%s/%s/collaborators/%s/permission",
		url.PathEscape(repo.RepoOwner()),
		url.PathEscape(repo.RepoName()),
		url.PathEscape(opts.Username),
	)

	var response accessWhyResponse
	client := api.NewClientFromHTTP(httpClient)
	if err := client.REST(repo.RepoHost(), "GET", path, nil, &response); err != nil {
		return err
	}

	diagnosis := buildAccessDiagnosis(repo, opts.Username, opts.RequiredPermission, response)
	if opts.Exporter != nil {
		return opts.Exporter.Write(opts.IO, diagnosis)
	}
	printAccessDiagnosis(opts.IO, diagnosis)
	return nil
}

func buildAccessDiagnosis(repo ghrepo.Interface, username, required string, response accessWhyResponse) AccessDiagnosis {
	login := response.User.Login
	if login == "" {
		login = username
	}

	permission := strings.ToLower(response.Permission)
	if permission == "" {
		permission = "none"
	}

	diagnosis := AccessDiagnosis{
		Login:              login,
		Repository:         ghrepo.FullName(repo),
		Host:               repo.RepoHost(),
		Permission:         permission,
		RoleName:           response.RoleName,
		RequiredPermission: required,
		Sources:            []string{"repository collaborator permission"},
	}
	if permissionRank(permission) >= permissionRank(required) {
		diagnosis.Decision = "allowed"
		diagnosis.NextAction = "no missing repository permission detected"
		return diagnosis
	}

	diagnosis.Decision = "blocked"
	diagnosis.Blocker = fmt.Sprintf("requires %s permission, observed %s", required, permission)
	diagnosis.NextAction = "check org role, team membership, outside collaborator access, SAML authorization, app installation, and token scopes"
	return diagnosis
}

func permissionRank(permission string) int {
	switch strings.ToLower(permission) {
	case "admin":
		return 5
	case "maintain":
		return 4
	case "write":
		return 3
	case "triage":
		return 2
	case "read":
		return 1
	default:
		return 0
	}
}

func printAccessDiagnosis(io *iostreams.IOStreams, d AccessDiagnosis) {
	fmt.Fprintf(io.Out, "%s on %s: %s\n", d.Login, d.Repository, d.Decision)
	fmt.Fprintf(io.Out, "permission: %s\n", d.Permission)
	if d.RoleName != "" {
		fmt.Fprintf(io.Out, "role: %s\n", d.RoleName)
	}
	fmt.Fprintf(io.Out, "required: %s\n", d.RequiredPermission)
	if d.Blocker != "" {
		fmt.Fprintf(io.Out, "blocker: %s\n", d.Blocker)
	}
	fmt.Fprintf(io.Out, "source: %s\n", strings.Join(d.Sources, ", "))
	fmt.Fprintf(io.Out, "next: %s\n", d.NextAction)
}
