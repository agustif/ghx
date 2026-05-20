package authssh

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/cli/cli/v2/internal/config"
	"github.com/cli/cli/v2/internal/gh"
	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/stretchr/testify/require"
)

func TestLinkRunStoresIdentity(t *testing.T) {
	cfg, _ := config.NewIsolatedTestConfig(t)
	_, err := cfg.Authentication().Login("github.com", "monalisa", "token", "ssh", false)
	require.NoError(t, err)

	identity := filepath.Join(t.TempDir(), "id_ed25519")
	require.NoError(t, os.WriteFile(identity, []byte("test"), 0o600))

	ios, _, _, stderr := iostreams.Test()
	opts := &SSHOptions{
		Config: func() (gh.Config, error) {
			return cfg, nil
		},
		IO:       ios,
		Hostname: "github.com",
		Username: "monalisa",
		Identity: identity,
	}

	err = linkRun(opts)
	require.NoError(t, err)
	require.Contains(t, stderr.String(), "Linked SSH identity for monalisa on github.com")

	entry, ok := cfg.GetOrDefault("github.com", "ssh_identity_monalisa").Value()
	require.True(t, ok)
	require.Equal(t, identity, entry.Value)
}

func TestUnlinkRunClearsIdentity(t *testing.T) {
	cfg, _ := config.NewIsolatedTestConfig(t)
	_, err := cfg.Authentication().Login("github.com", "monalisa", "token", "ssh", false)
	require.NoError(t, err)
	cfg.Set("github.com", "ssh_identity_monalisa", "/tmp/id_ed25519_monalisa")

	ios, _, _, stderr := iostreams.Test()
	opts := &SSHOptions{
		Config: func() (gh.Config, error) {
			return cfg, nil
		},
		IO:       ios,
		Hostname: "github.com",
		Username: "monalisa",
	}

	err = unlinkRun(opts)
	require.NoError(t, err)
	require.Contains(t, stderr.String(), "Unlinked SSH identity for monalisa on github.com")

	entry, ok := cfg.GetOrDefault("github.com", "ssh_identity_monalisa").Value()
	require.True(t, ok)
	require.Empty(t, entry.Value)
}

func TestListRunPrintsConfiguredIdentities(t *testing.T) {
	cfg, _ := config.NewIsolatedTestConfig(t)
	_, err := cfg.Authentication().Login("github.com", "monalisa", "token", "ssh", false)
	require.NoError(t, err)
	cfg.Set("github.com", "ssh_identity_monalisa", "/tmp/id_ed25519_monalisa")

	ios, _, stdout, _ := iostreams.Test()
	opts := &SSHOptions{
		Config: func() (gh.Config, error) {
			return cfg, nil
		},
		IO: ios,
	}

	err = listRun(opts)
	require.NoError(t, err)
	require.Equal(t, "github.com\n  monalisa: /tmp/id_ed25519_monalisa\n", stdout.String())
}

func TestListRunIncludesActiveUser(t *testing.T) {
	cfg, _ := config.NewIsolatedTestConfig(t)
	_, err := cfg.Authentication().Login("github.com", "monalisa", "token", "ssh", false)
	require.NoError(t, err)
	cfg.Authentication().SetHosts([]string{"github.com"})
	cfg.Set("github.com", "ssh_identity_monalisa", "/tmp/id_ed25519_monalisa")

	ios, _, stdout, _ := iostreams.Test()
	opts := &SSHOptions{
		Config: func() (gh.Config, error) {
			return cfg, nil
		},
		IO: ios,
	}

	err = listRun(opts)
	require.NoError(t, err)
	require.Equal(t, "github.com\n  monalisa: /tmp/id_ed25519_monalisa\n", stdout.String())
}

func TestNewCmdSSHParsesLinkFlags(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg, _ := config.NewIsolatedTestConfig(t)
	_, err := cfg.Authentication().Login("github.com", "monalisa", "token", "ssh", false)
	require.NoError(t, err)

	cmd := NewCmdSSH(&cmdutil.Factory{
		IOStreams: ios,
		Config: func() (gh.Config, error) {
			return cfg, nil
		},
	})
	for _, child := range cmd.Commands() {
		child.Flags().BoolP("help", "x", false, "")
	}
	cmd.SetArgs([]string{"link", "--hostname", "github.com", "--user", "monalisa", "--identity", "/tmp/key"})
	cmd.SetIn(&bytes.Buffer{})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	_, err = cmd.ExecuteC()
	require.Error(t, err)
	require.Contains(t, err.Error(), "could not read SSH identity")
}
