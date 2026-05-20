package coverage

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteMarkdown(t *testing.T) {
	report, err := BuildRESTReport(bytes.NewBufferString(restFixture), "fixture", RESTFilter{})
	require.NoError(t, err)

	var buf bytes.Buffer
	err = WriteMarkdown(&buf, Report{REST: report}, MarkdownOptions{Detail: true, Limit: 1})
	require.NoError(t, err)

	assert.Contains(t, buf.String(), "# ghx GitHub API coverage")
	assert.Contains(t, buf.String(), "| REST operations | 2 |")
	assert.Contains(t, buf.String(), "| `checks/list-for-ref` | `GET` |")
	assert.Contains(t, buf.String(), "_1 more not shown_")
}
