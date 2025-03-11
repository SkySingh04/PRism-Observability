package utils

import (
	"PRism/config"
	"regexp"
	"strings"
)

// ParseLLMSuggestions extracts file-based suggestions from Claude's response
func ParseLLMSuggestionsForObservability(llmResponse string) ([]config.FileSuggestion, error) {
	suggestions := []config.FileSuggestion{}

	// Check if response is LGTM
	if strings.Contains(llmResponse, "LGTM") {
		return suggestions, nil // Return empty suggestions for LGTM case
	}

	// Find all suggestion blocks
	suggestionPattern := regexp.MustCompile(`FILE: (.+?)\nLINE: (\d+)\nSUGGESTION:\n` +
		"```diff\n" + `((?s:.+?))` + "```")
	matches := suggestionPattern.FindAllStringSubmatch(llmResponse, -1)

	for _, match := range matches {
		if len(match) != 4 {
			continue
		}

		fileName := match[1]
		lineNum := match[2]
		diffContent := match[3]

		// Parse the diff content to get actual change
		suggestion := config.FileSuggestion{
			FileName: fileName,
			LineNum:  lineNum,
			Content:  ExtractActualContent(diffContent),
		}

		suggestions = append(suggestions, suggestion)
	}

	return suggestions, nil
}

// ParseLLMSuggestionsForDashboards extracts dashboard suggestions from Claude's response
func ParseLLMSuggestionsForDashboards(llmResponse string) ([]config.DashboardSuggestion, error) {
	suggestions := []config.DashboardSuggestion{}

	// Check if response is LGTM
	if strings.Contains(llmResponse, "LGTM") {
		return suggestions, nil
	}

	// Find all dashboard suggestion blocks
	dashboardPattern := regexp.MustCompile(`DASHBOARD: (.+?)\nTYPE: (.+?)\nPRIORITY: (.+?)\nQUERIES:\n` +
		"```json\n" + `((?s:.+?))` + "```\n" +
		`PANELS:\n` + "```json\n" + `((?s:.+?))` + "```\n" +
		`ALERTS:\n` + "```json\n" + `((?s:.+?))` + "```")

	matches := dashboardPattern.FindAllStringSubmatch(llmResponse, -1)

	for _, match := range matches {
		if len(match) != 7 {
			continue
		}

		name := match[1]
		dashboardType := match[2] // "grafana" or "amplitude"
		priority := match[3]
		queries := match[4]
		panels := match[5]
		alerts := match[6]

		suggestion := config.DashboardSuggestion{
			Name:     name,
			Type:     dashboardType,
			Priority: priority,
			Queries:  queries,
			Panels:   panels,
			Alerts:   alerts,
		}

		suggestions = append(suggestions, suggestion)
	}

	return suggestions, nil
}

func ParseLLMSummary(llmResponse string) (string, error) {
	// Correct pattern to capture content after "SUMMARY:"
	summaryPattern := regexp.MustCompile(`(?s)SUMMARY:\s*(.+?)\n\n|SUMMARY:\s*(.+)$`)

	matches := summaryPattern.FindStringSubmatch(llmResponse)
	if len(matches) < 2 {
		return "", nil
	}

	// Capture group may vary depending on the match pattern
	content := strings.TrimSpace(matches[1])
	if content == "" && len(matches) > 2 {
		content = strings.TrimSpace(matches[2])
	}

	return content, nil
}
