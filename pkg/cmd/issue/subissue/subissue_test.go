package subissue

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"

	"github.com/cli/cli/v2/internal/ghrepo"
	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/cli/v2/pkg/httpmock"
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/google/shlex"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCmdSubIssue(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "list", input: "list 1"},
		{name: "add", input: "add 1 2"},
		{name: "remove", input: "remove 1 2"},
		{name: "reprioritize before", input: "reprioritize 1 2 --before 3"},
		{name: "reprioritize after", input: "reprioritize 1 2 --after 3"},
		{name: "reprioritize missing anchor", input: "reprioritize 1 2", wantErr: true},
		{name: "reprioritize both anchors", input: "reprioritize 1 2 --before 3 --after 4", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ios, _, _, _ := iostreams.Test()
			f := &cmdutil.Factory{IOStreams: ios}
			cmd := NewCmdSubIssue(f)
			argv, err := shlex.Split(tt.input)
			require.NoError(t, err)
			if len(argv) > 0 {
				switch argv[0] {
				case "list":
					cmd = newCmdList(f, func(*options) error { return nil })
				case "add":
					cmd = newCmdAdd(f, func(*options) error { return nil })
				case "remove":
					cmd = newCmdRemove(f, func(*options) error { return nil })
				case "reprioritize":
					cmd = newCmdReprioritize(f, func(*options) error { return nil })
				}
				argv = argv[1:]
			}
			cmd.SetArgs(argv)
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})
			_, err = cmd.ExecuteC()
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestListRun(t *testing.T) {
	reg := &httpmock.Registry{}
	defer reg.Verify(t)
	reg.Register(
		httpmock.GraphQL(`query IssueByNumber\b`),
		httpmock.StringResponse(`{"data":{"repository":{
			"hasIssuesEnabled": true,
			"issue":{"__typename":"Issue","id":"PARENTID","number":1,"title":"parent","state":"OPEN","url":"https://github.com/OWNER/REPO/issues/1"}
		}}}`),
	)
	reg.Register(
		httpmock.GraphQL(`query IssueSubIssues\b`),
		httpmock.StringResponse(`{"data":{"repository":{
			"hasIssuesEnabled": true,
			"issue":{"subIssues":{"nodes":[
				{"id":"CHILDID","number":2,"title":"child","state":"OPEN","url":"https://github.com/OWNER/REPO/issues/2","repository":{"nameWithOwner":"OWNER/REPO"}}
			],"pageInfo":{"hasNextPage":false,"endCursor":""}}}
		}}}`),
	)

	ios, _, stdout, stderr := iostreams.Test()
	opts := &options{
		IO:         ios,
		HttpClient: func() (*http.Client, error) { return &http.Client{Transport: reg}, nil },
		BaseRepo:   func() (ghrepo.Interface, error) { return ghrepo.New("OWNER", "REPO"), nil },
		Issue:      "1",
	}
	err := listRun(opts)
	require.NoError(t, err)
	assert.Equal(t, "#2\tOPEN\tchild\thttps://github.com/OWNER/REPO/issues/2\n", stdout.String())
	assert.Equal(t, "", stderr.String())
}

func TestListRunJSON(t *testing.T) {
	reg := &httpmock.Registry{}
	defer reg.Verify(t)
	reg.Register(
		httpmock.GraphQL(`query IssueByNumber\b`),
		httpmock.StringResponse(`{"data":{"repository":{
			"hasIssuesEnabled": true,
			"issue":{"__typename":"Issue","id":"PARENTID","number":1,"title":"parent","state":"OPEN","url":"https://github.com/OWNER/REPO/issues/1"}
		}}}`),
	)
	reg.Register(
		httpmock.GraphQL(`query IssueSubIssues\b`),
		httpmock.StringResponse(`{"data":{"repository":{
			"hasIssuesEnabled": true,
			"issue":{"subIssues":{"nodes":[
				{"id":"CHILDID","number":2,"title":"child","state":"OPEN","url":"https://github.com/OWNER/REPO/issues/2","repository":{"nameWithOwner":"OWNER/REPO"}}
			],"pageInfo":{"hasNextPage":false,"endCursor":""}}}
		}}}`),
	)

	ios, _, stdout, _ := iostreams.Test()
	exporter := cmdutil.NewJSONExporter()
	exporter.SetFields([]string{"number", "repository"})
	opts := &options{
		IO:         ios,
		HttpClient: func() (*http.Client, error) { return &http.Client{Transport: reg}, nil },
		BaseRepo:   func() (ghrepo.Interface, error) { return ghrepo.New("OWNER", "REPO"), nil },
		Issue:      "1",
		Exporter:   exporter,
	}
	err := listRun(opts)
	require.NoError(t, err)
	assert.Equal(t, `[{"number":2,"repository":"OWNER/REPO"}]`+"\n", stdout.String())
}

func TestMutateRunAdd(t *testing.T) {
	reg := &httpmock.Registry{}
	defer reg.Verify(t)
	registerIssueLookup(reg, "PARENTID", 1)
	registerIssueLookup(reg, "CHILDID", 2)
	reg.Register(
		httpmock.GraphQL(`mutation AddSubIssue\b`),
		httpmock.StringResponse(`{"data":{"addSubIssue":{"issue":{"id":"PARENTID"},"subIssue":{"id":"CHILDID"}}}}`),
	)

	ios, _, _, stderr := iostreams.Test()
	opts := &options{
		IO:         ios,
		HttpClient: func() (*http.Client, error) { return &http.Client{Transport: reg}, nil },
		BaseRepo:   func() (ghrepo.Interface, error) { return ghrepo.New("OWNER", "REPO"), nil },
		Issue:      "1",
		SubIssue:   "2",
		Operation:  "add",
	}
	err := mutateRun(opts)
	require.NoError(t, err)
	assert.Equal(t, "Added subissue #2 for issue OWNER/REPO#1\n", stderr.String())
}

func TestReprioritizeRunBefore(t *testing.T) {
	reg := &httpmock.Registry{}
	defer reg.Verify(t)
	registerIssueLookup(reg, "PARENTID", 1)
	registerIssueLookup(reg, "CHILDID", 2)
	registerIssueLookup(reg, "ANCHORID", 3)
	reg.Register(
		httpmock.GraphQL(`mutation ReprioritizeSubIssue\b`),
		httpmock.GraphQLMutation(`{"data":{"reprioritizeSubIssue":{"issue":{"id":"PARENTID"}}}}`,
			func(inputs map[string]interface{}) {
				assert.Equal(t, "PARENTID", inputs["issueId"])
				assert.Equal(t, "CHILDID", inputs["subIssueId"])
				assert.Equal(t, "ANCHORID", inputs["beforeId"])
				assert.Nil(t, inputs["afterId"])
			}),
	)

	ios, _, _, stderr := iostreams.Test()
	opts := &options{
		IO:         ios,
		HttpClient: func() (*http.Client, error) { return &http.Client{Transport: reg}, nil },
		BaseRepo:   func() (ghrepo.Interface, error) { return ghrepo.New("OWNER", "REPO"), nil },
		Issue:      "1",
		SubIssue:   "2",
		Before:     "3",
	}
	err := reprioritizeRun(opts)
	require.NoError(t, err)
	assert.Equal(t, "Reprioritized subissue #2 for issue OWNER/REPO#1\n", stderr.String())
}

func registerIssueLookup(reg *httpmock.Registry, id string, number int) {
	reg.Register(
		httpmock.GraphQL(`query IssueByNumber\b`),
		httpmock.StringResponse(fmt.Sprintf(`{"data":{"repository":{
			"hasIssuesEnabled": true,
			"issue":{"__typename":"Issue","id":"%s","number":%d,"title":"issue","state":"OPEN","url":"https://github.com/OWNER/REPO/issues/%d"}
		}}}`, id, number, number)),
	)
}
