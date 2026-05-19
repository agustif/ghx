package version

import (
	"testing"
)

func TestFormat(t *testing.T) {
	expects := "gh version 1.4.0 (2020-12-15)\nhttps://github.com/cli/cli/releases/tag/v1.4.0\n"
	if got := Format("1.4.0", "2020-12-15"); got != expects {
		t.Errorf("Format() = %q, wants %q", got, expects)
	}
}

func TestFormatForCommand(t *testing.T) {
	tests := []struct {
		name        string
		commandName string
		version     string
		buildDate   string
		want        string
	}{
		{
			name:        "gh",
			commandName: "gh",
			version:     "1.4.0",
			buildDate:   "2020-12-15",
			want:        "gh version 1.4.0 (2020-12-15)\nhttps://github.com/cli/cli/releases/tag/v1.4.0\n",
		},
		{
			name:        "ghx",
			commandName: "ghx",
			version:     "1.4.0",
			buildDate:   "2020-12-15",
			want:        "ghx version 1.4.0 (2020-12-15)\nhttps://github.com/agustif/ghx/releases/tag/v1.4.0\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatForCommand(tt.commandName, tt.version, tt.buildDate); got != tt.want {
				t.Errorf("FormatForCommand() = %q, wants %q", got, tt.want)
			}
		})
	}
}

func TestChangelogURL(t *testing.T) {
	tag := "0.3.2"
	url := "https://github.com/cli/cli/releases/tag/v0.3.2"
	result := changelogURL("gh", tag)
	if result != url {
		t.Errorf("expected %s to create url %s but got %s", tag, url, result)
	}

	tag = "v0.3.2"
	url = "https://github.com/cli/cli/releases/tag/v0.3.2"
	result = changelogURL("gh", tag)
	if result != url {
		t.Errorf("expected %s to create url %s but got %s", tag, url, result)
	}

	tag = "0.3.2-pre.1"
	url = "https://github.com/cli/cli/releases/tag/v0.3.2-pre.1"
	result = changelogURL("gh", tag)
	if result != url {
		t.Errorf("expected %s to create url %s but got %s", tag, url, result)
	}

	tag = "0.3.5-90-gdd3f0e0"
	url = "https://github.com/cli/cli/releases/latest"
	result = changelogURL("gh", tag)
	if result != url {
		t.Errorf("expected %s to create url %s but got %s", tag, url, result)
	}

	tag = "deadbeef"
	url = "https://github.com/cli/cli/releases/latest"
	result = changelogURL("gh", tag)
	if result != url {
		t.Errorf("expected %s to create url %s but got %s", tag, url, result)
	}
}

func TestChangelogURL_ghx(t *testing.T) {
	tests := []struct {
		name string
		tag  string
		want string
	}{
		{
			name: "release tag",
			tag:  "0.3.2",
			want: "https://github.com/agustif/ghx/releases/tag/v0.3.2",
		},
		{
			name: "non release version",
			tag:  "deadbeef",
			want: "https://github.com/agustif/ghx/releases/latest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := changelogURL("ghx", tt.tag); got != tt.want {
				t.Errorf("changelogURL() = %q, wants %q", got, tt.want)
			}
		})
	}
}
