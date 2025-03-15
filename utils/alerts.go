package utils

import (
	"tracepr/config"
	"fmt"
	"strings"
	"time"
)

func ParseDuration(duration string) (int, error) {
	d, err := time.ParseDuration(duration)
	if err != nil {
		return 0, err
	}
	return int(d.Seconds()), nil
}

func FormatMessage(suggestion config.AlertSuggestion) string {
	message := fmt.Sprintf("{{#is_alert}}\n%s\n{{/is_alert}}\n\n", suggestion.Description)
	message += fmt.Sprintf("Priority: %s\n", suggestion.Priority)

	if suggestion.Notification != "" {
		message += fmt.Sprintf("\nNotifications: %s\n", suggestion.Notification)
	}

	if suggestion.RunbookLink != "" {
		message += fmt.Sprintf("\nRunbook: %s\n", suggestion.RunbookLink)
	}

	return message
}

func GetPriorityLevel(priority string) int {
	switch strings.ToLower(priority) {
	case "p1", "critical":
		return 1
	case "p2", "high":
		return 2
	case "p3", "medium":
		return 3
	case "p4", "low":
		return 4
	default:
		return 3
	}
}

func BuildPrometheusAlertRule(suggestion config.AlertSuggestion) string {
	return fmt.Sprintf(`groups:
- name: %s
  rules:
  - alert: %s
    expr: %s
    for: %s
    labels:
      severity: %s
      type: %s
    annotations:
      summary: "%s"
      description: "%s"
      runbook_url: "%s"
`,
		suggestion.Type,
		suggestion.Name,
		suggestion.Query,
		suggestion.Duration,
		strings.ToLower(suggestion.Priority),
		suggestion.Type,
		suggestion.Name,
		suggestion.Description,
		suggestion.RunbookLink)
}
