package access

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

func TestNewCmdWhy(t *testing.T) {
	tests := []struct {
		name     string
		args     string
		wantOpts WhyOptions
		wantErr  string
	}{
		{
			name:    "repo required",
			args:    "monalisa",
			wantErr: "required flag `--repo` not specified",
		},
		{
			name: "default required permission",
			args: "monalisa --repo OWNER/REPO",
			wantOpts: WhyOptions{
				Username:           "monalisa",
				Repository:         "OWNER/REPO",
				RequiredPermission: "write",
			},
		},
		{
			name: "custom required permission",
			args: "monalisa --repo OWNER/REPO --permission maintain",
			wantOpts: WhyOptions{
				Username:           "monalisa",
				Repository:         "OWNER/REPO",
				RequiredPermission: "maintain",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ios, _, _, _ := iostreams.Test()
			f := &cmdutil.Factory{IOStreams: ios}
			var gotOpts *WhyOptions
			cmd := NewCmdWhy(f, func(opts *WhyOptions) error {
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
			assert.Equal(t, tt.wantOpts.Username, gotOpts.Username)
			assert.Equal(t, tt.wantOpts.Repository, gotOpts.Repository)
			assert.Equal(t, tt.wantOpts.RequiredPermission, gotOpts.RequiredPermission)
		})
	}
}

func TestWhyRun(t *testing.T) {
	tests := []struct {
		name      string
		response  string
		required  string
		wantOut   string
		jsonField []string
	}{
		{
			name:     "allowed",
			required: "write",
			response: `{"permission":"maintain","role_name":"maintain","user":{"login":"monalisa"}}`,
			wantOut:  "monalisa on OWNER/REPO: allowed\npermission: maintain\nrole: maintain\nrequired: write\nsource: repository collaborator permission\nnext: no missing repository permission detected\n",
		},
		{
			name:     "blocked",
			required: "maintain",
			response: `{"permission":"read","role_name":"read","user":{"login":"monalisa"}}`,
			wantOut:  "monalisa on OWNER/REPO: blocked\npermission: read\nrole: read\nrequired: maintain\nblocker: requires maintain permission, observed read\nsource: repository collaborator permission\nnext: check org role, team membership, outside collaborator access, SAML authorization, app installation, and token scopes\n",
		},
		{
			name:      "json",
			required:  "write",
			response:  `{"permission":"admin","role_name":"admin","user":{"login":"monalisa"}}`,
			jsonField: []string{"login", "permission", "decision"},
			wantOut:   "{\"decision\":\"allowed\",\"login\":\"monalisa\",\"permission\":\"admin\"}\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := &httpmock.Registry{}
			defer reg.Verify(t)
			reg.Register(
				httpmock.REST("GET", "repos/OWNER/REPO/collaborators/monalisa/permission"),
				httpmock.StringResponse(tt.response),
			)

			ios, _, stdout, _ := iostreams.Test()
			opts := &WhyOptions{
				IO:                 ios,
				Config:             func() (gh.Config, error) { return config.NewBlankConfig(), nil },
				HttpClient:         func() (*http.Client, error) { return &http.Client{Transport: reg}, nil },
				Username:           "monalisa",
				Repository:         "OWNER/REPO",
				RequiredPermission: tt.required,
			}
			if tt.jsonField != nil {
				exporter := cmdutil.NewJSONExporter()
				exporter.SetFields(tt.jsonField)
				opts.Exporter = exporter
			}

			err := whyRun(opts)
			require.NoError(t, err)
			assert.Equal(t, tt.wantOut, stdout.String())
		})
	}
}
