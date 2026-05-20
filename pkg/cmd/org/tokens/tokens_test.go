package tokens

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/cli/cli/v2/internal/config"
	"github.com/cli/cli/v2/internal/gh"
	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/cli/v2/pkg/httpmock"
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/google/shlex"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCmdTokens(t *testing.T) {
	tests := []struct {
		name     string
		args     string
		wantOpts ListOptions
		wantErr  string
	}{
		{
			name: "defaults",
			args: "github",
			wantOpts: ListOptions{
				Org:   "github",
				Limit: 30,
			},
		},
		{
			name: "custom limit",
			args: "github --limit 5",
			wantOpts: ListOptions{
				Org:   "github",
				Limit: 5,
			},
		},
		{
			name:    "invalid limit",
			args:    "github --limit 0",
			wantErr: "invalid limit: 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ios, _, _, _ := iostreams.Test()
			f := &cmdutil.Factory{
				IOStreams: ios,
				Config:    func() (gh.Config, error) { return config.NewBlankConfig(), nil },
			}
			var gotOpts *ListOptions
			cmd := newCmdTokens(f, func(opts *ListOptions) error {
				gotOpts = opts
				return nil
			})
			argv, err := shlex.Split(tt.args)
			require.NoError(t, err)
			cmd.SetArgs(argv)
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})

			_, err = cmd.ExecuteC()
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Equal(t, tt.wantErr, err.Error())
				return
			}
			require.NoError(t, err)
			require.NotNil(t, gotOpts)
			assert.Equal(t, tt.wantOpts.Org, gotOpts.Org)
			assert.Equal(t, tt.wantOpts.Limit, gotOpts.Limit)
		})
	}
}

func TestTokensRunJSON(t *testing.T) {
	reg := &httpmock.Registry{}
	defer reg.Verify(t)
	reg.Register(
		httpmock.REST("GET", "orgs/github/personal-access-token-requests"),
		httpmock.StringResponse(`[
			{
				"id": 123,
				"reason": "CI access",
				"repository_selection": "selected",
				"token_name": "agent token",
				"token_expires_at": "2026-06-01T00:00:00Z",
				"token_last_used_at": "2026-05-19T00:00:00Z",
				"owner": {"login": "monalisa"}
			}
		]`),
	)

	ios, _, stdout, _ := iostreams.Test()
	exporter := cmdutil.NewJSONExporter()
	exporter.SetFields([]string{"id", "owner", "tokenName", "repositorySelection", "reason"})
	opts := &ListOptions{
		IO:         ios,
		Config:     func() (gh.Config, error) { return config.NewBlankConfig(), nil },
		HttpClient: func() (*http.Client, error) { return &http.Client{Transport: reg}, nil },
		Org:        "github",
		Limit:      1,
		Exporter:   exporter,
	}

	err := tokensRun(opts)
	require.NoError(t, err)
	assert.JSONEq(t, `[{"id":123,"owner":"monalisa","tokenName":"agent token","repositorySelection":"selected","reason":"CI access"}]`, stdout.String())
}

func TestReviewRunRequiresDryRun(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	f := &cmdutil.Factory{IOStreams: ios}
	cmd := newCmdReview(f, nil)
	cmd.SetArgs([]string{"github", "--request-id", "123", "--decision", "deny"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	_, err := cmd.ExecuteC()
	require.Error(t, err)
	assert.Equal(t, "`--dry-run` is required", err.Error())
}

func TestReviewRun(t *testing.T) {
	ios, _, stdout, _ := iostreams.Test()
	opts := &ReviewOptions{
		IO:        ios,
		Org:       "github",
		RequestID: 123,
		Decision:  "deny",
		DryRun:    true,
	}

	err := reviewRun(opts)
	require.NoError(t, err)
	assert.Equal(t, "Would deny fine-grained PAT request 123 in github\nendpoint: POST /orgs/github/personal-access-token-requests\nbody: {\"pat_request_ids\":[123],\"action\":\"deny\"}\ndry-run: true\nNo review was submitted.\n", stdout.String())
}
