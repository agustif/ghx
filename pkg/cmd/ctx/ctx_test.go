package ctx

import (
	"bytes"
	"io"
	"testing"

	"github.com/cli/cli/v2/internal/config"
	"github.com/cli/cli/v2/internal/gh"
	"github.com/cli/cli/v2/internal/ghrepo"
	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testConfig(t *testing.T) gh.Config {
	t.Helper()
	t.Chdir(t.TempDir())
	t.Setenv("GH_ACCOUNT", "")
	t.Setenv("GH_ACCOUNT_SESSION", "")
	cfg := config.NewBlankConfig()
	_, err := cfg.Authentication().Login("github.com", "monalisa", "test-token", "ssh", false)
	require.NoError(t, err)
	return cfg
}

func TestExplainRunJSON(t *testing.T) {
	ios, _, stdout, _ := iostreams.Test()
	exporter := cmdutil.NewJSONExporter()
	exporter.SetFields([]string{"host", "login", "loginSource", "tokenSource", "repository", "branch", "warnings"})

	opts := &ExplainOptions{
		Config: func() (gh.Config, error) {
			return testConfig(t), nil
		},
		IO:       ios,
		BaseRepo: func() (ghrepo.Interface, error) { return ghrepo.New("OWNER", "REPO"), nil },
		Branch:   func() (string, error) { return "trunk", nil },
		Exporter: exporter,
	}

	require.NoError(t, explainRun(opts, false))
	assert.JSONEq(t, `{
		"host": "github.com",
		"login": "monalisa",
		"loginSource": "host",
		"tokenSource": "oauth_token",
		"repository": {
			"nameWithOwner": "OWNER/REPO",
			"host": "github.com"
		},
		"branch": "trunk",
		"warnings": null
	}`, stdout.String())
}

func TestDoctorWarnsForMissingAuth(t *testing.T) {
	ios, _, stdout, _ := iostreams.Test()
	t.Chdir(t.TempDir())
	t.Setenv("GH_ACCOUNT", "")
	t.Setenv("GH_ACCOUNT_SESSION", "")
	cfg := config.NewBlankConfig()

	opts := &ExplainOptions{
		Config: func() (gh.Config, error) {
			return cfg, nil
		},
		IO: ios,
	}

	err := explainRun(opts, true)
	require.ErrorIs(t, err, cmdutil.SilentError)
	assert.Contains(t, stdout.String(), "ghx context has blockers")
	assert.Contains(t, stdout.String(), "no active account is configured for github.com")
	assert.Contains(t, stdout.String(), "no active token is configured for github.com")
}

func TestNewCmdExplainParsesJSONFields(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	f := &cmdutil.Factory{IOStreams: ios}

	var got *ExplainOptions
	cmd := NewCmdExplain(f, func(opts *ExplainOptions) error {
		got = opts
		return nil
	})
	cmd.SetArgs([]string{"--json", "host,login"})
	cmd.SetIn(&bytes.Buffer{})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	_, err := cmd.ExecuteC()
	require.NoError(t, err)
	require.NotNil(t, got)
	require.NotNil(t, got.Exporter)
	assert.Equal(t, []string{"host", "login"}, got.Exporter.Fields())
}

func TestNewCmdCtxRepoOverrideIsAppliedBeforeRun(t *testing.T) {
	ios, _, stdout, _ := iostreams.Test()
	f := &cmdutil.Factory{
		IOStreams: ios,
		BaseRepo:  func() (ghrepo.Interface, error) { return ghrepo.New("AMBIENT", "REPO"), nil },
		Config: func() (gh.Config, error) {
			return testConfig(t), nil
		},
	}

	cmd := NewCmdCtx(f)
	cmd.SetArgs([]string{"explain", "--repo", "OWNER/OVERRIDE", "--json", "repository"})
	cmd.SetIn(&bytes.Buffer{})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	_, err := cmd.ExecuteC()
	require.NoError(t, err)
	assert.JSONEq(t, `{
		"repository": {
			"nameWithOwner": "OWNER/OVERRIDE",
			"host": "github.com"
		}
	}`, stdout.String())
}
