package api

import (
	"fmt"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/cli/cli/v2/internal/ghapi/rest"
	"github.com/cli/cli/v2/internal/tableprinter"
	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/spf13/cobra"
)

// ExplainOptions captures options for `gh api explain`.
type ExplainOptions struct {
	IO       *iostreams.IOStreams
	Exporter cmdutil.Exporter

	OperationID string
	List        bool
	Tag         string
	Coverage    string
	Query       string
}

// NewCmdApiExplain creates the `gh api explain` command.
func NewCmdApiExplain(f *cmdutil.Factory, runF func(*ExplainOptions) error) *cobra.Command {
	opts := &ExplainOptions{
		IO: f.IOStreams,
	}

	cmd := &cobra.Command{
		Use:   "explain <operation-id>",
		Short: "Explain generated GitHub REST operation metadata",
		Long: heredoc.Doc(`
			Explain generated GitHub REST operation metadata and show the raw ghx api
			escape hatch for operations that do not have a first-class command yet.
		`),
		Example: heredoc.Doc(`
			# Show method, path, params, auth notes, pagination, and raw API command
			$ ghx api explain checks/list-for-ref

			# List known generated operations for Actions
			$ ghx api explain --list --tag actions

			# Emit stable JSON for scripts
			$ ghx api explain checks/list-for-ref --json operationId,method,path,rawCommand
		`),
		Args: func(cmd *cobra.Command, args []string) error {
			if opts.List {
				if len(args) != 0 {
					return cmdutil.FlagErrorf("operation id cannot be used with `--list`")
				}
				return nil
			}
			if len(args) != 1 {
				return cmdutil.FlagErrorf("expected one operation id")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if !opts.List {
				opts.OperationID = args[0]
			}
			if runF != nil {
				return runF(opts)
			}
			return apiExplainRun(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.List, "list", false, "List known generated REST operation metadata")
	cmd.Flags().StringVar(&opts.Tag, "tag", "", "Filter listed operations by REST tag")
	cmd.Flags().StringVar(&opts.Coverage, "coverage", "", "Filter listed operations by coverage state")
	cmd.Flags().StringVarP(&opts.Query, "search", "s", "", "Filter listed operations by keyword")
	cmdutil.AddJSONFlags(cmd, &opts.Exporter, rest.OperationJSONFields)

	return cmd
}

func apiExplainRun(opts *ExplainOptions) error {
	if opts.List {
		operations := rest.FindOperations(rest.OperationFilter{
			Tag:      opts.Tag,
			Coverage: opts.Coverage,
			Query:    opts.Query,
		})
		if opts.Exporter != nil {
			return opts.Exporter.Write(opts.IO, operations)
		}
		printOperationList(opts.IO, operations)
		return nil
	}

	operation, ok := rest.LookupOperation(opts.OperationID)
	if !ok {
		return fmt.Errorf("unknown GitHub REST operation %q; run `ghx api explain --list` to inspect known operations", opts.OperationID)
	}
	if opts.Exporter != nil {
		return opts.Exporter.Write(opts.IO, operation)
	}
	printOperation(opts.IO, operation)
	return nil
}

func printOperationList(ios *iostreams.IOStreams, operations []rest.Operation) {
	tp := tableprinter.New(ios, tableprinter.WithHeader("operation id", "method", "path", "coverage", "proposed command"))
	for _, operation := range operations {
		tp.AddField(operation.OperationID)
		tp.AddField(operation.Method)
		tp.AddField(operation.Path)
		tp.AddField(operation.Coverage.State)
		tp.AddField(operation.Coverage.ProposedCommand)
		tp.EndRow()
	}
	tp.Render()
}

func printOperation(ios *iostreams.IOStreams, operation rest.Operation) {
	fmt.Fprintf(ios.Out, "Operation: %s\n", operation.OperationID)
	fmt.Fprintf(ios.Out, "Summary: %s\n", operation.Summary)
	fmt.Fprintf(ios.Out, "Method: %s\n", operation.Method)
	fmt.Fprintf(ios.Out, "Path: %s\n", operation.Path)
	fmt.Fprintf(ios.Out, "API version: %s\n", operation.APIVersion)
	fmt.Fprintf(ios.Out, "Coverage: %s\n", coverageLine(operation.Coverage))
	fmt.Fprintf(ios.Out, "Raw API: %s\n", operation.RawCommand())
	if operation.DocsURL != "" {
		fmt.Fprintf(ios.Out, "Docs: %s\n", operation.DocsURL)
	}
	fmt.Fprintf(ios.Out, "Source: %s @ %s (%s)\n", operation.Source.Name, operation.Source.Ref, operation.Source.Checksum)

	printParameterGroup(ios, "Path params", operation.PathParameters())
	printParameterGroup(ios, "Query params", operation.QueryParameters())
	printParameterGroup(ios, "Body params", operation.BodyParameters())

	if len(operation.Scopes) > 0 {
		fmt.Fprintln(ios.Out, "Permissions/scopes:")
		for _, scope := range operation.Scopes {
			fmt.Fprintf(ios.Out, "  - %s\n", scope)
		}
	}

	fmt.Fprintln(ios.Out, "Pagination:")
	if operation.Pagination.Style == "none" {
		fmt.Fprintln(ios.Out, "  - none")
	} else {
		fmt.Fprintf(ios.Out, "  - %s", operation.Pagination.Style)
		if operation.Pagination.PerPageParam != "" {
			fmt.Fprintf(ios.Out, " via %s/%s", operation.Pagination.PageParam, operation.Pagination.PerPageParam)
		}
		if operation.Pagination.MaxPerPage > 0 {
			fmt.Fprintf(ios.Out, " max %d", operation.Pagination.MaxPerPage)
		}
		fmt.Fprintln(ios.Out)
		if operation.Pagination.Notes != "" {
			fmt.Fprintf(ios.Out, "  - %s\n", operation.Pagination.Notes)
		}
	}
}

func coverageLine(coverage rest.Coverage) string {
	var parts []string
	parts = append(parts, coverage.State)
	if coverage.LocalCommand != "" {
		parts = append(parts, "local "+coverage.LocalCommand)
	}
	if coverage.ProposedCommand != "" {
		parts = append(parts, "proposed "+coverage.ProposedCommand)
	}
	return strings.Join(parts, "; ")
}

func printParameterGroup(ios *iostreams.IOStreams, title string, params []rest.Parameter) {
	if len(params) == 0 {
		return
	}
	fmt.Fprintf(ios.Out, "%s:\n", title)
	for _, param := range params {
		required := "optional"
		if param.Required {
			required = "required"
		}
		fmt.Fprintf(ios.Out, "  - %s (%s %s)", param.Name, required, param.Type)
		if len(param.Enum) > 0 {
			fmt.Fprintf(ios.Out, " enum: %s", strings.Join(param.Enum, ", "))
		}
		if param.Description != "" {
			fmt.Fprintf(ios.Out, " - %s", param.Description)
		}
		fmt.Fprintln(ios.Out)
	}
}
