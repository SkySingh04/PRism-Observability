// cmd/dashboard.go
package cmd

import (
	"PRism/config"
	"PRism/dashboard"
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
	suggestions, err, _, summary := llm.CallClaudeAPIForDashboards(prompt, cfg)
	if err != nil {
		log.Fatalf("Error calling Claude API: %v", err)
	}

	if suggestions == nil || len(*suggestions) == 0 {
		log.Println("No dashboard suggestions found")
	} else {
		log.Printf("Found %d dashboard suggestions!", len(*suggestions))

		// Log the suggestions
		for i, suggestion := range *suggestions {
			log.Printf("Dashboard %d: %s (%s) - Priority: %s", i+1, suggestion.Name, suggestion.Type, suggestion.Priority)
		}

		// Create PR comments if suggestions exist
		err := github.CreateDashboardPRComments(*suggestions, prDetails, cfg, summary)
		if err != nil {
			log.Fatalf("Error creating Dashboard PR comments: %v", err)
		}

		// Ask user if they want to create the dashboards now
		reader := bufio.NewReader(os.Stdin)
		log.Println("\nDo you want to create these dashboards now? (y/n)")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		if input == "y" || input == "yes" {
			for _, suggestion := range *suggestions {
				if suggestion.Type == "grafana" {
					err := dashboard.CreateGrafanaDashboard(suggestion, cfg)
					if err != nil {
						log.Printf("Error creating Grafana dashboard '%s': %v", suggestion.Name, err)
					} else {
						log.Printf("Successfully created Grafana dashboard: %s", suggestion.Name)
					}
				} else if suggestion.Type == "amplitude" {
					err := dashboard.CreateAmplitudeDashboard(suggestion, cfg)
					if err != nil {
						log.Printf("Error creating Amplitude dashboard '%s': %v", suggestion.Name, err)
					} else {
						log.Printf("Successfully created Amplitude dashboard: %s", suggestion.Name)
					}
				}
			}
		} else {
			log.Println("Dashboard creation skipped. You can create them later from the PR comments.")
		}
	}
}
