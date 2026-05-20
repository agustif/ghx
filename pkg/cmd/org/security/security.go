package security

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/cli/cli/v2/api"
	"github.com/cli/cli/v2/internal/gh"
	"github.com/cli/cli/v2/internal/tableprinter"
	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/spf13/cobra"
)

var inboxFields = []string{
	"kind",
	"repository",
	"number",
	"state",
	"severity",
	"title",
	"url",
	"createdAt",
	"updatedAt",
}

var defaultAlertKinds = []string{"code-scanning", "secret-scanning", "dependabot"}

var alertKindStates = map[string][]string{
	"code-scanning":   {"open", "closed", "dismissed", "fixed", "all"},
	"secret-scanning": {"open", "resolved", "all"},
	"dependabot":      {"open", "dismissed", "fixed", "auto_dismissed", "all"},
}

type InboxOptions struct {
	IO         *iostreams.IOStreams
	Config     func() (gh.Config, error)
	HttpClient func() (*http.Client, error)

	Org      string
	Limit    int
	State    string
	Kinds    []string
	Exporter cmdutil.Exporter
}

type SecurityAlert struct {
	Kind       string
	Repository string
	Number     int
	State      string
	Severity   string
	Title      string
	URL        string
	CreatedAt  string
	UpdatedAt  string
}

func (a SecurityAlert) ExportData(fields []string) map[string]interface{} {
	return cmdutil.StructExportData(a, fields)
}

type repositoryRef struct {
	FullName string `json:"full_name"`
}

type codeScanningAlert struct {
	Number     int           `json:"number"`
	State      string        `json:"state"`
	HTMLURL    string        `json:"html_url"`
	CreatedAt  string        `json:"created_at"`
	UpdatedAt  string        `json:"updated_at"`
	Repository repositoryRef `json:"repository"`
	Rule       struct {
		ID                    string `json:"id"`
		Name                  string `json:"name"`
		Severity              string `json:"severity"`
		SecuritySeverityLevel string `json:"security_severity_level"`
		Description           string `json:"description"`
	} `json:"rule"`
}

type secretScanningAlert struct {
	Number     int           `json:"number"`
	State      string        `json:"state"`
	HTMLURL    string        `json:"html_url"`
	CreatedAt  string        `json:"created_at"`
	UpdatedAt  string        `json:"updated_at"`
	Repository repositoryRef `json:"repository"`
	SecretType string        `json:"secret_type"`
}

type dependabotAlert struct {
	Number           int           `json:"number"`
	State            string        `json:"state"`
	HTMLURL          string        `json:"html_url"`
	CreatedAt        string        `json:"created_at"`
	UpdatedAt        string        `json:"updated_at"`
	Repository       repositoryRef `json:"repository"`
	SecurityAdvisory struct {
		Summary  string `json:"summary"`
		Severity string `json:"severity"`
	} `json:"security_advisory"`
	Dependency struct {
		Package struct {
			Name string `json:"name"`
		} `json:"package"`
	} `json:"dependency"`
}

func NewCmdSecurity(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "security <command>",
		Short: "Inspect organization security state",
		Long: heredoc.Doc(`
			Inspect organization security state using read-only GitHub APIs.
		`),
	}
	cmd.AddCommand(NewCmdInbox(f, nil))
	return cmd
}

func NewCmdInbox(f *cmdutil.Factory, runF func(*InboxOptions) error) *cobra.Command {
	opts := &InboxOptions{
		IO:         f.IOStreams,
		Config:     f.Config,
		HttpClient: f.HttpClient,
		Limit:      30,
		State:      "open",
	}

	cmd := &cobra.Command{
		Use:   "inbox <org>",
		Short: "List organization security alerts",
		Long: heredoc.Doc(`
			List organization-level code scanning, secret scanning, and Dependabot
			alerts. The command is read-only and normalizes the alert source, repo,
			severity, state, and URL for triage queues.
		`),
		Example: heredoc.Doc(`
			$ gh org security inbox github --limit 20
			$ gh org security inbox github --kind code-scanning,secret-scanning --json kind,repository,severity,url
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Org = args[0]
			if opts.Limit < 1 {
				return cmdutil.FlagErrorf("invalid limit: %v", opts.Limit)
			}
			if len(opts.Kinds) == 0 {
				opts.Kinds = append([]string(nil), defaultAlertKinds...)
			}
			if err := validateInboxState(opts.State, opts.Kinds); err != nil {
				return err
			}
			if runF != nil {
				return runF(opts)
			}
			return inboxRun(opts)
		},
	}

	cmd.Flags().IntVarP(&opts.Limit, "limit", "L", opts.Limit, "Maximum number of alerts to fetch per alert kind")
	cmdutil.StringEnumFlag(cmd, &opts.State, "state", "s", opts.State, []string{"open", "closed", "dismissed", "fixed", "resolved", "auto_dismissed", "all"}, "Filter alerts by state where supported")
	cmdutil.StringSliceEnumFlag(cmd, &opts.Kinds, "kind", "", nil, defaultAlertKinds, "Alert kinds to include")
	cmdutil.AddJSONFlags(cmd, &opts.Exporter, inboxFields)

	return cmd
}

func validateInboxState(state string, kinds []string) error {
	for _, kind := range kinds {
		if !stringInSlice(state, alertKindStates[kind]) {
			return cmdutil.FlagErrorf("`--state %s` is not supported for `--kind %s`", state, kind)
		}
	}
	return nil
}

func stringInSlice(needle string, haystack []string) bool {
	for _, item := range haystack {
		if item == needle {
			return true
		}
	}
	return false
}

func inboxRun(opts *InboxOptions) error {
	httpClient, err := opts.HttpClient()
	if err != nil {
		return err
	}
	cfg, err := opts.Config()
	if err != nil {
		return err
	}
	host, _ := cfg.Authentication().DefaultHost()
	client := api.NewClientFromHTTP(httpClient)

	var alerts []SecurityAlert
	for _, kind := range opts.Kinds {
		kindAlerts, err := fetchAlerts(client, host, opts.Org, kind, opts.State, opts.Limit)
		if err != nil {
			return err
		}
		alerts = append(alerts, kindAlerts...)
	}

	if opts.Exporter != nil {
		return opts.Exporter.Write(opts.IO, alerts)
	}
	printAlerts(opts.IO, alerts)
	return nil
}

func fetchAlerts(client *api.Client, host, org, kind, state string, limit int) ([]SecurityAlert, error) {
	query := url.Values{}
	query.Set("per_page", fmt.Sprintf("%d", limit))
	if state != "all" {
		query.Set("state", state)
	}
	path := fmt.Sprintf("orgs/%s/%s?%s", url.PathEscape(org), alertPath(kind), query.Encode())

	switch kind {
	case "code-scanning":
		var response []codeScanningAlert
		if err := client.REST(host, "GET", path, nil, &response); err != nil {
			return nil, err
		}
		alerts := make([]SecurityAlert, 0, len(response))
		for _, alert := range response {
			severity := alert.Rule.SecuritySeverityLevel
			if severity == "" {
				severity = alert.Rule.Severity
			}
			title := alert.Rule.Description
			if title == "" {
				title = alert.Rule.Name
			}
			if title == "" {
				title = alert.Rule.ID
			}
			alerts = append(alerts, SecurityAlert{
				Kind:       kind,
				Repository: alert.Repository.FullName,
				Number:     alert.Number,
				State:      alert.State,
				Severity:   severity,
				Title:      title,
				URL:        alert.HTMLURL,
				CreatedAt:  alert.CreatedAt,
				UpdatedAt:  alert.UpdatedAt,
			})
		}
		return alerts, nil
	case "secret-scanning":
		var response []secretScanningAlert
		if err := client.REST(host, "GET", path, nil, &response); err != nil {
			return nil, err
		}
		alerts := make([]SecurityAlert, 0, len(response))
		for _, alert := range response {
			alerts = append(alerts, SecurityAlert{
				Kind:       kind,
				Repository: alert.Repository.FullName,
				Number:     alert.Number,
				State:      alert.State,
				Severity:   "secret",
				Title:      alert.SecretType,
				URL:        alert.HTMLURL,
				CreatedAt:  alert.CreatedAt,
				UpdatedAt:  alert.UpdatedAt,
			})
		}
		return alerts, nil
	case "dependabot":
		var response []dependabotAlert
		if err := client.REST(host, "GET", path, nil, &response); err != nil {
			return nil, err
		}
		alerts := make([]SecurityAlert, 0, len(response))
		for _, alert := range response {
			title := alert.SecurityAdvisory.Summary
			if title == "" {
				title = alert.Dependency.Package.Name
			}
			alerts = append(alerts, SecurityAlert{
				Kind:       kind,
				Repository: alert.Repository.FullName,
				Number:     alert.Number,
				State:      alert.State,
				Severity:   alert.SecurityAdvisory.Severity,
				Title:      title,
				URL:        alert.HTMLURL,
				CreatedAt:  alert.CreatedAt,
				UpdatedAt:  alert.UpdatedAt,
			})
		}
		return alerts, nil
	default:
		return nil, fmt.Errorf("unsupported alert kind: %s", kind)
	}
}

func alertPath(kind string) string {
	switch kind {
	case "code-scanning":
		return "code-scanning/alerts"
	case "secret-scanning":
		return "secret-scanning/alerts"
	case "dependabot":
		return "dependabot/alerts"
	default:
		return kind
	}
}

func printAlerts(io *iostreams.IOStreams, alerts []SecurityAlert) {
	if len(alerts) == 0 {
		fmt.Fprintln(io.Out, "no security alerts found")
		return
	}
	tp := tableprinter.New(io, tableprinter.WithHeader("Kind", "Repo", "Severity", "State", "Alert", "URL"))
	for _, alert := range alerts {
		tp.AddField(alert.Kind)
		tp.AddField(alert.Repository)
		tp.AddField(alert.Severity)
		tp.AddField(alert.State)
		tp.AddField(strings.TrimSpace(alert.Title))
		tp.AddField(alert.URL)
		tp.EndRow()
	}
	_ = tp.Render()
}
