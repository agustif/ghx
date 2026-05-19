package version

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewCmdVersion(f *cmdutil.Factory, version, buildDate string) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "version",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprint(f.IOStreams.Out, cmd.Root().Annotations["versionInfo"])
			return nil
		},
	}

	cmdutil.DisableAuthCheck(cmd)

	return cmd
}

func Format(version, buildDate string) string {
	return FormatForCommand("gh", version, buildDate)
}

func FormatForCommand(commandName, version, buildDate string) string {
	version = strings.TrimPrefix(version, "v")
	if commandName == "" {
		commandName = "gh"
	}

	var dateStr string
	if buildDate != "" {
		dateStr = fmt.Sprintf(" (%s)", buildDate)
	}

	return fmt.Sprintf("%s version %s%s\n%s\n", commandName, version, dateStr, changelogURL(commandName, version))
}

func changelogURL(commandName, version string) string {
	path := fmt.Sprintf("https://github.com/%s", releaseRepository(commandName))
	r := regexp.MustCompile(`^v?\d+\.\d+\.\d+(-[\w.]+)?$`)
	if !r.MatchString(version) {
		return fmt.Sprintf("%s/releases/latest", path)
	}

	url := fmt.Sprintf("%s/releases/tag/v%s", path, strings.TrimPrefix(version, "v"))
	return url
}

func releaseRepository(commandName string) string {
	if commandName == "ghx" {
		return "agustif/ghx"
	}
	return "cli/cli"
}
