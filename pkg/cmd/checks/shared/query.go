package shared

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/cli/cli/v2/api"
	"github.com/cli/cli/v2/internal/ghrepo"
)

type pullRequestNode struct {
	ID          string
	Number      int
	Title       string
	URL         string
	State       string
	HeadRefName string
	HeadRefOid  string
	IsDraft     bool
}

type checkContextNode struct {
	TypeName string `json:"__typename"`

	Context     string
	State       string
	TargetURL   string
	CreatedAt   time.Time
	Description string

	ID          string
	DatabaseID  int64
	Name        string
	Status      string
	Conclusion  string
	StartedAt   time.Time
	CompletedAt time.Time
	DetailsURL  string
	IsRequired  bool
	CheckSuite  struct {
		ID          string
		DatabaseID  int64
		App         CheckApp
		WorkflowRun struct {
			DatabaseID int64
			Event      string
			Workflow   struct {
				Name string
			}
		}
	}
}

// ListPullRequests returns recent pull requests for check inventory.
func ListPullRequests(httpClient *http.Client, repo ghrepo.Interface, state string, limit int) ([]PullRequestSummary, error) {
	client := api.NewClientFromHTTP(httpClient)

	type response struct {
		Repository struct {
			PullRequests struct {
				Nodes []pullRequestNode
			}
		}
	}

	query := `
	query CheckInventoryPullRequests($owner: String!, $repo: String!, $states: [PullRequestState!], $limit: Int!) {
		repository(owner: $owner, name: $repo) {
			pullRequests(states: $states, first: $limit, orderBy: { field: UPDATED_AT, direction: DESC }) {
				nodes {
					id
					number
					title
					url
					state
					headRefName
					headRefOid
					isDraft
				}
			}
		}
	}`

	variables := map[string]interface{}{
		"owner": repo.RepoOwner(),
		"repo":  repo.RepoName(),
		"limit": limit,
	}
	if states := pullRequestStates(state); len(states) > 0 {
		variables["states"] = states
	}

	var resp response
	if err := client.GraphQL(repo.RepoHost(), query, variables, &resp); err != nil {
		return nil, err
	}

	return summarizePullRequests(resp.Repository.PullRequests.Nodes), nil
}

// GetPullRequest returns one pull request for check inventory.
func GetPullRequest(httpClient *http.Client, repo ghrepo.Interface, number int) (PullRequestSummary, error) {
	client := api.NewClientFromHTTP(httpClient)

	type response struct {
		Repository struct {
			PullRequest pullRequestNode
		}
	}

	query := `
	query CheckInventoryPullRequest($owner: String!, $repo: String!, $number: Int!) {
		repository(owner: $owner, name: $repo) {
			pullRequest(number: $number) {
				id
				number
				title
				url
				state
				headRefName
				headRefOid
				isDraft
			}
		}
	}`

	var resp response
	variables := map[string]interface{}{
		"owner":  repo.RepoOwner(),
		"repo":   repo.RepoName(),
		"number": number,
	}
	if err := client.GraphQL(repo.RepoHost(), query, variables, &resp); err != nil {
		return PullRequestSummary{}, err
	}
	if resp.Repository.PullRequest.Number == 0 {
		return PullRequestSummary{}, fmt.Errorf("pull request %d not found", number)
	}
	return summarizePullRequest(resp.Repository.PullRequest), nil
}

// FetchChecksForPullRequests fetches check contexts for the supplied PR list.
func FetchChecksForPullRequests(httpClient *http.Client, repo ghrepo.Interface, prs []PullRequestSummary, filter Filter) ([]CheckItem, error) {
	var items []CheckItem
	for _, pr := range prs {
		prItems, err := FetchChecksForPullRequest(httpClient, repo, pr, filter)
		if err != nil {
			return nil, err
		}
		items = append(items, prItems...)
	}
	return items, nil
}

// FetchChecksForPullRequest fetches check contexts for one PR.
func FetchChecksForPullRequest(httpClient *http.Client, repo ghrepo.Interface, pr PullRequestSummary, filter Filter) ([]CheckItem, error) {
	client := api.NewClientFromHTTP(httpClient)

	type response struct {
		Node struct {
			Commits struct {
				Nodes []struct {
					Commit struct {
						StatusCheckRollup struct {
							Contexts struct {
								Nodes    []checkContextNode
								PageInfo struct {
									HasNextPage bool
									EndCursor   string
								}
							}
						}
					}
				}
			}
		}
	}

	query := `
	query PullRequestCheckInventory($id: ID!, $endCursor: String) {
		node(id: $id) {
			... on PullRequest {
				commits(last: 1) {
					nodes {
						commit {
							statusCheckRollup {
								contexts(first: 100, after: $endCursor) {
									nodes {
										__typename
										... on StatusContext {
											context
											state
											targetUrl
											createdAt
											description
											isRequired(pullRequestId: $id)
										}
										... on CheckRun {
											id
											databaseId
											name
											status
											conclusion
											startedAt
											completedAt
											detailsUrl
											isRequired(pullRequestId: $id)
											checkSuite {
												id
												databaseId
												app {
													name
													slug
												}
												workflowRun {
													databaseId
													event
													workflow {
														name
													}
												}
											}
										}
									}
									pageInfo {
										hasNextPage
										endCursor
									}
								}
							}
						}
					}
				}
			}
		}
	}`

	var items []CheckItem
	variables := map[string]interface{}{"id": prID(pr)}
	for {
		var resp response
		if err := client.GraphQL(repo.RepoHost(), query, variables, &resp); err != nil {
			return nil, err
		}
		if len(resp.Node.Commits.Nodes) == 0 {
			return items, nil
		}
		contexts := resp.Node.Commits.Nodes[0].Commit.StatusCheckRollup.Contexts
		for _, ctx := range contexts.Nodes {
			item := checkItem(repo, pr, ctx)
			if filter.Matches(item) {
				items = append(items, item)
			}
		}
		if !contexts.PageInfo.HasNextPage {
			break
		}
		variables["endCursor"] = contexts.PageInfo.EndCursor
	}

	return items, nil
}

func prID(pr PullRequestSummary) string {
	return pr.ID
}

func checkItem(repo ghrepo.Interface, pr PullRequestSummary, ctx checkContextNode) CheckItem {
	name := ctx.Name
	if name == "" {
		name = ctx.Context
	}
	detailsURL := ctx.DetailsURL
	if detailsURL == "" {
		detailsURL = ctx.TargetURL
	}
	state := stateFor(ctx.Status, ctx.Conclusion, ctx.State)
	item := CheckItem{
		Host:           repo.RepoHost(),
		Repository:     ghrepo.FullName(repo),
		PullRequest:    pr,
		HeadSHA:        pr.HeadSHA,
		Name:           name,
		State:          state,
		Status:         ctx.Status,
		Conclusion:     ctx.Conclusion,
		Bucket:         bucketFor(ctx.Status, ctx.Conclusion, ctx.State),
		DetailsURL:     detailsURL,
		App:            ctx.CheckSuite.App,
		CheckRunID:     ctx.DatabaseID,
		CheckSuiteID:   ctx.CheckSuite.DatabaseID,
		Workflow:       ctx.CheckSuite.WorkflowRun.Workflow.Name,
		WorkflowRunID:  ctx.CheckSuite.WorkflowRun.DatabaseID,
		Required:       ctx.IsRequired,
		Provider:       providerFromDetailsURL(detailsURL, ctx.CheckSuite.App),
		RerunSupported: false,
	}
	rerunShape(&item)
	return item
}

func summarizePullRequests(nodes []pullRequestNode) []PullRequestSummary {
	out := make([]PullRequestSummary, 0, len(nodes))
	for _, node := range nodes {
		out = append(out, summarizePullRequest(node))
	}
	return out
}

func summarizePullRequest(node pullRequestNode) PullRequestSummary {
	return PullRequestSummary{
		ID:          node.ID,
		Number:      node.Number,
		Title:       node.Title,
		URL:         node.URL,
		State:       node.State,
		HeadRefName: node.HeadRefName,
		HeadSHA:     node.HeadRefOid,
		IsDraft:     node.IsDraft,
	}
}

func pullRequestStates(state string) []string {
	switch state {
	case "open":
		return []string{"OPEN"}
	case "closed":
		return []string{"CLOSED"}
	case "merged":
		return []string{"MERGED"}
	default:
		return nil
	}
}

func int64String(v int64) string {
	return strconv.FormatInt(v, 10)
}
