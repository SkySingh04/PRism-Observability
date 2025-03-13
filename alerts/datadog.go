package alerts

import (
	"PRism/config"
	"PRism/utils"
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	datadog "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
)

// extractThresholdValue extracts numeric values from threshold strings
func extractThresholdValue(thresholdStr string) (float64, error) {
	patterns := []string{
		`^>\s*(\d+(?:\.\d+)?)`,
		`^<\s*(\d+(?:\.\d+)?)`,
		`^(\d+(?:\.\d+)?)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(thresholdStr)
		if len(matches) > 1 {
			return strconv.ParseFloat(matches[1], 64)
		}
	}

	if strings.Contains(strings.ToLower(thresholdStr), "any") {
		return 1.0, nil
	}

	return 0.0, fmt.Errorf("could not extract threshold value from: %s", thresholdStr)
}

// convertSplunkToDatadogQuery converts Splunk-style queries to Datadog format
func convertSplunkToDatadogQuery(query string, threshold float64) string {
	// Handle basic search queries
	if strings.Contains(query, "index=*") {
		// Extract search terms
		searchTerms := ""
		quotedTerms := regexp.MustCompile(`"([^"]+)"`).FindAllStringSubmatch(query, -1)
		for _, term := range quotedTerms {
			if len(term) > 1 {
				if searchTerms != "" {
					searchTerms += " "
				}
				searchTerms += term[1]
			}
		}

		// Handle NOT conditions
		notCondition := ""
		notMatch := regexp.MustCompile(`NOT\s+"([^"]+)"`).FindStringSubmatch(query)
		if len(notMatch) > 1 {
			notCondition = fmt.Sprintf(" -\"%s\"", notMatch[1])
			// Remove the NOT term from searchTerms
			searchTerms = strings.Replace(searchTerms, notMatch[1], "", 1)
			searchTerms = strings.TrimSpace(searchTerms)
		}

		// Build Datadog query
		datadogQuery := fmt.Sprintf("logs(\"%s%s\").index(\"*\").rollup(\"count\").last(\"15m\") > %g",
			searchTerms,
			notCondition,
			threshold)

		return datadogQuery
	}

	// If we can't parse it, return a default query format
	return fmt.Sprintf("logs(\"*\").index(\"*\").rollup(\"count\").last(\"15m\") > %g", threshold)
}

func CreateDatadogAlert(suggestion config.AlertSuggestion, cfg config.Config) error {
	configuration := datadog.NewConfiguration()
	configuration.Host = "api.ap1.datadoghq.com"
	configuration.AddDefaultHeader("DD-API-KEY", cfg.DatadogAPIKey)
	configuration.AddDefaultHeader("DD-APPLICATION-KEY", cfg.DatadogAppKey)
	client := datadog.NewAPIClient(configuration)

	// Parse threshold
	threshold, err := extractThresholdValue(suggestion.Threshold)
	if err != nil {
		log.Printf("Warning: %v. Using default threshold of 1.0", err)
		threshold = 1.0
	}

	// Convert Splunk query to Datadog format
	datadogQuery := convertSplunkToDatadogQuery(suggestion.Query, threshold)

	log.Printf("Creating alert '%s' with threshold value: %f", suggestion.Name, threshold)
	log.Printf("Original query: %s", suggestion.Query)
	log.Printf("Converted query: %s", datadogQuery)

	// Create message with runbook link if available
	message := suggestion.Description
	if suggestion.RunbookLink != "" {
		message += fmt.Sprintf("\n\nRunbook: %s", suggestion.RunbookLink)
	}
	if suggestion.Notification != "" {
		message += fmt.Sprintf("\n\n%s", suggestion.Notification)
	}

	// Create alert request
	alertRequest := datadog.Monitor{
		Name:     &suggestion.Name,
		Type:     "log alert",
		Query:    datadogQuery,
		Message:  &message,
		Priority: *datadog.NewNullableInt64(utils.Int64Ptr(getPriorityValue(suggestion.Priority))),
		Options: &datadog.MonitorOptions{
			NotifyNoData:      utils.BoolPtr(false),
			RenotifyInterval:  *datadog.NewNullableInt64(utils.Int64Ptr(60)),
			EvaluationDelay:   *datadog.NewNullableInt64(utils.Int64Ptr(300)),
			TimeoutH:          *datadog.NewNullableInt64(utils.Int64Ptr(0)),
			EscalationMessage: utils.StringPtr(fmt.Sprintf("Alert still triggered for %s", suggestion.Name)),
			Thresholds: &datadog.MonitorThresholds{
				Critical: &threshold,
			},
		},
		Tags: []string{fmt.Sprintf("type:%s", suggestion.Type), "source:prism", "auto-generated:true"},
	}

	monitor, resp, err := client.MonitorsApi.CreateMonitor(context.Background(), alertRequest)
	if err != nil {
		// Log more detailed error info
		log.Printf("Error response from Datadog: %v", resp)
		return fmt.Errorf("failed to create Datadog alert: %w", err)
	}
	log.Printf("Monitor created with ID: %d", *monitor.Id)
	return nil
}

// getPriorityValue converts string priority to numeric value
func getPriorityValue(priority string) int64 {
	switch strings.ToLower(priority) {
	case "p1", "critical", "high":
		return 1
	case "p2", "warning", "medium":
		return 3
	case "p3", "low":
		return 5
	default:
		return 3 // Default to medium priority
	}
}
