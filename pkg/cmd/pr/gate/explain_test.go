package gate

import (
	"net/http"
	"testing"

	"github.com/MakeNowJust/heredoc"
	"github.com/cli/cli/v2/api"
	"github.com/cli/cli/v2/internal/ghrepo"
	checksShared "github.com/cli/cli/v2/pkg/cmd/checks/shared"
	prShared "github.com/cli/cli/v2/pkg/cmd/pr/shared"
	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/cli/v2/pkg/httpmock"
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/stretchr/testify/require"
)

func TestExplainRunJSONWithCheckBlocker(t *testing.T) {
	reg := &httpmock.Registry{}
	defer reg.Verify(t)

	reg.Register(
		httpmock.GraphQL(`query PullRequestCheckInventory\b`),
		httpmock.StringResponse(explainChecksResponse),
	)

	repo := ghrepo.New("OWNER", "REPO")
	pr := &api.PullRequest{
		ID:               "PR_42",
		Number:           42,
		Title:            "Fix merge gate",
		State:            "OPEN",
		URL:              "https://github.com/OWNER/REPO/pull/42",
		HeadRefName:      "gate-fix",
		HeadRefOid:       "def456",
		Mergeable:        api.PullRequestMergeableMergeable,
		MergeStateStatus: "BLOCKED",
		ReviewDecision:   "APPROVED",
	}

	ios, _, stdout, _ := iostreams.Test()
	exporter := cmdutil.NewJSONExporter()
	exporter.SetFields([]string{"pullRequest", "ready", "summary", "checks", "blockers", "nextActions"})

	opts := &ExplainOptions{
		HttpClient: func() (*http.Client, error) {
			return &http.Client{Transport: reg}, nil
		},
		IO:          ios,
		Finder:      prShared.NewMockFinder("42", pr, repo),
		Exporter:    exporter,
		SelectorArg: "42",
	}

	require.NoError(t, explainRun(opts))
	require.JSONEq(t, heredoc.Doc(`
		{
		  "pullRequest": {
		    "id": "PR_42",
		    "number": 42,
		    "title": "Fix merge gate",
		    "url": "https://github.com/OWNER/REPO/pull/42",
		    "state": "OPEN",
		    "headRefName": "gate-fix",
		    "headSha": "def456",
		    "isDraft": false
		  },
		  "ready": false,
		  "summary": "1 merge blocker(s) found",
		  "checks": {
		    "total": 1,
		    "passing": 0,
		    "failing": 1,
		    "pending": 0,
		    "skipped": 0,
		    "canceled": 0
		  },
		  "blockers": [
		    {
		      "type": "checks",
		      "severity": "blocking",
		      "message": "1 check(s) are failing",
		      "detailsUrl": "https://github.com/OWNER/REPO/pull/42/checks",
		      "command": "gh checks inventory --pr 42 --bucket fail --repo OWNER/REPO --json pullRequest,name,state,detailsUrl,rerunCommand"
		    }
		  ],
		  "nextActions": [
		    "gh checks inventory --pr 42 --bucket fail --repo OWNER/REPO --json pullRequest,name,state,detailsUrl,rerunCommand"
		  ]
		}
	`), stdout.String())
}

func TestBuildReportReady(t *testing.T) {
	repo := ghrepo.New("OWNER", "REPO")
	pr := &api.PullRequest{
		ID:               "PR_7",
		Number:           7,
		Title:            "Ready PR",
		State:            "OPEN",
		URL:              "https://github.com/OWNER/REPO/pull/7",
		HeadRefName:      "ready",
		HeadRefOid:       "abc777",
		Mergeable:        api.PullRequestMergeableMergeable,
		MergeStateStatus: "CLEAN",
		ReviewDecision:   "APPROVED",
	}
	report := buildReport(&ExplainOptions{}, repo, pr, []checksShared.CheckItem{
		{Bucket: "pass"},
	})

	require.True(t, report.Ready)
	require.Equal(t, "pull request is merge-ready", report.Summary)
	require.Equal(t, []string{"gh pr merge 7 --repo OWNER/REPO"}, report.NextActions)
}

const explainChecksResponse = `{
  "data": {
    "node": {
      "commits": {
        "nodes": [
          {
            "commit": {
              "statusCheckRollup": {
                "contexts": {
                  "nodes": [
                    {
                      "__typename": "CheckRun",
                      "id": "CR_1001",
                      "databaseId": 1001,
                      "name": "build",
                      "status": "COMPLETED",
                      "conclusion": "FAILURE",
                      "startedAt": "2026-05-19T10:00:00Z",
                      "completedAt": "2026-05-19T10:01:00Z",
                      "detailsUrl": "https://github.com/OWNER/REPO/actions/runs/2002/job/1001",
                      "isRequired": true,
                      "checkSuite": {
                        "id": "CS_3003",
                        "databaseId": 3003,
                        "app": {
                          "name": "GitHub Actions",
                          "slug": "github-actions"
                        },
                        "workflowRun": {
                          "databaseId": 2002,
                          "event": "pull_request",
                          "workflow": {
                            "name": "CI"
                          }
                        }
                      }
                    }
                  ],
                  "pageInfo": {
                    "hasNextPage": false,
                    "endCursor": ""
                  }
                }
              }
            }
          }
        ]
      }
    }
  }
}`
