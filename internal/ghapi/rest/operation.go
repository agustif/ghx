package rest

import (
	"fmt"
	"sort"
	"strings"
)

// DefaultAPIVersion is the REST API version used by the current ghx HTTP client.
const DefaultAPIVersion = "2022-11-28"

// OperationJSONFields is the stable JSON field allowlist for generated REST
// operation metadata.
var OperationJSONFields = []string{
	"apiVersion",
	"coverage",
	"docsUrl",
	"method",
	"operationId",
	"pagination",
	"parameters",
	"path",
	"rawCommand",
	"scopes",
	"source",
	"summary",
	"tag",
}

// Parameter describes one generated REST operation parameter.
type Parameter struct {
	Name        string   `json:"name"`
	In          string   `json:"in"`
	Type        string   `json:"type,omitempty"`
	Required    bool     `json:"required"`
	Enum        []string `json:"enum,omitempty"`
	Description string   `json:"description,omitempty"`
	Example     string   `json:"example,omitempty"`
}

// Pagination describes the pagination contract for a REST operation.
type Pagination struct {
	Style        string `json:"style"`
	PageParam    string `json:"pageParam,omitempty"`
	PerPageParam string `json:"perPageParam,omitempty"`
	MaxPerPage   int    `json:"maxPerPage,omitempty"`
	Notes        string `json:"notes,omitempty"`
}

// Coverage describes how much local ghx coverage exists for an operation.
type Coverage struct {
	State           string `json:"state"`
	LocalCommand    string `json:"localCommand,omitempty"`
	ProposedCommand string `json:"proposedCommand,omitempty"`
	Notes           string `json:"notes,omitempty"`
}

// Source describes where the generated metadata came from.
type Source struct {
	Name     string `json:"name"`
	Ref      string `json:"ref"`
	URL      string `json:"url"`
	Checksum string `json:"checksum"`
}

// Operation is the generated metadata contract for one GitHub REST operation.
type Operation struct {
	OperationID string      `json:"operationId"`
	Method      string      `json:"method"`
	Path        string      `json:"path"`
	Tag         string      `json:"tag"`
	Summary     string      `json:"summary"`
	DocsURL     string      `json:"docsUrl"`
	APIVersion  string      `json:"apiVersion"`
	Scopes      []string    `json:"scopes,omitempty"`
	Parameters  []Parameter `json:"parameters,omitempty"`
	Pagination  Pagination  `json:"pagination"`
	Coverage    Coverage    `json:"coverage"`
	Source      Source      `json:"source"`
}

// ExportData returns the selected stable JSON fields for cmdutil.Exporter.
func (o Operation) ExportData(fields []string) map[string]interface{} {
	if len(fields) == 0 {
		fields = OperationJSONFields
	}

	data := make(map[string]interface{}, len(fields))
	for _, field := range fields {
		switch field {
		case "apiVersion":
			data[field] = o.APIVersion
		case "coverage":
			data[field] = o.Coverage
		case "docsUrl":
			data[field] = o.DocsURL
		case "method":
			data[field] = o.Method
		case "operationId":
			data[field] = o.OperationID
		case "pagination":
			data[field] = o.Pagination
		case "parameters":
			data[field] = o.Parameters
		case "path":
			data[field] = o.Path
		case "rawCommand":
			data[field] = o.RawCommand()
		case "scopes":
			data[field] = o.Scopes
		case "source":
			data[field] = o.Source
		case "summary":
			data[field] = o.Summary
		case "tag":
			data[field] = o.Tag
		}
	}
	return data
}

// PathParameters returns required path parameters in source order.
func (o Operation) PathParameters() []Parameter {
	return o.parametersByLocation("path")
}

// QueryParameters returns query parameters in source order.
func (o Operation) QueryParameters() []Parameter {
	return o.parametersByLocation("query")
}

// BodyParameters returns JSON body parameters in source order.
func (o Operation) BodyParameters() []Parameter {
	return o.parametersByLocation("body")
}

// RawCommand returns the raw ghx api escape hatch for this operation.
func (o Operation) RawCommand() string {
	parts := []string{"ghx", "api"}
	if o.Method != "" && !strings.EqualFold(o.Method, "GET") {
		parts = append(parts, "-X", strings.ToUpper(o.Method))
	}
	parts = append(parts, strings.TrimPrefix(o.Path, "/"))
	if strings.EqualFold(o.Method, "GET") && o.Pagination.Style == "page" {
		parts = append(parts, "--paginate")
	}
	for _, param := range o.BodyParameters() {
		if !param.Required {
			continue
		}
		parts = append(parts, "-F", param.rawField())
	}
	return strings.Join(parts, " ")
}

func (o Operation) parametersByLocation(location string) []Parameter {
	var params []Parameter
	for _, param := range o.Parameters {
		if param.In == location {
			params = append(params, param)
		}
	}
	return params
}

func (p Parameter) rawField() string {
	value := p.Example
	if value == "" {
		value = fmt.Sprintf("<%s>", p.Type)
	}
	if p.Type == "array" {
		return fmt.Sprintf("%s[]=%s", p.Name, value)
	}
	return fmt.Sprintf("%s=%s", p.Name, value)
}

// OperationFilter selects operations from the registry.
type OperationFilter struct {
	Tag      string
	Coverage string
	Query    string
}

// AllOperations returns all generated REST operation metadata in operation id order.
func AllOperations() []Operation {
	operations := append([]Operation(nil), operations...)
	sort.Slice(operations, func(i, j int) bool {
		return operations[i].OperationID < operations[j].OperationID
	})
	return operations
}

// FindOperations returns generated REST operations matching a filter.
func FindOperations(filter OperationFilter) []Operation {
	var matches []Operation
	query := strings.ToLower(filter.Query)
	for _, operation := range AllOperations() {
		if filter.Tag != "" && !strings.EqualFold(operation.Tag, filter.Tag) {
			continue
		}
		if filter.Coverage != "" && !strings.EqualFold(operation.Coverage.State, filter.Coverage) {
			continue
		}
		if query != "" && !operation.matchesQuery(query) {
			continue
		}
		matches = append(matches, operation)
	}
	return matches
}

// LookupOperation finds one generated REST operation by operation id.
func LookupOperation(operationID string) (Operation, bool) {
	for _, operation := range operations {
		if operation.OperationID == operationID {
			return operation, true
		}
	}
	return Operation{}, false
}

func (o Operation) matchesQuery(query string) bool {
	haystack := strings.ToLower(strings.Join([]string{
		o.OperationID,
		o.Method,
		o.Path,
		o.Tag,
		o.Summary,
		o.Coverage.State,
		o.Coverage.LocalCommand,
		o.Coverage.ProposedCommand,
	}, " "))
	return strings.Contains(haystack, query)
}
