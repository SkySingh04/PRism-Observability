package dashboard

import (
	"PRism/config"
	"log"
)

func CreateDatadogDashboard(suggestion config.DashboardSuggestion, cfg config.Config) error {
	log.Printf("Creating Datadog dashboard: %s", suggestion.Name)
	return nil
}
