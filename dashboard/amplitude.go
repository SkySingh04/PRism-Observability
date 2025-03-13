package dashboard

import (
	"PRism/config"
	"PRism/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

func CreateAmplitudeDashboard(suggestion config.DashboardSuggestion, cfg config.Config) error {
	log.Printf("Creating Amplitude dashboard: %s", suggestion.Name)

	if cfg.AmplitudeAPIKey == "" || cfg.AmplitudeSecretKey == "" {
		log.Printf("Error: Amplitude API key or secret key not configured")
		return fmt.Errorf("amplitude API key or secret key not configured")
	}

	// Parse the queries and panels
	var queries []map[string]interface{}
	var panels []map[string]interface{}

	log.Printf("Parsing dashboard queries and panels")
	if err := json.Unmarshal([]byte(suggestion.Queries), &queries); err != nil {
		log.Printf("Error parsing queries JSON: %v", err)
		return fmt.Errorf("error parsing queries JSON: %v", err)
	}

	if err := json.Unmarshal([]byte(suggestion.Panels), &panels); err != nil {
		log.Printf("Error parsing panels JSON: %v", err)
		return fmt.Errorf("error parsing panels JSON: %v", err)
	}

	// Build Amplitude dashboard request
	log.Printf("Building Amplitude dashboard request")
	dashboard := map[string]interface{}{
		"name":        suggestion.Name,
		"description": "Auto-generated from observability analysis",
		"charts":      []map[string]interface{}{},
	}

	// Convert panels to Amplitude chart format
	log.Printf("Converting %d panels to Amplitude chart format", len(panels))
	for i, panel := range panels {
		log.Printf("Processing panel %d: %s", i+1, panel["title"])
		chart := map[string]interface{}{
			"name": panel["title"],
			"type": utils.ConvertPanelType(panel["type"].(string)),
		}

		// Add chart-specific settings based on panel type
		if targetsInterface, ok := panel["targets"]; ok {
			if targets, ok := targetsInterface.([]interface{}); ok {
				queryIds := []string{}
				for _, target := range targets {
					if targetStr, ok := target.(string); ok {
						queryIds = append(queryIds, targetStr)
					}
				}

				// Find matching queries
				log.Printf("Finding matching queries for panel %s", panel["title"])
				chartQueries := []map[string]interface{}{}
				for _, queryId := range queryIds {
					for _, query := range queries {
						if refId, ok := query["refId"].(string); ok && refId == queryId {
							// Convert to Amplitude query format
							amplitudeQuery := utils.ConvertToAmplitudeQuery(query)
							chartQueries = append(chartQueries, amplitudeQuery)
						}
					}
				}

				chart["queries"] = chartQueries
			}
		}

		// Add chart to dashboard
		dashboard["charts"] = append(dashboard["charts"].([]map[string]interface{}), chart)
	}

	// Send to Amplitude API
	log.Printf("Marshaling dashboard JSON")
	dashboardJSON, err := json.Marshal(dashboard)
	if err != nil {
		log.Printf("Error marshaling dashboard JSON: %v", err)
		return fmt.Errorf("error marshaling dashboard JSON: %v", err)
	}

	// Use the correct Amplitude API endpoint
	url := "https://amplitude.com/api/2/dashboard"
	log.Printf("Sending request to Amplitude API: %s", url)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(dashboardJSON))
	if err != nil {
		log.Printf("Error creating HTTP request: %v", err)
		return fmt.Errorf("error creating HTTP request: %v", err)
	}

	// Use basic authentication instead of bearer token
	req.SetBasicAuth(cfg.AmplitudeAPIKey, cfg.AmplitudeSecretKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error making request to Amplitude API: %v", err)
		return fmt.Errorf("error making request to Amplitude API: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Amplitude API error (%d): %s", resp.StatusCode, string(body))
		return fmt.Errorf("amplitude API error (%d): %s", resp.StatusCode, string(body))
	}

	log.Printf("Successfully created Amplitude dashboard: %s", suggestion.Name)
	return nil
}
