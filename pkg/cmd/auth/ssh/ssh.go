package authssh

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/cli/cli/v2/git"
	"github.com/cli/cli/v2/internal/gh"
	"github.com/cli/cli/v2/pkg/cmd/auth/shared/sshidentity"
	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/spf13/cobra"
)

// SSHOptions contains dependencies and flags for auth ssh commands.
type SSHOptions struct {
	Config    func() (gh.Config, error)
	GitClient *git.Client
	IO        *iostreams.IOStreams

	Hostname string
	Username string
	Identity string
}

// NewCmdSSH creates auth ssh commands for account-specific SSH identities.
func NewCmdSSH(f *cmdutil.Factory) *cobra.Command {
	opts := &SSHOptions{
		Config:    f.Config,
		GitClient: f.GitClient,
		IO:        f.IOStreams,
	}

	cmd := &cobra.Command{
		Use:   "ssh <command>",
		Short: "Manage account-specific SSH identities",
		Long: heredoc.Doc(`
			Manage SSH identities for account-safe git remotes.

			SSH remotes authenticate with keys, not OAuth tokens. Link one key per
			GitHub account, then ghx can set repo-local core.sshCommand when you
			switch accounts inside an SSH-based checkout.
		`),
	}

	cmd.AddCommand(newCmdLink(opts))
	cmd.AddCommand(newCmdList(opts))
	cmd.AddCommand(newCmdSync(opts))
	cmd.AddCommand(newCmdUnlink(opts))

	return cmd
}

func newCmdLink(opts *SSHOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "link",
		Args:  cobra.NoArgs,
		Short: "Link an SSH identity to a GitHub account",
		Example: heredoc.Doc(`
			$ ghx auth ssh link --hostname github.com --user monalisa --identity ~/.ssh/id_ed25519_monalisa
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			return linkRun(opts)
		},
	}

	addHostUserFlags(cmd, opts)
	cmd.Flags().StringVar(&opts.Identity, "identity", "", "Path to the private SSH identity for this account")
	_ = cmd.MarkFlagRequired("identity")

	return cmd
}

func newCmdList(opts *SSHOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Args:  cobra.NoArgs,
		Short: "List linked SSH identities",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listRun(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Hostname, "hostname", "h", "", "Only list identities for this GitHub host")

	return cmd
}

func newCmdSync(opts *SSHOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Args:  cobra.NoArgs,
		Short: "Apply the active account SSH identity to the current repository",
		Example: heredoc.Doc(`
			$ ghx auth ssh sync
			$ ghx auth ssh sync --hostname github.com --user monalisa
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			return syncRun(opts)
		},
	}

	addHostUserFlags(cmd, opts)

	return cmd
}

func newCmdUnlink(opts *SSHOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unlink",
		Args:  cobra.NoArgs,
		Short: "Unlink an SSH identity from a GitHub account",
		Example: heredoc.Doc(`
			$ ghx auth ssh unlink --hostname github.com --user monalisa
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			return unlinkRun(opts)
		},
	}

	addHostUserFlags(cmd, opts)

	return cmd
}

func addHostUserFlags(cmd *cobra.Command, opts *SSHOptions) {
	cmd.Flags().StringVarP(&opts.Hostname, "hostname", "h", "", "The hostname of the GitHub instance")
	cmd.Flags().StringVarP(&opts.Username, "user", "u", "", "The GitHub account login")
}

func linkRun(opts *SSHOptions) error {
	cfg, hostname, username, err := resolveHostUser(opts)
	if err != nil {
		return err
	}

	identity := sshidentity.ExpandIdentityPath(opts.Identity)
	if _, err := os.Stat(identity); err != nil {
		return fmt.Errorf("could not read SSH identity %s: %w", identity, err)
	}

	if err := sshidentity.SetIdentity(cfg, hostname, username, identity); err != nil {
		return err
	}

	fmt.Fprintf(opts.IO.ErrOut, "%s Linked SSH identity for %s on %s: %s\n",
		opts.IO.ColorScheme().SuccessIcon(), username, hostname, identity)
	return nil
}

func unlinkRun(opts *SSHOptions) error {
	cfg, hostname, username, err := resolveHostUser(opts)
	if err != nil {
		return err
	}

	if err := sshidentity.ClearIdentity(cfg, hostname, username); err != nil {
		return err
	}

	fmt.Fprintf(opts.IO.ErrOut, "%s Unlinked SSH identity for %s on %s\n",
		opts.IO.ColorScheme().SuccessIcon(), username, hostname)
	return nil
}

func listRun(opts *SSHOptions) error {
	cfg, err := opts.Config()
	if err != nil {
		return err
	}

	authCfg := cfg.Authentication()
	hosts := authCfg.Hosts()
	if opts.Hostname != "" {
		hosts = []string{opts.Hostname}
	}
	sort.Strings(hosts)

	for _, host := range hosts {
		users := usersForHost(authCfg, host)
		sort.Strings(users)
		if len(users) == 0 {
			continue
		}

		fmt.Fprintln(opts.IO.Out, host)
		for _, user := range users {
			identity := sshidentity.Identity(cfg, host, user)
			if identity == "" {
				identity = "(none)"
			}
			fmt.Fprintf(opts.IO.Out, "  %s: %s\n", user, identity)
		}
	}

	return nil
}

func usersForHost(authCfg gh.AuthConfig, host string) []string {
	seen := map[string]struct{}{}
	var users []string
	for _, user := range authCfg.UsersForHost(host) {
		if strings.TrimSpace(user) == "" {
			continue
		}
		if _, ok := seen[user]; ok {
			continue
		}
		seen[user] = struct{}{}
		users = append(users, user)
	}
	if activeUser, err := authCfg.ActiveUser(host); err == nil && strings.TrimSpace(activeUser) != "" {
		if _, ok := seen[activeUser]; !ok {
			users = append(users, activeUser)
		}
	}
	return users
}

func syncRun(opts *SSHOptions) error {
	cfg, hostname, username, err := resolveHostUser(opts)
	if err != nil {
		return err
	}

	identity := sshidentity.Identity(cfg, hostname, username)
	if identity == "" {
		return fmt.Errorf("no SSH identity linked for %s on %s; run: ghx auth ssh link --hostname %s --user %s --identity ~/.ssh/<key>", username, hostname, hostname, username)
	}

	synced, err := sshidentity.SyncRepoSSHCommand(context.Background(), opts.GitClient, hostname, identity)
	if err != nil {
		return err
	}
	if !synced {
		fmt.Fprintf(opts.IO.ErrOut, "%s No SSH remote for %s in the current repository\n",
			opts.IO.ColorScheme().WarningIcon(), hostname)
		return nil
	}

	fmt.Fprintf(opts.IO.ErrOut, "%s Synced SSH identity for %s on %s to repo-local core.sshCommand\n",
		opts.IO.ColorScheme().SuccessIcon(), username, hostname)
	return nil
}

func resolveHostUser(opts *SSHOptions) (gh.Config, string, string, error) {
	cfg, err := opts.Config()
	if err != nil {
		return nil, "", "", err
	}

	authCfg := cfg.Authentication()
	hostname := opts.Hostname
	if hostname == "" {
		hostname, _ = authCfg.DefaultHost()
	}
	if hostname == "" {
		return nil, "", "", fmt.Errorf("hostname is required")
	}

	username := opts.Username
	if username == "" {
		user, err := authCfg.ActiveUser(hostname)
		if err != nil {
			return nil, "", "", err
		}
		username = user
	}
	if strings.TrimSpace(username) == "" {
		return nil, "", "", fmt.Errorf("user is required")
	}
	token, _, err := authCfg.TokenForUser(hostname, username)
	if err != nil || token == "" {
		return nil, "", "", fmt.Errorf("not logged in to %s account %s", hostname, username)
	}

	return cfg, hostname, username, nil
}
