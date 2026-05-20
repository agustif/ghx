package coverage

import (
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

// CommandReport describes the local Cobra command surface.
type CommandReport struct {
	Root             string       `json:"root"`
	TotalCommands    int          `json:"totalCommands"`
	VisibleCommands  int          `json:"visibleCommands"`
	RunnableCommands int          `json:"runnableCommands"`
	JSONCommands     int          `json:"jsonCommands"`
	Commands         []CommandRow `json:"commands"`
}

// CommandRow is one local command inventory row.
type CommandRow struct {
	Path       string   `json:"path"`
	Use        string   `json:"use"`
	Short      string   `json:"short,omitempty"`
	GroupID    string   `json:"groupId,omitempty"`
	Aliases    []string `json:"aliases,omitempty"`
	Hidden     bool     `json:"hidden"`
	Deprecated bool     `json:"deprecated"`
	Runnable   bool     `json:"runnable"`
	JSONFields []string `json:"jsonFields,omitempty"`
}

// BuildCommandReport walks a Cobra command tree into a stable inventory.
func BuildCommandReport(root *cobra.Command) *CommandReport {
	if root == nil {
		return nil
	}

	var rows []CommandRow
	walkCommands(root, &rows)
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].Path < rows[j].Path
	})

	report := &CommandReport{
		Root:          root.CommandPath(),
		TotalCommands: len(rows),
		Commands:      rows,
	}
	for _, row := range rows {
		if !row.Hidden {
			report.VisibleCommands++
		}
		if row.Runnable {
			report.RunnableCommands++
		}
		if len(row.JSONFields) > 0 {
			report.JSONCommands++
		}
	}
	return report
}

func walkCommands(cmd *cobra.Command, rows *[]CommandRow) {
	*rows = append(*rows, CommandRow{
		Path:       cmd.CommandPath(),
		Use:        cmd.Use,
		Short:      cmd.Short,
		GroupID:    cmd.GroupID,
		Aliases:    append([]string(nil), cmd.Aliases...),
		Hidden:     cmd.Hidden,
		Deprecated: cmd.Deprecated != "",
		Runnable:   cmd.Runnable(),
		JSONFields: commandJSONFields(cmd),
	})
	for _, child := range cmd.Commands() {
		walkCommands(child, rows)
	}
}

func commandJSONFields(cmd *cobra.Command) []string {
	if cmd.Annotations == nil {
		return nil
	}
	raw := cmd.Annotations["help:json-fields"]
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	fields := make([]string, 0, len(parts))
	for _, part := range parts {
		field := strings.TrimSpace(part)
		if field != "" {
			fields = append(fields, field)
		}
	}
	sort.Strings(fields)
	return fields
}
