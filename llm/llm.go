package llm

import (
	"tracepr/config"
	"tracepr/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

func CallClaudeAPIForObservability(prompt string, configStruct config.Config) (*[]config.FileSuggestion, error, string, string) {
	log.Printf("Calling Claude API for observability recommendations with model: %s", configStruct.ClaudeModel)

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
		log.Printf("Error marshaling Claude request: %v", err)
		return nil, fmt.Errorf("error marshaling Claude request: %v", err), "", ""
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", configStruct.ClaudeBaseURL, bytes.NewBuffer(reqBody))
	if err != nil {
		log.Printf("Error creating HTTP request: %v", err)
		return nil, fmt.Errorf("error creating HTTP request: %v", err), "", ""
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", configStruct.ClaudeAPIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	// Execute request
	log.Print("Sending request to Claude API")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error executing HTTP request: %v", err)
		return nil, fmt.Errorf("error executing HTTP request: %v", err), "", ""
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return nil, fmt.Errorf("error reading response body: %v", err), "", ""
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("Received non-200 status code from Claude API: %d", resp.StatusCode)
		return nil, fmt.Errorf("error from Claude API: %s", string(body)), "", ""
	}

	log.Print("Successfully received response from Claude API")

	// Parse Claude response
	var claudeResp config.ClaudeResponse
	if err := json.Unmarshal(body, &claudeResp); err != nil {
		log.Printf("Error parsing Claude response: %v", err)
		return nil, fmt.Errorf("error parsing Claude response: %v", err), "", ""
	}

	// Extract text from the array of content
	var responseText string
	for _, content := range claudeResp.Content {
		if content.Type == "text" {
			responseText += content.Text
		}
	}

	// Parse suggestions for PR comments
	log.Print("Parsing LLM suggestions for observability")
	suggestions, err := utils.ParseLLMSuggestionsForObservability(responseText)
	if err != nil {
		log.Printf("Error parsing suggestions: %v", err)
		return nil, fmt.Errorf("error parsing suggestions: %v", err), responseText, ""
	}

	log.Print("Parsing LLM summary")
	summary, err := utils.ParseLLMSummary(responseText)
	if err != nil {
		log.Printf("Error parsing summary: %v", err)
		return nil, fmt.Errorf("error parsing summary: %v", err), responseText, ""
	}

	log.Printf("Successfully processed Claude API response. Found %d suggestions", len(suggestions))
	return &suggestions, nil, responseText, summary
}

func CallClaudeAPIForDashboards(prompt string, configStruct config.Config) (*[]config.DashboardSuggestion, error, string, string) {
	log.Printf("Calling Claude API for dashboard recommendations with model: %s", configStruct.ClaudeModel)

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
		log.Printf("Error marshaling Claude request: %v", err)
		return nil, fmt.Errorf("error marshaling Claude request: %v", err), "", ""
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", configStruct.ClaudeBaseURL, bytes.NewBuffer(reqBody))
	if err != nil {
		log.Printf("Error creating HTTP request: %v", err)
		return nil, fmt.Errorf("error creating HTTP request: %v", err), "", ""
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", configStruct.ClaudeAPIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	// Execute request
	log.Print("Sending request to Claude API")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error executing HTTP request: %v", err)
		return nil, fmt.Errorf("error executing HTTP request: %v", err), "", ""
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return nil, fmt.Errorf("error reading response body: %v", err), "", ""
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("Received non-200 status code from Claude API: %d", resp.StatusCode)
		return nil, fmt.Errorf("error from Claude API: %s", string(body)), "", ""
	}

	log.Print("Successfully received response from Claude API")

	// Parse Claude response
	var claudeResp config.ClaudeResponse
	if err := json.Unmarshal(body, &claudeResp); err != nil {
		log.Printf("Error parsing Claude response: %v", err)
		return nil, fmt.Errorf("error parsing Claude response: %v", err), "", ""
	}

	// Extract text from the array of content
	var responseText string
	for _, content := range claudeResp.Content {
		if content.Type == "text" {
			responseText += content.Text
		}
	}

	// Parse suggestions for PR comments
	log.Print("Parsing LLM suggestions for dashboards")
	suggestions, err := utils.ParseLLMSuggestionsForDashboards(responseText)
	if err != nil {
		log.Printf("Error parsing suggestions: %v", err)
		return nil, fmt.Errorf("error parsing suggestions: %v", err), responseText, ""
	}

	log.Printf("Successfully processed Claude API response. Found %d dashboard suggestions", len(suggestions))
	return &suggestions, nil, responseText, ""
}

func CallClaudeAPIForAlerts(prompt string, configStruct config.Config) (*[]config.AlertSuggestion, error, string) {
	log.Printf("Calling Claude API for alert recommendations with model: %s", configStruct.ClaudeModel)

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
		log.Printf("Error marshaling Claude request: %v", err)
		return nil, fmt.Errorf("error marshaling Claude request: %v", err), ""
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", configStruct.ClaudeBaseURL, bytes.NewBuffer(reqBody))
	if err != nil {
		log.Printf("Error creating HTTP request: %v", err)
		return nil, fmt.Errorf("error creating HTTP request: %v", err), ""
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", configStruct.ClaudeAPIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	// Execute request
	log.Print("Sending request to Claude API")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error executing HTTP request: %v", err)
		return nil, fmt.Errorf("error executing HTTP request: %v", err), ""
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return nil, fmt.Errorf("error reading response body: %v", err), ""
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("Received non-200 status code from Claude API: %d", resp.StatusCode)
		return nil, fmt.Errorf("error from Claude API: %s", string(body)), ""
	}

	log.Print("Successfully received response from Claude API")

	// Parse Claude response
	var claudeResp config.ClaudeResponse
	if err := json.Unmarshal(body, &claudeResp); err != nil {
		log.Printf("Error parsing Claude response: %v", err)
		return nil, fmt.Errorf("error parsing Claude response: %v", err), ""
	}

	// Extract text from the array of content
	var responseText string
	for _, content := range claudeResp.Content {
		if content.Type == "text" {
			responseText += content.Text
		}
	}

	// Parse suggestions for PR comments
	log.Print("Parsing LLM suggestions for alerts")
	suggestions, err := utils.ParseLLMSuggestionsForAlerts(responseText)
	if err != nil {
		log.Printf("Error parsing suggestions: %v", err)
		return nil, fmt.Errorf("error parsing suggestions: %v", err), responseText
	}

	log.Printf("Successfully processed Claude API response. Found %d alert suggestions", len(suggestions))
	return &suggestions, nil, responseText
}

// SimpleClaudeChat sends the prompt to Claude API and returns the response
func SimpleClaudeChat(prompt string, cfg config.Config) (string, error) {
	log.Printf("Starting simple chat with Claude using model: claude-3-7-sonnet-20250219")

	// Prepare request body
	requestBody := map[string]interface{}{
		"model":       "claude-3-7-sonnet-20250219",
		"max_tokens":  1024,
		"temperature": 0.7,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		log.Printf("Error marshaling request: %v", err)
		return "", fmt.Errorf("error marshaling request: %v", err)
	}

	// Create request
	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonBody))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return "", fmt.Errorf("error creating request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", cfg.ClaudeAPIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	// Send request
	log.Print("Sending request to Claude API")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending request: %v", err)
		return "", fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Handle error responses
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("Received non-200 status code from Claude API: %d", resp.StatusCode)
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	log.Print("Successfully received response from Claude API")

	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("Error decoding response: %v", err)
		return "", fmt.Errorf("error decoding response: %v", err)
	}

	// Extract content
	content := ""
	if messages, ok := result["content"].([]interface{}); ok && len(messages) > 0 {
		if message, ok := messages[0].(map[string]interface{}); ok {
			if text, ok := message["text"].(string); ok {
				content = strings.TrimSpace(text)
			}
		}
	}

	if content == "" {
		log.Print("Could not extract content from response")
		return "", fmt.Errorf("could not extract content from response")
	}

	log.Print("Successfully extracted content from Claude API response")
	return content, nil
}
