package alerts

import (
	"tracepr/config"
	"tracepr/github"
	"tracepr/utils"
	"fmt"
	"os"
	"path/filepath"
)

func CreatePrometheusAlert(suggestion config.AlertSuggestion, cfg config.Config) error {
	// Build alert rule content in YAML format
	alertRule := utils.BuildPrometheusAlertRule(suggestion)

	// If running in CI mode, commit to repository
	if cfg.RunningInCI {
		return github.CommitAlertToRepository(suggestion, alertRule, cfg.PrometheusConfigPath, cfg)
	}

	// Otherwise, create local file as before
	rulesDir := cfg.PrometheusConfigPath
	if !filepath.IsAbs(rulesDir) {
		currentDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		rulesDir = filepath.Join(currentDir, cfg.PrometheusConfigPath)
	}

	// Write the rule to file
	rulePath := filepath.Join(rulesDir, utils.NormalizeFileName(suggestion.Name)+".yml")
	if err := os.WriteFile(rulePath, []byte(alertRule), 0644); err != nil {
		return fmt.Errorf("failed to write Prometheus alert rule file: %w", err)
	}

	fmt.Printf("Created Prometheus alert rule at: %s\n", rulePath)
	return nil
}
