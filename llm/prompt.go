package llm

import (
	"fmt"
	"path/filepath"
	"strings"
)

func BuildObservabilityPrompt(prDetails map[string]interface{}, prdContent string) string {
	var b strings.Builder

	b.WriteString("# Observability Instrumentation Analysis\n\n")
	b.WriteString("As an AI observability assistant, analyze the following PR and PRD to suggest code changes for:\n")
	b.WriteString("1. OpenTelemetry instrumentation (spans, metrics, attributes)\n")
	b.WriteString("2. Logging statements at appropriate locations\n")
	b.WriteString("3. Event tracking code (Amplitude)\n\n")

	// Add PR details
	b.WriteString("## Pull Request Details\n\n")
	b.WriteString(fmt.Sprintf("Title: %s\n", prDetails["title"]))
	b.WriteString(fmt.Sprintf("Description: %s\n", prDetails["description"]))
	b.WriteString(fmt.Sprintf("Author: %s\n", prDetails["author"]))
	b.WriteString(fmt.Sprintf("Created: %s\n\n", prDetails["created_at"]))

	// Add file diffs
	files := prDetails["files"].([]map[string]interface{})
	b.WriteString(fmt.Sprintf("## File Changes (%d files)\n\n", len(files)))

	for _, file := range files {
		filename := file["filename"].(string)
		status := file["status"].(string)
		additions := file["additions"].(int)
		deletions := file["deletions"].(int)
		patch := file["patch"].(string)

		// Only include .go files for detailed analysis
		if filepath.Ext(filename) == ".go" || len(files) < 5 {
			b.WriteString(fmt.Sprintf("### %s (%s, +%d, -%d)\n\n", filename, status, additions, deletions))
			b.WriteString("```diff\n")
			b.WriteString(patch)
			b.WriteString("\n```\n\n")
		} else {
			b.WriteString(fmt.Sprintf("### %s (%s, +%d, -%d) - Diff omitted\n\n", filename, status, additions, deletions))
		}
	}

	// Add PRD if provided
	if prdContent != "" {
		b.WriteString("## Product Requirements Document\n\n")
		b.WriteString(prdContent)
		b.WriteString("\n\n")
	}

	// Instructions focused only on code instrumentation
	b.WriteString("## Instructions\n\n")
	b.WriteString("For each file in the PR, suggest specific code changes as git diff format that can be posted as inline comments. Your suggestions should:\n\n")

	b.WriteString("1. Add OpenTelemetry instrumentation:\n")
	b.WriteString("   - Create spans for functions/methods\n")
	b.WriteString("   - Add attributes to spans for context\n")
	b.WriteString("   - Track errors and set status accordingly\n\n")

	b.WriteString("2. Add appropriate logging:\n")
	b.WriteString("   - Log entry/exit of important functions\n")
	b.WriteString("   - Log errors with context\n")
	b.WriteString("   - Add debug logs for complex operations\n\n")

	b.WriteString("3. Add event tracking where relevant:\n")
	b.WriteString("   - User actions\n")
	b.WriteString("   - System events\n")
	b.WriteString("   - Performance metrics\n\n")

	b.WriteString("EXTREMELY IMPORTANT CONSTRAINTS:\n")
	b.WriteString("1. ONLY suggest changes to code that appears in the diff patches above\n")
	b.WriteString("2. DO NOT suggest adding import statements or new files or functions that aren't in the diff\n")
	b.WriteString("3. Your suggestions should be insertions or modifications to the exact code blocks shown in the diff\n")
	b.WriteString("4. Always check if OpenTelemetry or logging packages are already imported before suggesting their use\n")
	b.WriteString("5. If imports are needed, only suggest them if the import section is visible in the diff\n\n")

	b.WriteString("Format each suggestion as follows:\n")
	b.WriteString("```\n")
	b.WriteString("FILE: filename.go\n")
	b.WriteString("LINE: 42\n")
	b.WriteString("SUGGESTION:\n")
	b.WriteString("```diff\n")
	b.WriteString("// Add after line 42\n")
	b.WriteString("+ span := otel.StartSpan(ctx, \"functionName\")\n")
	b.WriteString("+ defer span.End()\n")
	b.WriteString("```\n")

	b.WriteString("Follow Go best practices and match the existing code style. Only suggest changes related to observability instrumentation.")
	b.WriteString("IMPORTANT: Also, provide a summary paragraph of all the suggested changes starting with SUMMARY:, along with the reason for each change and sort them by priority (High, Medium, Low).\n\n")

	return b.String()
}

func BuildDashboardPrompt(prDetails map[string]interface{}, prdContent string) string {
	var b strings.Builder

	b.WriteString("# Observability Dashboard Analysis\n\n")
	b.WriteString("As an AI observability assistant, analyze the following PR and PRD to suggest dashboard improvements for:\n")
	b.WriteString("1. Grafana dashboards based on OpenTelemetry instrumentation\n")
	b.WriteString("2. Amplitude dashboards for event tracking\n")
	b.WriteString("3. Log-based dashboards and alerts\n\n")

	// Add PR details
	b.WriteString("## Pull Request Details\n\n")
	b.WriteString(fmt.Sprintf("Title: %s\n", prDetails["title"]))
	b.WriteString(fmt.Sprintf("Description: %s\n", prDetails["description"]))
	b.WriteString(fmt.Sprintf("Author: %s\n", prDetails["author"]))
	b.WriteString(fmt.Sprintf("Created: %s\n\n", prDetails["created_at"]))

	// Add file diffs
	files := prDetails["files"].([]map[string]interface{})
	b.WriteString(fmt.Sprintf("## File Changes (%d files)\n\n", len(files)))

	for _, file := range files {
		filename := file["filename"].(string)
		status := file["status"].(string)
		additions := file["additions"].(int)
		deletions := file["deletions"].(int)
		patch := file["patch"].(string)

		// Include all files for dashboard analysis
		b.WriteString(fmt.Sprintf("### %s (%s, +%d, -%d)\n\n", filename, status, additions, deletions))
		b.WriteString("```diff\n")
		b.WriteString(patch)
		b.WriteString("\n```\n\n")
	}

	// Add PRD if provided
	if prdContent != "" {
		b.WriteString("## Product Requirements Document\n\n")
		b.WriteString(prdContent)
		b.WriteString("\n\n")
	}

	// Instructions for dashboard suggestions
	b.WriteString("## Instructions\n\n")
	b.WriteString("Analyze the provided code changes and identify all observability instrumentation including OpenTelemetry spans, metrics, logs, and Amplitude events. Then suggest appropriate dashboards that could be created.\n\n")

	b.WriteString("For each type of telemetry data, suggest:\n\n")

	b.WriteString("1. Grafana Dashboards for OpenTelemetry data:\n")
	b.WriteString("   - Service-level dashboards showing request rates, latencies, and error rates\n")
	b.WriteString("   - Process-level dashboards showing resource utilization\n")
	b.WriteString("   - Custom dashboards for business metrics\n\n")

	b.WriteString("2. Amplitude Dashboards for event tracking:\n")
	b.WriteString("   - User journey funnels\n")
	b.WriteString("   - Feature adoption metrics\n")
	b.WriteString("   - User engagement patterns\n\n")

	b.WriteString("3. Log-based Dashboards:\n")
	b.WriteString("   - Error rate dashboards\n")
	b.WriteString("   - Log volume anomaly detection\n")
	b.WriteString("   - Critical path monitoring\n\n")

	// API-specific format
	b.WriteString("Format EACH dashboard suggestion in EXACTLY this format for parsing:\n\n")

	b.WriteString("DASHBOARD: [Dashboard name]\n")
	b.WriteString("TYPE: [grafana or amplitude]\n")
	b.WriteString("PRIORITY: [High, Medium, or Low]\n")
	b.WriteString("QUERIES:\n")
	b.WriteString("```json\n")
	b.WriteString("[\n")
	b.WriteString("  {\n")
	b.WriteString("    \"refId\": \"A\",\n")
	b.WriteString("    \"datasource\": \"Prometheus\",\n")
	b.WriteString("    \"expr\": \"sum(rate(span_count{service_name=\\\"service_name\\\"}[5m])) by (operation)\",\n")
	b.WriteString("    \"legendFormat\": \"{{operation}}\",\n")
	b.WriteString("    \"interval\": \"30s\"\n")
	b.WriteString("  }\n")
	b.WriteString("]\n")
	b.WriteString("```\n")

	b.WriteString("PANELS:\n")
	b.WriteString("```json\n")
	b.WriteString("[\n")
	b.WriteString("  {\n")
	b.WriteString("    \"title\": \"Request Rate\",\n")
	b.WriteString("    \"type\": \"timeseries\",\n")
	b.WriteString("    \"gridPos\": { \"h\": 8, \"w\": 12, \"x\": 0, \"y\": 0 },\n")
	b.WriteString("    \"targets\": [\"A\"]\n")
	b.WriteString("  }\n")
	b.WriteString("]\n")
	b.WriteString("```\n")

	b.WriteString("ALERTS:\n")
	b.WriteString("```json\n")
	b.WriteString("[\n")
	b.WriteString("  {\n")
	b.WriteString("    \"name\": \"High Error Rate\",\n")
	b.WriteString("    \"expr\": \"sum(rate(span_count{status_code=\\\"ERROR\\\"}[5m])) / sum(rate(span_count[5m])) > 0.05\",\n")
	b.WriteString("    \"for\": \"5m\",\n")
	b.WriteString("    \"severity\": \"warning\"\n")
	b.WriteString("  }\n")
	b.WriteString("]\n")
	b.WriteString("```\n\n")

	b.WriteString("IMPORTANT GUIDELINES:\n")
	b.WriteString("1. Only suggest dashboards based on telemetry data present in the code\n")
	b.WriteString("2. Focus on actionable insights, not just vanity metrics\n")
	b.WriteString("3. For Grafana, use valid Prometheus or Loki queries based on the instrumentation\n")
	b.WriteString("4. For Amplitude, use valid event names and properties from the code\n")
	b.WriteString("5. Provide dashboard configuration in EXACTLY the format specified above\n")
	b.WriteString("6. Include at least the minimum required fields for API creation\n\n")

	b.WriteString("## Identified Telemetry\n")
	b.WriteString("Before providing dashboard suggestions, list all identified spans, metrics, logs, and events with their attributes.\n\n")

	b.WriteString("## Dashboard Suggestions\n")
	b.WriteString("List each dashboard suggestion in the specified format.\n\n")

	b.WriteString("SUMMARY:\n")
	b.WriteString("After your detailed analysis, provide a prioritized summary of all suggested dashboards with business justification and expected value.\n\n")

	return b.String()
}

func BuildAlertsPrompt(prDetails map[string]interface{}, prdContent string) string {
	var b strings.Builder

	b.WriteString("# Observability Alerts Analysis\n\n")
	b.WriteString("As an AI observability assistant, analyze the following PR and PRD to suggest alerts for:\n")
	b.WriteString("1. OpenTelemetry-based metrics and trace alerts\n")
	b.WriteString("2. Log-based alerts for error patterns\n")
	b.WriteString("3. SLO/SLI monitoring alerts\n\n")

	// Add PR details
	b.WriteString("## Pull Request Details\n\n")
	b.WriteString(fmt.Sprintf("Title: %s\n", prDetails["title"]))
	b.WriteString(fmt.Sprintf("Description: %s\n", prDetails["description"]))
	b.WriteString(fmt.Sprintf("Author: %s\n", prDetails["author"]))
	b.WriteString(fmt.Sprintf("Created: %s\n\n", prDetails["created_at"]))

	// Add file diffs
	files := prDetails["files"].([]map[string]interface{})
	b.WriteString(fmt.Sprintf("## File Changes (%d files)\n\n", len(files)))

	for _, file := range files {
		filename := file["filename"].(string)
		status := file["status"].(string)
		additions := file["additions"].(int)
		deletions := file["deletions"].(int)
		patch := file["patch"].(string)

		// Include all files for alert analysis
		b.WriteString(fmt.Sprintf("### %s (%s, +%d, -%d)\n\n", filename, status, additions, deletions))
		b.WriteString("```diff\n")
		b.WriteString(patch)
		b.WriteString("\n```\n\n")
	}

	// Add PRD if provided
	if prdContent != "" {
		b.WriteString("## Product Requirements Document\n\n")
		b.WriteString(prdContent)
		b.WriteString("\n\n")
	}

	// Instructions for alert suggestions
	b.WriteString("## Instructions\n\n")
	b.WriteString("Analyze the provided code changes and identify all observability instrumentation including OpenTelemetry spans, metrics, logs, and events. Then suggest appropriate alerts that should be configured.\n\n")

	b.WriteString("For each type of telemetry data, suggest:\n\n")

	b.WriteString("1. OpenTelemetry Metric and Trace Alerts:\n")
	b.WriteString("   - High error rates or latency\n")
	b.WriteString("   - Unusual traffic patterns\n")
	b.WriteString("   - Dependency failures\n\n")

	b.WriteString("2. Log-based Alerts:\n")
	b.WriteString("   - Critical error patterns\n")
	b.WriteString("   - Authentication failures\n")
	b.WriteString("   - Data integrity issues\n\n")

	b.WriteString("3. SLO/SLI Alerts:\n")
	b.WriteString("   - Error budget burn rate\n")
	b.WriteString("   - Latency percentile thresholds\n")
	b.WriteString("   - Availability metrics\n\n")

	// API-specific format for parsing
	b.WriteString("Format EACH alert suggestion in EXACTLY this format for parsing:\n\n")

	b.WriteString("ALERT: [Alert name]\n")
	b.WriteString("TYPE: [metric, log, or slo]\n")
	b.WriteString("PRIORITY: [P0, P1, or P2]\n")
	b.WriteString("QUERY:\n")
	b.WriteString("```\n")
	b.WriteString("sum(rate(span_count{status_code=\"ERROR\"}[5m])) / sum(rate(span_count[5m])) > 0.05\n")
	b.WriteString("```\n")
	b.WriteString("DESCRIPTION: [Brief description of what the alert means]\n")
	b.WriteString("THRESHOLD: [Numerical threshold or condition]\n")
	b.WriteString("DURATION: [How long condition must be true, e.g. 5m]\n")
	b.WriteString("NOTIFICATION: [Where alert should be sent, e.g. slack-sre-channel]\n")
	b.WriteString("RUNBOOK_LINK: [Link to runbook or troubleshooting guide]\n\n")

	b.WriteString("IMPORTANT GUIDELINES:\n")
	b.WriteString("1. Only suggest alerts based on telemetry data present in the code\n")
	b.WriteString("2. Focus on actionable alerts, avoid noise\n")
	b.WriteString("3. Use valid PromQL, LogQL, or other appropriate query languages\n")
	b.WriteString("4. Prioritize alerts: P0=critical, P1=warning, P2=info\n")
	b.WriteString("5. Provide alert configuration in EXACTLY the format specified above\n")
	b.WriteString("6. Include all required fields\n\n")

	b.WriteString("## Identified Telemetry\n")
	b.WriteString("Before providing alert suggestions, list all identified spans, metrics, logs, and events with their attributes.\n\n")

	b.WriteString("## Alert Suggestions\n")
	b.WriteString("List each alert suggestion in the specified format.\n\n")

	b.WriteString("SUMMARY:\n")
	b.WriteString("After your detailed analysis, provide a prioritized summary of all suggested alerts with business justification and expected value.\n\n")

	return b.String()
}
