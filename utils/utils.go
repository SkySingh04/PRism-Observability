package utils

import (
	"PRism/config"
	"os"
	"regexp"
	"strings"
)

// Helper function to get environment variable with fallback default
func GetEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

// ParseLLMSuggestions extracts file-based suggestions from Claude's response
func ParseLLMSuggestions(llmResponse string) ([]config.FileSuggestion, error) {
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
			Content:  extractActualContent(diffContent),
		}

		suggestions = append(suggestions, suggestion)
	}

	return suggestions, nil
}

// extractActualContent extracts the content from diff format
func extractActualContent(diff string) string {
	var result strings.Builder

	// Process each line
	lines := strings.Split(diff, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "++") {
			// Remove the + prefix and add to result
			result.WriteString(strings.TrimPrefix(line, "+"))
			result.WriteString("\n")
		}
	}

	return strings.TrimSpace(result.String())
}

func ExtractJSONFromText(text string) string {
	// Look for JSON between ```json and ``` or just {}
	startIdx := strings.Index(text, "```json")
	if startIdx != -1 {
		startIdx += 7 // Length of "```json"
		endIdx := strings.Index(text[startIdx:], "```")
		if endIdx != -1 {
			return strings.TrimSpace(text[startIdx : startIdx+endIdx])
		}
	}

	// Try finding JSON between { and }
	startIdx = strings.Index(text, "{")
	if startIdx != -1 {
		// Find the matching closing brace
		braceCount := 1
		for i := startIdx + 1; i < len(text); i++ {
			if text[i] == '{' {
				braceCount++
			} else if text[i] == '}' {
				braceCount--
				if braceCount == 0 {
					return text[startIdx : i+1]
				}
			}
		}
	}

	return ""
}
