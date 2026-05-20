package sshidentity

import (
	"context"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/cli/cli/v2/git"
	"github.com/cli/cli/v2/internal/config"
	"github.com/stretchr/testify/require"
)

func TestIdentityRoundTrip(t *testing.T) {
	cfg, _ := config.NewIsolatedTestConfig(t)
	_, err := cfg.Authentication().Login("github.com", "agustif", "token", "ssh", false)
	require.NoError(t, err)

	home := t.TempDir()
	t.Setenv("HOME", home)

	err = SetIdentity(cfg, "github.com", "agustif", "~/.ssh/id_ed25519_agustif")
	require.NoError(t, err)

	require.Equal(t, filepath.Join(home, ".ssh", "id_ed25519_agustif"), Identity(cfg, "github.com", "agustif"))
	require.Empty(t, Identity(cfg, "github.com", "agustiobvious"))

	err = ClearIdentity(cfg, "github.com", "agustif")
	require.NoError(t, err)
	require.Empty(t, Identity(cfg, "github.com", "agustif"))
}

func TestSSHCommandQuotesIdentity(t *testing.T) {
	require.Equal(t, "ssh -i /tmp/key -o IdentitiesOnly=yes", SSHCommand("/tmp/key"))
	require.Equal(t, "ssh -i '/tmp/key with spaces' -o IdentitiesOnly=yes", SSHCommand("/tmp/key with spaces"))
	require.Equal(t, "ssh -i '/tmp/key'\\''s' -o IdentitiesOnly=yes", SSHCommand("/tmp/key's"))
}

func TestURLIsSSHForHost(t *testing.T) {
	tests := []struct {
		name     string
		rawURL   string
		hostname string
		want     bool
	}{
		{
			name:     "scp-style ssh",
			rawURL:   "git@github.com:OWNER/REPO.git",
			hostname: "github.com",
			want:     true,
		},
		{
			name:     "ssh url",
			rawURL:   "ssh://git@github.com/OWNER/REPO.git",
			hostname: "github.com",
			want:     true,
		},
		{
			name:     "https url",
			rawURL:   "https://github.com/OWNER/REPO.git",
			hostname: "github.com",
			want:     false,
		},
		{
			name:     "different ssh host",
			rawURL:   "git@ghe.io:OWNER/REPO.git",
			hostname: "github.com",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := git.ParseURL(tt.rawURL)
			require.NoError(t, err)
			require.Equal(t, tt.want, URLIsSSHForHost(parsed, tt.hostname))
		})
	}
}

func TestSyncRepoSSHCommand(t *testing.T) {
	repoDir := t.TempDir()
	runGit(t, repoDir, "init", "--quiet")
	runGit(t, repoDir, "remote", "add", "origin", "git@github.com:OWNER/REPO.git")

	identity := filepath.Join(t.TempDir(), "id_ed25519")
	require.NoError(t, os.WriteFile(identity, []byte("test"), 0o600))

	gitClient := &git.Client{RepoDir: repoDir}
	synced, err := SyncRepoSSHCommand(context.Background(), gitClient, "github.com", identity)
	require.NoError(t, err)
	require.True(t, synced)

	got := runGitOutput(t, repoDir, "config", "--local", "core.sshCommand")
	require.Equal(t, "ssh -i "+identity+" -o IdentitiesOnly=yes\n", got)
}

func TestSyncRepoSSHCommandSkipsHTTPSRemote(t *testing.T) {
	repoDir := t.TempDir()
	runGit(t, repoDir, "init", "--quiet")
	runGit(t, repoDir, "remote", "add", "origin", "https://github.com/OWNER/REPO.git")

	gitClient := &git.Client{RepoDir: repoDir}
	synced, err := SyncRepoSSHCommand(context.Background(), gitClient, "github.com", "/tmp/id_ed25519")
	require.NoError(t, err)
	require.False(t, synced)
}

func TestURLIsSSHForHostWithNilURL(t *testing.T) {
	require.False(t, URLIsSSHForHost((*url.URL)(nil), "github.com"))
}

func runGit(t *testing.T, repoDir string, args ...string) {
	t.Helper()
	client := &git.Client{RepoDir: repoDir}
	cmd, err := client.Command(context.Background(), args...)
	require.NoError(t, err)
	require.NoError(t, cmd.Run())
}

func runGitOutput(t *testing.T, repoDir string, args ...string) string {
	t.Helper()
	client := &git.Client{RepoDir: repoDir}
	cmd, err := client.Command(context.Background(), args...)
	require.NoError(t, err)
	output, err := cmd.Output()
	require.NoError(t, err)
	return string(output)
}
