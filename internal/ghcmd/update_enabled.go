//go:build updateable

package ghcmd

// `updateable` is a build tag set in the gh formula within homebrew/homebrew-core
// and is used to control whether users are notified of newer GitHub CLI releases.
//
// Currently, updaterEnabled defaults to 'cli/cli' as it affects where
// update.CheckForUpdate() checks for releases. ghx builds remap this default to
// agustif/ghx at runtime so forked binaries do not read upstream release
// metadata. Other injected values are still left available for packagers.
//
// Development builds do not generate update messages by default.
//
// For more information, see:
// - the Homebrew formula for gh: <https://github.com/Homebrew/homebrew-core/blob/master/Formula/g/gh.rb>.
// - a discussion about adding this build tag: <https://github.com/cli/cli/pull/11024#discussion_r2107597618>.
var updaterEnabled = "cli/cli"
