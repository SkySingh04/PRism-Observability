package github

import (
	"PRism/config"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/google/go-github/v53/github"
	"golang.org/x/oauth2"
)

// CreatePRComments handles inline comments and summary posting
func CreateObservabilityPRComments(suggestions []config.FileSuggestion, prDetails map[string]interface{}, configStruct config.Config, summary string) error {
	log.Printf("Creating observability PR comments for PR #%d", configStruct.PRNumber)
	ctx := context.Background()
	client := github.NewClient(nil)

	if configStruct.GithubToken != "" {
		log.Printf("Configuring GitHub client with authentication token")
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: configStruct.GithubToken})
		tc := oauth2.NewClient(ctx, ts)
		client = github.NewClient(tc)
	}

	// Fetch PR to get HEAD SHA
	log.Printf("Fetching PR to get HEAD SHA")
	pr, _, err := client.PullRequests.Get(ctx, configStruct.RepoOwner, configStruct.RepoName, configStruct.PRNumber)
	if err != nil {
		log.Printf("Error fetching PR to get HEAD SHA: %v", err)
		return fmt.Errorf("error fetching PR to get HEAD SHA: %v", err)
	}

	headSHA := pr.GetHead().GetSHA()
	if headSHA == "" {
		log.Printf("Could not get HEAD SHA from PR")
		return fmt.Errorf("could not get HEAD SHA from PR")
	}

	// Post the summary comment
	log.Printf("Posting summary comment")
	if err := PostSummaryComment(configStruct.RepoOwner, configStruct.RepoName, configStruct.PRNumber, summary, configStruct.GithubToken); err != nil {
		return err
	}

	// Post inline comments
	log.Printf("Processing %d inline suggestions", len(suggestions))
	for _, suggestion := range suggestions {
		log.Printf("Creating inline comment for file %s at line %s", suggestion.FileName, suggestion.LineNum)
		commentBody := fmt.Sprintf("```suggestion\n%s\n```", suggestion.Content)

		lineNum, err := strconv.Atoi(suggestion.LineNum)
		if err != nil {
			log.Printf("Invalid line number %s: %v", suggestion.LineNum, err)
			return fmt.Errorf("invalid line number: %v", err)
		}

		payload := map[string]interface{}{
			"commit_id": headSHA,
			"path":      suggestion.FileName,
			"line":      lineNum,
			"body":      commentBody,
		}

		jsonData, err := json.Marshal(payload)
		if err != nil {
			log.Printf("Error marshaling comment payload: %v", err)
			return fmt.Errorf("error marshaling comment payload: %v", err)
		}

		url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%d/comments", configStruct.RepoOwner, configStruct.RepoName, configStruct.PRNumber)
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			log.Printf("Error creating HTTP request: %v", err)
			return fmt.Errorf("error creating HTTP request: %v", err)
		}

		req.Header.Set("Accept", "application/vnd.github.comfort-fade-preview+json")
		req.Header.Set("Authorization", fmt.Sprintf("token %s", configStruct.GithubToken))
		req.Header.Set("Content-Type", "application/json")

		httpClient := &http.Client{}
		resp, err := httpClient.Do(req)
		if err != nil {
			log.Printf("Error posting PR comment: %v", err)
			return fmt.Errorf("error posting PR comment: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			log.Printf("GitHub API error (%d): %s", resp.StatusCode, string(jsonData))
			return fmt.Errorf("GitHub API error (%d): %s", resp.StatusCode, string(jsonData))
		}
	}

	log.Printf("Successfully created all observability PR comments")
	return nil
}

func CreateDashboardPRComments(suggestions []config.DashboardSuggestion, prDetails map[string]interface{}, configStruct config.Config, summary string) error {
	log.Printf("Creating dashboard PR comments for PR #%d", configStruct.PRNumber)
	ctx := context.Background()
	client := github.NewClient(nil)

	if configStruct.GithubToken != "" {
		log.Printf("Configuring GitHub client with authentication token")
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: configStruct.GithubToken})
		tc := oauth2.NewClient(ctx, ts)
		client = github.NewClient(tc)
	}

	// Post the summary comment first
	log.Printf("Posting summary comment")
	if err := PostSummaryComment(configStruct.RepoOwner, configStruct.RepoName, configStruct.PRNumber, summary, configStruct.GithubToken); err != nil {
		return err
	}

	// Create a detailed comment for each dashboard suggestion
	log.Printf("Processing %d dashboard suggestions", len(suggestions))
	for _, suggestion := range suggestions {
		log.Printf("Creating comment for dashboard suggestion: %s", suggestion.Name)
		// Format a readable dashboard suggestion comment
		commentBody := fmt.Sprintf("## Dashboard Suggestion: %s\n\n", suggestion.Name)
		commentBody += fmt.Sprintf("**Type:** %s\n", suggestion.Type)
		commentBody += fmt.Sprintf("**Priority:** %s\n\n", suggestion.Priority)

		commentBody += "### Queries\n```json\n" + suggestion.Queries + "\n```\n\n"
		commentBody += "### Panels\n```json\n" + suggestion.Panels + "\n```\n\n"
		commentBody += "### Alerts\n```json\n" + suggestion.Alerts + "\n```\n\n"

		// Add action buttons - these will be parsed by the GitHub action
		commentBody += "<details>\n"
		commentBody += "<summary>Click to create this dashboard</summary>\n\n"
		commentBody += fmt.Sprintf("To create this dashboard, comment with:\n\n`prism dashboard --create  %s`\n\n", suggestion.Name)
		commentBody += fmt.Sprintf("<!-- DASHBOARD_CREATE:%s:%s -->\n", suggestion.Type, suggestion.Name)
		commentBody += "</details>\n"

		// Post the comment
		issueComment := &github.IssueComment{
			Body: github.String(commentBody),
		}

		_, _, err := client.Issues.CreateComment(ctx, configStruct.RepoOwner, configStruct.RepoName, configStruct.PRNumber, issueComment)
		if err != nil {
			log.Printf("Error posting dashboard comment: %v", err)
			return fmt.Errorf("error posting dashboard comment: %v", err)
		}
	}

	// Add a comment for creating all dashboards at once
	allDashboardsComment := "## Create All Dashboards\n\n"
	allDashboardsComment += "To create all suggested dashboards, comment with:\n\n`prism dashboard --create-all`\n\n"

	issueComment := &github.IssueComment{
		Body: github.String(allDashboardsComment),
	}

	_, _, err := client.Issues.CreateComment(ctx, configStruct.RepoOwner, configStruct.RepoName, configStruct.PRNumber, issueComment)
	if err != nil {
		log.Printf("Error posting all dashboards comment: %v", err)
		return fmt.Errorf("error posting all dashboards comment: %v", err)
	}

	log.Printf("Successfully created all dashboard PR comments")
	return nil
}

func CreateAlertsPRComments(suggestions []config.AlertSuggestion, prDetails map[string]interface{}, configStruct config.Config) error {
	log.Printf("Creating alerts PR comments for PR #%d", configStruct.PRNumber)
	ctx := context.Background()
	client := github.NewClient(nil)

	if configStruct.GithubToken != "" {
		log.Printf("Configuring GitHub client with authentication token")
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: configStruct.GithubToken})
		tc := oauth2.NewClient(ctx, ts)
		client = github.NewClient(tc)
	}

	// Create a detailed comment for each alert suggestion
	log.Printf("Processing %d alert suggestions", len(suggestions))
	for _, suggestion := range suggestions {
		log.Printf("Creating comment for alert suggestion: %s", suggestion.Name)
		// Format a readable alert suggestion comment
		commentBody := fmt.Sprintf("## Alert Suggestion: %s\n\n", suggestion.Name)
		commentBody += fmt.Sprintf("**Type:** %s\n", suggestion.Type)
		commentBody += fmt.Sprintf("**Priority:** %s\n\n", suggestion.Priority)

		commentBody += "### Query\n```json\n" + suggestion.Query + "\n```\n\n"
		commentBody += fmt.Sprintf("### Description\n%s\n\n", suggestion.Description)
		commentBody += fmt.Sprintf("### Threshold\n%s\n\n", suggestion.Threshold)
		commentBody += fmt.Sprintf("### Duration\n%s\n\n", suggestion.Duration)
		commentBody += fmt.Sprintf("### Notification\n%s\n\n", suggestion.Notification)

		if suggestion.RunbookLink != "" {
			commentBody += fmt.Sprintf("### Runbook\n[Link to Runbook](%s)\n\n", suggestion.RunbookLink)
		}

		// Add action buttons - these will be parsed by the GitHub action
		commentBody += "<details>\n"
		commentBody += "<summary>Click to create this alert</summary>\n\n"
		commentBody += fmt.Sprintf("To create this alert, comment with:\n\n`prism alert --create %s`\n\n", suggestion.Name)
		commentBody += fmt.Sprintf("<!-- ALERT_CREATE:%s:%s -->\n", suggestion.Type, suggestion.Name)
		commentBody += "</details>\n"

		// Post the comment
		issueComment := &github.IssueComment{
			Body: github.String(commentBody),
		}

		_, _, err := client.Issues.CreateComment(ctx, configStruct.RepoOwner, configStruct.RepoName, configStruct.PRNumber, issueComment)
		if err != nil {
			log.Printf("Error posting alert comment: %v", err)
			return fmt.Errorf("error posting alert comment: %v", err)
		}
	}

	// Add a comment for creating all alerts at once
	allAlertsComment := "## Create All Alerts\n\n"
	allAlertsComment += "To create all suggested alerts, comment with:\n\n`prism alert --create-all`\n\n"

	issueComment := &github.IssueComment{
		Body: github.String(allAlertsComment),
	}

	_, _, err := client.Issues.CreateComment(ctx, configStruct.RepoOwner, configStruct.RepoName, configStruct.PRNumber, issueComment)
	if err != nil {
		log.Printf("Error posting all alerts comment: %v", err)
		return fmt.Errorf("error posting all alerts comment: %v", err)
	}

	log.Printf("Successfully created all alert PR comments")
	return nil
}

// PostSummaryComment posts a summary comment to the PR's conversation
func PostSummaryComment(owner, repo string, prNumber int, summary, token string) error {
	if summary == "" {
		log.Printf("No summary provided, skipping comment creation")
		return nil // No summary to post
	}

	log.Printf("Posting summary comment to PR #%d in %s/%s", prNumber, owner, repo)
	summaryPayload := map[string]interface{}{
		"body": summary,
	}

	summaryJSON, err := json.Marshal(summaryPayload)
	if err != nil {
		log.Printf("Error marshaling summary payload: %v", err)
		return fmt.Errorf("error marshaling summary payload: %v", err)
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d/comments", owner, repo, prNumber)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(summaryJSON))
	if err != nil {
		log.Printf("Error creating HTTP request for summary: %v", err)
		return fmt.Errorf("error creating HTTP request for summary: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("token %s", token))
	req.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Printf("Error posting summary comment: %v", err)
		return fmt.Errorf("error posting summary comment: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		log.Printf("GitHub API error (%d) for summary comment", resp.StatusCode)
		return fmt.Errorf("GitHub API error (%d) for summary comment", resp.StatusCode)
	}

	log.Printf("Successfully posted summary comment")
	return nil
}
