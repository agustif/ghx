package shared

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFilterMatchesPassingChecks(t *testing.T) {
	item := CheckItem{Name: "unit", Bucket: "pass", Conclusion: "SUCCESS", State: "SUCCESS"}

	require.False(t, Filter{}.Matches(item))
	require.True(t, Filter{IncludePass: true}.Matches(item))
	require.True(t, Filter{Bucket: "pass"}.Matches(item))
	require.True(t, Filter{Conclusion: "success"}.Matches(item))
}
