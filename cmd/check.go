// cmd/check.go
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

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check a pull request for observability issues",
	Long: `Analyzes a GitHub pull request using Claude AI to identify 
potential observability issues and suggests improvements.`,
	Run: func(cmd *cobra.Command, args []string) {
		runCheck()
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
}

func runCheck() {
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
	prompt := llm.BuildObservabilityPrompt(prDetails, prdContent)

	// Call Claude API
	suggestions, err, _, summary := llm.CallClaudeAPIForObservability(prompt, cfg)
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
		err := github.CreateObservabilityPRComments(*suggestions, prDetails, cfg, summary)
		if err != nil {
			log.Fatalf("Error creating Observability PR comments: %v", err)
		}
	}
}
