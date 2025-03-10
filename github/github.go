package github

import (
	"PRism/config"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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

// CreatePRComments posts the suggestions as inline comments in the PR
func CreatePRComments(suggestions []config.FileSuggestion, prDetails map[string]interface{}, configStruct config.Config) error {
	// Get PR details
	prNumber := configStruct.PRNumber
	owner := configStruct.RepoOwner
	repo := configStruct.RepoName
	headSHA := prDetails["head_sha"].(string)

	// For each suggestion, create a review comment
	for _, suggestion := range suggestions {
		commentBody := fmt.Sprintf("```suggestion\n%s\n```", suggestion.Content)

		// Create request payload
		payload := map[string]interface{}{
			"commit_id": headSHA,
			"path":      suggestion.FileName,
			"line":      suggestion.LineNum,
			"body":      commentBody,
		}

		// Convert to JSON
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("error marshaling comment payload: %v", err)
		}

		// Create GitHub API request to post comment
		url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%d/comments",
			owner, repo, prNumber)

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("error creating HTTP request: %v", err)
		}

		req.Header.Set("Authorization", fmt.Sprintf("token %s", configStruct.GithubToken))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/vnd.github.comfort-fade-preview+json")

		// Execute request
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("error posting PR comment: %v", err)
		}
		defer resp.Body.Close()

		// Check response
		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			bodyBytes, _ := ioutil.ReadAll(resp.Body)
			return fmt.Errorf("GitHub API error (%d): %s", resp.StatusCode, string(bodyBytes))
		}
	}

	return nil
}
