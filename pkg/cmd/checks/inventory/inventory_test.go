package inventory

import (
	"net/http"
	"testing"

	"github.com/MakeNowJust/heredoc"
	"github.com/cli/cli/v2/internal/ghrepo"
	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/cli/v2/pkg/httpmock"
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/stretchr/testify/require"
)

func TestInventoryRunJSON(t *testing.T) {
	reg := &httpmock.Registry{}
	defer reg.Verify(t)

	reg.Register(
		httpmock.GraphQL(`query CheckInventoryPullRequests\b`),
		httpmock.StringResponse(checkInventoryPullRequestsResponse),
	)
	reg.Register(
		httpmock.GraphQL(`query PullRequestCheckInventory\b`),
		httpmock.StringResponse(checkInventoryChecksResponse),
	)

	ios, _, stdout, _ := iostreams.Test()
	exporter := cmdutil.NewJSONExporter()
	exporter.SetFields([]string{
		"pullRequest",
		"name",
		"state",
		"bucket",
		"detailsUrl",
		"checkRunId",
		"checkSuiteId",
		"app",
		"rerunSupported",
		"rerunCommand",
	})

	opts := &InventoryOptions{
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
		Bucket:   "fail",
	}

	require.NoError(t, inventoryRun(opts))
	require.JSONEq(t, heredoc.Doc(`
		[
		  {
		    "app": {
		      "name": "GitHub Actions",
		      "slug": "github-actions"
		    },
		    "bucket": "fail",
		    "checkRunId": 111,
		    "checkSuiteId": 333,
		    "detailsUrl": "https://github.com/OWNER/REPO/actions/runs/222/job/111",
		    "name": "lint",
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
		    "rerunCommand": "gh api -X POST repos/OWNER/REPO/check-runs/111/rerequest",
		    "rerunSupported": true,
		    "state": "FAILURE"
		  }
		]
	`), stdout.String())
}

const checkInventoryPullRequestsResponse = `{
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

const checkInventoryChecksResponse = `{
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
                    },
                    {
                      "__typename": "CheckRun",
                      "id": "CR_112",
                      "databaseId": 112,
                      "name": "unit",
                      "status": "COMPLETED",
                      "conclusion": "SUCCESS",
                      "startedAt": "2026-05-19T10:00:00Z",
                      "completedAt": "2026-05-19T10:01:00Z",
                      "detailsUrl": "https://github.com/OWNER/REPO/actions/runs/222/job/112",
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
