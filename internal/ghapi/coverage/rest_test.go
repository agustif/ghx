package coverage

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildRESTReportMergesKnownMetadata(t *testing.T) {
	report, err := BuildRESTReport(strings.NewReader(restFixture), "fixture", RESTFilter{})
	require.NoError(t, err)

	assert.Equal(t, 2, report.TotalOperations)
	assert.Equal(t, 2, report.MatchingOperations)
	assert.Equal(t, 1, report.RegisteredOperations)
	assert.Equal(t, 1, report.MatchingRegisteredOperations)
	assert.Equal(t, 1, report.RemainingExplicitMetadataGap)
	assert.Equal(t, 50.0, report.ExplicitMetadataCoveragePct)

	require.Len(t, report.Operations, 2)
	known := report.Operations[0]
	assert.Equal(t, "checks/list-for-ref", known.OperationID)
	assert.True(t, known.Registered)
	assert.Equal(t, StateThin, known.CoverageState)
	assert.Equal(t, "ghx pr checks", known.LocalCommand)
	assert.Equal(t, "ghx api repos/{owner}/{repo}/commits/{ref}/check-runs --paginate", known.RawCommand)

	missing := report.Operations[1]
	assert.Equal(t, "orgs/list-foo-audit-events", missing.OperationID)
	assert.False(t, missing.Registered)
	assert.Equal(t, StateMissing, missing.CoverageState)
	assert.Equal(t, "ghx org", missing.ProposedCommand)
	assert.Equal(t, "ghx api orgs/{org}/foo-audit-events --paginate", missing.RawCommand)
}

func TestBuildRESTReportFiltersRows(t *testing.T) {
	report, err := BuildRESTReport(strings.NewReader(restFixture), "fixture", RESTFilter{
		Tag:   "orgs",
		State: StateMissing,
		Query: "audit",
	})
	require.NoError(t, err)

	require.Len(t, report.Operations, 1)
	assert.Equal(t, "orgs/list-foo-audit-events", report.Operations[0].OperationID)
	assert.Equal(t, 2, report.TotalOperations)
	assert.Equal(t, 1, report.MatchingOperations)
	assert.Equal(t, 1, report.RegisteredOperations)
}

func TestBuildRESTReportRejectsUnknownState(t *testing.T) {
	_, err := BuildRESTReport(strings.NewReader(restFixture), "fixture", RESTFilter{State: "nope"})
	require.EqualError(t, err, `unknown coverage state "nope"`)
}

const restFixture = `{
  "paths": {
    "/repos/{owner}/{repo}/commits/{ref}/check-runs": {
      "get": {
        "operationId": "checks/list-for-ref",
        "tags": ["checks"],
        "summary": "List check runs for a Git reference",
        "externalDocs": {
          "url": "https://docs.github.com/rest/checks/runs#list-check-runs-for-a-git-reference"
        },
        "parameters": [
          {"name": "per_page", "in": "query"},
          {"name": "page", "in": "query"}
        ]
      }
    },
    "/orgs/{org}/foo-audit-events": {
      "get": {
        "operationId": "orgs/list-foo-audit-events",
        "tags": ["orgs"],
        "summary": "List organization foo audit events",
        "externalDocs": {
          "url": "https://docs.github.com/rest/orgs/foo#list-organization-foo-audit-events"
        },
        "parameters": [
          {"name": "per_page", "in": "query"}
        ]
      }
    }
  }
}`
