package api

import (
	"bytes"
	"testing"

	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCmdApiExplain(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		want     ExplainOptions
		wantErr  string
		wantJSON bool
	}{
		{
			name: "operation id",
			args: []string{"checks/list-for-ref"},
			want: ExplainOptions{
				OperationID: "checks/list-for-ref",
			},
		},
		{
			name: "list filters",
			args: []string{"--list", "--tag", "actions", "--coverage", "missing", "--search", "hosted"},
			want: ExplainOptions{
				List:     true,
				Tag:      "actions",
				Coverage: "missing",
				Query:    "hosted",
			},
		},
		{
			name:     "json exporter",
			args:     []string{"checks/list-for-ref", "--json", "operationId,method,path"},
			wantJSON: true,
			want: ExplainOptions{
				OperationID: "checks/list-for-ref",
			},
		},
		{
			name:    "missing operation id",
			args:    []string{},
			wantErr: "expected one operation id",
		},
		{
			name:    "operation id with list",
			args:    []string{"checks/list-for-ref", "--list"},
			wantErr: "operation id cannot be used with `--list`",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ios, _, _, _ := iostreams.Test()
			f := &cmdutil.Factory{IOStreams: ios}
			var got *ExplainOptions
			cmd := NewCmdApiExplain(f, func(opts *ExplainOptions) error {
				got = opts
				return nil
			})
			cmd.SetArgs(tt.args)
			cmd.SetIn(&bytes.Buffer{})
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})

			_, err := cmd.ExecuteC()
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, tt.want.OperationID, got.OperationID)
			assert.Equal(t, tt.want.List, got.List)
			assert.Equal(t, tt.want.Tag, got.Tag)
			assert.Equal(t, tt.want.Coverage, got.Coverage)
			assert.Equal(t, tt.want.Query, got.Query)
			assert.Equal(t, tt.wantJSON, got.Exporter != nil)
		})
	}
}

func TestApiExplainRunHuman(t *testing.T) {
	ios, _, stdout, _ := iostreams.Test()
	opts := &ExplainOptions{
		IO:          ios,
		OperationID: "checks/list-for-ref",
	}

	err := apiExplainRun(opts)

	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "Operation: checks/list-for-ref\n")
	assert.Contains(t, stdout.String(), "Method: GET\n")
	assert.Contains(t, stdout.String(), "Path params:\n")
	assert.Contains(t, stdout.String(), "Raw API: ghx api repos/{owner}/{repo}/commits/{ref}/check-runs --paginate\n")
	assert.Contains(t, stdout.String(), "Permissions/scopes:\n")
}

func TestApiExplainRunList(t *testing.T) {
	ios, _, stdout, _ := iostreams.Test()
	opts := &ExplainOptions{
		IO:       ios,
		List:     true,
		Tag:      "actions",
		Coverage: "missing",
		Query:    "hosted",
	}

	err := apiExplainRun(opts)

	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "actions/get-hosted-runners-limits-for-org")
	assert.Contains(t, stdout.String(), "actions/list-hosted-runners-for-org")
	assert.NotContains(t, stdout.String(), "checks/list-for-ref")
}

func TestApiExplainRunJSON(t *testing.T) {
	ios, _, stdout, _ := iostreams.Test()
	exporter := cmdutil.NewJSONExporter()
	exporter.SetFields([]string{"operationId", "method", "path", "rawCommand"})
	opts := &ExplainOptions{
		IO:          ios,
		Exporter:    exporter,
		OperationID: "repos/create-deployment-status",
	}

	err := apiExplainRun(opts)

	require.NoError(t, err)
	assert.JSONEq(t, `{
		"operationId": "repos/create-deployment-status",
		"method": "POST",
		"path": "/repos/{owner}/{repo}/deployments/{deployment_id}/statuses",
		"rawCommand": "ghx api -X POST repos/{owner}/{repo}/deployments/{deployment_id}/statuses -F state=success"
	}`, stdout.String())
}

func TestApiExplainRunUnknown(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	opts := &ExplainOptions{
		IO:          ios,
		OperationID: "actions/nope",
	}

	err := apiExplainRun(opts)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown GitHub REST operation")
}
