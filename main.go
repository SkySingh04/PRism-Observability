package main

import (
	"PRism/config"
	"PRism/llm"
	"PRism/utils"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v53/github"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found or could not be loaded")
	}
	config := parseFlags()

	// Initialize GitHub client
	ctx := context.Background()
	githubClient := github.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: config.GithubToken},
	)))

	// Fetch PR details including diff
	prDetails, err := fetchPRDetails(githubClient, config)
	if err != nil {
		log.Fatalf("Error fetching PR details: %v", err)
	}

	// Read PRD content if provided
	prdContent := ""
	if config.PRDFilePath != "" {
		content, err := ioutil.ReadFile(config.PRDFilePath)
		if err != nil {
			log.Printf("Warning: Could not read PRD file: %v", err)
		} else {
			prdContent = string(content)
		}
	}

	// Prepare prompt for Claude
	prompt := llm.BuildPrompt(prDetails, prdContent)

	// Call Claude API
	recommendations, err := llm.CallClaudeAPI(prompt, config)
	if err != nil {
		log.Fatalf("Error calling Claude API: %v", err)
	}

	// Output recommendations
	outputRecommendations(recommendations, config)
}

func parseFlags() config.Config {
	config := config.Config{}

	// Set defaults from environment variables
	githubToken := utils.GetEnv("GITHUB_TOKEN", "")
	claudeAPIKey := utils.GetEnv("CLAUDE_API_KEY", "")
	repoOwner := utils.GetEnv("REPO_OWNER", "")
	repoName := utils.GetEnv("REPO_NAME", "")
	prNumberStr := utils.GetEnv("PR_NUMBER", "0")
	prNumber, _ := strconv.Atoi(prNumberStr)
	prdFile := utils.GetEnv("PRD_FILE", "")
	outputFormat := utils.GetEnv("OUTPUT_FORMAT", "json")
	maxDiffSizeStr := utils.GetEnv("MAX_DIFF_SIZE", "10000")
	maxDiffSize, _ := strconv.Atoi(maxDiffSizeStr)
	claudeModel := utils.GetEnv("CLAUDE_MODEL", "claude-3-7-sonnet-20250219")
	claudeBaseURL := utils.GetEnv("CLAUDE_BASE_URL", "https://api.anthropic.com/v1/messages")

	// Define flags with environment variable defaults
	flag.StringVar(&config.GithubToken, "github-token", githubToken, "GitHub API token")
	flag.StringVar(&config.ClaudeAPIKey, "claude-api-key", claudeAPIKey, "Claude API key")
	flag.StringVar(&config.RepoOwner, "repo-owner", repoOwner, "GitHub repository owner")
	flag.StringVar(&config.RepoName, "repo-name", repoName, "GitHub repository name")
	flag.IntVar(&config.PRNumber, "pr-number", prNumber, "GitHub PR number")
	flag.StringVar(&config.PRDFilePath, "prd-file", prdFile, "Path to PRD file")
	flag.StringVar(&config.OutputFormat, "output", outputFormat, "Output format (json, markdown)")
	flag.IntVar(&config.MaxDiffSize, "max-diff-size", maxDiffSize, "Maximum diff size to analyze")
	flag.StringVar(&config.ClaudeModel, "claude-model", claudeModel, "Claude model to use")
	flag.StringVar(&config.ClaudeBaseURL, "claude-base-url", claudeBaseURL, "Claude API base URL")

	flag.Parse()

	// Validate required parameters
	if config.GithubToken == "" {
		log.Fatal("GitHub token is required. Set GITHUB_TOKEN env var or use --github-token flag")
	}
	if config.ClaudeAPIKey == "" {
		log.Fatal("Claude API key is required. Set CLAUDE_API_KEY env var or use --claude-api-key flag")
	}
	if config.RepoOwner == "" || config.RepoName == "" || config.PRNumber == 0 {
		log.Fatal("Repository details and PR number are required. Set REPO_OWNER, REPO_NAME, PR_NUMBER env vars or use flags")
	}

	return config
}

func fetchPRDetails(client *github.Client, config config.Config) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// Fetch PR details
	pr, _, err := client.PullRequests.Get(
		context.Background(),
		config.RepoOwner,
		config.RepoName,
		config.PRNumber,
	)
	if err != nil {
		return nil, fmt.Errorf("error fetching PR details: %v", err)
	}

	result["title"] = pr.GetTitle()
	result["description"] = pr.GetBody()
	result["author"] = pr.GetUser().GetLogin()
	result["created_at"] = pr.GetCreatedAt().Format(time.RFC3339)

	// Fetch PR diff
	opt := &github.ListOptions{}
	ctx := context.Background()
	commits, _, err := client.PullRequests.ListCommits(
		ctx,
		config.RepoOwner,
		config.RepoName,
		config.PRNumber,
		opt,
	)
	if err != nil {
		return nil, fmt.Errorf("error fetching PR commits: %v", err)
	}

	// Get PR files (diff)
	files, _, err := client.PullRequests.ListFiles(
		context.Background(),
		config.RepoOwner,
		config.RepoName,
		config.PRNumber,
		opt,
	)
	if err != nil {
		return nil, fmt.Errorf("error fetching PR files: %v", err)
	}

	// Process files
	fileDetails := []map[string]interface{}{}
	totalDiffSize := 0

	for _, file := range files {
		// Check if we're exceeding max diff size
		patchSize := len(file.GetPatch())
		if totalDiffSize+patchSize > config.MaxDiffSize {
			continue
		}
		totalDiffSize += patchSize

		fileDetail := map[string]interface{}{
			"filename":  file.GetFilename(),
			"status":    file.GetStatus(),
			"additions": file.GetAdditions(),
			"deletions": file.GetDeletions(),
			"patch":     file.GetPatch(),
		}
		fileDetails = append(fileDetails, fileDetail)
	}

	result["files"] = fileDetails
	result["commits"] = len(commits)

	return result, nil
}

func outputRecommendations(recommendations *config.ObservabilityRecommendation, config config.Config) {
	if config.OutputFormat == "json" {
		output, err := json.MarshalIndent(recommendations, "", "  ")
		if err != nil {
			log.Fatalf("Error marshaling recommendations: %v", err)
		}
		fmt.Println(string(output))
	} else if config.OutputFormat == "markdown" {
		outputMarkdown(recommendations)
	} else {
		log.Fatalf("Unsupported output format: %s", config.OutputFormat)
	}
}

func outputMarkdown(recommendations *config.ObservabilityRecommendation) {
	fmt.Println("# Observability Recommendations\n")

	// Event tracking
	fmt.Println("## Event Tracking Recommendations\n")
	for _, rec := range recommendations.EventTrackingRecommendations {
		fmt.Printf("### %s\n\n", rec.EventName)
		fmt.Printf("**Properties**: %s\n\n", strings.Join(rec.Properties, ", "))
		fmt.Printf("**Implementation**:\n```go\n%s\n```\n\n", rec.Implementation)
		fmt.Printf("**Context**: %s\n\n", rec.ContextualInfo)
		fmt.Printf("**Location**: %s\n\n", rec.Location)
		fmt.Println("---\n")
	}

	// Alerting rules
	fmt.Println("## Alerting Rules\n")
	for _, rule := range recommendations.AlertingRules {
		fmt.Printf("### %s\n\n", rule.Name)
		fmt.Printf("**Description**: %s\n\n", rule.Description)
		fmt.Printf("**Query**: `%s`\n\n", rule.Query)
		fmt.Printf("**Threshold**: %s\n\n", rule.Threshold)
		fmt.Printf("**Severity**: %s\n\n", rule.Severity)
		fmt.Printf("**Implementation**:\n```\n%s\n```\n\n", rule.Implementation)
		fmt.Println("---\n")
	}

	// Dashboards
	fmt.Println("## Dashboard Recommendations\n")
	for _, dashboard := range recommendations.DashboardRecommendations {
		fmt.Printf("### %s (%s)\n\n", dashboard.Name, dashboard.Platform)
		fmt.Printf("**Description**: %s\n\n", dashboard.Description)

		fmt.Println("#### Charts\n")
		for _, chart := range dashboard.Charts {
			fmt.Printf("- **%s**: %s\n", chart.Title, chart.Description)
			fmt.Printf("  - Query: `%s`\n", chart.Query)
			fmt.Printf("  - Type: %s\n\n", chart.ChartType)
		}
		fmt.Println("---\n")
	}

	// General advice
	fmt.Println("## General Advice\n")
	fmt.Println(recommendations.GeneralAdvice)
}
