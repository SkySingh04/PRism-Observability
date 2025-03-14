package github

import (
	"PRism/config"
	"PRism/utils"
	"context"
	"fmt"
	"log"
	"path/filepath"
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

func FetchPRDetails(client *github.Client, config config.Config) (config.Config, map[string]interface{}, error) {
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
		return config, nil, fmt.Errorf("error fetching PR details: %v", err)
	}

	result["title"] = pr.GetTitle()
	result["description"] = pr.GetBody()
	result["author"] = pr.GetUser().GetLogin()
	result["created_at"] = pr.GetCreatedAt().Format(time.RFC3339)
	config.PRBranch = pr.GetHead().GetRef()
	log.Println("PR branch:", config.PRBranch)

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
		return config, nil, fmt.Errorf("error fetching PR commits: %v", err)
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
		return config, nil, fmt.Errorf("error fetching PR files: %v", err)
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
	return config, result, nil
}

// Shared function to commit an alert rule to the repository
func CommitAlertToRepository(suggestion config.AlertSuggestion, alertRule string, configPath string, cfg config.Config) error {
	fileName := utils.NormalizeFileName(suggestion.Name) + ".yml"
	repoPath := filepath.Join(configPath, fileName)

	ctx := context.Background()
	client := InitializeGithubClient(cfg, ctx)

	ref, _, err := client.Git.GetRef(ctx, cfg.RepoOwner, cfg.RepoName, fmt.Sprintf("refs/heads/%s", cfg.PRBranch))
	if err != nil {
		return fmt.Errorf("failed to get reference to branch: %w", err)
	}

	commit, _, err := client.Git.GetCommit(ctx, cfg.RepoOwner, cfg.RepoName, *ref.Object.SHA)
	if err != nil {
		return fmt.Errorf("failed to get commit: %w", err)
	}

	blob, _, err := client.Git.CreateBlob(ctx, cfg.RepoOwner, cfg.RepoName, &github.Blob{
		Content:  github.String(alertRule),
		Encoding: github.String("utf-8"),
	})
	if err != nil {
		return fmt.Errorf("failed to create blob: %w", err)
	}

	entries := []*github.TreeEntry{
		{
			Path: github.String(repoPath),
			Mode: github.String("100644"),
			Type: github.String("blob"),
			SHA:  blob.SHA,
		},
	}

	tree, _, err := client.Git.CreateTree(ctx, cfg.RepoOwner, cfg.RepoName, *commit.Tree.SHA, entries)
	if err != nil {
		return fmt.Errorf("failed to create tree: %w", err)
	}

	alertType := suggestion.Type
	newCommit, _, err := client.Git.CreateCommit(ctx, cfg.RepoOwner, cfg.RepoName, &github.Commit{
		Message: github.String(fmt.Sprintf("Add %s alert rule for %s", alertType, suggestion.Name)),
		Tree:    tree,
		Parents: []*github.Commit{commit},
	})
	if err != nil {
		return fmt.Errorf("failed to create commit: %w", err)
	}

	ref.Object.SHA = newCommit.SHA
	_, _, err = client.Git.UpdateRef(ctx, cfg.RepoOwner, cfg.RepoName, ref, false)
	if err != nil {
		return fmt.Errorf("failed to update reference: %w", err)
	}

	fmt.Printf("Added %s alert rule for %s to PR branch\n", alertType, suggestion.Name)
	return nil
}
