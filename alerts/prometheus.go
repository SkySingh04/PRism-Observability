package alerts

import (
	"PRism/config"
	"PRism/utils"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func CreatePrometheusAlert(suggestion config.AlertSuggestion, cfg config.Config) error {
	// Build alert rule content in YAML format
	alertRule := utils.BuildPrometheusAlertRule(suggestion)

	// Ensure we have a valid path
	rulesDir := cfg.PrometheusConfigPath
	if !filepath.IsAbs(rulesDir) {
		// If not absolute, use current directory
		currentDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		rulesDir = filepath.Join(currentDir, cfg.PrometheusConfigPath)
	}

	// Create rules directory if it doesn't exist
	rulesPath := filepath.Join(rulesDir, "rules")
	if err := os.MkdirAll(rulesPath, 0755); err != nil {
		return fmt.Errorf("failed to create rules directory: %w", err)
	}

	// Write the rule to file
	rulePath := filepath.Join(rulesPath, normalizeFileName(suggestion.Name)+".yml")
	if err := os.WriteFile(rulePath, []byte(alertRule), 0644); err != nil {
		return fmt.Errorf("failed to write Prometheus alert rule file: %w", err)
	}

	fmt.Printf("Created Prometheus alert rule at: %s\n", rulePath)

	// Optionally trigger reload
	if cfg.ReloadPrometheus {
		reloadURL := fmt.Sprintf("%s/-/reload", cfg.PrometheusAlertmanagerURL)
		req, err := http.NewRequest("POST", reloadURL, nil)
		if err != nil {
			return fmt.Errorf("failed to create reload request: %w", err)
		}

		if cfg.PrometheusAuthToken != "" {
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.PrometheusAuthToken))
		}

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to reload Prometheus: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 300 {
			return fmt.Errorf("failed to reload Prometheus (status %d)", resp.StatusCode)
		}
	}

	return nil
}

// Helper function to create safe filenames
func normalizeFileName(name string) string {
	// Replace spaces and special characters with underscores
	normalized := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return r
		}
		return '_'
	}, name)

	return strings.ToLower(normalized)
}
