package config

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
