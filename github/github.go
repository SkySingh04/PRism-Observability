package github

import (
	"PRism/config"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/go-github/v53/github"
	"golang.org/x/oauth2"
)

// initializeGithubClient creates and returns a GitHub client with proper authentication
func InitializeGithubClient(config config.Config, ctx context.Context) *github.Client {
	log.Printf("Initializing GitHub client")
	return github.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: config.GithubToken},
	)))
}

func FetchPRDetails(client *github.Client, config config.Config) (map[string]interface{}, error) {
	log.Printf("Fetching PR details for PR #%d in %s/%s", config.PRNumber, config.RepoOwner, config.RepoName)
	result := make(map[string]interface{})

	// Fetch PR details
	pr, _, err := client.PullRequests.Get(
		context.Background(),
		config.RepoOwner,
		config.RepoName,
		config.PRNumber,
	)
	if err != nil {
		log.Printf("Error fetching PR details: %v", err)
		return nil, fmt.Errorf("error fetching PR details: %v", err)
	}

	result["title"] = pr.GetTitle()
	result["description"] = pr.GetBody()
	result["author"] = pr.GetUser().GetLogin()
	result["created_at"] = pr.GetCreatedAt().Format(time.RFC3339)

	// Fetch PR diff
	opt := &github.ListOptions{}
	ctx := context.Background()
	log.Printf("Fetching commits for PR #%d", config.PRNumber)
	commits, _, err := client.PullRequests.ListCommits(
		ctx,
		config.RepoOwner,
		config.RepoName,
		config.PRNumber,
		opt,
	)
	if err != nil {
		log.Printf("Error fetching PR commits: %v", err)
		return nil, fmt.Errorf("error fetching PR commits: %v", err)
	}

	// Get PR files (diff)
	log.Printf("Fetching files for PR #%d", config.PRNumber)
	files, _, err := client.PullRequests.ListFiles(
		context.Background(),
		config.RepoOwner,
		config.RepoName,
		config.PRNumber,
		opt,
	)
	if err != nil {
		log.Printf("Error fetching PR files: %v", err)
		return nil, fmt.Errorf("error fetching PR files: %v", err)
	}

	// Process files
	fileDetails := []map[string]interface{}{}
	totalDiffSize := 0

	log.Printf("Processing %d files from PR", len(files))
	for _, file := range files {
		// Check if we're exceeding max diff size
		patchSize := len(file.GetPatch())
		if totalDiffSize+patchSize > config.MaxDiffSize {
			log.Printf("Skipping file %s as it would exceed max diff size", file.GetFilename())
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

	log.Printf("Successfully fetched PR details with %d files and %d commits", len(fileDetails), len(commits))
	return result, nil
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
