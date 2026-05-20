package coverage

import (
	"fmt"
	"io"
	"strings"
)

// MarkdownOptions controls Markdown report rendering.
type MarkdownOptions struct {
	Detail bool
	Limit  int
}

// WriteMarkdown writes a Markdown coverage report.
func WriteMarkdown(w io.Writer, report Report, opts MarkdownOptions) error {
	if _, err := fmt.Fprintln(w, "# ghx GitHub API coverage"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "Generated from live source data and the local ghx coverage registry."); err != nil {
		return err
	}

	if report.Commands != nil {
		if err := writeCommandsMarkdown(w, report.Commands, opts); err != nil {
			return err
		}
	}
	if report.REST != nil {
		if err := writeRESTMarkdown(w, report.REST, opts); err != nil {
			return err
		}
	}
	if report.GraphQL != nil {
		if err := writeGraphQLMarkdown(w, report.GraphQL, opts); err != nil {
			return err
		}
	}
	return nil
}

func writeCommandsMarkdown(w io.Writer, report *CommandReport, opts MarkdownOptions) error {
	if _, err := fmt.Fprintln(w, "\n## Local Commands"); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "\nRoot: `%s`\n\n", mdEscape(report.Root)); err != nil {
		return err
	}
	rows := []struct {
		metric string
		value  interface{}
	}{
		{"Total commands", report.TotalCommands},
		{"Visible commands", report.VisibleCommands},
		{"Runnable commands", report.RunnableCommands},
		{"Commands with JSON fields", report.JSONCommands},
	}
	if _, err := fmt.Fprintln(w, "| Metric | Value |"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "| --- | ---: |"); err != nil {
		return err
	}
	for _, row := range rows {
		if _, err := fmt.Fprintf(w, "| %s | %v |\n", row.metric, row.value); err != nil {
			return err
		}
	}
	if opts.Detail {
		return writeCommandTable(w, report.Commands, opts.Limit)
	}
	return nil
}

func writeRESTMarkdown(w io.Writer, report *RESTReport, opts MarkdownOptions) error {
	if _, err := fmt.Fprintln(w, "\n## REST"); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "\nSource: `%s`\n\n", mdEscape(report.Source)); err != nil {
		return err
	}
	rows := []struct {
		metric string
		value  interface{}
	}{
		{"REST operations", report.TotalOperations},
		{"Matching operations", report.MatchingOperations},
		{"Explicit metadata entries", report.RegisteredOperations},
		{"Remaining explicit metadata gap", report.RemainingExplicitMetadataGap},
		{"Explicit metadata coverage", fmt.Sprintf("%.1f%%", report.ExplicitMetadataCoveragePct)},
		{"Remaining explicit metadata gap", fmt.Sprintf("%.1f%%", report.RemainingExplicitMetadataGapPct)},
	}
	if _, err := fmt.Fprintln(w, "| Metric | Value |"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "| --- | ---: |"); err != nil {
		return err
	}
	for _, row := range rows {
		if _, err := fmt.Fprintf(w, "| %s | %v |\n", row.metric, row.value); err != nil {
			return err
		}
	}
	if err := writeBucketTable(w, "Coverage States", "State", report.StateCounts); err != nil {
		return err
	}
	if err := writeBucketTable(w, "Top REST Tags", "Tag", limitBuckets(report.TagCounts, 25)); err != nil {
		return err
	}
	if opts.Detail {
		return writeRESTOperationTable(w, report.Operations, opts.Limit)
	}
	return nil
}

func writeGraphQLMarkdown(w io.Writer, report *GraphQLReport, opts MarkdownOptions) error {
	if _, err := fmt.Fprintln(w, "\n## GraphQL"); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "\nHost: `%s`\n\n", mdEscape(report.Host)); err != nil {
		return err
	}
	rows := []struct {
		metric string
		value  interface{}
	}{
		{"Schema hash", report.SchemaHash},
		{"Schema types", report.TotalTypes},
		{"Object types", report.ObjectTypes},
		{"Input types", report.InputTypes},
		{"Enum types", report.EnumTypes},
		{"Interface types", report.InterfaceTypes},
		{"Union types", report.UnionTypes},
		{"Total fields", report.TotalFields},
		{"Query fields", report.QueryFields},
		{"Mutation fields", report.MutationFields},
		{"Deprecated fields", report.DeprecatedFields},
		{"Tracked query and mutation fields", report.TrackedFields},
		{"Matching query and mutation fields", report.MatchingFields},
		{"Explicit coverage", fmt.Sprintf("%.1f%%", report.ExplicitCoveragePct)},
		{"Remaining explicit gap", fmt.Sprintf("%.1f%%", report.RemainingGapPct)},
	}
	if _, err := fmt.Fprintln(w, "| Metric | Value |"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "| --- | ---: |"); err != nil {
		return err
	}
	for _, row := range rows {
		if _, err := fmt.Fprintf(w, "| %s | %v |\n", row.metric, row.value); err != nil {
			return err
		}
	}
	if err := writeBucketTable(w, "GraphQL Coverage States", "State", report.StateCounts); err != nil {
		return err
	}
	if opts.Detail {
		return writeGraphQLFieldTable(w, report.Fields, opts.Limit)
	}
	return nil
}

func writeBucketTable(w io.Writer, title, column string, rows []BucketCount) error {
	if _, err := fmt.Fprintf(w, "\n### %s\n\n", title); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "| %s | Count |\n", column); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "| --- | ---: |"); err != nil {
		return err
	}
	for _, row := range rows {
		if _, err := fmt.Fprintf(w, "| `%s` | %d |\n", mdEscape(row.Name), row.Count); err != nil {
			return err
		}
	}
	return nil
}

func writeCommandTable(w io.Writer, rows []CommandRow, limit int) error {
	if _, err := fmt.Fprint(w, "\n### Commands\n\n"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "| Command | Runnable | JSON Fields |"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "| --- | --- | --- |"); err != nil {
		return err
	}
	for i, row := range rows {
		if limit > 0 && i >= limit {
			if _, err := fmt.Fprintf(w, "| _%d more not shown_ |  |  |\n", len(rows)-limit); err != nil {
				return err
			}
			break
		}
		if _, err := fmt.Fprintf(w, "| `%s` | `%t` | `%s` |\n",
			mdEscape(row.Path),
			row.Runnable,
			mdEscape(strings.Join(row.JSONFields, ", ")),
		); err != nil {
			return err
		}
	}
	return nil
}

func writeRESTOperationTable(w io.Writer, rows []RESTOperation, limit int) error {
	if _, err := fmt.Fprint(w, "\n### REST Operations\n\n"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "| Operation | Method | Path | Tag | State | Command |"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "| --- | --- | --- | --- | --- | --- |"); err != nil {
		return err
	}
	for i, row := range rows {
		if limit > 0 && i >= limit {
			if _, err := fmt.Fprintf(w, "| _%d more not shown_ |  |  |  |  |  |\n", len(rows)-limit); err != nil {
				return err
			}
			break
		}
		command := row.LocalCommand
		if command == "" {
			command = row.ProposedCommand
		}
		if _, err := fmt.Fprintf(w, "| `%s` | `%s` | `%s` | `%s` | `%s` | `%s` |\n",
			mdEscape(row.OperationID),
			mdEscape(row.Method),
			mdEscape(row.Path),
			mdEscape(row.Tag),
			mdEscape(row.CoverageState),
			mdEscape(command),
		); err != nil {
			return err
		}
	}
	return nil
}

func writeGraphQLFieldTable(w io.Writer, rows []GraphQLField, limit int) error {
	if _, err := fmt.Fprint(w, "\n### GraphQL Fields\n\n"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "| Kind | Field | Return | Args | State | Command |"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "| --- | --- | --- | --- | --- | --- |"); err != nil {
		return err
	}
	for i, row := range rows {
		if limit > 0 && i >= limit {
			if _, err := fmt.Fprintf(w, "| _%d more not shown_ |  |  |  |  |  |\n", len(rows)-limit); err != nil {
				return err
			}
			break
		}
		if _, err := fmt.Fprintf(w, "| `%s` | `%s` | `%s` | `%s` | `%s` | `%s` |\n",
			mdEscape(row.Kind),
			mdEscape(row.Coordinate),
			mdEscape(row.ReturnType),
			mdEscape(graphQLArgsSummary(row.Args)),
			mdEscape(row.CoverageState),
			mdEscape(row.ProposedCommand),
		); err != nil {
			return err
		}
	}
	return nil
}

func limitBuckets(rows []BucketCount, limit int) []BucketCount {
	if limit <= 0 || len(rows) <= limit {
		return rows
	}
	return rows[:limit]
}

func graphQLArgsSummary(args []GraphQLArgument) string {
	if len(args) == 0 {
		return ""
	}
	parts := make([]string, 0, len(args))
	for _, arg := range args {
		part := arg.Name
		if arg.Type != "" {
			part += ": " + arg.Type
		}
		parts = append(parts, part)
	}
	return strings.Join(parts, ", ")
}

func mdEscape(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "|", "\\|")
	return s
}
