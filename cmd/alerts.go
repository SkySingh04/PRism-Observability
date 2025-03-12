// cmd/alerts.go
package cmd

import (
	"PRism/alerts"
	"PRism/config"
	"PRism/github"
	"PRism/llm"
	"bufio"
	"context"
	"io/ioutil"
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
	cfg := config.LoadConfig()

	// Initialize GitHub client
	ctx := context.Background()
	githubClient := github.InitializeGithubClient(cfg, ctx)

	// Fetch PR details including diff
	prDetails, err := github.FetchPRDetails(githubClient, cfg)
	if err != nil {
		log.Fatalf("Error fetching PR details: %v", err)
	}

	// Read PRD content if provided
	prdContent := ""
	if cfg.PRDFilePath != "" {
		content, err := ioutil.ReadFile(cfg.PRDFilePath)
		if err != nil {
			log.Printf("Warning: Could not read PRD file: %v", err)
		} else {
			prdContent = string(content)
		}
	}

	// Prepare prompt for Claude
	prompt := llm.BuildAlertsPrompt(prDetails, prdContent)

	// Call Claude API
	suggestions, err, responseText := llm.CallClaudeAPIForAlerts(prompt, cfg)
	if err != nil {
		log.Fatalf("Error calling Claude API: %v", err)
	}

	if suggestions == nil || len(*suggestions) == 0 {
		log.Println("No alert suggestions found")
		log.Println("Response text:")
		log.Println(responseText)
	} else {
		log.Printf("Found %d alert suggestions!", len(*suggestions))

		// Log the suggestions
		for i, suggestion := range *suggestions {
			log.Printf("Alert %d: %s (%s) - Priority: %s", i+1, suggestion.Name, suggestion.Type, suggestion.Priority)
		}

		// Create PR comments if suggestions exist
		err := github.CreateAlertsPRComments(*suggestions, prDetails, cfg)
		if err != nil {
			log.Fatalf("Error creating Alerts PR comments: %v", err)
		}

		// Ask user if they want to create the alerts now
		reader := bufio.NewReader(os.Stdin)
		log.Println("\nDo you want to create these alerts now? (y/n)")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		if input == "y" || input == "yes" {
			for _, suggestion := range *suggestions {
				if suggestion.Type == "prometheus" || suggestion.Type == "metric" {
					err := alerts.CreatePrometheusAlert(suggestion, cfg)
					if err != nil {
						log.Printf("Error creating Prometheus alert '%s': %v", suggestion.Name, err)
					} else {
						log.Printf("Successfully created Prometheus alert: %s", suggestion.Name)
					}
				} else if suggestion.Type == "datadog" || suggestion.Type == "log" {
					err := alerts.CreateDatadogAlert(suggestion, cfg)
					if err != nil {
						log.Printf("Error creating Datadog alert '%s': %v", suggestion.Name, err)
					} else {
						log.Printf("Successfully created Datadog alert: %s", suggestion.Name)
					}
				}
			}
		} else {
			log.Println("Alert creation skipped. You can create them later from the PR comments.")
		}
	}
}
