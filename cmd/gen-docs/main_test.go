package main

import (
	"os"
	"strings"
	"testing"
)

func Test_run(t *testing.T) {
	dir := t.TempDir()
	args := []string{"--man-page", "--website", "--doc-path", dir}
	err := run(args)
	if err != nil {
		t.Fatalf("got error: %v", err)
	}

	manPage, err := os.ReadFile(dir + "/gh-issue-create.1")
	if err != nil {
		t.Fatalf("error reading `gh-issue-create.1`: %v", err)
	}
	if !strings.Contains(string(manPage), `\fBgh issue create`) {
		t.Fatal("man page corrupted")
	}

	markdownPage, err := os.ReadFile(dir + "/gh_issue_create.md")
	if err != nil {
		t.Fatalf("error reading `gh_issue_create.md`: %v", err)
	}
	if !strings.Contains(string(markdownPage), `## gh issue create`) {
		t.Fatal("markdown page corrupted")
	}
}

func Test_runWithCommandName(t *testing.T) {
	dir := t.TempDir()
	args := []string{"--man-page", "--website", "--doc-path", dir, "--command-name", "ghx"}
	err := run(args)
	if err != nil {
		t.Fatalf("got error: %v", err)
	}

	manPage, err := os.ReadFile(dir + "/ghx-issue-create.1")
	if err != nil {
		t.Fatalf("error reading `ghx-issue-create.1`: %v", err)
	}
	if !strings.Contains(string(manPage), `\fBghx issue create`) {
		t.Fatal("ghx man page corrupted")
	}

	markdownPage, err := os.ReadFile(dir + "/ghx_issue_create.md")
	if err != nil {
		t.Fatalf("error reading `ghx_issue_create.md`: %v", err)
	}
	if !strings.Contains(string(markdownPage), `## ghx issue create`) {
		t.Fatal("ghx markdown page corrupted")
	}
	if !strings.Contains(string(markdownPage), `$ ghx issue create`) {
		t.Fatal("ghx markdown examples corrupted")
	}
	if strings.Contains(string(markdownPage), `$ gh issue create`) {
		t.Fatal("gh markdown example leaked into ghx docs")
	}

	parentMarkdownPage, err := os.ReadFile(dir + "/ghx_issue.md")
	if err != nil {
		t.Fatalf("error reading `ghx_issue.md`: %v", err)
	}
	if !strings.Contains(string(parentMarkdownPage), `[ghx issue create](./ghx_issue_create)`) {
		t.Fatal("ghx markdown links corrupted")
	}

	helpTopicPage, err := os.ReadFile(dir + "/ghx_help_environment.md")
	if err != nil {
		t.Fatalf("error reading `ghx_help_environment.md`: %v", err)
	}
	if !strings.Contains(string(helpTopicPage), `## ghx environment`) {
		t.Fatal("ghx help topic page corrupted")
	}
	if _, err := os.Stat(dir + "/gh_help_environment.md"); !os.IsNotExist(err) {
		t.Fatal("upstream gh help topic page leaked into ghx docs")
	}
}
