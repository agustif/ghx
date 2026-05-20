package shared

import (
	"fmt"
	"strings"

	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/spf13/cobra"
)

// ResolveBodyFlags normalizes inline, file, and literal issue body flags.
func ResolveBodyFlags(cmd *cobra.Command, io *iostreams.IOStreams, body, bodyFile, bodyLiteral string) (string, bool, error) {
	bodyProvided := cmd.Flags().Changed("body")
	bodyFileProvided := bodyFile != ""
	bodyLiteralProvided := cmd.Flags().Changed("body-literal")

	if err := cmdutil.MutuallyExclusive(
		"specify only one of `--body`, `--body-file`, or `--body-literal`",
		bodyProvided,
		bodyFileProvided,
		bodyLiteralProvided,
	); err != nil {
		return "", false, err
	}

	switch {
	case bodyFileProvided:
		b, err := cmdutil.ReadFile(bodyFile, io.In)
		if err != nil {
			return "", false, err
		}
		return string(b), true, nil
	case bodyLiteralProvided:
		return bodyLiteral, true, nil
	case bodyProvided:
		WarnInlineBodyRisk(io, body)
		return body, true, nil
	default:
		return "", false, nil
	}
}

// WarnInlineBodyRisk points users at stdin or files for markdown that shells often mangle.
func WarnInlineBodyRisk(io *iostreams.IOStreams, body string) {
	if io == nil || !io.IsStderrTTY() {
		return
	}
	if strings.Contains(body, "\n") || strings.Contains(body, "`") || strings.Contains(body, "$(") {
		fmt.Fprintln(io.ErrOut, "warning: inline body text can be interpreted by your shell; prefer --body-file - for multi-line Markdown or code blocks")
	}
}
