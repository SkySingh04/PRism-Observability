package dashboard

import (
	"PRism/config"
	"fmt"
	"log"
)

func CreateAmplitudeDashboard(suggestion config.DashboardSuggestion, cfg config.Config) error {
	log.Printf("Creating Amplitude dashboard: %s", suggestion.Name)

	// Note: Amplitude does not currently support creating dashboards via their API.
	// Dashboards must be created manually through the Amplitude web interface.
	// See: https://www.docs.developers.amplitude.com/analytics/apis/
	return fmt.Errorf("amplitude does not support creating dashboards via API - please create the dashboard manually in the Amplitude UI")
}
