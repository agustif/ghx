package coverage

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/cli/cli/v2/internal/ghapi/rest"
)

var restMethods = map[string]bool{
	"delete":  true,
	"get":     true,
	"patch":   true,
	"post":    true,
	"put":     true,
	"head":    true,
	"options": true,
}

// RESTFilter selects REST operations in a generated report.
type RESTFilter struct {
	Tag   string
	State string
	Query string
}

// RESTReport describes generated REST API coverage.
type RESTReport struct {
	Source                          string          `json:"source"`
	TotalOperations                 int             `json:"totalOperations"`
	MatchingOperations              int             `json:"matchingOperations"`
	RegisteredOperations            int             `json:"registeredOperations"`
	MatchingRegisteredOperations    int             `json:"matchingRegisteredOperations"`
	RemainingExplicitMetadataGap    int             `json:"remainingExplicitMetadataGap"`
	ExplicitMetadataCoveragePct     float64         `json:"explicitMetadataCoveragePct"`
	RemainingExplicitMetadataGapPct float64         `json:"remainingExplicitMetadataGapPct"`
	StateCounts                     []BucketCount   `json:"stateCounts"`
	TagCounts                       []BucketCount   `json:"tagCounts"`
	Operations                      []RESTOperation `json:"operations"`
}

// RESTOperation is one generated REST coverage row.
type RESTOperation struct {
	OperationID     string `json:"operationId"`
	Method          string `json:"method"`
	Path            string `json:"path"`
	Tag             string `json:"tag"`
	Summary         string `json:"summary,omitempty"`
	DocsURL         string `json:"docsUrl,omitempty"`
	CoverageState   string `json:"coverageState"`
	Registered      bool   `json:"registered"`
	LocalCommand    string `json:"localCommand,omitempty"`
	ProposedCommand string `json:"proposedCommand,omitempty"`
	RawCommand      string `json:"rawCommand"`
	Pagination      string `json:"pagination"`
	Notes           string `json:"notes,omitempty"`
}

type openAPISpec struct {
	Paths map[string]map[string]json.RawMessage `json:"paths"`
}

type openAPIOperation struct {
	OperationID  string              `json:"operationId"`
	Tags         []string            `json:"tags"`
	Summary      string              `json:"summary"`
	ExternalDocs *openAPIExternalDoc `json:"externalDocs"`
	Parameters   []openAPIParameter  `json:"parameters"`
}

type openAPIExternalDoc struct {
	URL string `json:"url"`
}

type openAPIParameter struct {
	Name string `json:"name"`
	In   string `json:"in"`
}

// BuildRESTReport parses GitHub REST OpenAPI and merges local ghx metadata.
func BuildRESTReport(r io.Reader, source string, filter RESTFilter) (*RESTReport, error) {
	if !ValidCoverageState(filter.State) {
		return nil, fmt.Errorf("unknown coverage state %q", filter.State)
	}

	var spec openAPISpec
	if err := json.NewDecoder(r).Decode(&spec); err != nil {
		return nil, fmt.Errorf("failed to parse REST OpenAPI: %w", err)
	}
	if len(spec.Paths) == 0 {
		return nil, fmt.Errorf("REST OpenAPI did not contain any paths")
	}

	allOperations := restOperationsFromSpec(spec)
	registeredInSpec := 0
	for _, operation := range allOperations {
		if _, ok := rest.LookupOperation(operation.OperationID); ok {
			registeredInSpec++
		}
	}

	var matches []RESTOperation
	for _, operation := range allOperations {
		if !operationMatchesRESTFilter(operation, filter) {
			continue
		}
		matches = append(matches, operation)
	}

	matchingRegistered := 0
	for _, operation := range matches {
		if operation.Registered {
			matchingRegistered++
		}
	}

	return &RESTReport{
		Source:                          source,
		TotalOperations:                 len(allOperations),
		MatchingOperations:              len(matches),
		RegisteredOperations:            registeredInSpec,
		MatchingRegisteredOperations:    matchingRegistered,
		RemainingExplicitMetadataGap:    len(allOperations) - registeredInSpec,
		ExplicitMetadataCoveragePct:     percent(registeredInSpec, len(allOperations)),
		RemainingExplicitMetadataGapPct: percent(len(allOperations)-registeredInSpec, len(allOperations)),
		StateCounts:                     bucketCounts(matches, func(operation RESTOperation) string { return operation.CoverageState }),
		TagCounts:                       bucketCounts(matches, func(operation RESTOperation) string { return operation.Tag }),
		Operations:                      matches,
	}, nil
}

func restOperationsFromSpec(spec openAPISpec) []RESTOperation {
	var rows []RESTOperation
	for path, methods := range spec.Paths {
		for method, raw := range methods {
			method = strings.ToLower(method)
			if !restMethods[method] {
				continue
			}
			var specOperation openAPIOperation
			if err := json.Unmarshal(raw, &specOperation); err != nil || specOperation.OperationID == "" {
				continue
			}
			row := restOperationFromSpec(method, path, specOperation)
			rows = append(rows, row)
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].OperationID < rows[j].OperationID
	})
	return rows
}

func restOperationFromSpec(method, path string, specOperation openAPIOperation) RESTOperation {
	tag := ""
	if len(specOperation.Tags) > 0 {
		tag = specOperation.Tags[0]
	}
	docsURL := ""
	if specOperation.ExternalDocs != nil {
		docsURL = specOperation.ExternalDocs.URL
	}
	pagination := detectRESTPagination(specOperation.Parameters)
	row := RESTOperation{
		OperationID:     specOperation.OperationID,
		Method:          strings.ToUpper(method),
		Path:            path,
		Tag:             tag,
		Summary:         specOperation.Summary,
		DocsURL:         docsURL,
		CoverageState:   StateMissing,
		Registered:      false,
		ProposedCommand: proposedRESTCommand(tag),
		RawCommand:      rawRESTCommand(strings.ToUpper(method), path, pagination),
		Pagination:      pagination,
	}
	if known, ok := rest.LookupOperation(specOperation.OperationID); ok {
		row.Registered = true
		row.CoverageState = known.Coverage.State
		row.LocalCommand = known.Coverage.LocalCommand
		row.ProposedCommand = known.Coverage.ProposedCommand
		row.Notes = known.Coverage.Notes
		row.RawCommand = known.RawCommand()
		row.Pagination = known.Pagination.Style
		if row.Pagination == "" {
			row.Pagination = "none"
		}
	}
	return row
}

func detectRESTPagination(parameters []openAPIParameter) string {
	hasPage := false
	hasPerPage := false
	for _, param := range parameters {
		if param.In != "query" {
			continue
		}
		switch param.Name {
		case "page":
			hasPage = true
		case "per_page":
			hasPerPage = true
		}
	}
	if hasPage || hasPerPage {
		return "page"
	}
	return "none"
}

func rawRESTCommand(method, path, pagination string) string {
	parts := []string{"ghx", "api"}
	if method != "" && method != "GET" {
		parts = append(parts, "-X", method)
	}
	parts = append(parts, strings.TrimPrefix(path, "/"))
	if method == "GET" && pagination == "page" {
		parts = append(parts, "--paginate")
	}
	return strings.Join(parts, " ")
}

func proposedRESTCommand(tag string) string {
	switch tag {
	case "actions":
		return "ghx actions"
	case "checks":
		return "ghx checks"
	case "code-scanning", "code-security", "dependabot", "secret-scanning", "security-advisories":
		return "ghx sec"
	case "issues":
		return "ghx issue"
	case "orgs":
		return "ghx org"
	case "projects":
		return "ghx board"
	case "pulls":
		return "ghx pr"
	case "repos":
		return "ghx repo"
	case "teams":
		return "ghx team"
	case "agents", "agent-tasks":
		return "ghx agent"
	default:
		return "ghx api"
	}
}

func operationMatchesRESTFilter(operation RESTOperation, filter RESTFilter) bool {
	if filter.Tag != "" && !strings.EqualFold(filter.Tag, operation.Tag) {
		return false
	}
	if filter.State != "" && !strings.EqualFold(filter.State, operation.CoverageState) {
		return false
	}
	query := strings.ToLower(filter.Query)
	if query == "" {
		return true
	}
	haystack := strings.ToLower(strings.Join([]string{
		operation.OperationID,
		operation.Method,
		operation.Path,
		operation.Tag,
		operation.Summary,
		operation.CoverageState,
		operation.LocalCommand,
		operation.ProposedCommand,
	}, " "))
	return strings.Contains(haystack, query)
}

func bucketCounts[T any](items []T, key func(T) string) []BucketCount {
	counts := map[string]int{}
	for _, item := range items {
		name := key(item)
		if name == "" {
			name = "unknown"
		}
		counts[name]++
	}
	rows := make([]BucketCount, 0, len(counts))
	for name, count := range counts {
		rows = append(rows, BucketCount{Name: name, Count: count})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Count == rows[j].Count {
			return rows[i].Name < rows[j].Name
		}
		return rows[i].Count > rows[j].Count
	})
	return rows
}
