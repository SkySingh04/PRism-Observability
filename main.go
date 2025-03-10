package main

import (
	"bytes"
	"context" 
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v53/github"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

// Config holds configuration for the application
type Config struct {
	GithubToken   string
	ClaudeAPIKey  string
	RepoOwner     string
	RepoName      string
	PRNumber      int
	PRDFilePath   string
	OutputFormat  string
	MaxDiffSize   int
	ClaudeModel   string
	ClaudeBaseURL string
}

// ObservabilityRecommendation represents the recommendations from Claude
type ObservabilityRecommendation struct {
	EventTrackingRecommendations []EventTrackingRec `json:"event_tracking"`
	AlertingRules               []AlertingRule     `json:"alerting_rules"`
	DashboardRecommendations     []Dashboard        `json:"dashboards"`
	GeneralAdvice               string             `json:"general_advice"`
}

type EventTrackingRec struct {
	EventName       string   `json:"event_name"`
	Properties      []string `json:"properties"`
	Implementation  string   `json:"implementation"`
	ContextualInfo  string   `json:"contextual_info"`
	Location        string   `json:"location"`
}

type AlertingRule struct {
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	Query          string   `json:"query"`
	Threshold      string   `json:"threshold"`
	Severity       string   `json:"severity"`
	Implementation string   `json:"implementation"`
}

type Dashboard struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Charts      []Chart  `json:"charts"`
	Platform    string   `json:"platform"`
}

type Chart struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Query       string `json:"query"`
	ChartType   string `json:"chart_type"`
}

// ClaudeRequest represents the request structure for Claude API
type ClaudeRequest struct {
	Model            string    `json:"model"`
	MaxTokens        int       `json:"max_tokens"`
	Temperature      float64   `json:"temperature"`
	Messages         []Message `json:"messages"`
	System           string    `json:"system"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ClaudeResponse represents the response structure from Claude API
type ClaudeResponse struct {
	ID      string `json:"id"`
	Content struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

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
	prompt := buildPrompt(prDetails, prdContent)

	// Call Claude API
	recommendations, err := callClaudeAPI(prompt, config)
	if err != nil {
		log.Fatalf("Error calling Claude API: %v", err)
	}

	// Output recommendations
	outputRecommendations(recommendations, config)
}

func parseFlags() Config {
	config := Config{}

	// Set defaults from environment variables
	githubToken := getEnv("GITHUB_TOKEN", "")
	claudeAPIKey := getEnv("CLAUDE_API_KEY", "")
	repoOwner := getEnv("REPO_OWNER", "")
	repoName := getEnv("REPO_NAME", "")
	prNumberStr := getEnv("PR_NUMBER", "0")
	prNumber, _ := strconv.Atoi(prNumberStr)
	prdFile := getEnv("PRD_FILE", "")
	outputFormat := getEnv("OUTPUT_FORMAT", "json")
	maxDiffSizeStr := getEnv("MAX_DIFF_SIZE", "10000")
	maxDiffSize, _ := strconv.Atoi(maxDiffSizeStr)
	claudeModel := getEnv("CLAUDE_MODEL", "claude-3-7-sonnet-20250219")
	claudeBaseURL := getEnv("CLAUDE_BASE_URL", "https://api.anthropic.com/v1/messages")

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

// Helper function to get environment variable with fallback default
func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func fetchPRDetails(client *github.Client, config Config) (map[string]interface{}, error) {
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
		if totalDiffSize + patchSize > config.MaxDiffSize {
			continue
		}
		totalDiffSize += patchSize
		
		fileDetail := map[string]interface{}{
			"filename": file.GetFilename(),
			"status": file.GetStatus(),
			"additions": file.GetAdditions(),
			"deletions": file.GetDeletions(),
			"patch": file.GetPatch(),
		}
		fileDetails = append(fileDetails, fileDetail)
	}
	
	result["files"] = fileDetails
	result["commits"] = len(commits)
	
	return result, nil
}

func buildPrompt(prDetails map[string]interface{}, prdContent string) string {
	var b strings.Builder
	
	b.WriteString("# Observability Analysis Request\n\n")
	b.WriteString("As an AI observability assistant, analyze the following PR and PRD to suggest:\n")
	b.WriteString("1. Event tracking recommendations (Amplitude, OpenTelemetry)\n")
	b.WriteString("2. Alerting rules (Datadog, Grafana)\n")
	b.WriteString("3. Dashboards and charts (Amplitude, Datadog, Grafana)\n\n")
	
	// Add PR details
	b.WriteString("## Pull Request Details\n\n")
	b.WriteString(fmt.Sprintf("Title: %s\n", prDetails["title"]))
	b.WriteString(fmt.Sprintf("Description: %s\n", prDetails["description"]))
	b.WriteString(fmt.Sprintf("Author: %s\n", prDetails["author"]))
	b.WriteString(fmt.Sprintf("Created: %s\n\n", prDetails["created_at"]))
	
	// Add file diffs
	files := prDetails["files"].([]map[string]interface{})
	b.WriteString(fmt.Sprintf("## File Changes (%d files)\n\n", len(files)))
	
	for _, file := range files {
		filename := file["filename"].(string)
		status := file["status"].(string)
		additions := file["additions"].(int)
		deletions := file["deletions"].(int)
		patch := file["patch"].(string)
		
		// Only include .go files for detailed analysis
		if filepath.Ext(filename) == ".go" || len(files) < 5 {
			b.WriteString(fmt.Sprintf("### %s (%s, +%d, -%d)\n\n", filename, status, additions, deletions))
			b.WriteString("```diff\n")
			b.WriteString(patch)
			b.WriteString("\n```\n\n")
		} else {
			b.WriteString(fmt.Sprintf("### %s (%s, +%d, -%d) - Diff omitted\n\n", filename, status, additions, deletions))
		}
	}
	
	// Add PRD if provided
	if prdContent != "" {
		b.WriteString("## Product Requirements Document\n\n")
		b.WriteString(prdContent)
		b.WriteString("\n\n")
	}
	
	// Add instructions for output format
	b.WriteString("## Instructions\n\n")
	b.WriteString("Please analyze the PR and PRD to provide recommendations for observability.")
	b.WriteString(" Respond with a JSON object containing event tracking recommendations, alerting rules, dashboard recommendations, and general advice.")
	b.WriteString(" Consider what events should be tracked, what metrics should be alerted on, and what dashboards would be useful for monitoring this code.")
	b.WriteString(" Focus on Go best practices for observability using OpenTelemetry, Amplitude SDKs, and Datadog/Grafana integrations.")
	
	return b.String()
}

func callClaudeAPI(prompt string, config Config) (*ObservabilityRecommendation, error) {
	// Prepare Claude request
	claudeReq := ClaudeRequest{
		Model:       config.ClaudeModel,
		MaxTokens:   4000,
		Temperature: 0.3,
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		System: "You are an AI observability assistant that analyzes Go code changes and PRDs to suggest event tracking, alerting rules, and dashboards. Provide specific, actionable recommendations that follow observability best practices. Your recommendations should be relevant to the changes and detailed enough to implement.",
	}
	
	reqBody, err := json.Marshal(claudeReq)
	if err != nil {
		return nil, fmt.Errorf("error marshaling Claude request: %v", err)
	}
	
	// Create HTTP request
	req, err := http.NewRequest("POST", config.ClaudeBaseURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("error creating HTTP request: %v", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", config.ClaudeAPIKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	
	// Execute request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing HTTP request: %v", err)
	}
	defer resp.Body.Close()
	
	// Read response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error from Claude API: %s", string(body))
	}
	
	// Parse Claude response
	var claudeResp ClaudeResponse
	if err := json.Unmarshal(body, &claudeResp); err != nil {
		return nil, fmt.Errorf("error parsing Claude response: %v", err)
	}
	
	// Extract JSON from Claude's response
	jsonStr := extractJSONFromText(claudeResp.Content.Text)
	if jsonStr == "" {
		return nil, fmt.Errorf("no JSON found in Claude's response")
	}
	
	// Parse the recommendations
	var recommendations ObservabilityRecommendation
	if err := json.Unmarshal([]byte(jsonStr), &recommendations); err != nil {
		return nil, fmt.Errorf("error parsing recommendations: %v", err)
	}
	
	return &recommendations, nil
}

func extractJSONFromText(text string) string {
	// Look for JSON between ```json and ``` or just {}
	startIdx := strings.Index(text, "```json")
	if startIdx != -1 {
		startIdx += 7 // Length of "```json"
		endIdx := strings.Index(text[startIdx:], "```")
		if endIdx != -1 {
			return strings.TrimSpace(text[startIdx : startIdx+endIdx])
		}
	}
	
	// Try finding JSON between { and }
	startIdx = strings.Index(text, "{")
	if startIdx != -1 {
		// Find the matching closing brace
		braceCount := 1
		for i := startIdx + 1; i < len(text); i++ {
			if text[i] == '{' {
				braceCount++
			} else if text[i] == '}' {
				braceCount--
				if braceCount == 0 {
					return text[startIdx : i+1]
				}
			}
		}
	}
	
	return ""
}

func outputRecommendations(recommendations *ObservabilityRecommendation, config Config) {
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

func outputMarkdown(recommendations *ObservabilityRecommendation) {
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