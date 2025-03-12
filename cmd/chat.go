// cmd/chat.go
package cmd

import (
	"PRism/config"
	"PRism/github"
	"PRism/llm"
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Interactive chat about your repository",
	Long: `Start an interactive chat session with Claude AI about your repository.
You can ask questions about code, PRs, best practices, and more.`,
	Run: func(cmd *cobra.Command, args []string) {
		runChat()
	},
}

func init() {
	rootCmd.AddCommand(chatCmd)

	chatCmd.Flags().Bool("with-context", true, "Include repository context in the conversation")
	chatCmd.Flags().Int("pr-context", 0, "Specific PR to use as context (defaults to current PR if set)")
}

func runChat() {
	cfg := config.LoadConfig()

	log.Println("Starting PRism chat session. Type 'exit' or 'quit' to end the session.")
	log.Printf("Connected to repository: %s/%s\n", cfg.RepoOwner, cfg.RepoName)

	// Initialize context
	ctx := context.Background()
	githubClient := github.InitializeGithubClient(cfg, ctx)

	// Fetch PR details including diff
	prDetails, err := github.FetchPRDetails(githubClient, cfg)
	if err != nil {
		log.Fatalf("Error fetching PR details: %v", err)
	}

	// Start chat loop
	reader := bufio.NewReader(os.Stdin)
	conversation := []string{}

	if prDetails != nil {
		conversation = append(conversation,
			"System: The following is context about a GitHub PR being discussed:",
			fmt.Sprintf("%+v", prDetails))
	}

	for {
		fmt.Print("\nYou: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "exit" || input == "quit" {
			log.Println("Exiting chat session. Goodbye!")
			break
		}

		// Add user message to conversation
		conversation = append(conversation, "User: "+input)

		// Build complete prompt with conversation history
		prompt := strings.Join(conversation, "\n")

		// Call Claude API with conversation
		response, err := llm.SimpleClaudeChat(prompt, cfg)
		if err != nil {
			log.Printf("Error: %v\n", err)
			continue
		}

		log.Println("\nPRism: " + response)

		// Add assistant response to conversation
		conversation = append(conversation, "Assistant: "+response)

		// Limit conversation history if it gets too long
		if len(conversation) > 20 {
			conversation = conversation[len(conversation)-20:]
		}
	}
}
