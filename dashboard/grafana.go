package dashboard

import (
	"PRism/config"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// CreateGrafanaDashboard makes an API call to create a Grafana dashboard
func CreateGrafanaDashboard(suggestion config.DashboardSuggestion, cfg config.Config) error {
	if cfg.GrafanaAPIToken == "" || cfg.GrafanaURL == "" {
		return fmt.Errorf("Grafana API token or URL not configured")
	}

	// Parse the queries and panels into proper JSON objects
	var queries []map[string]interface{}
	var panels []map[string]interface{}
	var alerts []map[string]interface{}

	if err := json.Unmarshal([]byte(suggestion.Queries), &queries); err != nil {
		return fmt.Errorf("error parsing queries JSON: %v", err)
	}

	if err := json.Unmarshal([]byte(suggestion.Panels), &panels); err != nil {
		return fmt.Errorf("error parsing panels JSON: %v", err)
	}

	if err := json.Unmarshal([]byte(suggestion.Alerts), &alerts); err != nil {
		return fmt.Errorf("error parsing alerts JSON: %v", err)
	}

	// Build Grafana dashboard JSON
	dashboard := map[string]interface{}{
		"dashboard": map[string]interface{}{
			"id":            nil,
			"title":         suggestion.Name,
			"tags":          []string{"auto-generated", "observability"},
			"timezone":      "browser",
			"schemaVersion": 16,
			"version":       1,
			"refresh":       "5s",
			"panels":        panels,
		},
		"folderId":  0,
		"overwrite": true,
	}

	// Add queries to panels
	for i, panel := range panels {
		if targets, ok := panel["targets"].([]interface{}); ok {
			for j, target := range targets {
				if targetStr, ok := target.(string); ok {
					// Find matching query by refId
					for _, query := range queries {
						if refId, ok := query["refId"].(string); ok && refId == targetStr {
							if panels[i]["targets"] == nil {
								panels[i]["targets"] = make([]interface{}, 0)
							}
							// Replace string reference with actual query
							newTargets := make([]interface{}, len(targets))
							copy(newTargets, targets)
							newTargets[j] = query
							panels[i]["targets"] = newTargets
							break
						}
					}
				}
			}
		}
	}

	// Send to Grafana API
	dashboardJSON, err := json.Marshal(dashboard)
	if err != nil {
		return fmt.Errorf("error marshaling dashboard JSON: %v", err)
	}

	url := fmt.Sprintf("%s/api/dashboards/db", cfg.GrafanaURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(dashboardJSON))
	if err != nil {
		return fmt.Errorf("error creating HTTP request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.GrafanaAPIToken))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error making request to Grafana API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("Grafana API error (%d): %s", resp.StatusCode, string(body))
	}

	return nil
}
