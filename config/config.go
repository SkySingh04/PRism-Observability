package config

import (
	"log"

	"github.com/spf13/viper"
)

func LoadConfig() Config {
	cfg := Config{
		GithubToken:                viper.GetString("github_token"),
		ClaudeAPIKey:               viper.GetString("claude_api_key"),
		RepoOwner:                  viper.GetString("repo_owner"),
		RepoName:                   viper.GetString("repo_name"),
		PRNumber:                   viper.GetInt("pr_number"),
		PRDFilePath:                viper.GetString("prd_file"),
		OutputFormat:               viper.GetString("output_format"),
		MaxDiffSize:                viper.GetInt("max_diff_size"),
		ClaudeModel:                viper.GetString("claude_model"),
		ClaudeBaseURL:              viper.GetString("claude_base_url"),
		AmplitudeSecretKey:         viper.GetString("amplitude_secret_key"),
		AmplitudeAPIKey:            viper.GetString("amplitude_api_key"),
		AmplitudeAPIToken:          viper.GetString("amplitude_api_token"),
		GrafanaServiceAccountToken: viper.GetString("grafana_service_account_token"),
		GrafanaURL:                 viper.GetString("grafana_url"),
		PrometheusAlertmanagerURL:  viper.GetString("prometheus_alertmanager_url"),
		PrometheusConfigPath:       viper.GetString("prometheus_config_path"),
		PrometheusAuthToken:        viper.GetString("prometheus_auth_token"),
		DatadogAPIKey:              viper.GetString("datadog_api_key"),
		DatadogAppKey:              viper.GetString("datadog_app_key"),
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

// Config holds configuration for the application
type Config struct {
	GithubToken                string
	ClaudeAPIKey               string
	RepoOwner                  string
	RepoName                   string
	PRNumber                   int
	PRDFilePath                string
	OutputFormat               string
	MaxDiffSize                int
	ClaudeModel                string
	ClaudeBaseURL              string
	GrafanaServiceAccountToken string
	GrafanaURL                 string
	AmplitudeAPIKey            string
	AmplitudeSecretKey         string
	AmplitudeAPIToken          string
	PrometheusAlertmanagerURL  string
	PrometheusAuthToken        string
	DatadogAPIKey              string
	DatadogAppKey              string
	PrometheusConfigPath       string
	PRBranch                   string
	RunningInCI                bool
}

// ObservabilityRecommendation represents the recommendations from Claude
type ObservabilityRecommendation struct {
	EventTrackingRecommendations []EventTrackingRec `json:"event_tracking"`
	AlertingRules                []AlertingRule     `json:"alerting_rules"`
	DashboardRecommendations     []Dashboard        `json:"dashboards"`
	GeneralAdvice                string             `json:"general_advice"`
}

type EventTrackingRec struct {
	EventName      string   `json:"event_name"`
	Properties     []string `json:"properties"`
	Implementation string   `json:"implementation"`
	ContextualInfo string   `json:"contextual_info"`
	Location       string   `json:"location"`
}

type AlertingRule struct {
	Name           string `json:"name"`
	Description    string `json:"description"`
	Query          string `json:"query"`
	Threshold      string `json:"threshold"`
	Severity       string `json:"severity"`
	Implementation string `json:"implementation"`
}

type Dashboard struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Charts      []Chart `json:"charts"`
	Platform    string  `json:"platform"`
}

type Chart struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Query       string `json:"query"`
	ChartType   string `json:"chart_type"`
}

// ClaudeRequest represents the request structure for Claude API
type ClaudeRequest struct {
	Model       string    `json:"model"`
	MaxTokens   int       `json:"max_tokens"`
	Temperature float64   `json:"temperature"`
	Messages    []Message `json:"messages"`
	System      string    `json:"system"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ClaudeResponse represents the response structure from Claude API
type ClaudeResponse struct {
	ID      string `json:"id"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

// FileSuggestion represents a suggested change for a specific file and line
type FileSuggestion struct {
	FileName string
	LineNum  string
	Content  string
}

// Example DashboardSuggestion struct for the config package
type DashboardSuggestion struct {
	Name     string
	Type     string
	Priority string
	Queries  string
	Panels   string
	Alerts   string
}

type AlertSuggestion struct {
	Name         string
	Type         string
	Priority     string
	Query        string
	Description  string
	Threshold    string
	Duration     string
	Notification string
	RunbookLink  string
}

// CodeEmbedding represents an embedding for a code file
type CodeEmbedding struct {
	FilePath  string    `json:"file_path"`
	Content   string    `json:"content"`
	Embedding []float32 `json:"embedding"`
}
