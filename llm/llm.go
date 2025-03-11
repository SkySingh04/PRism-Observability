package llm

import (
	"PRism/config"
	"PRism/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func CallClaudeAPIForObservability(prompt string, configStruct config.Config) (*[]config.FileSuggestion, error, string, string) {
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

	// log.Printf("Response from Claude API: %s", string(body))
	// fmt.Printf("Response from Claude API: %s", string(body))

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
	suggestions, err := utils.ParseLLMSuggestionsForObservability(responseText)
	if err != nil {
		return nil, fmt.Errorf("error parsing suggestions: %v", err), responseText, ""
	}

	summary, err := utils.ParseLLMSummary(responseText)
	// log.Println("Summary:")
	// log.Println(summary)
	if err != nil {
		return nil, fmt.Errorf("error parsing summary: %v", err), responseText, ""
	}

	return &suggestions, nil, responseText, summary
}
func CallClaudeAPIForDashboards(prompt string, configStruct config.Config) (*[]config.DashboardSuggestion, error, string, string) {
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
	// log.Printf("Response from Claude API: %s", string(body))
	// fmt.Printf("Response from Claude API: %s", string(body))

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
	suggestions, err := utils.ParseLLMSuggestionsForDashboards(responseText)
	if err != nil {
		return nil, fmt.Errorf("error parsing suggestions: %v", err), responseText, ""
	}

	summary, err := utils.ParseLLMSummary(responseText)
	// log.Println("Summary:")
	// log.Println(summary)
	if err != nil {
		return nil, fmt.Errorf("error parsing summary: %v", err), responseText, ""
	}

	return &suggestions, nil, responseText, summary
}
func CallClaudeAPIForAlerts(prompt string, configStruct config.Config) (*[]config.AlertSuggestion, error, string, string) {
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
	// log.Printf("Response from Claude API: %s", string(body))
	// fmt.Printf("Response from Claude API: %s", string(body))

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
	suggestions, err := utils.ParseLLMSuggestionsForAlerts(responseText)
	if err != nil {
		return nil, fmt.Errorf("error parsing suggestions: %v", err), responseText, ""
	}

	summary, err := utils.ParseLLMSummary(responseText)
	// log.Println("Summary:")
	// log.Println(summary)
	if err != nil {
		return nil, fmt.Errorf("error parsing summary: %v", err), responseText, ""
	}

	return &suggestions, nil, responseText, summary
}
