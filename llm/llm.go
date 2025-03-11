package llm

import (
	"PRism/config"
	"PRism/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
)

func BuildObservabilityPrompt(prDetails map[string]interface{}, prdContent string) string {
	var b strings.Builder

	b.WriteString("# Observability Instrumentation Analysis\n\n")
	b.WriteString("As an AI observability assistant, analyze the following PR and PRD to suggest code changes for:\n")
	b.WriteString("1. OpenTelemetry instrumentation (spans, metrics, attributes)\n")
	b.WriteString("2. Logging statements at appropriate locations\n")
	b.WriteString("3. Event tracking code (Amplitude)\n\n")

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

	// Instructions focused only on code instrumentation
	b.WriteString("## Instructions\n\n")
	b.WriteString("For each file in the PR, suggest specific code changes as git diff format that can be posted as inline comments. Your suggestions should:\n\n")

	b.WriteString("1. Add OpenTelemetry instrumentation:\n")
	b.WriteString("   - Create spans for functions/methods\n")
	b.WriteString("   - Add attributes to spans for context\n")
	b.WriteString("   - Track errors and set status accordingly\n\n")

	b.WriteString("2. Add appropriate logging:\n")
	b.WriteString("   - Log entry/exit of important functions\n")
	b.WriteString("   - Log errors with context\n")
	b.WriteString("   - Add debug logs for complex operations\n\n")

	b.WriteString("3. Add event tracking where relevant:\n")
	b.WriteString("   - User actions\n")
	b.WriteString("   - System events\n")
	b.WriteString("   - Performance metrics\n\n")

	b.WriteString("IMPORTANT : Also, provide a summary paragraph of all the suggested changes starting with SUMMARY:, along with the reason for each change and sort them by priority (High, Medium, Low).\n\n")

	b.WriteString("Format each suggestion as follows:\n")
	b.WriteString("```\n")
	b.WriteString("FILE: filename.go\n")
	b.WriteString("LINE: 42\n")
	b.WriteString("SUGGESTION:\n")
	b.WriteString("```diff\n")
	b.WriteString("// Add after line 42\n")
	b.WriteString("+ span := otel.StartSpan(ctx, \"functionName\")\n")
	b.WriteString("+ defer span.End()\n")
	b.WriteString("```\n")

	b.WriteString("Follow Go best practices and match the existing code style. Only suggest changes related to observability instrumentation.")

	return b.String()
}

func CallClaudeAPI(prompt string, configStruct config.Config) (*[]config.FileSuggestion, error, string, string) {
	// Prepare Claude request
	claudeReq := config.ClaudeRequest{
		Model:       configStruct.ClaudeModel,
		MaxTokens:   4000,
		Temperature: 0.3,
		Messages: []config.Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		System: "You are an AI observability assistant that analyzes Go code changes and PRDs to suggest event tracking, alerting rules, and dashboards. Provide specific, actionable recommendations that follow observability best practices. Your recommendations should be relevant to the changes and detailed enough to implement.",
	}

	reqBody, err := json.Marshal(claudeReq)
	if err != nil {
		return nil, fmt.Errorf("error marshaling Claude request: %v", err), "", ""
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", configStruct.ClaudeBaseURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("error creating HTTP request: %v", err), "", ""
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", configStruct.ClaudeAPIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	// Execute request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing HTTP request: %v", err), "", ""
	}
	defer resp.Body.Close()

	// Read response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err), "", ""
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error from Claude API: %s", string(body)), "", ""
	}

	// Parse Claude response
	var claudeResp config.ClaudeResponse
	if err := json.Unmarshal(body, &claudeResp); err != nil {
		return nil, fmt.Errorf("error parsing Claude response: %v", err), "", ""
	}

	// Extract text from the array of content
	var responseText string
	for _, content := range claudeResp.Content {
		if content.Type == "text" {
			responseText += content.Text
		}
	}

	// // Check if response is LGTM
	// if strings.Contains(responseText, "LGTM") {
	// 	// Return empty recommendations for LGTM case
	// 	return &config.ObservabilityRecommendation{}, nil, responseText
	// }

	// Parse suggestions for PR comments
	suggestions, err := utils.ParseLLMSuggestions(responseText)
	if err != nil {
		return nil, fmt.Errorf("error parsing suggestions: %v", err), responseText, ""
	}

	summary, err := utils.ParseLLMSummary(responseText)
	if err != nil {
		return nil, fmt.Errorf("error parsing summary: %v", err), responseText, ""
	}

	return &suggestions, nil, responseText, summary
}
