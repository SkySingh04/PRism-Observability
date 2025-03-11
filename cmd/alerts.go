// cmd/alerts.go
package cmd

import (
	"PRism/config"
	"PRism/github"
	"PRism/llm"
	"context"
	"io/ioutil"
	"log"

	"github.com/spf13/cobra"
)

var alertsCmd = &cobra.Command{
	Use:   "alerts",
	Short: "Manage PR alerts",
	Long: `Configure and view alerts for PR issues and observability concerns.
Alerts can be set for specific patterns or thresholds.`,
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
	suggestions, err, _, summary := llm.CallClaudeAPIForAlerts(prompt, cfg)
	if err != nil {
		log.Fatalf("Error calling Claude API: %v", err)
	}

	if suggestions == nil {
		log.Println("No suggestions found")
		// log.Println("Response text:")
		// log.Println(responseText)
	} else {
		log.Println("Suggestions found!")
		// log.Println(suggestions)

		// Create PR comments if suggestions exist
		err := github.CreateAlertsPRComments(*suggestions, prDetails, cfg, summary)
		if err != nil {
			log.Fatalf("Error creating Alerts PR comments: %v", err)
		}
	}
}
