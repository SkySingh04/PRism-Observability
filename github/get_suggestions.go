package github

import (
	"PRism/config"
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/google/go-github/v53/github"
)

func GetDashboardSuggestionsFromPR(client *github.Client, cfg config.Config) (*[]config.DashboardSuggestion, error) {
	ctx := context.Background()
	var allSuggestions []config.DashboardSuggestion

	// Get all comments on the PR
	comments, _, err := client.Issues.ListComments(
		ctx,
		cfg.RepoOwner,
		cfg.RepoName,
		cfg.PRNumber,
		&github.IssueListCommentsOptions{},
	)
	log.Println(comments)
	if err != nil {
		return nil, fmt.Errorf("error getting PR comments: %v", err)
	}

	// Process each comment to find dashboard suggestions
	for _, comment := range comments {
		body := comment.GetBody()

		// Look for our dashboard suggestion marker format
		if strings.Contains(body, "Dashboard Suggestion") {
			suggestion := parseDashboardSuggestionFromComment(body)
			if suggestion != nil {
				allSuggestions = append(allSuggestions, *suggestion)
			}
		}
	}
	log.Println(allSuggestions)

	return &allSuggestions, nil
}

func GetAlertSuggestionsFromPR(client *github.Client, cfg config.Config) (*[]config.AlertSuggestion, error) {
	ctx := context.Background()
	var allSuggestions []config.AlertSuggestion

	// Get all comments on the PR
	comments, _, err := client.Issues.ListComments(
		ctx,
		cfg.RepoOwner,
		cfg.RepoName,
		cfg.PRNumber,
		&github.IssueListCommentsOptions{},
	)
	if err != nil {
		return nil, fmt.Errorf("error getting PR comments: %v", err)
	}

	// Process each comment to find alert suggestions
	for _, comment := range comments {
		body := comment.GetBody()

		// Look for our alert suggestion marker format
		if strings.Contains(body, "Alert Suggestion") {
			suggestion := parseAlertSuggestionFromComment(body)
			if suggestion != nil {
				allSuggestions = append(allSuggestions, *suggestion)
			}
		}
	}

	return &allSuggestions, nil
}
func parseDashboardSuggestionFromComment(commentBody string) *config.DashboardSuggestion {
	// Extract name from the title line
	nameMatch := regexp.MustCompile(`## Dashboard Suggestion: (.+)`).FindStringSubmatch(commentBody)
	if len(nameMatch) < 2 {
		return nil
	}
	name := strings.TrimSpace(nameMatch[1])

	// Extract type
	typeMatch := regexp.MustCompile(`\*\*Type:\*\* (.+)`).FindStringSubmatch(commentBody)
	if len(typeMatch) < 2 {
		return nil
	}
	dashboardType := strings.TrimSpace(typeMatch[1])

	// Extract priority
	priorityMatch := regexp.MustCompile(`\*\*Priority:\*\* (.+)`).FindStringSubmatch(commentBody)
	priority := "medium" // Default
	if len(priorityMatch) >= 2 {
		priority = strings.TrimSpace(priorityMatch[1])
	}

	// Extract queries
	queriesMatch := regexp.MustCompile(`### Queries\n` + "```json\n" + `([\s\S]*?)\n` + "```").FindStringSubmatch(commentBody)
	queries := ""
	if len(queriesMatch) >= 2 {
		queries = queriesMatch[1]
	}

	// Extract panels
	panelsMatch := regexp.MustCompile(`### Panels\n` + "```json\n" + `([\s\S]*?)\n` + "```").FindStringSubmatch(commentBody)
	panels := ""
	if len(panelsMatch) >= 2 {
		panels = panelsMatch[1]
	}

	// Extract alerts
	alertsMatch := regexp.MustCompile(`### Alerts\n` + "```json\n" + `([\s\S]*?)\n` + "```").FindStringSubmatch(commentBody)
	alerts := ""
	if len(alertsMatch) >= 2 {
		alerts = alertsMatch[1]
	}

	return &config.DashboardSuggestion{
		Name:     name,
		Type:     dashboardType,
		Priority: priority,
		Queries:  queries,
		Panels:   panels,
		Alerts:   alerts,
	}
}

func parseAlertSuggestionFromComment(commentBody string) *config.AlertSuggestion {
	// Extract name from the title line
	nameMatch := regexp.MustCompile(`Alert Suggestion: (.+)`).FindStringSubmatch(commentBody)
	if len(nameMatch) < 2 {
		return nil
	}
	name := strings.TrimSpace(nameMatch[1])

	// Extract type and priority
	typeMatch := regexp.MustCompile(`\*\*Type:\*\* ([^\s]+)`).FindStringSubmatch(commentBody)
	if len(typeMatch) < 2 {
		return nil
	}
	alertType := strings.TrimSpace(typeMatch[1])

	priorityMatch := regexp.MustCompile(`\*\*Priority:\*\* ([^\s]+)`).FindStringSubmatch(commentBody)
	if len(priorityMatch) < 2 {
		return nil
	}
	priority := strings.TrimSpace(priorityMatch[1])

	// Extract query
	queryMatch := regexp.MustCompile(`Query\n\n` + "```\n" + `([\s\S]*?)\n` + "```").FindStringSubmatch(commentBody)
	query := ""
	if len(queryMatch) >= 2 {
		query = strings.TrimSpace(queryMatch[1])
	}

	// Extract description
	descMatch := regexp.MustCompile(`Description\n([^\n]+)`).FindStringSubmatch(commentBody)
	description := ""
	if len(descMatch) >= 2 {
		description = strings.TrimSpace(descMatch[1])
	}

	// Extract threshold
	thresholdMatch := regexp.MustCompile(`Threshold\n([^\n]+)`).FindStringSubmatch(commentBody)
	threshold := ""
	if len(thresholdMatch) >= 2 {
		threshold = strings.TrimSpace(thresholdMatch[1])
	}

	// Extract duration
	durationMatch := regexp.MustCompile(`Duration\n([^\n]+)`).FindStringSubmatch(commentBody)
	duration := ""
	if len(durationMatch) >= 2 {
		duration = strings.TrimSpace(durationMatch[1])
	}

	// Extract notification
	notificationMatch := regexp.MustCompile(`Notification\n([^\n]+)`).FindStringSubmatch(commentBody)
	notification := ""
	if len(notificationMatch) >= 2 {
		notification = strings.TrimSpace(notificationMatch[1])
	}

	// Extract runbook
	runbookMatch := regexp.MustCompile(`Runbook\n([^\n]+)`).FindStringSubmatch(commentBody)
	runbook := ""
	if len(runbookMatch) >= 2 {
		runbook = strings.TrimSpace(runbookMatch[1])
	}

	return &config.AlertSuggestion{
		Name:         name,
		Type:         alertType,
		Priority:     priority,
		Query:        query,
		Description:  description,
		Threshold:    threshold,
		Duration:     duration,
		Notification: notification,
		RunbookLink:  runbook,
	}
}
