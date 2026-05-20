package rerun

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/MakeNowJust/heredoc"
	"github.com/cli/cli/v2/internal/ghrepo"
	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/cli/v2/pkg/httpmock"
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/stretchr/testify/require"
)

func TestNewCmdRerunRequiresDryRun(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cmd := NewCmdRerun(&cmdutil.Factory{IOStreams: ios}, func(opts *RerunOptions) error {
		return nil
	})
	cmd.SetArgs([]string{"--name", "lint"})
	cmd.SetIn(&bytes.Buffer{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	_, err := cmd.ExecuteC()
	require.EqualError(t, err, "only `--dry-run` is supported for `checks rerun` in this release")
}

func TestRerunRunDryRunJSON(t *testing.T) {
	reg := &httpmock.Registry{}
	defer reg.Verify(t)

	reg.Register(
		httpmock.GraphQL(`query CheckInventoryPullRequests\b`),
		httpmock.StringResponse(rerunPullRequestsResponse),
	)
	reg.Register(
		httpmock.GraphQL(`query PullRequestCheckInventory\b`),
		httpmock.StringResponse(rerunChecksResponse),
	)

	ios, _, stdout, _ := iostreams.Test()
	exporter := cmdutil.NewJSONExporter()
	exporter.SetFields([]string{"dryRun", "repository", "targetCount", "runnableCount", "targets"})

	opts := &RerunOptions{
		HttpClient: func() (*http.Client, error) {
			return &http.Client{Transport: reg}, nil
		},
		IO: ios,
		BaseRepo: func() (ghrepo.Interface, error) {
			return ghrepo.New("OWNER", "REPO"), nil
		},
		Exporter: exporter,
		State:    "open",
		Limit:    30,
		Name:     "lint",
		Bucket:   "fail",
		DryRun:   true,
	}

	require.NoError(t, rerunRun(opts))
	require.JSONEq(t, heredoc.Doc(`
		{
		  "dryRun": true,
		  "repository": "OWNER/REPO",
		  "targetCount": 1,
		  "runnableCount": 1,
		  "targets": [
		    {
		      "host": "github.com",
		      "repository": "OWNER/REPO",
		      "pullRequest": {
		        "id": "PR_1",
		        "number": 1,
		        "title": "Fix the thing",
		        "url": "https://github.com/OWNER/REPO/pull/1",
		        "state": "OPEN",
		        "headRefName": "feature",
		        "headSha": "abc123",
		        "isDraft": false
		      },
		      "headSha": "abc123",
		      "name": "lint",
		      "state": "FAILURE",
		      "status": "COMPLETED",
		      "conclusion": "FAILURE",
		      "bucket": "fail",
		      "detailsUrl": "https://github.com/OWNER/REPO/actions/runs/222/job/111",
		      "app": {
		        "name": "GitHub Actions",
		        "slug": "github-actions"
		      },
		      "checkRunId": 111,
		      "checkSuiteId": 333,
		      "workflow": "CI",
		      "workflowRunId": 222,
		      "required": true,
		      "rerunSupported": true,
		      "rerunKind": "check-run-rerequest",
		      "rerunEndpoint": "repos/OWNER/REPO/check-runs/111/rerequest",
		      "rerunCommand": "gh api -X POST repos/OWNER/REPO/check-runs/111/rerequest",
		      "rerunLimitation": "uses the Checks API check-run rerequest endpoint; use gh run rerun --job only after resolving the Actions job database id",
		      "provider": {
		        "kind": "github-actions",
		        "host": "github.com",
		        "url": "https://github.com/OWNER/REPO/actions/runs/222/job/111"
		      }
		    }
		  ]
		}
	`), stdout.String())
}

const rerunPullRequestsResponse = `{
  "data": {
    "repository": {
      "pullRequests": {
        "nodes": [
          {
            "id": "PR_1",
            "number": 1,
            "title": "Fix the thing",
            "url": "https://github.com/OWNER/REPO/pull/1",
            "state": "OPEN",
            "headRefName": "feature",
            "headRefOid": "abc123",
            "isDraft": false
          }
        ]
      }
    }
  }
}`

const rerunChecksResponse = `{
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
                      "id": "CR_111",
                      "databaseId": 111,
                      "name": "lint",
                      "status": "COMPLETED",
                      "conclusion": "FAILURE",
                      "startedAt": "2026-05-19T10:00:00Z",
                      "completedAt": "2026-05-19T10:01:00Z",
                      "detailsUrl": "https://github.com/OWNER/REPO/actions/runs/222/job/111",
                      "isRequired": true,
                      "checkSuite": {
                        "id": "CS_333",
                        "databaseId": 333,
                        "app": {
                          "name": "GitHub Actions",
                          "slug": "github-actions"
                        },
                        "workflowRun": {
                          "databaseId": 222,
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
