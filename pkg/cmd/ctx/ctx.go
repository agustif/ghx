package ctx

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/cli/cli/v2/git"
	"github.com/cli/cli/v2/internal/gh"
	"github.com/cli/cli/v2/internal/ghrepo"
	"github.com/cli/cli/v2/pkg/cmd/auth/shared/sshidentity"
	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/spf13/cobra"
)

// ExplainOptions contains dependencies and flags for ctx explain.
type ExplainOptions struct {
	Config     func() (gh.Config, error)
	HttpClient func() (*http.Client, error)
	IO         *iostreams.IOStreams
	GitClient  *git.Client
	BaseRepo   func() (ghrepo.Interface, error)
	Branch     func() (string, error)
	Exporter   cmdutil.Exporter
}

type explainResult struct {
	Host          string       `json:"host"`
	HostSource    string       `json:"hostSource"`
	Login         string       `json:"login"`
	LoginSource   string       `json:"loginSource"`
	LoginSelector string       `json:"loginSelector,omitempty"`
	LoginScoped   bool         `json:"loginScoped"`
	TokenSource   string       `json:"tokenSource"`
	GitProtocol   string       `json:"gitProtocol"`
	SSHIdentity   string       `json:"sshIdentity,omitempty"`
	SSHRemote     bool         `json:"sshRemote,omitempty"`
	SSHCommand    string       `json:"sshCommand,omitempty"`
	Repository    *repoContext `json:"repository,omitempty"`
	Branch        string       `json:"branch,omitempty"`
	Warnings      []string     `json:"warnings,omitempty"`
}

func (r explainResult) ExportData(fields []string) map[string]interface{} {
	return cmdutil.StructExportData(r, fields)
}

type repoContext struct {
	NameWithOwner string `json:"nameWithOwner"`
	Host          string `json:"host"`
}

var explainFields = []string{
	"host",
	"hostSource",
	"login",
	"loginSource",
	"loginSelector",
	"loginScoped",
	"tokenSource",
	"gitProtocol",
	"sshIdentity",
	"sshRemote",
	"sshCommand",
	"repository",
	"branch",
	"warnings",
}

type activeUserInfoProvider interface {
	ActiveUserInfo(string) (gh.ActiveUserInfo, error)
}

// NewCmdCtx builds the ghx context command group.
func NewCmdCtx(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ctx <command>",
		Short:   "Explain the active ghx account and repository context",
		Long:    "Explain the active ghx account, host, token source, repository, and branch context before running account-sensitive commands.",
		GroupID: "core",
	}
	cmdutil.DisableAuthCheck(cmd)
	cmdutil.EnableRepoOverride(cmd, f)

	cmd.AddCommand(NewCmdExplain(f, nil))
	cmd.AddCommand(NewCmdDoctor(f, nil))

	return cmd
}

// NewCmdExplain creates the ctx explain command.
func NewCmdExplain(f *cmdutil.Factory, runF func(*ExplainOptions) error) *cobra.Command {
	opts := &ExplainOptions{
		Config:     f.Config,
		HttpClient: f.HttpClient,
		GitClient:  f.GitClient,
		IO:         f.IOStreams,
	}

	cmd := &cobra.Command{
		Use:   "explain",
		Short: "Show the active ghx account and repository context",
		Long: heredoc.Doc(`
			Show the active ghx account, host, token source, repository, and branch context.

			This command is read-only and is intended to be used before mutations or
			cross-repository automation.
		`),
		Example: heredoc.Doc(`
			$ ghx ctx explain
			$ ghx ctx explain --json host,login,loginSource,repository,warnings
		`),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.BaseRepo = f.BaseRepo
			opts.Branch = f.Branch
			if runF != nil {
				return runF(opts)
			}
			return explainRun(opts, false)
		},
	}

	cmdutil.AddJSONFlags(cmd, &opts.Exporter, explainFields)
	return cmd
}

// NewCmdDoctor creates the ctx doctor command.
func NewCmdDoctor(f *cmdutil.Factory, runF func(*ExplainOptions) error) *cobra.Command {
	opts := &ExplainOptions{
		Config:     f.Config,
		HttpClient: f.HttpClient,
		GitClient:  f.GitClient,
		IO:         f.IOStreams,
	}

	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check whether the active ghx context is ready for automation",
		Long: heredoc.Doc(`
			Check whether the active ghx host, account, repository, and branch context
			are specific enough for agent and automation workflows.
		`),
		Example: heredoc.Doc(`
			$ ghx ctx doctor
			$ ghx ctx doctor --json warnings,repository,loginSource
		`),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.BaseRepo = f.BaseRepo
			opts.Branch = f.Branch
			if runF != nil {
				return runF(opts)
			}
			return explainRun(opts, true)
		},
	}

	cmdutil.AddJSONFlags(cmd, &opts.Exporter, explainFields)
	return cmd
}

func explainRun(opts *ExplainOptions, doctor bool) error {
	result, err := buildExplainResult(opts)
	if err != nil {
		return err
	}

	if opts.Exporter != nil {
		return opts.Exporter.Write(opts.IO, result)
	}

	printExplainResult(opts.IO, result, doctor)
	if doctor && hasHardContextWarnings(result.Warnings) {
		return cmdutil.SilentError
	}
	return nil
}

func buildExplainResult(opts *ExplainOptions) (explainResult, error) {
	cfg, err := opts.Config()
	if err != nil {
		return explainResult{}, err
	}

	authCfg := cfg.Authentication()
	host, hostSource := authCfg.DefaultHost()
	result := explainResult{
		Host:        host,
		HostSource:  hostSource,
		GitProtocol: cfg.GitProtocol(host).Value,
	}

	if host == "" {
		result.Warnings = append(result.Warnings, "no default GitHub host is configured")
		return result, nil
	}

	if provider, ok := authCfg.(activeUserInfoProvider); ok {
		if info, err := provider.ActiveUserInfo(host); err == nil {
			result.Login = info.Username
			result.LoginSource = info.Source
			result.LoginSelector = info.Selector
			result.LoginScoped = info.Scoped
		}
	}
	if result.Login == "" {
		if user, err := authCfg.ActiveUser(host); err == nil {
			result.Login = user
			result.LoginSource = gh.ActiveUserSourceHost
		}
	}
	if result.Login == "" {
		result.Warnings = append(result.Warnings, fmt.Sprintf("no active account is configured for %s", host))
	}

	token, tokenSource := authCfg.ActiveToken(host)
	result.TokenSource = tokenSource
	if token == "" {
		result.Warnings = append(result.Warnings, fmt.Sprintf("no active token is configured for %s", host))
	}
	if result.LoginSource == gh.ActiveUserSourceHost && len(authCfg.UsersForHost(host)) > 1 {
		result.Warnings = append(result.Warnings, "active account is host-global while multiple accounts are configured; prefer GH_ACCOUNT, GH_ACCOUNT_SESSION, .ghaccount, or cwd scope")
	}
	if result.Login != "" {
		result.SSHIdentity = sshidentity.Identity(cfg, host, result.Login)
	}
	if opts.GitClient != nil {
		hasSSHRemote, err := sshidentity.HasSSHRemoteForHost(context.Background(), opts.GitClient, host)
		if err == nil {
			result.SSHRemote = hasSSHRemote
			result.SSHCommand = sshidentity.CoreSSHCommand(context.Background(), opts.GitClient)
			if hasSSHRemote && result.SSHIdentity == "" {
				result.Warnings = append(result.Warnings, fmt.Sprintf("SSH remote is configured for %s but no identity is linked for %s; run ghx auth ssh link", host, result.Login))
			}
			if hasSSHRemote && result.SSHIdentity != "" && !strings.Contains(result.SSHCommand, result.SSHIdentity) {
				result.Warnings = append(result.Warnings, "SSH remote identity is linked but repo-local core.sshCommand is not synced; run ghx auth ssh sync")
			}
		}
	}

	if opts.BaseRepo != nil {
		if repo, err := opts.BaseRepo(); err == nil {
			result.Repository = &repoContext{
				NameWithOwner: ghrepo.FullName(repo),
				Host:          repo.RepoHost(),
			}
			if repo.RepoHost() != "" && host != "" && !strings.EqualFold(repo.RepoHost(), host) {
				result.Warnings = append(result.Warnings, fmt.Sprintf("repository host %s differs from active host %s", repo.RepoHost(), host))
			}
		} else {
			result.Warnings = append(result.Warnings, fmt.Sprintf("repository context unavailable: %v", err))
		}
	}

	if opts.Branch != nil {
		if branch, err := opts.Branch(); err == nil {
			result.Branch = branch
		}
	}

	return result, nil
}

func printExplainResult(io *iostreams.IOStreams, result explainResult, doctor bool) {
	cs := io.ColorScheme()
	if doctor && hasHardContextWarnings(result.Warnings) {
		fmt.Fprintf(io.Out, "%s ghx context has blockers\n", cs.FailureIcon())
	} else if doctor {
		fmt.Fprintf(io.Out, "%s ghx context is ready\n", cs.SuccessIcon())
	} else {
		fmt.Fprintln(io.Out, "ghx context")
	}

	fmt.Fprintf(io.Out, "- Host: %s", result.Host)
	if result.HostSource != "" {
		fmt.Fprintf(io.Out, " (%s)", result.HostSource)
	}
	fmt.Fprintln(io.Out)

	fmt.Fprintf(io.Out, "- Account: %s", valueOrPlaceholder(result.Login))
	if result.LoginSource != "" {
		fmt.Fprintf(io.Out, " (%s)", result.LoginSource)
		if result.LoginSelector != "" {
			fmt.Fprintf(io.Out, ": %s", result.LoginSelector)
		}
	}
	fmt.Fprintln(io.Out)

	fmt.Fprintf(io.Out, "- Token source: %s\n", valueOrPlaceholder(result.TokenSource))
	fmt.Fprintf(io.Out, "- Git protocol: %s\n", valueOrPlaceholder(result.GitProtocol))
	if result.SSHRemote {
		fmt.Fprintf(io.Out, "- SSH identity: %s\n", valueOrPlaceholder(result.SSHIdentity))
		fmt.Fprintf(io.Out, "- SSH command: %s\n", valueOrPlaceholder(result.SSHCommand))
	}

	if result.Repository != nil {
		fmt.Fprintf(io.Out, "- Repository: %s (%s)\n", result.Repository.NameWithOwner, result.Repository.Host)
	}
	if result.Branch != "" {
		fmt.Fprintf(io.Out, "- Branch: %s\n", result.Branch)
	}
	if len(result.Warnings) > 0 {
		fmt.Fprintln(io.Out, "- Warnings:")
		for _, warning := range result.Warnings {
			fmt.Fprintf(io.Out, "  - %s\n", warning)
		}
	}
}

func valueOrPlaceholder(value string) string {
	if value == "" {
		return "(none)"
	}
	return value
}

func hasHardContextWarnings(warnings []string) bool {
	for _, warning := range warnings {
		if strings.HasPrefix(warning, "no active account") ||
			strings.HasPrefix(warning, "no active token") ||
			strings.HasPrefix(warning, "no default GitHub host") ||
			strings.HasPrefix(warning, "repository host") {
			return true
		}
	}
	return false
}
