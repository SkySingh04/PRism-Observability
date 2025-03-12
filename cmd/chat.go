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

	// Generate repo embeddings for context
	repoURL := fmt.Sprintf("https://github.com/%s/%s", cfg.RepoOwner, cfg.RepoName)
	log.Printf("Generating code embeddings for repository: %s\n", repoURL)
	embeddings, err := llm.GenerateCodeEmbeddingsFromGitHub(cfg, repoURL)
	if err != nil {
		log.Printf("Warning: Error generating code embeddings: %v", err)
	} else {
		log.Printf("Successfully generated embeddings for %d files", len(embeddings))
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

		// Find relevant code files for context based on the query
		var relevantFiles []string
		var relevantFileContents []string
		if embeddings != nil && len(embeddings) > 0 {
			relevantFiles, err = llm.FindRelevantFiles(input, embeddings, cfg, 3)
			if err != nil {
				log.Printf("Warning: Error finding relevant files: %v", err)
			} else if len(relevantFiles) > 0 {
				log.Printf("Found %d relevant files for query", len(relevantFiles))

				// Add file contents to context
				for _, filePath := range relevantFiles {
					// Find the matching embedding to get content
					for _, emb := range embeddings {
						if emb.FilePath == filePath {
							relevantFileContents = append(relevantFileContents,
								fmt.Sprintf("File: %s\n%s", filePath, emb.Content))
							break
						}
					}
				}
			}
		}

		// Add user message to conversation
		conversation = append(conversation, "User: "+input)

		// Build complete prompt with conversation history and code context
		prompt := strings.Join(conversation, "\n")

		// Add code context if available
		if len(relevantFileContents) > 0 {
			codeContext := "System: The following code files are relevant to the query:\n" +
				strings.Join(relevantFileContents, "\n\n")
			prompt = codeContext + "\n\n" + prompt
		}

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
