package utils

import (
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

// extractActualContent extracts the content from diff format
func ExtractActualContent(diff string) string {
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

func StringPtr(s string) *string {
	return &s
}

func BoolPtr(b bool) *bool {
	return &b
}

func Int64Ptr(i int64) *int64 {
	return &i
}

// Helper function to convert Grafana panel types to Amplitude chart types
func ConvertPanelType(panelType string) string {
	switch panelType {
	case "timeseries":
		return "line"
	case "bar":
		return "bar"
	case "table":
		return "table"
	case "stat":
		return "number"
	case "pie":
		return "pie"
	default:
		return "line"
	}
}

// Helper function to convert Grafana queries to Amplitude format
func ConvertToAmplitudeQuery(query map[string]interface{}) map[string]interface{} {
	// This is a simplified conversion - in a real implementation,
	// you would need more sophisticated conversion logic
	amplitudeQuery := map[string]interface{}{
		"event_type": "custom",
	}

	// Extract event name from the Grafana query
	if expr, ok := query["expr"].(string); ok {
		// Extract event name from Prometheus-style query
		// This is just a simple extraction and would need to be more robust
		regexEvent := regexp.MustCompile(`\{([^}]+)\}`)
		if matches := regexEvent.FindStringSubmatch(expr); len(matches) > 1 {
			parts := strings.Split(matches[1], ",")
			for _, part := range parts {
				kv := strings.Split(part, "=")
				if len(kv) == 2 && strings.TrimSpace(kv[0]) == "event_name" {
					amplitudeQuery["event_type"] = strings.Trim(kv[1], "\"'")
					break
				}
			}
		}
	}

	return amplitudeQuery
}
