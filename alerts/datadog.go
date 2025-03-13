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

func convertPrometheusToDatadogQuery(query string, threshold float64) string {
	// Handle simple count queries with message regex
	countRegex := regexp.MustCompile(`count\(message=~"([^"]+)"\)\s*>\s*([\d.]+)`)
	if matches := countRegex.FindStringSubmatch(query); len(matches) > 2 {
		messageRegex := strings.ReplaceAll(matches[1], ".*", "*")
		return fmt.Sprintf("logs(\"@job:prism message:~\\\"%s\\\"\").index(\"*\").rollup(\"count\").last(\"5m\") > %s",
			messageRegex, matches[2])
	}

	// Handle composite queries with AND/OR
	if strings.Contains(query, " AND ") || strings.Contains(query, " OR ") {
		parts := strings.Split(query, " ")
		var ddParts []string
		currentPart := ""

		for _, part := range parts {
			if strings.EqualFold(part, "AND") || strings.EqualFold(part, "OR") {
				if currentPart != "" {
					ddParts = append(ddParts, convertQueryFragment(currentPart))
					currentPart = ""
				}
				ddParts = append(ddParts, strings.ToLower(part))
				continue
			}
			currentPart += " " + part
		}
		if currentPart != "" {
			ddParts = append(ddParts, convertQueryFragment(currentPart))
		}

		return fmt.Sprintf("%s > %.2f", strings.Join(ddParts, " "), threshold)
	}

	// Handle ratio queries with error rate calculation
	if strings.Contains(query, "/") {
		parts := strings.Split(query, "/")
		if len(parts) == 2 {
			numerator := convertQueryFragment(parts[0])
			denominator := convertQueryFragment(parts[1])
			return fmt.Sprintf("(%s / %s) > %.2f", numerator, denominator, threshold)
		}
	}

	// Default case for simple thresholds
	return fmt.Sprintf("logs(\"@job:prism\").index(\"*\").rollup(\"count\").last(\"5m\") > %.2f", threshold)
}

func convertQueryFragment(fragment string) string {
	// Handle count_over_time with message regex
	if strings.Contains(fragment, "count_over_time") {
		regex := regexp.MustCompile(`count_over_time\(\{([^}]+)\}\s+\|=~?"([^"]+)"\s*\[([^\]]+)\]\)`)
		if matches := regex.FindStringSubmatch(fragment); len(matches) > 3 {
			labels := strings.ReplaceAll(matches[1], "\"", "")
			message := strings.ReplaceAll(matches[2], ".*", "*")
			timeWindow := matches[3]

			var facets []string
			for _, label := range strings.Split(labels, ",") {
				parts := strings.Split(strings.TrimSpace(label), "=")
				if len(parts) == 2 {
					facets = append(facets, fmt.Sprintf("@%s:%s", parts[0], strings.Trim(parts[1], "\"")))
				}
			}

			return fmt.Sprintf("logs(\"%s message:~\\\"%s\\\"\").rollup(\"count\", \"%s\").last(\"%s\")",
				strings.Join(facets, " "), message, timeWindow, timeWindow)
		}
	}

	// Handle simple message matches
	if strings.Contains(fragment, "message=") {
		regex := regexp.MustCompile(`message=~?"([^"]+)"`)
		if matches := regex.FindStringSubmatch(fragment); len(matches) > 1 {
			message := strings.ReplaceAll(matches[1], ".*", "*")
			return fmt.Sprintf("logs(\"@job:prism message:~\\\"%s\\\"\").rollup(\"count\").last(\"5m\")", message)
		}
	}

	return fragment
}

func extractThreshold(query string) float64 {
	// First try to extract from comparison operators
	thresholdRegex := regexp.MustCompile(`>\s*([\d.]+)`)
	if matches := thresholdRegex.FindStringSubmatch(query); len(matches) > 1 {
		if val, err := strconv.ParseFloat(matches[1], 64); err == nil {
			return val
		}
	}

	// Then check for ratio thresholds
	ratioRegex := regexp.MustCompile(`>\s*([\d.]+)`)
	if matches := ratioRegex.FindStringSubmatch(query); len(matches) > 1 {
		if val, err := strconv.ParseFloat(matches[1], 64); err == nil {
			return val
		}
	}

	return 1.0 // Default threshold
}

func CreateDatadogAlert(alertSuggestion config.AlertSuggestion, cfg config.Config) error {
	threshold := extractThreshold(alertSuggestion.Query)

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
