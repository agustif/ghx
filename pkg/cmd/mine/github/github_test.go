package github

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MakeNowJust/heredoc"
	"github.com/cli/cli/v2/internal/config"
	"github.com/cli/cli/v2/internal/gh"
	"github.com/cli/cli/v2/internal/ghapi/coverage"
	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/cli/v2/pkg/httpmock"
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCmdGithub(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	f := &cmdutil.Factory{
		IOStreams: ios,
	}

	var got *Options
	cmd := NewCmdGithub(f, func(opts *Options) error {
		got = opts
		return nil
	})

	cmd.SetArgs([]string{
		"--source", "all",
		"--format", "json",
		"--rest-openapi", "fixture.json",
		"--graphql-schema", "schema.json",
		"--tag", "actions",
		"--state", "missing",
		"--search", "runner",
		"--detail",
		"--limit", "0",
	})
	require.NoError(t, cmd.Execute())

	require.NotNil(t, got)
	assert.Equal(t, "all", got.Source)
	assert.Equal(t, "json", got.Format)
	assert.Equal(t, "fixture.json", got.RESTOpenAPI)
	assert.Equal(t, "schema.json", got.GraphQLSchema)
	assert.Equal(t, "actions", got.Tag)
	assert.Equal(t, coverage.StateMissing, got.State)
	assert.Equal(t, "runner", got.Search)
	assert.True(t, got.Detail)
	assert.Equal(t, 0, got.Limit)
}

func TestNewCmdGithubRejectsMisleadingFlags(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	f := &cmdutil.Factory{
		IOStreams: ios,
	}
	cmd := NewCmdGithub(f, func(opts *Options) error { return nil })

	cmd.SetArgs([]string{"--source", "rest", "--hostname", "github.com"})
	err := cmd.Execute()
	require.EqualError(t, err, "cannot use `--hostname` with `--source rest`")
}

func TestGithubRunRESTJSON(t *testing.T) {
	openAPIPath := writeOpenAPIFixture(t)
	ios, _, stdout, _ := iostreams.Test()

	opts := &Options{
		IO:          ios,
		Source:      "rest",
		Format:      "json",
		RESTOpenAPI: openAPIPath,
		HttpClient: func() (*http.Client, error) {
			return http.DefaultClient, nil
		},
	}

	require.NoError(t, githubRun(opts))
	got := stdout.String()
	got = strings.ReplaceAll(got, openAPIPath, "OPENAPI_PATH")
	got = strings.ReplaceAll(got, strings.ReplaceAll(openAPIPath, "\\", "\\\\"), "OPENAPI_PATH")
	assert.JSONEq(t, heredoc.Doc(`
		{
		  "rest": {
		    "source": "OPENAPI_PATH",
		    "totalOperations": 2,
		    "matchingOperations": 2,
		    "registeredOperations": 1,
		    "matchingRegisteredOperations": 1,
		    "remainingExplicitMetadataGap": 1,
		    "explicitMetadataCoveragePct": 50,
		    "remainingExplicitMetadataGapPct": 50,
		    "stateCounts": [
		      {
		        "name": "missing",
		        "count": 1
		      },
		      {
		        "name": "thin",
		        "count": 1
		      }
		    ],
		    "tagCounts": [
		      {
		        "name": "checks",
		        "count": 1
		      },
		      {
		        "name": "orgs",
		        "count": 1
		      }
		    ],
		    "operations": [
		      {
		        "operationId": "checks/list-for-ref",
		        "method": "GET",
		        "path": "/repos/{owner}/{repo}/commits/{ref}/check-runs",
		        "tag": "checks",
		        "summary": "List check runs for a Git reference",
		        "docsUrl": "https://docs.github.com/rest/checks/runs#list-check-runs-for-a-git-reference",
		        "coverageState": "thin",
		        "registered": true,
		        "localCommand": "ghx pr checks",
		        "proposedCommand": "ghx pr gate explain, ghx checks inventory",
		        "rawCommand": "ghx api repos/{owner}/{repo}/commits/{ref}/check-runs --paginate",
		        "pagination": "page",
		        "notes": "Existing commands expose checks but do not explain merge gate blockers across refs."
		      },
		      {
		        "operationId": "orgs/list-foo-audit-events",
		        "method": "GET",
		        "path": "/orgs/{org}/foo-audit-events",
		        "tag": "orgs",
		        "summary": "List organization foo audit events",
		        "docsUrl": "https://docs.github.com/rest/orgs/foo#list-organization-foo-audit-events",
		        "coverageState": "missing",
		        "registered": false,
		        "proposedCommand": "ghx org",
		        "rawCommand": "ghx api orgs/{org}/foo-audit-events --paginate",
		        "pagination": "page"
		      }
		    ]
		  }
		}
	`), got)
}

func TestGithubRunRESTMarkdown(t *testing.T) {
	openAPIPath := writeOpenAPIFixture(t)
	ios, _, stdout, _ := iostreams.Test()

	opts := &Options{
		IO:          ios,
		Source:      "rest",
		Format:      "md",
		RESTOpenAPI: openAPIPath,
		Detail:      true,
		Limit:       1,
		HttpClient: func() (*http.Client, error) {
			return http.DefaultClient, nil
		},
	}

	require.NoError(t, githubRun(opts))
	assert.Contains(t, stdout.String(), "# ghx GitHub API coverage")
	assert.Contains(t, stdout.String(), "## REST")
	assert.Contains(t, stdout.String(), "| REST operations | 2 |")
	assert.Contains(t, stdout.String(), "_1 more not shown_")
}

func TestGithubRunGraphQLJSON(t *testing.T) {
	reg := &httpmock.Registry{}
	defer reg.Verify(t)

	reg.Register(
		httpmock.GraphQL(`query GhxMineGraphQLSchema\b`),
		httpmock.GraphQLQuery(graphQLSchemaResponse, func(query string, variables map[string]interface{}) {
			assert.Contains(t, query, "GhxMineGraphQLSchema")
			assert.Contains(t, query, "fields(includeDeprecated: true)")
			assert.Empty(t, variables)
		}),
	)

	ios, _, stdout, _ := iostreams.Test()
	opts := &Options{
		IO:       ios,
		Source:   "graphql",
		Format:   "json",
		Hostname: "github.com",
		Config: func() (gh.Config, error) {
			return config.NewFromString("hosts:\n  github.com:\n    oauth_token: token\n"), nil
		},
		HttpClient: func() (*http.Client, error) {
			return &http.Client{Transport: reg}, nil
		},
	}

	require.NoError(t, githubRun(opts))
	assert.Contains(t, stdout.String(), `"source": "graphql"`)
	assert.Contains(t, stdout.String(), `"queryFields": 1`)
	assert.Contains(t, stdout.String(), `"mutationFields": 1`)
	assert.NotContains(t, stdout.String(), "token")
	assert.NotContains(t, stdout.String(), "Authorization")
}

func writeOpenAPIFixture(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "openapi.json")
	require.NoError(t, os.WriteFile(path, []byte(coverageTestRESTFixture), 0600))
	return path
}

const coverageTestRESTFixture = `{
  "paths": {
    "/repos/{owner}/{repo}/commits/{ref}/check-runs": {
      "get": {
        "operationId": "checks/list-for-ref",
        "tags": ["checks"],
        "summary": "List check runs for a Git reference",
        "externalDocs": {
          "url": "https://docs.github.com/rest/checks/runs#list-check-runs-for-a-git-reference"
        },
        "parameters": [
          {"name": "per_page", "in": "query"},
          {"name": "page", "in": "query"}
        ]
      }
    },
    "/orgs/{org}/foo-audit-events": {
      "get": {
        "operationId": "orgs/list-foo-audit-events",
        "tags": ["orgs"],
        "summary": "List organization foo audit events",
        "externalDocs": {
          "url": "https://docs.github.com/rest/orgs/foo#list-organization-foo-audit-events"
        },
        "parameters": [
          {"name": "per_page", "in": "query"}
        ]
      }
    }
  }
}`

const graphQLSchemaResponse = `{
  "data": {
    "__schema": {
      "queryType": {"name": "Query"},
      "mutationType": {"name": "Mutation"},
      "subscriptionType": null,
      "directives": [],
      "types": [
        {
          "kind": "OBJECT",
          "name": "Query",
          "description": "Root query",
          "fields": [
            {
              "name": "repository",
              "description": "Look up a repository",
              "isDeprecated": false,
              "deprecationReason": null,
              "args": [
                {
                  "name": "owner",
                  "description": "Owner",
                  "defaultValue": null,
                  "type": {"kind": "NON_NULL", "name": null, "ofType": {"kind": "SCALAR", "name": "String", "ofType": null}}
                }
              ],
              "type": {"kind": "OBJECT", "name": "Repository", "ofType": null}
            }
          ],
          "inputFields": null,
          "interfaces": [],
          "possibleTypes": [],
          "enumValues": null
        },
        {
          "kind": "OBJECT",
          "name": "Mutation",
          "description": "Root mutation",
          "fields": [
            {
              "name": "addSubIssue",
              "description": "Add a sub-issue",
              "isDeprecated": false,
              "deprecationReason": null,
              "args": [],
              "type": {"kind": "OBJECT", "name": "Issue", "ofType": null}
            }
          ],
          "inputFields": null,
          "interfaces": [],
          "possibleTypes": [],
          "enumValues": null
        },
        {
          "kind": "OBJECT",
          "name": "__Schema",
          "description": "Introspection",
          "fields": [
            {
              "name": "types",
              "description": "Types",
              "isDeprecated": false,
              "deprecationReason": null,
              "args": [],
              "type": {"kind": "LIST", "name": null, "ofType": {"kind": "OBJECT", "name": "__Type", "ofType": null}}
            }
          ],
          "inputFields": null,
          "interfaces": [],
          "possibleTypes": [],
          "enumValues": null
        }
      ]
    }
  }
}`
