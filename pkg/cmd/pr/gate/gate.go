package gate

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// NewCmdGate builds the PR merge gate command group.
func NewCmdGate(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gate <command>",
		Short: "Explain pull request merge gates",
		Long:  "Explain pull request merge readiness without mutating remote state.",
		Example: heredoc.Doc(`
			# Explain why a pull request is not merge-ready
			$ gh pr gate explain 42 --json pullRequest,ready,blockers,nextActions
		`),
	}

	cmd.AddCommand(NewCmdExplain(f, nil))
	return cmd
}
