package coverage

import "math"

const (
	// StateFirstClass means a purpose-built ghx command owns the workflow.
	StateFirstClass = "first-class"
	// StateThin means ghx has partial command coverage, but not full workflow coverage.
	StateThin = "thin"
	// StateRawAPI means ghx intentionally exposes the operation through raw api usage.
	StateRawAPI = "raw-api"
	// StateMissing means ghx has not classified or wrapped the operation yet.
	StateMissing = "missing"
)

// DefaultRESTOpenAPIURL is the pinned GitHub REST OpenAPI source used for github.com.
const DefaultRESTOpenAPIURL = "https://raw.githubusercontent.com/github/rest-api-description/133d385dfbee06825d4d4136a82dd2b4c79813ba/descriptions/api.github.com/api.github.com.json"

// Report is the generated GitHub API coverage report.
type Report struct {
	Commands *CommandReport `json:"commands,omitempty"`
	REST     *RESTReport    `json:"rest,omitempty"`
	GraphQL  *GraphQLReport `json:"graphql,omitempty"`
}

// BucketCount is a stable count row for tags, states, or other buckets.
type BucketCount struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// ValidCoverageState reports whether state is a known coverage state.
func ValidCoverageState(state string) bool {
	switch state {
	case "", StateFirstClass, StateThin, StateRawAPI, StateMissing:
		return true
	default:
		return false
	}
}

func percent(part, total int) float64 {
	if total == 0 {
		return 0
	}
	return math.Round((float64(part)/float64(total))*1000) / 10
}
