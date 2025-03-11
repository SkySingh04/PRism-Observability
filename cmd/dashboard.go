// cmd/dashboard.go
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

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Generate Grafana and Amplitude dashboard of PR metrics and data",
	Long: `Generates a Grafana and Amplitude dashboard of PR metrics and data.
You can view visualizations of PR data, trends, and other insights.`,
	Run: func(cmd *cobra.Command, args []string) {
		runDashboard()
	},
}

func init() {
	rootCmd.AddCommand(dashboardCmd)
}

func runDashboard() {
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
	prompt := llm.BuildDashboardPrompt(prDetails, prdContent)

	// Call Claude API
	suggestions, err, _, summary := llm.CallClaudeAPI(prompt, cfg)
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
		err := github.CreateDashboardPRComments(*suggestions, prDetails, cfg, summary)
		if err != nil {
			log.Fatalf("Error creating Dashboard PR comments: %v", err)
		}
	}
}
