package sshidentity

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/cli/cli/v2/git"
	"github.com/cli/cli/v2/internal/gh"
)

const configKeyPrefix = "ssh_identity_"

// ConfigKey returns the host-scoped config key used for an account SSH identity.
func ConfigKey(login string) string {
	return configKeyPrefix + strings.ToLower(login)
}

// Identity returns the configured SSH identity path for a host account.
func Identity(cfg gh.Config, hostname, login string) string {
	if cfg == nil || hostname == "" || login == "" {
		return ""
	}

	entry, ok := cfg.GetOrDefault(hostname, ConfigKey(login)).Value()
	if !ok {
		return ""
	}
	return strings.TrimSpace(entry.Value)
}

// SetIdentity stores the SSH identity path for a host account.
func SetIdentity(cfg gh.Config, hostname, login, identity string) error {
	if cfg == nil {
		return fmt.Errorf("config is required")
	}
	if hostname == "" {
		return fmt.Errorf("hostname is required")
	}
	if login == "" {
		return fmt.Errorf("user is required")
	}
	if strings.TrimSpace(identity) == "" {
		return fmt.Errorf("identity is required")
	}

	cfg.Set(hostname, ConfigKey(login), ExpandIdentityPath(identity))
	return cfg.Write()
}

// ClearIdentity clears the SSH identity path for a host account.
func ClearIdentity(cfg gh.Config, hostname, login string) error {
	if cfg == nil {
		return fmt.Errorf("config is required")
	}
	if hostname == "" {
		return fmt.Errorf("hostname is required")
	}
	if login == "" {
		return fmt.Errorf("user is required")
	}

	cfg.Set(hostname, ConfigKey(login), "")
	return cfg.Write()
}

// ExpandIdentityPath expands a leading tilde in an SSH identity path.
func ExpandIdentityPath(identity string) string {
	identity = strings.TrimSpace(identity)
	if identity == "~" {
		if home, err := os.UserHomeDir(); err == nil {
			return home
		}
		return identity
	}
	if strings.HasPrefix(identity, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, strings.TrimPrefix(identity, "~/"))
		}
	}
	return identity
}

// SSHCommand returns the repo-local core.sshCommand value for an identity.
func SSHCommand(identity string) string {
	return fmt.Sprintf("ssh -i %s -o IdentitiesOnly=yes", shellQuote(identity))
}

// URLIsSSHForHost reports whether a parsed remote URL is an SSH URL for hostname.
func URLIsSSHForHost(u *url.URL, hostname string) bool {
	if u == nil || hostname == "" {
		return false
	}
	if u.Scheme != "ssh" {
		return false
	}
	return strings.EqualFold(u.Hostname(), hostname) || strings.EqualFold(u.Host, hostname)
}

// HasSSHRemoteForHost reports whether the current repository has an SSH remote for hostname.
func HasSSHRemoteForHost(ctx context.Context, gitClient *git.Client, hostname string) (bool, error) {
	remotes, err := sshRemotesForHost(ctx, gitClient, hostname)
	return len(remotes) > 0, err
}

// SyncRepoSSHCommand configures core.sshCommand for a repository with an SSH remote.
func SyncRepoSSHCommand(ctx context.Context, gitClient *git.Client, hostname, identity string) (bool, error) {
	if gitClient == nil {
		return false, nil
	}
	if strings.TrimSpace(identity) == "" {
		return false, fmt.Errorf("identity is required")
	}

	hasSSHRemote, err := HasSSHRemoteForHost(ctx, gitClient, hostname)
	if err != nil || !hasSSHRemote {
		return false, err
	}

	cmd, err := gitClient.Command(ctx, "config", "--local", "core.sshCommand", SSHCommand(identity))
	if err != nil {
		return false, err
	}
	if err := cmd.Run(); err != nil {
		return false, err
	}
	return true, nil
}

// CoreSSHCommand returns the repository core.sshCommand value when it is set.
func CoreSSHCommand(ctx context.Context, gitClient *git.Client) string {
	if gitClient == nil {
		return ""
	}
	value, err := gitClient.Config(ctx, "core.sshCommand")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(value)
}

func sshRemotesForHost(ctx context.Context, gitClient *git.Client, hostname string) ([]string, error) {
	if gitClient == nil || hostname == "" {
		return nil, nil
	}

	remotes, err := gitClient.Remotes(ctx)
	if err != nil {
		if isRepositoryUnavailable(err) {
			return nil, nil
		}
		return nil, err
	}

	var names []string
	for _, remote := range remotes {
		if URLIsSSHForHost(remote.FetchURL, hostname) || URLIsSSHForHost(remote.PushURL, hostname) {
			names = append(names, remote.Name)
		}
	}
	return names, nil
}

func isRepositoryUnavailable(err error) bool {
	var gitErr *git.GitError
	if !errors.As(err, &gitErr) {
		return false
	}
	message := strings.ToLower(gitErr.Error())
	return gitErr.ExitCode == 128 && strings.Contains(message, "not a git repository")
}

func shellQuote(value string) string {
	if value == "" {
		return "''"
	}
	if !strings.ContainsAny(value, " \t\n'\"\\$`;&|<>(){}[]*?!") {
		return value
	}
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}
