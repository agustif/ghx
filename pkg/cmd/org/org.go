package org

import (
	"github.com/MakeNowJust/heredoc"
	orgAccessCmd "github.com/cli/cli/v2/pkg/cmd/org/access"
	orgListCmd "github.com/cli/cli/v2/pkg/cmd/org/list"
	orgSecurityCmd "github.com/cli/cli/v2/pkg/cmd/org/security"
	orgTokensCmd "github.com/cli/cli/v2/pkg/cmd/org/tokens"
	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewCmdOrg(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "org <command>",
		Short: "Manage organizations",
		Long:  "Work with GitHub organizations.",
		Example: heredoc.Doc(`
			$ gh org list
		`),
		GroupID: "core",
	}

	cmdutil.AddGroup(cmd, "General commands", orgListCmd.NewCmdList(f, nil))
	cmdutil.AddGroup(cmd, "Diagnostics",
		orgAccessCmd.NewCmdAccess(f),
		orgSecurityCmd.NewCmdSecurity(f),
		orgTokensCmd.NewCmdTokens(f),
	)

	return cmd
}
