package main

import (
	"PRism/cmd"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found or could not be loaded")
	}

	// config := parseFlags()

	cmd.Execute()
}

// func parseFlags() config.Config {
// 	config := config.Config{}

// 	// Set defaults from environment variables
// 	githubToken := utils.GetEnv("GITHUB_TOKEN", "")
// 	claudeAPIKey := utils.GetEnv("CLAUDE_API_KEY", "")
// 	repoOwner := utils.GetEnv("REPO_OWNER", "")
// 	repoName := utils.GetEnv("REPO_NAME", "")
// 	prNumberStr := utils.GetEnv("PR_NUMBER", "0")
// 	prNumber, _ := strconv.Atoi(prNumberStr)
// 	prdFile := utils.GetEnv("PRD_FILE", "")
// 	outputFormat := utils.GetEnv("OUTPUT_FORMAT", "json")
// 	maxDiffSizeStr := utils.GetEnv("MAX_DIFF_SIZE", "10000")
// 	maxDiffSize, _ := strconv.Atoi(maxDiffSizeStr)
// 	claudeModel := utils.GetEnv("CLAUDE_MODEL", "claude-3-7-sonnet-20250219")
// 	claudeBaseURL := utils.GetEnv("CLAUDE_BASE_URL", "https://api.anthropic.com/v1/messages")

// 	// Define flags with environment variable defaults
// 	flag.StringVar(&config.GithubToken, "github-token", githubToken, "GitHub API token")
// 	flag.StringVar(&config.ClaudeAPIKey, "claude-api-key", claudeAPIKey, "Claude API key")
// 	flag.StringVar(&config.RepoOwner, "repo-owner", repoOwner, "GitHub repository owner")
// 	flag.StringVar(&config.RepoName, "repo-name", repoName, "GitHub repository name")
// 	flag.IntVar(&config.PRNumber, "pr-number", prNumber, "GitHub PR number")
// 	flag.StringVar(&config.PRDFilePath, "prd-file", prdFile, "Path to PRD file")
// 	flag.StringVar(&config.OutputFormat, "output", outputFormat, "Output format (json, markdown)")
// 	flag.IntVar(&config.MaxDiffSize, "max-diff-size", maxDiffSize, "Maximum diff size to analyze")
// 	flag.StringVar(&config.ClaudeModel, "claude-model", claudeModel, "Claude model to use")
// 	flag.StringVar(&config.ClaudeBaseURL, "claude-base-url", claudeBaseURL, "Claude API base URL")

// 	flag.Parse()

// 	// Validate required parameters
// 	if config.GithubToken == "" {
// 		log.Fatal("GitHub token is required. Set GITHUB_TOKEN env var or use --github-token flag")
// 	}
// 	if config.ClaudeAPIKey == "" {
// 		log.Fatal("Claude API key is required. Set CLAUDE_API_KEY env var or use --claude-api-key flag")
// 	}
// 	if config.RepoOwner == "" || config.RepoName == "" || config.PRNumber == 0 {
// 		log.Fatal("Repository details and PR number are required. Set REPO_OWNER, REPO_NAME, PR_NUMBER env vars or use flags")
// 	}

// 	return config
// }
