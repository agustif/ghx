package rest

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLookupOperation(t *testing.T) {
	operation, ok := LookupOperation("checks/list-for-ref")
	require.True(t, ok)

	assert.Equal(t, "GET", operation.Method)
	assert.Equal(t, "/repos/{owner}/{repo}/commits/{ref}/check-runs", operation.Path)
	assert.Equal(t, "checks", operation.Tag)
	assert.Equal(t, "page", operation.Pagination.Style)
	assert.Equal(t, "ghx api repos/{owner}/{repo}/commits/{ref}/check-runs --paginate", operation.RawCommand())
}

func TestLookupOperationUnknown(t *testing.T) {
	_, ok := LookupOperation("actions/definitely-not-real")
	assert.False(t, ok)
}

func TestAllOperationsSorted(t *testing.T) {
	operations := AllOperations()
	require.NotEmpty(t, operations)

	for i := 1; i < len(operations); i++ {
		assert.LessOrEqual(t, operations[i-1].OperationID, operations[i].OperationID)
	}
}

func TestFindOperations(t *testing.T) {
	operations := FindOperations(OperationFilter{
		Tag:      "actions",
		Coverage: "missing",
		Query:    "hosted runners",
	})

	require.Len(t, operations, 2)
	assert.Equal(t, "actions/get-hosted-runners-limits-for-org", operations[0].OperationID)
	assert.Equal(t, "actions/list-hosted-runners-for-org", operations[1].OperationID)
}

func TestOperationExportData(t *testing.T) {
	operation, ok := LookupOperation("repos/create-deployment-status")
	require.True(t, ok)

	data := operation.ExportData([]string{"operationId", "method", "path", "rawCommand"})

	assert.Equal(t, map[string]interface{}{
		"operationId": "repos/create-deployment-status",
		"method":      "POST",
		"path":        "/repos/{owner}/{repo}/deployments/{deployment_id}/statuses",
		"rawCommand":  "ghx api -X POST repos/{owner}/{repo}/deployments/{deployment_id}/statuses -F state=success",
	}, data)
}

func TestOperationExportDataJSON(t *testing.T) {
	operation, ok := LookupOperation("actions/list-jobs-for-workflow-run")
	require.True(t, ok)

	b, err := json.Marshal(operation.ExportData([]string{"operationId", "pagination", "source"}))
	require.NoError(t, err)

	assert.Contains(t, string(b), `"operationId":"actions/list-jobs-for-workflow-run"`)
	assert.Contains(t, string(b), `"style":"page"`)
	assert.Contains(t, string(b), sourceRef)
	assert.Contains(t, string(b), sourceChecksum)
}
