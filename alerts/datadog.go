package alerts

import (
	"PRism/config"
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	datadog "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
)

// convertPrometheusToDatadogQuery converts Prometheus/Loki-style queries to Datadog format
func convertPrometheusToDatadogQuery(query string, threshold float64) string {
	// Handle rate queries with app tag and contains filter
	rateRegex := regexp.MustCompile(`rate\(\(\{app="([^"]+)"\} \|= "([^"]+)"\)\[([^\]]+)\]\)`)
	if matches := rateRegex.FindStringSubmatch(query); len(matches) > 3 {
		app := matches[1]
		searchTerm := matches[2]
		// Convert to Datadog log query format
		return fmt.Sprintf("logs(\"@app:%s %s\").index(\"*\").rollup(\"count\").last(\"%s\") > %g",
			app, searchTerm, matches[3], threshold)
	}

	// Handle absent queries
	absentRegex := regexp.MustCompile(`absent\(\(\{app="([^"]+)"\} \|= "([^"]+)"\)`)
	if matches := absentRegex.FindStringSubmatch(query); len(matches) > 2 {
		app := matches[1]
		searchTerm := matches[2]
		// Convert to Datadog absence query
		return fmt.Sprintf("logs(\"@app:%s %s\").index(\"*\").rollup(\"count\").last(\"24h\") <= 0",
			app, searchTerm)
	}

	// Handle ratio queries (error rate calculation)
	if strings.Contains(query, "sum(rate(") && strings.Contains(query, ")) / sum(rate((") {
		parts := strings.Split(query, ")) / sum(rate((")
		if len(parts) == 2 {
			// Extract the numerator and denominator parts
			numeratorPart := strings.TrimPrefix(parts[0], "sum(rate((")
			denominatorPart := strings.TrimSuffix(parts[1], ")) > 0.2")

			// Extract app and search terms
			numRegex := regexp.MustCompile(`\{app="([^"]+)"\} \|= "([^"]+)"\)\[([^\]]+)`)
			denRegex := regexp.MustCompile(`\{app="([^"]+)"\} \|= "([^"]+)"\)\[([^\]]+)`)

			numMatches := numRegex.FindStringSubmatch(numeratorPart)
			denMatches := denRegex.FindStringSubmatch(denominatorPart)

			if len(numMatches) > 3 && len(denMatches) > 3 {
				// Build a Datadog query that calculates the ratio
				return fmt.Sprintf("(logs(\"@app:%s %s\").index(\"*\").rollup(\"count\").last(\"%s\") / logs(\"@app:%s %s\").index(\"*\").rollup(\"count\").last(\"%s\")) > 0.2",
					numMatches[1], numMatches[2], numMatches[3],
					denMatches[1], denMatches[2], denMatches[3])
			}
		}
	}

	// If we can't parse it, return a default query format
	return fmt.Sprintf("logs(\"*\").index(\"*\").rollup(\"count\").last(\"15m\") > %g", threshold)
}

// Update the CreateDatadogAlert function to use this conversion
func CreateDatadogAlert(alertSuggestion config.AlertSuggestion, cfg config.Config) error {
	// Extract threshold from description if available
	threshold := 1.0 // Default threshold
	thresholdRegex := regexp.MustCompile(`threshold:\s*(\d+(?:\.\d+)?)`)
	thresholdMatches := thresholdRegex.FindStringSubmatch(alertSuggestion.Description)
	if len(thresholdMatches) > 1 {
		parsedThreshold, err := strconv.ParseFloat(thresholdMatches[1], 64)
		if err != nil {
			log.Printf("Warning: could not parse threshold value from: %s. Using default threshold of 1.0", thresholdMatches[0])
		} else {
			threshold = parsedThreshold
		}
	} else {
		log.Printf("Warning: could not extract threshold value from: %s. Using default threshold of 1.0", alertSuggestion.Description)
	}

	log.Printf("Creating alert '%s' with threshold value: %f", alertSuggestion.Name, threshold)

	// Log the original query
	log.Printf("Original query: %s", alertSuggestion.Query)

	// Convert the query to Datadog format
	query := convertPrometheusToDatadogQuery(alertSuggestion.Query, threshold)
	log.Printf("Converted query: %s", query)

	// Simple validation - ensure the query isn't empty
	if query == "" {
		query = "logs(\"*\").index(\"*\").rollup(\"count\").last(\"15m\") > " + fmt.Sprintf("%g", threshold)
		log.Printf("Warning: empty query, using default: %s", query)
	}

	// Initialize Datadog client
	configuration := datadog.NewConfiguration()
	configuration.Host = "api.ap1.datadoghq.com" // Use ap1 for Asia Pacific
	configuration.AddDefaultHeader("DD-API-KEY", cfg.DatadogAPIKey)
	configuration.AddDefaultHeader("DD-APPLICATION-KEY", cfg.DatadogAppKey)
	apiClient := datadog.NewAPIClient(configuration)

	// Create monitor options
	options := &datadog.MonitorOptions{
		NotifyNoData:      datadog.PtrBool(false),
		RequireFullWindow: datadog.PtrBool(false),
		TimeoutH:          *datadog.NewNullableInt64(datadog.PtrInt64(0)),
	}

	// Create monitor request
	monitorType := datadog.MONITORTYPE_LOG_ALERT
	monitorName := alertSuggestion.Name
	message := fmt.Sprintf("%s\n\nDescription: %s", alertSuggestion.Name, alertSuggestion.Description)
	monitorRequest := datadog.Monitor{
		Name:    &monitorName,
		Type:    monitorType,
		Query:   query,
		Message: &message,
		Options: options,
	}

	// Create the monitor
	ctx := context.Background()
	monitor, resp, err := apiClient.MonitorsApi.CreateMonitor(ctx, monitorRequest)
	if err != nil {
		log.Printf("Error response from Datadog: %v", resp)
		return fmt.Errorf("failed to create Datadog alert: %w", err)
	}

	log.Printf("Successfully created Datadog alert with ID: %d", monitor.GetId())
	return nil
}
