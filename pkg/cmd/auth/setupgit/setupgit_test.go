package setupgit

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/cli/cli/v2/internal/config"
	"github.com/cli/cli/v2/internal/gh"
	"github.com/cli/cli/v2/pkg/cmd/auth/shared/gitcredentials"
	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/google/shlex"
	"github.com/stretchr/testify/require"
)

type gitCredentialsConfigurerSpy struct {
	hosts    []string
	setupErr error
}

func (gf *gitCredentialsConfigurerSpy) ConfigureOurs(hostname string) error {
	gf.hosts = append(gf.hosts, hostname)
	return gf.setupErr
}

func TestNewCmdSetupGit(t *testing.T) {
	tests := []struct {
		name     string
		cli      string
		wantsErr bool
		errMsg   string
	}{
		{
			name:     "--force without hostname",
			cli:      "--force",
			wantsErr: true,
			errMsg:   "`--force` must be used in conjunction with `--hostname`",
		},
		{
			name:     "no error when --force used with hostname",
			cli:      "--force --hostname ghe.io",
			wantsErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &cmdutil.Factory{}

			argv, err := shlex.Split(tt.cli)
			require.NoError(t, err)

			cmd := NewCmdSetupGit(f, func(opts *SetupGitOptions) error {
				return nil
			})

			// TODO cobra hack-around
			cmd.Flags().BoolP("help", "x", false, "")

			cmd.SetArgs(argv)
			cmd.SetIn(&bytes.Buffer{})
			cmd.SetOut(io.Discard)
			cmd.SetErr(io.Discard)

			_, err = cmd.ExecuteC()
			if tt.wantsErr {
				require.Error(t, err)
				require.Equal(t, err.Error(), tt.errMsg)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestNewCmdSetupGitUsesExecutableName(t *testing.T) {
	f := &cmdutil.Factory{
		ExecutablePath: "/opt/homebrew/bin/ghx",
	}

	var gotCommandName string
	var gotSelfExecutablePath string

	cmd := NewCmdSetupGit(f, func(opts *SetupGitOptions) error {
		gotCommandName = opts.CommandName
		helperConfig, ok := opts.CredentialsHelperConfig.(*gitcredentials.HelperConfig)
		require.True(t, ok)
		gotSelfExecutablePath = helperConfig.SelfExecutablePath
		return nil
	})

	cmd.SetArgs([]string{"--hostname", "github.com", "--force"})
	cmd.SetIn(&bytes.Buffer{})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	// TODO cobra hack-around
	cmd.Flags().BoolP("help", "x", false, "")

	_, err := cmd.ExecuteC()
	require.NoError(t, err)
	require.Equal(t, "ghx", gotCommandName)
	require.Equal(t, "/opt/homebrew/bin/ghx", gotSelfExecutablePath)
}

func TestNewCmdSetupGitHelpUsesExecutableName(t *testing.T) {
	f := &cmdutil.Factory{
		ExecutablePath: "/path/to/ghx",
	}

	cmd := NewCmdSetupGit(f, nil)
	stdout := &bytes.Buffer{}
	cmd.SetOut(stdout)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--help"})
	// TODO cobra hack-around
	cmd.Flags().BoolP("help", "x", false, "")

	_, err := cmd.ExecuteC()
	require.NoError(t, err)

	help := stdout.String()
	require.Equal(t, "Setup git with ghx", cmd.Short)
	require.Contains(t, help, "This command configures `git` to use ghx as a credential helper.")
	require.Contains(t, help, "$ ghx auth setup-git")
	require.False(t, strings.Contains(help, "$ gh auth setup-git"), help)
}

func Test_setupGitRun(t *testing.T) {
	tests := []struct {
		name               string
		opts               *SetupGitOptions
		setupErr           error
		cfgStubs           func(*testing.T, gh.Config)
		expectedHostsSetup []string
		expectedErr        error
		expectedErrOut     string
	}{
		{
			name: "opts.Config returns an error",
			opts: &SetupGitOptions{
				Config: func() (gh.Config, error) {
					return nil, fmt.Errorf("oops")
				},
			},
			expectedErr: errors.New("oops"),
		},
		{
			name: "when an unknown hostname is provided without forcing, return an error",
			opts: &SetupGitOptions{
				Hostname: "ghe.io",
			},
			cfgStubs: func(t *testing.T, cfg gh.Config) {
				login(t, cfg, "github.com", "test-user", "gho_ABCDEFG", "https", false)
			},
			expectedErr: errors.New("You are not logged into the GitHub host \"ghe.io\". Run gh auth login -h ghe.io to authenticate or provide `--force`"),
		},
		{
			name: "when an unknown hostname is provided without forcing, use the configured command name",
			opts: &SetupGitOptions{
				CommandName: "ghx",
				Hostname:    "ghe.io",
			},
			cfgStubs: func(t *testing.T, cfg gh.Config) {
				login(t, cfg, "github.com", "test-user", "gho_ABCDEFG", "https", false)
			},
			expectedErr: errors.New("You are not logged into the GitHub host \"ghe.io\". Run ghx auth login -h ghe.io to authenticate or provide `--force`"),
		},
		{
			name: "when an unknown hostname is provided with forcing, set it up",
			opts: &SetupGitOptions{
				Hostname: "ghe.io",
				Force:    true,
			},
			expectedHostsSetup: []string{"ghe.io"},
		},
		{
			name: "when a known hostname is provided without forcing, set it up",
			opts: &SetupGitOptions{
				Hostname: "ghe.io",
			},
			cfgStubs: func(t *testing.T, cfg gh.Config) {
				login(t, cfg, "ghe.io", "test-user", "gho_ABCDEFG", "https", false)
			},
			expectedHostsSetup: []string{"ghe.io"},
		},
		{
			name: "when a hostname is provided but setting it up errors, that error is bubbled",
			opts: &SetupGitOptions{
				Hostname: "ghe.io",
			},
			setupErr: fmt.Errorf("broken"),
			cfgStubs: func(t *testing.T, cfg gh.Config) {
				login(t, cfg, "ghe.io", "test-user", "gho_ABCDEFG", "https", false)
			},
			expectedErr:    errors.New("failed to set up git credential helper: broken"),
			expectedErrOut: "",
		},
		{
			name:           "when there are no known hosts and no hostname is provided, return an error",
			opts:           &SetupGitOptions{},
			expectedErr:    cmdutil.SilentError,
			expectedErrOut: "You are not logged into any GitHub hosts. Run gh auth login to authenticate.\n",
		},
		{
			name: "when there are no known hosts and no hostname is provided, use the configured command name",
			opts: &SetupGitOptions{
				CommandName: "ghx",
			},
			expectedErr:    cmdutil.SilentError,
			expectedErrOut: "You are not logged into any GitHub hosts. Run ghx auth login to authenticate.\n",
		},
		{
			name: "when there are known hosts, and no hostname is provided, set them all up",
			opts: &SetupGitOptions{},
			cfgStubs: func(t *testing.T, cfg gh.Config) {
				login(t, cfg, "ghe.io", "test-user", "gho_ABCDEFG", "https", false)
				login(t, cfg, "github.com", "test-user", "gho_ABCDEFG", "https", false)
			},
			expectedHostsSetup: []string{"github.com", "ghe.io"},
		},
		{
			name:     "when no hostname is provided but setting one up errors, that error is bubbled",
			opts:     &SetupGitOptions{},
			setupErr: fmt.Errorf("broken"),
			cfgStubs: func(t *testing.T, cfg gh.Config) {
				login(t, cfg, "ghe.io", "test-user", "gho_ABCDEFG", "https", false)
			},
			expectedErr:    errors.New("failed to set up git credential helper: broken"),
			expectedErrOut: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ios, _, _, stderr := iostreams.Test()

			ios.SetStdinTTY(true)
			ios.SetStderrTTY(true)
			ios.SetStdoutTTY(true)
			tt.opts.IO = ios

			cfg, _ := config.NewIsolatedTestConfig(t)
			if tt.cfgStubs != nil {
				tt.cfgStubs(t, cfg)
			}

			if tt.opts.Config == nil {
				tt.opts.Config = func() (gh.Config, error) {
					return cfg, nil
				}
			}

			credentialsConfigurerSpy := &gitCredentialsConfigurerSpy{setupErr: tt.setupErr}
			tt.opts.CredentialsHelperConfig = credentialsConfigurerSpy

			err := setupGitRun(tt.opts)
			if tt.expectedErr != nil {
				require.Equal(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
			}

			if tt.expectedHostsSetup != nil {
				require.Equal(t, tt.expectedHostsSetup, credentialsConfigurerSpy.hosts)
			}

			require.Equal(t, tt.expectedErrOut, stderr.String())
		})
	}
}

func login(t *testing.T, c gh.Config, hostname, username, token, gitProtocol string, secureStorage bool) {
	t.Helper()
	_, err := c.Authentication().Login(hostname, username, token, gitProtocol, secureStorage)
	require.NoError(t, err)
}
