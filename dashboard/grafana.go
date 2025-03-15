package dashboard

import (
	"tracepr/config"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

func CreateGrafanaDashboard(suggestion config.DashboardSuggestion, cfg config.Config) error {
	log.Printf("Creating Grafana dashboard: %s", suggestion.Name)

	if cfg.GrafanaServiceAccountToken == "" || cfg.GrafanaURL == "" {
		log.Printf("Error: Grafana service account token or URL not configured")
		return fmt.Errorf("grafana service account token or URL not configured")
	}

	// Parse the queries and panels into proper JSON objects
	var queries []map[string]interface{}
	var panels []map[string]interface{}
	var alerts []map[string]interface{}

	log.Printf("Parsing dashboard queries, panels and alerts")
	if err := json.Unmarshal([]byte(suggestion.Queries), &queries); err != nil {
		log.Printf("Error parsing queries JSON: %v", err)
		return fmt.Errorf("error parsing queries JSON: %v", err)
	}

	if err := json.Unmarshal([]byte(suggestion.Panels), &panels); err != nil {
		log.Printf("Error parsing panels JSON: %v", err)
		return fmt.Errorf("error parsing panels JSON: %v", err)
	}

	if err := json.Unmarshal([]byte(suggestion.Alerts), &alerts); err != nil {
		log.Printf("Error parsing alerts JSON: %v", err)
		return fmt.Errorf("error parsing alerts JSON: %v", err)
	}

	// Build Grafana dashboard JSON
	log.Printf("Building Grafana dashboard JSON")
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
	log.Printf("Adding queries to %d panels", len(panels))
	for i, panel := range panels {
		if targets, ok := panel["targets"].([]interface{}); ok {
			log.Printf("Processing panel %d: %v", i+1, panel["title"])
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
	log.Printf("Marshaling dashboard JSON")
	dashboardJSON, err := json.Marshal(dashboard)
	if err != nil {
		log.Printf("Error marshaling dashboard JSON: %v", err)
		return fmt.Errorf("error marshaling dashboard JSON: %v", err)
	}

	url := fmt.Sprintf("%s/api/dashboards/db", cfg.GrafanaURL)
	log.Printf("Sending request to Grafana API: %s", url)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(dashboardJSON))
	if err != nil {
		log.Printf("Error creating HTTP request: %v", err)
		return fmt.Errorf("error creating HTTP request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.GrafanaServiceAccountToken))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error making request to Grafana API: %v", err)
		return fmt.Errorf("error making request to Grafana API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Grafana API error (%d): %s", resp.StatusCode, string(body))
		return fmt.Errorf("grafana API error (%d): %s", resp.StatusCode, string(body))
	}

	log.Printf("Successfully created Grafana dashboard: %s", suggestion.Name)
	return nil
}
