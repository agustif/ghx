package coverage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildGraphQLReport(t *testing.T) {
	report, err := BuildGraphQLReport("github.com", GraphQLSchema{
		QueryType:    &GraphQLRootType{Name: "Query"},
		MutationType: &GraphQLRootType{Name: "Mutation"},
		Types: []GraphQLType{
			{
				Kind: "OBJECT",
				Name: "Query",
				Fields: []GraphQLMember{
					{Name: "repository"},
					{Name: "node"},
				},
			},
			{
				Kind: "OBJECT",
				Name: "Mutation",
				Fields: []GraphQLMember{
					{Name: "markPullRequestReadyForReview"},
				},
			},
			{
				Kind: "OBJECT",
				Name: "Repository",
				Fields: []GraphQLMember{
					{Name: "name"},
					{Name: "pullRequests"},
				},
			},
		},
	}, GraphQLFilter{})
	require.NoError(t, err)

	assert.Equal(t, "github.com", report.Host)
	assert.Equal(t, 3, report.TotalTypes)
	assert.Equal(t, 3, report.ObjectTypes)
	assert.Equal(t, 5, report.TotalFields)
	assert.Equal(t, 2, report.QueryFields)
	assert.Equal(t, 1, report.MutationFields)
	assert.Equal(t, 0, report.TrackedFields)
	assert.Equal(t, 3, report.MatchingFields)
	assert.Equal(t, 0.0, report.ExplicitCoveragePct)
	assert.Equal(t, 100.0, report.RemainingGapPct)
	require.Len(t, report.Fields, 3)
	assert.Equal(t, "mutation", report.Fields[0].Kind)
	assert.Equal(t, "markPullRequestReadyForReview", report.Fields[0].Name)
	assert.Equal(t, "ghx pr", report.Fields[0].ProposedCommand)
}

func TestBuildGraphQLReportFiltersFields(t *testing.T) {
	report, err := BuildGraphQLReport("github.com", GraphQLSchema{
		QueryType: &GraphQLRootType{Name: "Query"},
		Types: []GraphQLType{
			{
				Kind: "OBJECT",
				Name: "Query",
				Fields: []GraphQLMember{
					{Name: "repository"},
					{Name: "viewer"},
				},
			},
		},
	}, GraphQLFilter{State: StateMissing, Query: "repo"})
	require.NoError(t, err)

	require.Len(t, report.Fields, 1)
	assert.Equal(t, "repository", report.Fields[0].Name)
}
