package utils

import (
	"os"
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
