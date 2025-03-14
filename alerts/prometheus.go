package alerts

import (
	"PRism/config"
	"PRism/utils"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func CreatePrometheusAlert(suggestion config.AlertSuggestion, cfg config.Config) error {
	// Build alert rule content in YAML format
	alertRule := utils.BuildPrometheusAlertRule(suggestion)

	// Ensure we have a valid path
	rulesDir := cfg.PrometheusConfigPath
	log.Println("rulesDir: ", rulesDir)
	if !filepath.IsAbs(rulesDir) {
		// If not absolute, use current directory
		log.Println("rulesDir is not absolute")
		currentDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		rulesDir = filepath.Join(currentDir, cfg.PrometheusConfigPath)
		log.Println("rulesDir: ", rulesDir)
	}

	// Write the rule to file
	rulePath := filepath.Join(rulesDir, normalizeFileName(suggestion.Name)+".yml")
	if err := os.WriteFile(rulePath, []byte(alertRule), 0644); err != nil {
		return fmt.Errorf("failed to write Prometheus alert rule file: %w", err)
	}

	fmt.Printf("Created Prometheus alert rule at: %s\n", rulePath)

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
