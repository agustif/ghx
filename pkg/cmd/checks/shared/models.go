package shared

import (
	"net/url"
	"strings"

	"github.com/cli/cli/v2/pkg/cmdutil"
)

// PullRequestSummary identifies the PR whose head commit produced a check.
type PullRequestSummary struct {
	ID          string `json:"id"`
	Number      int    `json:"number"`
	Title       string `json:"title"`
	URL         string `json:"url"`
	State       string `json:"state"`
	HeadRefName string `json:"headRefName"`
	HeadSHA     string `json:"headSha"`
	IsDraft     bool   `json:"isDraft"`
}

// CheckApp identifies the app that owns a check run, when GitHub reports one.
type CheckApp struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// ProviderIdentity is a best-effort parse of the check details URL.
type ProviderIdentity struct {
	Kind string `json:"kind"`
	Host string `json:"host"`
	URL  string `json:"url"`
}

// CheckItem is the stable JSON unit emitted by checks inventory and rerun dry-run.
type CheckItem struct {
	Host            string             `json:"host"`
	Repository      string             `json:"repository"`
	PullRequest     PullRequestSummary `json:"pullRequest"`
	HeadSHA         string             `json:"headSha"`
	Name            string             `json:"name"`
	State           string             `json:"state"`
	Status          string             `json:"status"`
	Conclusion      string             `json:"conclusion"`
	Bucket          string             `json:"bucket"`
	DetailsURL      string             `json:"detailsUrl"`
	App             CheckApp           `json:"app"`
	CheckRunID      int64              `json:"checkRunId"`
	CheckSuiteID    int64              `json:"checkSuiteId"`
	Workflow        string             `json:"workflow"`
	WorkflowRunID   int64              `json:"workflowRunId"`
	Required        bool               `json:"required"`
	RerunSupported  bool               `json:"rerunSupported"`
	RerunKind       string             `json:"rerunKind"`
	RerunEndpoint   string             `json:"rerunEndpoint"`
	RerunCommand    string             `json:"rerunCommand"`
	RerunLimitation string             `json:"rerunLimitation"`
	Provider        ProviderIdentity   `json:"provider"`
}

// ExportData supports gh's --json field selection.
func (c *CheckItem) ExportData(fields []string) map[string]interface{} {
	return cmdutil.StructExportData(c, fields)
}

// Filter narrows inventory and rerun target lists.
type Filter struct {
	Name        string
	Conclusion  string
	Bucket      string
	Required    bool
	Rerunnable  bool
	IncludePass bool
}

// Matches reports whether a check item matches this filter.
func (f Filter) Matches(item CheckItem) bool {
	if f.Name != "" && !strings.Contains(strings.ToLower(item.Name), strings.ToLower(f.Name)) {
		return false
	}
	if f.Conclusion != "" {
		want := strings.ToLower(f.Conclusion)
		if strings.ToLower(item.Conclusion) != want &&
			strings.ToLower(item.State) != want &&
			strings.ToLower(item.Status) != want {
			return false
		}
	}
	if f.Bucket != "" && item.Bucket != f.Bucket {
		return false
	}
	if f.Required && !item.Required {
		return false
	}
	if f.Rerunnable && !item.RerunSupported {
		return false
	}
	passRequested := f.Bucket == "pass" || strings.EqualFold(f.Conclusion, "success")
	if !f.IncludePass && !passRequested && item.Bucket == "pass" {
		return false
	}
	return true
}

func providerFromDetailsURL(detailsURL string, app CheckApp) ProviderIdentity {
	provider := ProviderIdentity{URL: detailsURL}
	if parsed, err := url.Parse(detailsURL); err == nil {
		provider.Host = parsed.Hostname()
	}
	switch {
	case app.Slug == "github-actions":
		provider.Kind = "github-actions"
	case app.Slug != "":
		provider.Kind = "github-app"
	case provider.Host != "" && provider.Host != "github.com":
		provider.Kind = "external"
	default:
		provider.Kind = "unknown"
	}
	return provider
}

func rerunShape(item *CheckItem) {
	if item.CheckRunID == 0 {
		item.RerunLimitation = "status contexts do not have check-run rerun targets"
		return
	}
	item.RerunSupported = true
	item.RerunKind = "check-run-rerequest"
	item.RerunEndpoint = "repos/" + item.Repository + "/check-runs/" + int64String(item.CheckRunID) + "/rerequest"
	item.RerunCommand = "gh api -X POST " + item.RerunEndpoint
	switch {
	case item.App.Slug == "github-actions":
		item.RerunLimitation = "uses the Checks API check-run rerequest endpoint; use gh run rerun --job only after resolving the Actions job database id"
	case item.App.Slug != "":
		item.RerunLimitation = "check-run rerequest may require permissions owned by the check app"
	}
}

func bucketFor(status, conclusion, state string) string {
	value := strings.ToUpper(state)
	if value == "" {
		value = strings.ToUpper(conclusion)
	}
	if value == "" {
		value = strings.ToUpper(status)
	}
	if strings.ToUpper(status) != "COMPLETED" && conclusion == "" && state == "" {
		switch strings.ToUpper(status) {
		case "QUEUED", "IN_PROGRESS", "PENDING", "REQUESTED", "WAITING":
			return "pending"
		}
	}
	switch value {
	case "SUCCESS":
		return "pass"
	case "SKIPPED", "NEUTRAL":
		return "skipping"
	case "ERROR", "FAILURE", "TIMED_OUT", "ACTION_REQUIRED", "STARTUP_FAILURE", "STALE":
		return "fail"
	case "CANCELLED":
		return "cancel"
	case "EXPECTED", "REQUESTED", "WAITING", "QUEUED", "PENDING", "IN_PROGRESS":
		return "pending"
	default:
		return "pending"
	}
}

func stateFor(status, conclusion, state string) string {
	if state != "" {
		return state
	}
	if strings.ToUpper(status) == "COMPLETED" && conclusion != "" {
		return conclusion
	}
	if status != "" {
		return status
	}
	return conclusion
}
