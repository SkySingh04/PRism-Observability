# PRism - PR Observability and Analysis Tool

PRism is an AI-powered tool that analyzes GitHub pull requests using Claude AI to provide observability recommendations, create dashboards, and manage alerts. It helps improve code quality and observability standards in your projects.

## Project Structure

```
.
├── .env                # Environment variables configuration
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
├── prd.md              # Product Requirements Document
└── utils/              # Utility functions
    ├── parse.go        # Parsing utilities for LLM responses
    └── utils.go        # General utilities
```

## Features

-   **Observability Analysis:** Analyzes pull requests to suggest improvements for:
    -   OpenTelemetry instrumentation
    -   Logging statements
    -   Event tracking
-   **AI-Powered Recommendations:** Uses Claude AI to provide context-aware suggestions based on code changes and PRD
-   **Dashboard Generation:** Creates Grafana and Amplitude dashboards with:
    -   Service-level metrics visualization
    -   Error rate tracking
    -   Performance monitoring
-   **GitHub Integration:** Directly interacts with GitHub to:
    -   Analyze PR diffs
    -   Add inline code suggestions
    -   Post summary comments

## Installation

1.  Clone the repository:

    ```bash
    git clone https://github.com/yourusername/prism.git
    cd prism
    ```
2.  Create a `.env` file with your configuration (see Configuration section)
3.  Build the application:

    ```bash
    go build -o prism
    ```

## Usage

### Check a Pull Request for Observability Improvements

```bash
./prism check --repo-owner=<owner> --repo-name=<repo> --pr-number=<number>
```

### Generate Dashboards from PR Analysis

```bash
./prism dashboard --repo-owner=<owner> --repo-name=<repo> --pr-number=<number> --dashboard-type=<grafana|amplitude>
```

### Manage PR Alerts

```bash
./prism alerts --repo-owner=<owner> --repo-name=<repo> --pr-number=<number> --alert-type=<missing-logs|missing-metrics>
```

### Interactive Chat Mode

```bash
./prism chat --repo-owner=<owner> --repo-name=<repo> --pr-number=<number>
```

## Configuration

PRism can be configured using environment variables, command-line flags, or a config file.

### Environment Variables

Create a `.env` file with the following variables:

```
# GitHub Configuration
GITHUB_TOKEN=your_github_token
REPO_OWNER=repository_owner
REPO_NAME=repository_name
PR_NUMBER=pull_request_number

# Claude AI Configuration
CLAUDE_API_KEY=your_claude_api_key
CLAUDE_MODEL=claude-3-7-sonnet-20250219
CLAUDE_BASE_URL=https://api.anthropic.com/v1/messages

# Application Configuration
PRD_FILE=./prd.md
OUTPUT_FORMAT=markdown
MAX_DIFF_SIZE=10000

# Dashboard Configuration
GRAFANA_SERVICE_ACCOUNT_TOKEN=your_grafana_token
GRAFANA_URL=your_grafana_url
AMPLITUDE_API_KEY=your_amplitude_api_key
AMPLITUDE_SECRET_KEY=your_amplitude_secret_key
```

### Command-line Flags

All environment variables can also be set via command-line flags:

```bash
./prism check \
  --github-token=your_token \
  --repo-owner=owner \
  --repo-name=repo \
  --pr-number=123 \
  --prd-file=./prd.md \
  --output=markdown
```

### Setting Up Prometheus for Testing

```bash
docker run --rm --detach \
  --name my-prometheus \
  --publish 9090:9090 \
  --volume prometheus-volume:/prometheus \
  --volume "$(pwd)"/alerts/prometheus.yml:/etc/prometheus/prometheus.yml \
  --volume "$(pwd)"/alerts/prometheus/rules:/etc/prometheus/rules \
  prom/prometheus
```

## Architecture

PRism follows a modular architecture:

-   **cmd:** Contains command implementations
-   **config:** Manages application configuration
-   **dashboard:** Handles dashboard creation for Grafana and Amplitude
-   **github:** Interfaces with GitHub API
-   **llm:** Manages interactions with Claude AI
-   **utils:** Provides helper functions and parsers
-   **alerts:** Contains Prometheus alert configurations and rules

## Requirements

-   Go 1.24+
-   GitHub access token
-   Claude AI API key
-   Grafana and/or Amplitude credentials (for dashboard creation)
-   Docker (for running Prometheus locally)

## Troubleshooting

- **API Rate Limits**: If you encounter GitHub API rate limits, try authenticating with a token that has higher rate limits
- **Large PRs**: For PRs with very large diffs, use the `--max-diff-size` flag to limit the analysis size
- **LLM Errors**: If Claude API returns errors, check your API key and ensure your account has sufficient credits
