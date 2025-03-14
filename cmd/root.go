// cmd/root.go
package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile       string
	githubToken   string
	claudeAPIKey  string
	repoOwner     string
	repoName      string
	prNumber      int
	prdFilePath   string
	outputFormat  string
	maxDiffSize   int
	claudeModel   string
	claudeBaseURL string
)
var asciiLogo = `

░▒▓███████▓▒░░▒▓███████▓▒░░▒▓█▓▒░░▒▓███████▓▒░▒▓██████████████▓▒░  
░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░▒▓█▓▒░      ░▒▓█▓▒░░▒▓█▓▒░░▒▓█▓▒░ 
░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░▒▓█▓▒░      ░▒▓█▓▒░░▒▓█▓▒░░▒▓█▓▒░ 
░▒▓███████▓▒░░▒▓███████▓▒░░▒▓█▓▒░░▒▓██████▓▒░░▒▓█▓▒░░▒▓█▓▒░░▒▓█▓▒░ 
░▒▓█▓▒░      ░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░      ░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░░▒▓█▓▒░ 
░▒▓█▓▒░      ░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░      ░▒▓█▓▒░▒▓█▓▒░░▒▓█▓▒░░▒▓█▓▒░ 
░▒▓█▓▒░      ░▒▓█▓▒░░▒▓█▓▒░▒▓█▓▒░▒▓███████▓▒░░▒▓█▓▒░░▒▓█▓▒░░▒▓█▓▒░ 
`

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "prism",
	Short: "PRism - PR Observability and Analysis Tool",
	Long: `PRism is a tool for analyzing GitHub pull requests using Claude AI.
It provides observability, alerts, and dashboard features to help you
maintain code quality and standards.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	fmt.Println(asciiLogo)
	if err := rootCmd.Execute(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.prism.yaml)")
	rootCmd.PersistentFlags().StringVar(&githubToken, "github-token", "", "GitHub API token")
	rootCmd.PersistentFlags().StringVar(&claudeAPIKey, "claude-api-key", "", "Claude API key")
	rootCmd.PersistentFlags().StringVar(&repoOwner, "repo-owner", "", "GitHub repository owner")
	rootCmd.PersistentFlags().StringVar(&repoName, "repo-name", "", "GitHub repository name")
	rootCmd.PersistentFlags().IntVar(&prNumber, "pr-number", 0, "GitHub PR number")
	rootCmd.PersistentFlags().StringVar(&prdFilePath, "prd-file", "", "Path to PRD file")
	rootCmd.PersistentFlags().StringVar(&outputFormat, "output", "json", "Output format (json, markdown)")
	rootCmd.PersistentFlags().IntVar(&maxDiffSize, "max-diff-size", 10000, "Maximum diff size to analyze")
	rootCmd.PersistentFlags().StringVar(&claudeModel, "claude-model", "claude-3-7-sonnet-20250219", "Claude model to use")
	rootCmd.PersistentFlags().StringVar(&claudeBaseURL, "claude-base-url", "https://api.anthropic.com/v1/messages", "Claude API base URL")

	// Bind flags to viper
	viper.BindPFlag("github_token", rootCmd.PersistentFlags().Lookup("github-token"))
	viper.BindPFlag("claude_api_key", rootCmd.PersistentFlags().Lookup("claude-api-key"))
	viper.BindPFlag("repo_owner", rootCmd.PersistentFlags().Lookup("repo-owner"))
	viper.BindPFlag("repo_name", rootCmd.PersistentFlags().Lookup("repo-name"))
	viper.BindPFlag("pr_number", rootCmd.PersistentFlags().Lookup("pr-number"))
	viper.BindPFlag("prd_file", rootCmd.PersistentFlags().Lookup("prd-file"))
	viper.BindPFlag("output_format", rootCmd.PersistentFlags().Lookup("output"))
	viper.BindPFlag("max_diff_size", rootCmd.PersistentFlags().Lookup("max-diff-size"))
	viper.BindPFlag("claude_model", rootCmd.PersistentFlags().Lookup("claude-model"))
	viper.BindPFlag("claude_base_url", rootCmd.PersistentFlags().Lookup("claude-base-url"))
	viper.BindPFlag("amplitude_secret_key", rootCmd.PersistentFlags().Lookup("amplitude_secret_key"))
	viper.BindPFlag("amplitude_api_key", rootCmd.PersistentFlags().Lookup("amplitude_api_key"))
	viper.BindPFlag("grafana_service_account_token", rootCmd.PersistentFlags().Lookup("grafana_service_account_token"))
	viper.BindPFlag("grafana_url", rootCmd.PersistentFlags().Lookup("grafana_url"))
	viper.BindPFlag("amplitude_api_token", rootCmd.PersistentFlags().Lookup("amplitude_api_token"))
	viper.BindPFlag("prometheus_alertmanager_url", rootCmd.PersistentFlags().Lookup("prometheus_alertmanager_url"))
	viper.BindPFlag("prometheus_auth_token", rootCmd.PersistentFlags().Lookup("prometheus_auth_token"))
	viper.BindPFlag("datadog_api_key", rootCmd.PersistentFlags().Lookup("datadog_api_key"))
	viper.BindPFlag("datadog_app_key", rootCmd.PersistentFlags().Lookup("datadog_app_key"))
	viper.BindPFlag("prometheus_config_path", rootCmd.PersistentFlags().Lookup("prometheus_config_path"))

	// Bind env variables
	viper.BindEnv("github_token", "GITHUB_TOKEN")
	viper.BindEnv("claude_api_key", "CLAUDE_API_KEY")
	viper.BindEnv("repo_owner", "REPO_OWNER")
	viper.BindEnv("repo_name", "REPO_NAME")
	viper.BindEnv("pr_number", "PR_NUMBER")
	viper.BindEnv("prd_file", "PRD_FILE")
	viper.BindEnv("output_format", "OUTPUT_FORMAT")
	viper.BindEnv("max_diff_size", "MAX_DIFF_SIZE")
	viper.BindEnv("claude_model", "CLAUDE_MODEL")
	viper.BindEnv("claude_base_url", "CLAUDE_BASE_URL")
	viper.BindEnv("amplitude_secret_key", "AMPLITUDE_SECRET_KEY")
	viper.BindEnv("amplitude_api_key", "AMPLITUDE_API_KEY")
	viper.BindEnv("grafana_service_account_token", "GRAFANA_SERVICE_ACCOUNT_TOKEN")
	viper.BindEnv("grafana_url", "GRAFANA_URL")
	viper.BindEnv("amplitude_api_token", "AMPLITUDE_API_TOKEN")
	viper.BindEnv("prometheus_alertmanager_url", "PROMETHEUS_ALERTMANAGER_URL")
	viper.BindEnv("prometheus_auth_token", "PROMETHEUS_AUTH_TOKEN")
	viper.BindEnv("datadog_api_key", "DATADOG_API_KEY")
	viper.BindEnv("datadog_app_key", "DATADOG_APP_KEY")
	viper.BindEnv("prometheus_config_path", "PROMETHEUS_CONFIG_PATH")
}

// initConfig reads in config file and ENV variables if set
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory
		home, err := os.UserHomeDir()
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".prism" (without extension)
		viper.AddConfigPath(home)
		viper.SetConfigName(".prism")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in
	if err := viper.ReadInConfig(); err == nil {
		log.Println("Using config file:", viper.ConfigFileUsed())
	}
}
