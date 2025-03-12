package alerts

import (
	"PRism/config"
	"PRism/utils"
	"context"
	"fmt"
	"log"
	"strconv"

	datadog "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
)

func CreateDatadogAlert(suggestion config.AlertSuggestion, cfg config.Config) error {
	configuration := datadog.NewConfiguration()
	configuration.Host = "api.ap1.datadoghq.com"
	configuration.AddDefaultHeader("DD-API-KEY", cfg.DatadogAPIKey)
	configuration.AddDefaultHeader("DD-APPLICATION-KEY", cfg.DatadogAppKey)
	client := datadog.NewAPIClient(configuration)

	// Format duration string to integer seconds
	_, err := utils.ParseDuration(suggestion.Duration)
	if err != nil {
		return fmt.Errorf("failed to parse duration: %w", err)
	}

	// Parse threshold to appropriate type
	threshold, err := strconv.ParseFloat(suggestion.Threshold, 64)
	if err != nil {
		return fmt.Errorf("failed to parse threshold: %w", err)
	}
	log.Printf("Attempting to create alert with query: %s", suggestion.Query)
	// Create alert request
	alertRequest := datadog.Monitor{
		Name:    &suggestion.Name,
		Type:    "query alert",
		Query:   suggestion.Query,
		Message: utils.StringPtr(utils.FormatMessage(suggestion)),
		Options: &datadog.MonitorOptions{
			NotifyNoData:      utils.BoolPtr(false),
			RenotifyInterval:  *datadog.NewNullableInt64(utils.Int64Ptr(60)),
			EvaluationDelay:   *datadog.NewNullableInt64(utils.Int64Ptr(900)),
			TimeoutH:          *datadog.NewNullableInt64(utils.Int64Ptr(0)),
			EscalationMessage: utils.StringPtr(fmt.Sprintf("Alert still triggered for %s", suggestion.Name)),
			Thresholds: &datadog.MonitorThresholds{
				Critical: &threshold,
			},
		},
		Tags: []string{fmt.Sprintf("type:%s", suggestion.Type)},
	}

	monitor, _, err := client.MonitorsApi.CreateMonitor(context.Background(), alertRequest)
	if err != nil {
		return fmt.Errorf("failed to create Datadog alert: %w", err)
	}
	fmt.Println("Monitor created with ID:", monitor.Id)
	return nil
}
