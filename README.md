# TracePR - AI-Powered PR Observability and Analysis Tool

TracePR is an AI-powered tool that analyzes GitHub pull requests using Claude AI to provide observability recommendations, create dashboards, and manage alerts. It helps improve code quality and observability standards in your projects by suggesting improvements for logging, metrics, tracing, and event tracking.

## Demo Videos:

### Github Integration Demo on Python PR:

https://github.com/user-attachments/assets/89e10a08-7b7e-4c46-af3c-4991662f9ab5

### CLI Demo on Typescript PR: 

https://github.com/user-attachments/assets/8c0e29c9-ad53-4d40-806c-c4824e259ce6

## Table of Contents

- [Project Overview](#project-overview)
- [Features](#features)
- [Project Structure](#project-structure)
- [Installation](#installation)
- [Usage](#usage)
  - [Check Command](#check-command)
  - [Dashboard Command](#dashboard-command)
  - [Alerts Command](#alerts-command)
  - [Chat Command](#chat-command)
- [Configuration](#configuration)
  - [Environment Variables](#environment-variables)
  - [Command-line Flags](#command-line-flags)
- [CI/CD Integration](#cicd-integration)
- [Monitoring Integrations](#monitoring-integrations)
  - [Grafana](#grafana)
  - [Amplitude](#amplitude)
  - [Prometheus](#prometheus)
  - [Datadog](#datadog)
- [How TracePR Works](#how-TracePR-works)
  - [Prompt Engineering](#prompt-engineering)
  - [Response Parsing](#response-parsing)
  - [Git Diff Generation](#git-diff-generation)
- [Architecture](#architecture)
- [Requirements](#requirements)
- [Troubleshooting](#troubleshooting)
- [Contributing](#contributing)
- [License](#license)

## Project Overview

TracePR analyzes pull requests to identify missing or inadequate observability instrumentation and suggests improvements. It leverages Claude AI to provide context-aware recommendations based on code changes and product requirements. The tool can automatically generate dashboards and alerts based on the analysis, and can be integrated into CI/CD pipelines for automated feedback.

## Features

### Observability Analysis
- **Code Review:** Analyzes pull requests to suggest improvements for:
  - OpenTelemetry instrumentation
  - Logging statements
  - Metrics collection
  - Event tracking
  - Tracing
- **AI-Powered Recommendations:** Uses Claude AI to provide context-aware suggestions based on code changes and PRD
- **Inline Comments:** Adds specific, actionable comments directly to the PR

### Dashboard Generation
- **Automated Creation:** Generates dashboards based on PR analysis
- **Multi-Platform Support:**
  - Grafana dashboards with service-level metrics
  - Amplitude dashboards for user analytics
  - Datadog dashboards for application performance
- **Customizable Panels:** Creates panels for:
  - Error rate tracking
  - Performance monitoring
  - Service health metrics
  - Custom business metrics

### Alert Management
- **Alert Rule Generation:** Creates alert rules based on PR analysis
- **Multi-Platform Support:**
  - Prometheus alerts
  - Datadog monitors
- **Threshold Configuration:** Suggests appropriate thresholds based on context
- **Alert Notifications:** Configures notification channels

### GitHub Integration
- **PR Analysis:** Analyzes PR diffs to understand code changes
- **Comment Creation:** Adds inline code suggestions and summary comments
- **Webhook Support:** Integrates with GitHub webhooks for automated analysis

### Interactive Chat
- **Context-Aware Conversations:** Chat with Claude AI about your repository
- **Code Context:** Automatically includes relevant code files in the conversation
- **PR Context:** Includes PR details for more relevant responses

## Project Structure

```
.
├── .env                # Environment variables configuration
├── .github/            # GitHub Actions workflows
│   └── workflows/      # CI/CD workflow definitions
├── .gitignore          # Git ignore file
├── alerts/             # Alert configuration and rules
│   ├── prometheus.yml  # Prometheus configuration
│   └── prometheus/     # Prometheus alert rules directory
├── cmd/                # Command-line interface commands
│   ├── alerts.go       # Manages PR alerts
│   ├── chat.go         # Interactive chat functionality
│   ├── check.go        # Checks PRs for observability issues
│   ├── dashboard.go    # Generates dashboards
│   └── root.go         # Base command setup
├── codegen/            # Code generation utilities
├── config/             # Configuration management
│   └── config.go       # Configuration loading and structures
├── dashboard/          # Dashboard creation modules
│   ├── amplitude.go    # Amplitude dashboard integration
│   └── grafana.go      # Grafana dashboard integration
├── github/             # GitHub API integration
│   └── github.go       # GitHub client and API functions
├── go.mod              # Go module definition
├── go.sum              # Go module checksums
├── llm/                # Large language model integration
│   └── llm.go          # Claude AI integration
├── main.go             # Application entry point
├── mcp/                # MCP server for Cursor integration
├── prd.md              # Product Requirements Document
├── TracePR               # Compiled binary
├── requirements.txt    # Python dependencies
└── utils/              # Utility functions
    ├── parse.go        # Parsing utilities for LLM responses
    └── utils.go        # General utilities
```

## Installation

### Prerequisites
- Go 1.21+
- GitHub access token
- Claude AI API key
- Grafana and/or Amplitude credentials (for dashboard creation)
- Prometheus and/or Datadog credentials (for alert creation)
- Docker (for running Prometheus locally)

### Steps

1. Clone the repository:
   ```bash
   git clone https://github.com/SkySingh04/TracePR.git
   cd TracePR
   ```

2. Create a `.env` file with your configuration (see [Configuration](#configuration) section)

3. Build the application:
   ```bash
   go build -o TracePR
   ```

4. Verify the installation:
   ```bash
   ./TracePR --help
   ```

## Usage

### Check Command

The `check` command analyzes a pull request for observability issues and suggests improvements.

```bash
./TracePR check --repo-owner=<owner> --repo-name=<repo> --pr-number=<number>
```

Example:
```bash
./TracePR check --repo-owner=SkySingh04 --repo-name=TracePR-observability --pr-number=6
```

### Dashboard Command

The `dashboard` command generates dashboards based on PR analysis.

```bash
./TracePR dashboard [flags]
```

Flags:
- `--create`: Create a specific dashboard
- `--create-all`: Create all suggested dashboards
- `--name`: Name of the dashboard to create (used with `--create`)
- `--type`: Type of dashboard (grafana, amplitude, datadog)
- `--skip-prompt`: Skip interactive prompts (for CI/CD)

Examples:
```bash
# Generate dashboard suggestions
./TracePR dashboard --repo-owner=<owner> --repo-name=<repo> --pr-number=<number>

# Create a specific dashboard
./TracePR dashboard --create --name="Service Metrics" --type=grafana

# Create all suggested dashboards
./TracePR dashboard --create-all
```

### Alerts Command

The `alerts` command creates alert rules based on PR analysis.

```bash
./TracePR alerts [flags]
```

Flags:
- `--create`: Create a specific alert
- `--create-all`: Create all suggested alerts
- `--name`: Name of the alert to create (used with `--create`)
- `--type`: Type of alert (prometheus, datadog)
- `--skip-prompt`: Skip interactive prompts (for CI/CD)
- `--running-in-ci`: Specify if tool is running in CI

Examples:
```bash
# Generate alert suggestions
./TracePR alerts --repo-owner=<owner> --repo-name=<repo> --pr-number=<number>

# Create a specific alert
./TracePR alerts --create --name="High Error Rate" --type=prometheus

# Create all suggested alerts
./TracePR alerts --create-all
```

### Chat Command

The `chat` command starts an interactive chat session with Claude AI about your repository.

```bash
./TracePR chat [flags]
```

Flags:
- `--with-context`: Include repository context in the conversation (default: true)
- `--pr-context`: Specific PR to use as context (defaults to current PR if set)
- `--mcp`: Run as an MCP server for Cursor integration

Examples:
```bash
# Start a chat session
./TracePR chat --repo-owner=<owner> --repo-name=<repo>

# Start a chat session with a specific PR context
./TracePR chat --pr-context=6

# Run as an MCP server for Cursor integration
./TracePR chat --mcp
```

## Configuration

TracePR can be configured using environment variables, command-line flags, or a config file.

### Environment Variables

Create a `.env` file with the following variables:

```
# GitHub Configuration
GITHUB_TOKEN=your_github_token
REPO_OWNER=repository_owner
REPO_NAME=repository_name
PR_NUMBER=pull_request_number
MAX_COMMENTS=10

# Claude AI Configuration
CLAUDE_API_KEY=your_claude_api_key
CLAUDE_MODEL=claude-3-7-sonnet-20250219
CLAUDE_BASE_URL=https://api.anthropic.com/v1/messages

# Application Configuration
PRD_FILE=./prd.md
OUTPUT_FORMAT=markdown
MAX_DIFF_SIZE=10000

# Grafana Configuration
GRAFANA_SERVICE_ACCOUNT_TOKEN=your_grafana_token
GRAFANA_URL=your_grafana_url

# Amplitude Configuration
AMPLITUDE_API_KEY=your_amplitude_api_key
AMPLITUDE_SECRET_KEY=your_amplitude_secret_key
AMPLITUDE_API_TOKEN=your_amplitude_api_token
AMPLITUDE_URL=your_amplitude_url

# Prometheus Configuration
PROMETHEUS_ALERTMANAGER_URL=http://localhost:9090
PROMETHEUS_CONFIG_PATH=./alerts/prometheus/rules

# Datadog Configuration
DATADOG_API_KEY=your_datadog_api_key
DATADOG_APP_KEY=your_datadog_app_key
```

### Command-line Flags

All environment variables can also be set via command-line flags:

```bash
./TracePR check \
  --github-token=your_token \
  --repo-owner=owner \
  --repo-name=repo \
  --pr-number=123 \
  --prd-file=./prd.md \
  --output=markdown \
  --max-diff-size=10000
```

## CI/CD Integration

TracePR can be integrated into CI/CD pipelines using GitHub Actions workflows.

### Available Workflows

1. **PR Trigger Workflow** (`.github/workflows/TracePR-pr-trigger.yml`):
   - Triggered on PR open, synchronize, or reopen
   - Runs the `check`, `dashboard`, and `alerts` commands
   - Posts results as PR comments

2. **Comment Trigger Workflow** (`.github/workflows/TracePR-comment-trigger.yml`):
   - Triggered by PR comments with specific commands
   - Supports commands like `/TracePR check`, `/TracePR dashboard`, `/TracePR alerts`
   - Posts results as PR comments

3. **Dashboard Creation Workflow** (`.github/workflows/TracePR-dashboard-creation.yml`):
   - Creates dashboards based on TracePR suggestions
   - Can be triggered manually or by other workflows

4. **Alert Creation Workflow** (`.github/workflows/TracePR-alert-creation.yml`):
   - Creates alert rules based on TracePR suggestions
   - Can be triggered manually or by other workflows

### Setting Up CI/CD

1. Add the required secrets to your GitHub repository:
   - `CLAUDE_API_KEY`: Your Claude API key
   - `GRAFANA_SERVICE_ACCOUNT_TOKEN`: Your Grafana service account token
   - `GRAFANA_URL`: Your Grafana URL
   - `AMPLITUDE_API_KEY`: Your Amplitude API key
   - `AMPLITUDE_SECRET_KEY`: Your Amplitude secret key
   - `DATADOG_API_KEY`: Your Datadog API key
   - `DATADOG_APP_KEY`: Your Datadog application key

2. The workflows will automatically use the GitHub token provided by GitHub Actions.

## Monitoring Integrations

### Grafana

TracePR can create Grafana dashboards with panels for:
- Service-level metrics
- Error rates
- Performance metrics
- Custom business metrics

Requirements:
- Grafana service account token
- Grafana URL

### Amplitude

TracePR can create Amplitude dashboards for:
- User analytics
- Event tracking
- Conversion funnels
- Retention metrics

Requirements:
- Amplitude API key
- Amplitude secret key
- Amplitude API token
- Amplitude URL

### Prometheus

TracePR can create Prometheus alert rules for:
- High error rates
- Latency thresholds
- Resource utilization
- Custom metrics

Requirements:
- Prometheus Alertmanager URL
- Prometheus configuration path

#### Setting Up Prometheus for Testing

```bash
docker run --rm --detach \
  --name my-prometheus \
  --publish 9090:9090 \
  --volume prometheus-volume:/prometheus \
  --volume "$(pwd)"/alerts/prometheus.yml:/etc/prometheus/prometheus.yml \
  --volume "$(pwd)"/alerts/prometheus/rules:/etc/prometheus/rules \
  prom/prometheus
```

### Datadog

TracePR can create Datadog monitors for:
- Service health
- Error rates
- Performance metrics
- Custom metrics

Requirements:
- Datadog API key
- Datadog application key

## How TracePR Works

TracePR leverages Claude AI to analyze pull requests and generate recommendations. This section explains how TracePR processes prompts, parses responses, and converts them to Git diffs.

### Prompt Engineering

TracePR constructs specialized prompts for Claude AI based on the task at hand:

1. **Observability Analysis Prompt Structure**:
   ```
   System: You are an observability expert reviewing a GitHub pull request.
   
   Context:
   - Repository: {repo_owner}/{repo_name}
   - PR Number: {pr_number}
   - PR Title: {pr_title}
   - PR Description: {pr_description}
   
   PR Diff:
   {pr_diff}
   
   Product Requirements (if available):
   {prd_content}
   
   Task: Analyze this PR for observability issues. Look for:
   1. Missing or inadequate logging
   2. Lack of metrics instrumentation
   3. Missing tracing spans
   4. Incomplete error handling
   5. Missing event tracking
   
   For each issue, provide:
   - File path and line number
   - Description of the issue
   - Suggested code improvement
   - Explanation of why this improvement is important
   
   Format your response as JSON with the following structure:
   {
     "summary": "Overall assessment of observability in this PR",
     "suggestions": [
       {
         "file": "path/to/file.go",
         "line": 42,
         "issue": "Description of the issue",
         "suggestion": "Suggested code improvement",
         "explanation": "Why this improvement is important"
       }
     ]
   }
   ```

2. **Dashboard Generation Prompt Structure**:
   ```
   System: You are a monitoring expert designing dashboards for a service.
   
   Context:
   - Repository: {repo_owner}/{repo_name}
   - PR Number: {pr_number}
   - PR Title: {pr_title}
   - PR Description: {pr_description}
   
   PR Diff:
   {pr_diff}
   
   Product Requirements (if available):
   {prd_content}
   
   Task: Design dashboards to monitor the changes in this PR. Create:
   1. Service-level metrics dashboards
   2. Error tracking dashboards
   3. Performance monitoring dashboards
   
   For each dashboard, provide:
   - Dashboard name
   - Description
   - Panels (with queries)
   - Alerts (with thresholds)
   
   Format your response as JSON with the following structure:
   {
     "dashboards": [
       {
         "name": "Dashboard name",
         "description": "Dashboard description",
         "type": "grafana|amplitude|datadog",
         "panels": [...],
         "queries": [...],
         "alerts": [...]
       }
     ]
   }
   ```

3. **Alert Generation Prompt Structure**:
   ```
   System: You are a monitoring expert designing alerts for a service.
   
   Context:
   - Repository: {repo_owner}/{repo_name}
   - PR Number: {pr_number}
   - PR Title: {pr_title}
   - PR Description: {pr_description}
   
   PR Diff:
   {pr_diff}
   
   Product Requirements (if available):
   {prd_content}
   
   Task: Design alerts for the changes in this PR. Create:
   1. Error rate alerts
   2. Latency alerts
   3. Resource utilization alerts
   
   For each alert, provide:
   - Alert name
   - Description
   - Query
   - Threshold
   - Severity
   
   Format your response as JSON with the following structure:
   {
     "alerts": [
       {
         "name": "Alert name",
         "description": "Alert description",
         "type": "prometheus|datadog",
         "query": "Alert query",
         "threshold": "Alert threshold",
         "severity": "critical|warning|info"
       }
     ]
   }
   ```

The prompts are constructed in the `llm/llm.go` file using the `BuildObservabilityPrompt`, `BuildDashboardPrompt`, and `BuildAlertPrompt` functions.

### Response Parsing

After receiving a response from Claude AI, TracePR parses it to extract structured data:

1. **JSON Extraction**:
   TracePR uses regular expressions to extract JSON objects from Claude's responses, which may contain explanatory text alongside the structured data.

   ```go
   func ExtractJSONFromResponse(response string) (string, error) {
       // Find JSON content between triple backticks
       re := regexp.MustCompile("```json\\s*([\\s\\S]*?)\\s*```")
       matches := re.FindStringSubmatch(response)
       
       if len(matches) > 1 {
           return matches[1], nil
       }
       
       // Try without language specifier
       re = regexp.MustCompile("```\\s*([\\s\\S]*?)\\s*```")
       matches = re.FindStringSubmatch(response)
       
       if len(matches) > 1 {
           return matches[1], nil
       }
       
       // If no triple backticks, try to find JSON object directly
       re = regexp.MustCompile("\\{[\\s\\S]*\\}")
       matches = re.FindStringSubmatch(response)
       
       if len(matches) > 0 {
           return matches[0], nil
       }
       
       return "", fmt.Errorf("no JSON found in response")
   }
   ```

2. **Structured Data Parsing**:
   The extracted JSON is then unmarshaled into Go structs for further processing:

   ```go
   func ParseObservabilitySuggestions(response string) (*[]ObservabilitySuggestion, error, string, string) {
       jsonStr, err := ExtractJSONFromResponse(response)
       if err != nil {
           return nil, err, "", ""
       }
       
       var result struct {
           Summary     string                    `json:"summary"`
           Suggestions []ObservabilitySuggestion `json:"suggestions"`
       }
       
       err = json.Unmarshal([]byte(jsonStr), &result)
       if err != nil {
           return nil, err, "", ""
       }
       
       return &result.Suggestions, nil, jsonStr, result.Summary
   }
   ```

The parsing functions are implemented in the `utils/parse.go` file.

### Git Diff Generation

TracePR converts the parsed suggestions into Git diffs for inline PR comments:

1. **Locating Code Context**:
   TracePR uses the file path and line number from each suggestion to locate the exact position in the code where the comment should be placed.

2. **Creating Inline Comments**:
   For each suggestion, TracePR creates an inline comment on the PR with:
   - The issue description
   - The suggested code improvement
   - An explanation of why the improvement is important

3. **Generating Diff Suggestions**:
   For code suggestions, TracePR generates a Git-compatible diff format that can be included in the PR comment, allowing users to directly apply the suggestion:

   ```go
   func CreateDiffSuggestion(originalCode string, suggestedCode string) string {
       // Create a diff between original and suggested code
       originalLines := strings.Split(originalCode, "\n")
       suggestedLines := strings.Split(suggestedCode, "\n")
       
       diff := difflib.UnifiedDiff{
           A:        originalLines,
           B:        suggestedLines,
           FromFile: "original",
           ToFile:   "suggested",
           Context:  3,
       }
       
       diffText, _ := difflib.GetUnifiedDiffString(diff)
       return diffText
   }
   ```

4. **Posting Comments**:
   TracePR uses the GitHub API to post the comments on the PR:

   ```go
   func CreateObservabilityPRComments(suggestions []ObservabilitySuggestion, prDetails map[string]interface{}, cfg config.Config, summary string) error {
       // Create GitHub client
       ctx := context.Background()
       client := github.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(
           &oauth2.Token{AccessToken: cfg.GithubToken},
       )))
       
       // Post summary comment
       _, _, err := client.Issues.CreateComment(
           ctx,
           cfg.RepoOwner,
           cfg.RepoName,
           cfg.PRNumber,
           &github.IssueComment{Body: github.String(summary)},
       )
       
       // Post inline comments for each suggestion
       for _, suggestion := range suggestions {
           // Create review comment with suggestion
           _, _, err = client.PullRequests.CreateComment(
               ctx,
               cfg.RepoOwner,
               cfg.RepoName,
               cfg.PRNumber,
               &github.PullRequestComment{
                   Path:     github.String(suggestion.File),
                   Line:     github.Int(suggestion.Line),
                   Body:     github.String(formatSuggestionComment(suggestion)),
                   CommitID: github.String(prDetails["head"].(map[string]interface{})["sha"].(string)),
               },
           )
       }
       
       return err
   }
   ```

The GitHub integration functions are implemented in the `github/github.go` file.

## Architecture

TracePR follows a modular architecture:

- **cmd:** Contains command implementations for the CLI
- **config:** Manages application configuration and settings
- **dashboard:** Handles dashboard creation for Grafana, Amplitude, and Datadog
- **github:** Interfaces with GitHub API for PR analysis and comment creation
- **llm:** Manages interactions with Claude AI for analysis and recommendations
- **utils:** Provides helper functions and parsers for LLM responses
- **alerts:** Contains alert configurations and rules for Prometheus and Datadog
- **mcp:** Implements MCP server for Cursor integration

## Requirements

- Go 1.21+
- GitHub access token with repo scope
- Claude AI API key
- Grafana and/or Amplitude credentials (for dashboard creation)
- Prometheus and/or Datadog credentials (for alert creation)
- Docker (for running Prometheus locally)
- Python 3.8+ with torch, numpy, and transformers (for embedding generation)

## Troubleshooting

### API Rate Limits
- If you encounter GitHub API rate limits, try authenticating with a token that has higher rate limits
- For Claude API rate limits, consider upgrading your plan or implementing rate limiting in your code

### Large PRs
- For PRs with very large diffs, use the `--max-diff-size` flag to limit the analysis size
- Consider breaking large PRs into smaller, more focused changes

### LLM Errors
- If Claude API returns errors, check your API key and ensure your account has sufficient credits
- Verify that your prompt is not exceeding the model's context window

### Dashboard Creation Errors
- Ensure your Grafana or Amplitude credentials are correct
- Check that the dashboard JSON is valid and follows the platform's schema

### Alert Creation Errors
- Verify that your Prometheus or Datadog credentials are correct
- Ensure the alert rules follow the correct syntax for the platform

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE.md) file for details.
