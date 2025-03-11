// cmd/check.go
package cmd

import (
	"PRism/config"
	"PRism/github"
	"PRism/llm"
	"context"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	cfg := loadConfig()

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
	suggestions, err, responseText := llm.CallClaudeAPI(prompt, cfg)
	if err != nil {
		log.Fatalf("Error calling Claude API: %v", err)
	}

	if suggestions == nil {
		fmt.Println("No suggestions found")
		fmt.Println("Response text:")
		fmt.Println(responseText)
	} else {
		fmt.Println("Suggestions found:")
		fmt.Println(suggestions)

		// Create PR comments if suggestions exist
		err := github.CreatePRComments(*suggestions, prDetails, cfg)
		if err != nil {
			log.Fatalf("Error creating PR comments: %v", err)
		}
	}
}

func loadConfig() config.Config {
	cfg := config.Config{
		GithubToken:   viper.GetString("github_token"),
		ClaudeAPIKey:  viper.GetString("claude_api_key"),
		RepoOwner:     viper.GetString("repo_owner"),
		RepoName:      viper.GetString("repo_name"),
		PRNumber:      viper.GetInt("pr_number"),
		PRDFilePath:   viper.GetString("prd_file"),
		OutputFormat:  viper.GetString("output_format"),
		MaxDiffSize:   viper.GetInt("max_diff_size"),
		ClaudeModel:   viper.GetString("claude_model"),
		ClaudeBaseURL: viper.GetString("claude_base_url"),
	}

	// Validate required parameters
	if cfg.GithubToken == "" {
		log.Fatal("GitHub token is required. Set GITHUB_TOKEN env var or use --github-token flag")
	}
	if cfg.ClaudeAPIKey == "" {
		log.Fatal("Claude API key is required. Set CLAUDE_API_KEY env var or use --claude-api-key flag")
	}
	if cfg.RepoOwner == "" || cfg.RepoName == "" || cfg.PRNumber == 0 {
		log.Fatal("Repository details and PR number are required. Set REPO_OWNER, REPO_NAME, PR_NUMBER env vars or use flags")
	}

	return cfg
}
