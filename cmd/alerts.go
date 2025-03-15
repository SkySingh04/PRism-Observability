// cmd/alerts.go
package cmd

import (
	"tracepr/alerts"
	"tracepr/config"
	"tracepr/github"
	"tracepr/llm"
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	createAlertFlag     bool
	createAllAlertsFlag bool
	alertName           string
	alertType           string
	skipAlertPromptFlag bool
	runningInCIFlag     bool
)

var alertsCmd = &cobra.Command{
	Use:   "alerts",
	Short: "Create alerts based on PR changes",
	Long:  `Create alerts based on PR changes`,
	Run: func(cmd *cobra.Command, args []string) {
		runAlerts()
	},
}

func init() {
	rootCmd.AddCommand(alertsCmd)

	// Add CI/CD compatible flags similar to dashboard command
	alertsCmd.Flags().BoolVar(&createAlertFlag, "create", false, "Create a specific alert")
	alertsCmd.Flags().BoolVar(&createAllAlertsFlag, "create-all", false, "Create all suggested alerts")
	alertsCmd.Flags().StringVar(&alertName, "name", "", "Name of the alert to create (used with --create)")
	alertsCmd.Flags().StringVar(&alertType, "type", "", "Type of alert (prometheus, datadog)")
	alertsCmd.Flags().BoolVar(&skipAlertPromptFlag, "skip-prompt", false, "Skip interactive prompts (for CI/CD)")
	alertsCmd.Flags().BoolVar(&runningInCIFlag, "running-in-ci", false, "Specify if tool is running in CI")

}

func runAlerts() {
	log.Println("INFO: Starting alerts analysis...")
	cfg := config.LoadConfig()

	cfg.RunningInCI = runningInCIFlag

	// Initialize GitHub client
	log.Println("INFO: Initializing GitHub client...")
	ctx := context.Background()
	githubClient := github.InitializeGithubClient(cfg, ctx)

	// Fetch PR details including diff
	log.Printf("INFO: Fetching PR details for PR #%d...", cfg.PRNumber)
	cfg, prDetails, err := github.FetchPRDetails(githubClient, cfg)
	if err != nil {
		log.Fatalf("ERROR: Failed to fetch PR details: %v", err)
	}
	log.Printf("INFO: Successfully fetched PR details for '%s'", prDetails["title"])

	// Check for specific alert creation first
	if createAlertFlag && alertName != "" {
		log.Printf("INFO: Creating specific alert: %s", alertName)
		createSpecificAlert(cfg, alertName, alertType)
		return
	}

	if createAllAlertsFlag {
		log.Println("INFO: Creating all suggested alerts...")
		// First load saved suggestions
		savedAlerts, err := loadSavedAlertSuggestions(cfg)
		if err != nil || savedAlerts == nil || len(*savedAlerts) == 0 {
			log.Fatalf("ERROR: No saved alert suggestions found for PR #%d", cfg.PRNumber)
		}
		createAllAlerts(*savedAlerts, cfg)
		return
	}

	// Read PRD content if provided
	prdContent := ""
	if cfg.PRDFilePath != "" {
		log.Printf("INFO: Reading PRD file from %s...", cfg.PRDFilePath)
		content, err := os.ReadFile(cfg.PRDFilePath)
		if err != nil {
			log.Printf("WARN: Could not read PRD file: %v", err)
		} else {
			prdContent = string(content)
			log.Printf("INFO: Successfully read PRD file (%d bytes)", len(prdContent))
		}
	}

	// Prepare prompt for Claude
	log.Println("INFO: Building alerts analysis prompt...")
	prompt := llm.BuildAlertsPrompt(prDetails, prdContent)

	// Call Claude API
	log.Println("INFO: Calling Claude API for alerts analysis...")
	suggestions, err, responseText := llm.CallClaudeAPIForAlerts(prompt, cfg)
	if err != nil {
		log.Fatalf("ERROR: Failed to call Claude API: %v", err)
	}

	if suggestions == nil || len(*suggestions) == 0 {
		log.Println("INFO: No alert suggestions found")
		log.Println("DEBUG: Claude API response:")
		log.Println(responseText)
		return
	}

	log.Printf("INFO: Found %d alert suggestions!", len(*suggestions))

	// Log the suggestions
	for i, suggestion := range *suggestions {
		log.Printf("INFO: Alert %d: %s (%s) - Priority: %s", i+1, suggestion.Name, suggestion.Type, suggestion.Priority)
	}

	// Create PR comments if suggestions exist
	log.Println("INFO: Creating PR comments for alert suggestions...")
	err = github.CreateAlertsPRComments(*suggestions, prDetails, cfg)
	if err != nil {
		log.Fatalf("ERROR: Failed to create Alerts PR comments: %v", err)
	}
	log.Println("INFO: Successfully created PR comments")

	// Interactive prompt if not in CI/CD mode
	if !skipAlertPromptFlag {
		reader := bufio.NewReader(os.Stdin)
		fmt.Println("\nDo you want to create these alerts now? (y/n)")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		if input == "y" || input == "yes" {
			createAllAlerts(*suggestions, cfg)
		} else {
			log.Println("INFO: Alert creation skipped. You can create them later from the PR comments.")
		}
	}

	log.Println("INFO: Alerts analysis complete")
}

// createSpecificAlert attempts to load and create a specific alert by name
func createSpecificAlert(cfg config.Config, name string, alertType string) {
	// Try to load saved alert suggestions from storage
	savedAlerts, err := loadSavedAlertSuggestions(cfg)
	if err != nil || savedAlerts == nil || len(*savedAlerts) == 0 {
		log.Fatalf("ERROR: No saved alert suggestions found for PR #%d", cfg.PRNumber)
	}

	// Find the matching alert
	var targetAlert config.AlertSuggestion
	found := false

	for _, alert := range *savedAlerts {
		if alert.Name == name && (alertType == "" || alert.Type == alertType) {
			targetAlert = alert
			found = true
			break
		}
	}

	if !found {
		log.Fatalf("ERROR: No alert found with name: %s", name)
	}

	log.Printf("INFO: Creating %s alert: %s", targetAlert.Type, targetAlert.Name)
	err = createAlert(targetAlert, cfg)
	if err != nil {
		log.Fatalf("ERROR: Failed to create alert: %v", err)
	}
	log.Printf("INFO: Successfully created alert: %s", name)
}

// createAllAlerts creates all alerts in the provided suggestions
func createAllAlerts(suggestions []config.AlertSuggestion, cfg config.Config) {
	log.Println("INFO: Starting alert creation process...")
	for _, suggestion := range suggestions {
		log.Printf("INFO: Creating %s alert: %s", suggestion.Type, suggestion.Name)
		err := createAlert(suggestion, cfg)
		if err != nil {
			log.Printf("ERROR: Failed to create %s alert '%s': %v", suggestion.Type, suggestion.Name, err)
		} else {
			log.Printf("INFO: Successfully created %s alert: %s", suggestion.Type, suggestion.Name)
		}
	}
	log.Println("INFO: Alert creation process completed")
}

// createAlert creates an alert based on its type
func createAlert(suggestion config.AlertSuggestion, cfg config.Config) error {
	switch suggestion.Type {
	case "prometheus", "metric":
		return alerts.CreatePrometheusAlert(suggestion, cfg)
	case "datadog":
		return alerts.CreateDatadogAlert(suggestion, cfg)
	default:
		return fmt.Errorf("unsupported alert type: %s", suggestion.Type)
	}
}

// loadSavedAlertSuggestions loads previously generated alert suggestions
func loadSavedAlertSuggestions(cfg config.Config) (*[]config.AlertSuggestion, error) {
	// Implement storage/retrieval of alert suggestions
	// This could be from a file, database, or fetched from GitHub PR comments

	// For now, let's assume we're getting this from the PR
	ctx := context.Background()
	githubClient := github.InitializeGithubClient(cfg, ctx)

	// Implementation needed: Parse alert suggestions from PR comments
	suggestions, err := github.GetAlertSuggestionsFromPR(githubClient, cfg)
	if err != nil {
		return nil, fmt.Errorf("error loading saved alert suggestions: %v", err)
	}

	return suggestions, nil
}
