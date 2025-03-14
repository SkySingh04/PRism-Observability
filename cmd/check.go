// cmd/check.go
package cmd

import (
	"PRism/config"
	"PRism/github"
	"PRism/llm"
	"context"
	"os"

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
	log.Println("INFO: Starting PR observability check...")
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
	log.Println("INFO: Building observability analysis prompt...")
	prompt := llm.BuildObservabilityPrompt(prDetails, prdContent)

	// Call Claude API
	log.Println("INFO: Calling Claude API for observability analysis...")
	suggestions, err, _, summary := llm.CallClaudeAPIForObservability(prompt, cfg)
	if err != nil {
		log.Fatalf("ERROR: Failed to call Claude API: %v", err)
	}

	if suggestions == nil {
		log.Println("INFO: No observability suggestions found")
	} else {
		log.Printf("INFO: Found %d observability suggestions!", len(*suggestions))

		// Create PR comments if suggestions exist
		log.Println("INFO: Creating PR comments for observability suggestions...")
		err := github.CreateObservabilityPRComments(*suggestions, prDetails, cfg, summary)
		if err != nil {
			log.Fatalf("ERROR: Failed to create observability PR comments: %v", err)
		}
		log.Println("INFO: Successfully created PR comments")
	}
}
