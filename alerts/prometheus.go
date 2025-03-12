package alerts

import (
	"PRism/config"
	"PRism/utils"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func CreatePrometheusAlert(suggestion config.AlertSuggestion, cfg config.Config) error {
	// Build alert rule content in YAML format
	alertRule := utils.BuildPrometheusAlertRule(suggestion)

	// Make API call to Prometheus Alertmanager
	url := fmt.Sprintf("%s/api/v1/rules", cfg.PrometheusAlertmanagerURL)
	req, err := http.NewRequest("POST", url, strings.NewReader(alertRule))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/yaml")
	if cfg.PrometheusAuthToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.PrometheusAuthToken))
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to Prometheus: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create Prometheus alert (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}
