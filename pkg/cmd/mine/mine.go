package mine

import (
	"github.com/MakeNowJust/heredoc"
	githubCmd "github.com/cli/cli/v2/pkg/cmd/mine/github"
	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// NewCmdMine creates the `gh mine` command.
func NewCmdMine(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "mine <command>",
		Short:   "Generate ghx control-plane coverage maps",
		GroupID: "core",
		Long: heredoc.Doc(`
			Generate coverage maps for API and automation surfaces that ghx should
			make discoverable for agents and operators.
		`),
	}
	cmdutil.DisableAuthCheck(cmd)

	cmdutil.AddGroup(cmd, "Targets", githubCmd.NewCmdGithub(f, nil))

	return cmd
}
