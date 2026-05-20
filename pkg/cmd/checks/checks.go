package checks

import (
	"github.com/MakeNowJust/heredoc"
	cmdInventory "github.com/cli/cli/v2/pkg/cmd/checks/inventory"
	cmdRerun "github.com/cli/cli/v2/pkg/cmd/checks/rerun"
	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// NewCmdChecks builds the top-level checks operations command.
func NewCmdChecks(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "checks <command>",
		Short:   "Inventory and plan check operations",
		Long:    "Inventory pull request checks and plan safe rerun operations.",
		GroupID: "actions",
		Example: heredoc.Doc(`
			# List failing checks on open pull requests
			$ gh checks inventory --bucket fail --json pullRequest,name,state,detailsUrl,rerunSupported

			# Plan a targeted rerun without mutating remote state
			$ gh checks rerun --name lint --dry-run --json targets,targetCount,runnableCount
		`),
	}
	cmdutil.EnableRepoOverride(cmd, f)

	cmd.AddCommand(cmdInventory.NewCmdInventory(f, nil))
	cmd.AddCommand(cmdRerun.NewCmdRerun(f, nil))

	return cmd
}
