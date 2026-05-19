package setupgit

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/cli/cli/v2/internal/gh"
	"github.com/cli/cli/v2/pkg/cmd/auth/shared/gitcredentials"
	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/spf13/cobra"
)

type gitCredentialsConfigurer interface {
	ConfigureOurs(hostname string) error
}

type SetupGitOptions struct {
	IO                      *iostreams.IOStreams
	Config                  func() (gh.Config, error)
	CommandName             string
	Hostname                string
	Force                   bool
	CredentialsHelperConfig gitCredentialsConfigurer
}

func NewCmdSetupGit(f *cmdutil.Factory, runF func(*SetupGitOptions) error) *cobra.Command {
	opts := &SetupGitOptions{
		IO:     f.IOStreams,
		Config: f.Config,
	}
	commandName := commandNameForExecutable(f.ExecutablePath)
	commandDisplayName := displayNameForCommand(commandName)
	opts.CommandName = commandName

	cmd := &cobra.Command{
		Use:   "setup-git",
		Short: fmt.Sprintf("Setup git with %s", commandDisplayName),
		Long: heredoc.Docf(`
			This command configures %[1]sgit%[1]s to use %[2]s as a credential helper.
			For more information on git credential helpers please reference:
			<https://git-scm.com/docs/gitcredentials>.

			By default, %[2]s will be set as the credential helper for all authenticated hosts.
			If there is no authenticated hosts the command fails with an error.

			Alternatively, use the %[1]s--hostname%[1]s flag to specify a single host to be configured.
			If the host is not authenticated with, the command fails with an error.
		`, "`", commandDisplayName),
		Example: heredoc.Docf(`
			# Configure git to use %[1]s as the credential helper for all authenticated hosts
			$ %[2]s auth setup-git

			# Configure git to use %[1]s as the credential helper for enterprise.internal host
			$ %[2]s auth setup-git --hostname enterprise.internal
		`, commandDisplayName, commandName),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.CredentialsHelperConfig = &gitcredentials.HelperConfig{
				SelfExecutablePath: f.ExecutablePath,
				GitClient:          f.GitClient,
			}
			if opts.Hostname == "" && opts.Force {
				return cmdutil.FlagErrorf("`--force` must be used in conjunction with `--hostname`")
			}
			if runF != nil {
				return runF(opts)
			}
			return setupGitRun(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Hostname, "hostname", "h", "", "The hostname to configure git for")
	cmd.Flags().BoolVarP(&opts.Force, "force", "f", false, "Force setup even if the host is not known. Must be used in conjunction with `--hostname`")

	return cmd
}

func setupGitRun(opts *SetupGitOptions) error {
	cfg, err := opts.Config()
	if err != nil {
		return err
	}
	authCfg := cfg.Authentication()
	hostnames := authCfg.Hosts()

	stderr := opts.IO.ErrOut
	cs := opts.IO.ColorScheme()
	commandName := commandNameOrDefault(opts.CommandName)

	// If a hostname was provided, we'll set up just that one
	if opts.Hostname != "" {
		if !opts.Force && !has(opts.Hostname, hostnames) {
			return fmt.Errorf("You are not logged into the GitHub host %q. Run %s to authenticate or provide `--force`",
				opts.Hostname,
				cs.Bold(fmt.Sprintf("%s auth login -h %s", commandName, opts.Hostname)),
			)
		}

		if err := opts.CredentialsHelperConfig.ConfigureOurs(opts.Hostname); err != nil {
			return fmt.Errorf("failed to set up git credential helper: %s", err)
		}

		return nil
	}

	// Otherwise we'll set up any known hosts
	if len(hostnames) == 0 {
		fmt.Fprintf(
			stderr,
			"You are not logged into any GitHub hosts. Run %s to authenticate.\n",
			cs.Bold(fmt.Sprintf("%s auth login", commandName)),
		)

		return cmdutil.SilentError
	}

	for _, hostname := range hostnames {
		if err := opts.CredentialsHelperConfig.ConfigureOurs(hostname); err != nil {
			return fmt.Errorf("failed to set up git credential helper: %s", err)
		}
	}

	return nil
}

func commandNameForExecutable(executablePath string) string {
	commandName := strings.TrimSuffix(filepath.Base(executablePath), ".exe")
	return commandNameOrDefault(commandName)
}

func commandNameOrDefault(commandName string) string {
	if commandName == "" || commandName == "." {
		return "gh"
	}
	return commandName
}

func displayNameForCommand(commandName string) string {
	if commandName == "gh" {
		return "GitHub CLI"
	}
	return commandName
}

func has(needle string, haystack []string) bool {
	for _, s := range haystack {
		if strings.EqualFold(s, needle) {
			return true
		}
	}
	return false
}
