package security

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

func TestNewCmdInbox(t *testing.T) {
	tests := []struct {
		name     string
		args     string
		wantOpts InboxOptions
		wantErr  string
	}{
		{
			name: "defaults",
			args: "github",
			wantOpts: InboxOptions{
				Org:   "github",
				Limit: 30,
				State: "open",
				Kinds: []string{"code-scanning", "secret-scanning", "dependabot"},
			},
		},
		{
			name: "kind and limit",
			args: "github --kind code-scanning,secret-scanning --limit 5 --state all",
			wantOpts: InboxOptions{
				Org:   "github",
				Limit: 5,
				State: "all",
				Kinds: []string{"code-scanning", "secret-scanning"},
			},
		},
		{
			name: "secret scanning resolved state",
			args: "github --kind secret-scanning --state resolved",
			wantOpts: InboxOptions{
				Org:   "github",
				Limit: 30,
				State: "resolved",
				Kinds: []string{"secret-scanning"},
			},
		},
		{
			name:    "reject state unsupported by selected kind",
			args:    "github --kind secret-scanning --state fixed",
			wantErr: "`--state fixed` is not supported for `--kind secret-scanning`",
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
			f := &cmdutil.Factory{IOStreams: ios}
			var gotOpts *InboxOptions
			cmd := NewCmdInbox(f, func(opts *InboxOptions) error {
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
			assert.Equal(t, tt.wantOpts.State, gotOpts.State)
			assert.Equal(t, tt.wantOpts.Kinds, gotOpts.Kinds)
		})
	}
}

func TestInboxRunJSON(t *testing.T) {
	reg := &httpmock.Registry{}
	defer reg.Verify(t)
	reg.Register(
		httpmock.REST("GET", "orgs/github/code-scanning/alerts"),
		httpmock.StringResponse(`[
			{
				"number": 7,
				"state": "open",
				"html_url": "https://github.com/github/cli/security/code-scanning/7",
				"created_at": "2026-05-19T00:00:00Z",
				"updated_at": "2026-05-19T01:00:00Z",
				"repository": {"full_name": "github/cli"},
				"rule": {"id": "js/sql", "security_severity_level": "high", "description": "SQL injection"}
			}
		]`),
	)
	reg.Register(
		httpmock.REST("GET", "orgs/github/secret-scanning/alerts"),
		httpmock.StringResponse(`[
			{
				"number": 9,
				"state": "open",
				"html_url": "https://github.com/github/cli/security/secret-scanning/9",
				"created_at": "2026-05-18T00:00:00Z",
				"updated_at": "2026-05-18T01:00:00Z",
				"repository": {"full_name": "github/cli"},
				"secret_type": "github_pat"
			}
		]`),
	)
	reg.Register(
		httpmock.REST("GET", "orgs/github/dependabot/alerts"),
		httpmock.StringResponse(`[
			{
				"number": 11,
				"state": "open",
				"html_url": "https://github.com/github/cli/security/dependabot/11",
				"created_at": "2026-05-17T00:00:00Z",
				"updated_at": "2026-05-17T01:00:00Z",
				"repository": {"full_name": "github/cli"},
				"security_advisory": {"summary": "vulnerable dependency", "severity": "critical"}
			}
		]`),
	)

	ios, _, stdout, _ := iostreams.Test()
	exporter := cmdutil.NewJSONExporter()
	exporter.SetFields([]string{"kind", "repository", "number", "severity", "title"})
	opts := &InboxOptions{
		IO:         ios,
		Config:     func() (gh.Config, error) { return config.NewBlankConfig(), nil },
		HttpClient: func() (*http.Client, error) { return &http.Client{Transport: reg}, nil },
		Org:        "github",
		Limit:      3,
		State:      "open",
		Kinds:      []string{"code-scanning", "secret-scanning", "dependabot"},
		Exporter:   exporter,
	}

	err := inboxRun(opts)
	require.NoError(t, err)
	assert.JSONEq(t, `[
		{"kind":"code-scanning","repository":"github/cli","number":7,"severity":"high","title":"SQL injection"},
		{"kind":"secret-scanning","repository":"github/cli","number":9,"severity":"secret","title":"github_pat"},
		{"kind":"dependabot","repository":"github/cli","number":11,"severity":"critical","title":"vulnerable dependency"}
	]`, stdout.String())
}
