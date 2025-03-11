package github

import (
	"PRism/config"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/go-github/v53/github"
	"golang.org/x/oauth2"
)

// initializeGithubClient creates and returns a GitHub client with proper authentication
func InitializeGithubClient(config config.Config, ctx context.Context) *github.Client {
	return github.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: config.GithubToken},
	)))
}

func FetchPRDetails(client *github.Client, config config.Config) (map[string]interface{}, error) {
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

// PostSummaryComment posts a summary comment to the PR's conversation
func PostSummaryComment(owner, repo string, prNumber int, summary, token string) error {
	if summary == "" {
		return nil // No summary to post
	}

	summaryPayload := map[string]interface{}{
		"body": summary,
	}

	summaryJSON, err := json.Marshal(summaryPayload)
	if err != nil {
		return fmt.Errorf("error marshaling summary payload: %v", err)
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d/comments", owner, repo, prNumber)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(summaryJSON))
	if err != nil {
		return fmt.Errorf("error creating HTTP request for summary: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("token %s", token))
	req.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error posting summary comment: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("GitHub API error (%d) for summary comment", resp.StatusCode)
	}

	return nil
}

// CreatePRComments handles inline comments and summary posting
func CreateObservabilityPRComments(suggestions []config.FileSuggestion, prDetails map[string]interface{}, configStruct config.Config, summary string) error {
	ctx := context.Background()
	client := github.NewClient(nil)

	if configStruct.GithubToken != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: configStruct.GithubToken})
		tc := oauth2.NewClient(ctx, ts)
		client = github.NewClient(tc)
	}

	// Fetch PR to get HEAD SHA
	pr, _, err := client.PullRequests.Get(ctx, configStruct.RepoOwner, configStruct.RepoName, configStruct.PRNumber)
	if err != nil {
		return fmt.Errorf("error fetching PR to get HEAD SHA: %v", err)
	}

	headSHA := pr.GetHead().GetSHA()
	if headSHA == "" {
		return fmt.Errorf("could not get HEAD SHA from PR")
	}

	// Post the summary comment
	if err := PostSummaryComment(configStruct.RepoOwner, configStruct.RepoName, configStruct.PRNumber, summary, configStruct.GithubToken); err != nil {
		return err
	}

	// Post inline comments
	for _, suggestion := range suggestions {
		commentBody := fmt.Sprintf("```suggestion\n%s\n```", suggestion.Content)

		lineNum, err := strconv.Atoi(suggestion.LineNum)
		if err != nil {
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
			return fmt.Errorf("error marshaling comment payload: %v", err)
		}

		url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%d/comments", configStruct.RepoOwner, configStruct.RepoName, configStruct.PRNumber)
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("error creating HTTP request: %v", err)
		}

		req.Header.Set("Accept", "application/vnd.github.comfort-fade-preview+json")
		req.Header.Set("Authorization", fmt.Sprintf("token %s", configStruct.GithubToken))
		req.Header.Set("Content-Type", "application/json")

		httpClient := &http.Client{}
		resp, err := httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("error posting PR comment: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			return fmt.Errorf("GitHub API error (%d): %s", resp.StatusCode, string(jsonData))
		}
	}

	return nil
}

func CreateDashboardPRComments(suggestions []config.DashboardSuggestion, prDetails map[string]interface{}, configStruct config.Config, summary string) error {
	ctx := context.Background()
	client := github.NewClient(nil)

	if configStruct.GithubToken != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: configStruct.GithubToken})
		tc := oauth2.NewClient(ctx, ts)
		client = github.NewClient(tc)
	}

	// Post the summary comment first
	if err := PostSummaryComment(configStruct.RepoOwner, configStruct.RepoName, configStruct.PRNumber, summary, configStruct.GithubToken); err != nil {
		return err
	}

	// Create a detailed comment for each dashboard suggestion
	for _, suggestion := range suggestions {
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
		commentBody += fmt.Sprintf("<!-- DASHBOARD_CREATE:%s:%s -->\n", suggestion.Type, suggestion.Name)
		commentBody += "</details>\n"

		// Post the comment
		issueComment := &github.IssueComment{
			Body: github.String(commentBody),
		}

		_, _, err := client.Issues.CreateComment(ctx, configStruct.RepoOwner, configStruct.RepoName, configStruct.PRNumber, issueComment)
		if err != nil {
			return fmt.Errorf("error posting dashboard comment: %v", err)
		}
	}

	return nil
}

func CreateAlertsPRComments(suggestions []config.AlertSuggestion, prDetails map[string]interface{}, configStruct config.Config, summary string) error {
	ctx := context.Background()
	client := github.NewClient(nil)

	if configStruct.GithubToken != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: configStruct.GithubToken})
		tc := oauth2.NewClient(ctx, ts)
		client = github.NewClient(tc)
	}

	// Post the summary comment first
	if err := PostSummaryComment(configStruct.RepoOwner, configStruct.RepoName, configStruct.PRNumber, summary, configStruct.GithubToken); err != nil {
		return err
	}

	// Create a detailed comment for each alert suggestion
	for _, suggestion := range suggestions {
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
		commentBody += fmt.Sprintf("<!-- ALERT_CREATE:%s:%s -->\n", suggestion.Type, suggestion.Name)
		commentBody += "</details>\n"

		// Post the comment
		issueComment := &github.IssueComment{
			Body: github.String(commentBody),
		}

		_, _, err := client.Issues.CreateComment(ctx, configStruct.RepoOwner, configStruct.RepoName, configStruct.PRNumber, issueComment)
		if err != nil {
			return fmt.Errorf("error posting alert comment: %v", err)
		}
	}

	return nil
}
