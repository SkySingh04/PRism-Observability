// cmd/alerts.go
package cmd

import (
	"PRism/alerts"
	"PRism/config"
	"PRism/github"
	"PRism/llm"
	"bufio"
	"context"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
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

	// Add alerts-specific flags here if needed
	alertsCmd.Flags().Bool("show-all", false, "Show all alerts including resolved ones")
	alertsCmd.Flags().String("severity", "all", "Filter alerts by severity (high, medium, low, all)")
}

func runAlerts() {
	log.Println("INFO: Starting alerts analysis...")
	cfg := config.LoadConfig()

	// Initialize GitHub client
	log.Println("INFO: Initializing GitHub client...")
	ctx := context.Background()
	githubClient := github.InitializeGithubClient(cfg, ctx)

	// Fetch PR details including diff
	log.Printf("INFO: Fetching PR details for PR #%d...", cfg.PRNumber)
	prDetails, err := github.FetchPRDetails(githubClient, cfg)
	if err != nil {
		log.Fatalf("ERROR: Failed to fetch PR details: %v", err)
	}
	log.Printf("INFO: Successfully fetched PR details for '%s'", prDetails["title"])

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
	} else {
		log.Printf("INFO: Found %d alert suggestions!", len(*suggestions))

		// Log the suggestions
		for i, suggestion := range *suggestions {
			log.Printf("INFO: Alert %d: %s (%s) - Priority: %s", i+1, suggestion.Name, suggestion.Type, suggestion.Priority)
		}

		// Create PR comments if suggestions exist
		log.Println("INFO: Creating PR comments for alert suggestions...")
		err := github.CreateAlertsPRComments(*suggestions, prDetails, cfg)
		if err != nil {
			log.Fatalf("ERROR: Failed to create Alerts PR comments: %v", err)
		}
		log.Println("INFO: Successfully created PR comments")

		// Ask user if they want to create the alerts now
		reader := bufio.NewReader(os.Stdin)
		log.Println("\nINFO: Do you want to create these alerts now? (y/n)")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		if input == "y" || input == "yes" {
			log.Println("INFO: Creating alerts...")
			for _, suggestion := range *suggestions {
				if suggestion.Type == "prometheus" || suggestion.Type == "metric" {
					log.Printf("INFO: Creating Prometheus alert '%s'...", suggestion.Name)
					err := alerts.CreatePrometheusAlert(suggestion, cfg)
					if err != nil {
						log.Printf("ERROR: Failed to create Prometheus alert '%s': %v", suggestion.Name, err)
					} else {
						log.Printf("INFO: Successfully created Prometheus alert: %s", suggestion.Name)
					}
				} else if suggestion.Type == "datadog" {
					log.Printf("INFO: Creating Datadog alert '%s'...", suggestion.Name)
					err := alerts.CreateDatadogAlert(suggestion, cfg)
					if err != nil {
						log.Printf("ERROR: Failed to create Datadog alert '%s': %v", suggestion.Name, err)
					} else {
						log.Printf("INFO: Successfully created Datadog alert: %s", suggestion.Name)
					}
				}
			}
			log.Println("INFO: Completed alert creation")
		} else {
			log.Println("INFO: Alert creation skipped. You can create them later from the PR comments.")
		}
	}
	log.Println("INFO: Alerts analysis complete")
}
