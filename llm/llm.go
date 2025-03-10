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

func BuildPrompt(prDetails map[string]interface{}, prdContent string) string {
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

func CallClaudeAPI(prompt string, configStruct config.Config) (*config.ObservabilityRecommendation, error) {
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
		return nil, fmt.Errorf("error marshaling Claude request: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", configStruct.ClaudeBaseURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("error creating HTTP request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", configStruct.ClaudeAPIKey)
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
	var claudeResp config.ClaudeResponse
	if err := json.Unmarshal(body, &claudeResp); err != nil {
		return nil, fmt.Errorf("error parsing Claude response: %v", err)
	}
	// Extract text from the array of content
	var responseText string
	for _, content := range claudeResp.Content {
		if content.Type == "text" {
			responseText += content.Text
		}
	}

	// Extract JSON from Claude's response
	jsonStr := utils.ExtractJSONFromText(responseText)
	if jsonStr == "" {
		return nil, fmt.Errorf("no JSON found in Claude's response")
	}

	// Parse the recommendations
	var recommendations config.ObservabilityRecommendation
	if err := json.Unmarshal([]byte(jsonStr), &recommendations); err != nil {
		return nil, fmt.Errorf("error parsing recommendations: %v", err)
	}

	return &recommendations, nil
}
