package contenthealth

import (
	"fmt"
	"strings"
)

// Failed reports whether the unmuted summary should fail a check.
// Info and muted hits do not fail. failOn "warning" also fails on warnings.
func Failed(report Report, failOn string) bool {
	if report.Summary.Errors > 0 {
		return true
	}
	if strings.EqualFold(failOn, "warning") && report.Summary.Warnings > 0 {
		return true
	}
	return false
}

// FormatText renders a report. Each rule shows at most limit hits.
func FormatText(report Report, limit int) string {
	if limit <= 0 {
		limit = 30
	}
	var b strings.Builder
	commit := report.ContentCommit
	if len(commit) > 7 {
		commit = commit[:7]
	}
	if commit == "" {
		commit = "none"
	}
	fmt.Fprintf(&b, "Content health  %s\n", commit)
	fmt.Fprintf(&b, "Errors: %d  Warnings: %d  Info: %d  Muted: %d  Drift: %d  Live anomalies: %d\n\n",
		report.Summary.Errors, report.Summary.Warnings, report.Summary.Info, report.Summary.Muted, report.Summary.Drift, report.Summary.LiveAnomalies)
	for _, rule := range report.Rules {
		flag := "ok"
		if rule.Muted {
			flag = "muted"
		} else if len(rule.Hits) > 0 {
			flag = rule.Severity
		}
		fmt.Fprintf(&b, "[%s] %s  %s  (%d)  %s\n", flag, rule.ID, rule.Title, len(rule.Hits), rule.Summary)
		shown := rule.Hits
		extra := 0
		if len(shown) > limit {
			extra = len(shown) - limit
			shown = shown[:limit]
		}
		for _, hit := range shown {
			name := hit.EntityID
			if hit.EntityName != "" {
				name = hit.EntityID + " " + hit.EntityName
			}
			fmt.Fprintf(&b, "  - %s %s %s\n", hit.EntityType, name, hit.Message)
		}
		if extra > 0 {
			fmt.Fprintf(&b, "  ... and %d more\n", extra)
		}
	}
	if len(report.LiveAnomalies) > 0 {
		b.WriteString("\nLive anomalies\n")
		for _, anomaly := range report.LiveAnomalies {
			fmt.Fprintf(&b, "  - %s %s %s\n", anomaly.EntityType, anomaly.EntityID, anomaly.Message)
		}
	}
	return b.String()
}
