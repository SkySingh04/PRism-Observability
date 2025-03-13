package cmd

import (
	"PRism/config"
	"PRism/github"
	"PRism/llm"
	"PRism/mcp"
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
		mcpMode, _ := cmd.Flags().GetBool("mcp")
		if mcpMode {
			runMCPServer()
		} else {
			runChat()
		}
	},
}

func init() {
	rootCmd.AddCommand(chatCmd)
	chatCmd.Flags().Bool("with-context", true, "Include repository context in the conversation")
	chatCmd.Flags().Int("pr-context", 0, "Specific PR to use as context (defaults to current PR if set)")
	chatCmd.Flags().Bool("mcp", false, "Run as an MCP server for Cursor integration")
}

func runMCPServer() {
	log.Println("INFO: Starting PRism in MCP server mode...")
	mcp.RunMCPServer()
}

func runChat() {
	cfg := config.LoadConfig()

	log.Println("INFO: Starting PRism chat session. Type 'exit' or 'quit' to end the session.")
	log.Printf("INFO: Connected to repository: %s/%s", cfg.RepoOwner, cfg.RepoName)

	// Initialize context
	ctx := context.Background()
	githubClient := github.InitializeGithubClient(cfg, ctx)

	// Fetch PR details including diff
	prDetails, err := github.FetchPRDetails(githubClient, cfg)
	if err != nil {
		log.Fatalf("ERROR: Failed to fetch PR details: %v", err)
	}

	// Generate repo embeddings for context
	repoURL := fmt.Sprintf("https://github.com/%s/%s", cfg.RepoOwner, cfg.RepoName)
	log.Printf("INFO: Generating code embeddings for repository: %s", repoURL)
	embeddings, err := llm.GenerateCodeEmbeddingsFromGitHub(cfg, repoURL)
	if err != nil {
		log.Printf("WARN: Failed to generate code embeddings: %v", err)
	} else {
		log.Printf("INFO: Successfully generated embeddings for %d files", len(embeddings))
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
			log.Println("INFO: Exiting chat session. Goodbye!")
			break
		}

		// Find relevant code files for context based on the query
		var relevantFiles []string
		var relevantFileContents []string
		if embeddings != nil {
			relevantFiles, err = llm.FindRelevantFiles(input, embeddings, cfg, 3)
			if err != nil {
				log.Printf("WARN: Failed to find relevant files: %v", err)
			} else if len(relevantFiles) > 0 {
				log.Printf("DEBUG: Found %d relevant files for query", len(relevantFiles))

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
		log.Printf("DEBUG: Sending prompt to Claude API with %d conversation entries", len(conversation))
		response, err := llm.SimpleClaudeChat(prompt, cfg)
		if err != nil {
			log.Printf("ERROR: Failed to get response from Claude API: %v", err)
			continue
		}

		log.Println("\nPRism: " + response)

		// Add assistant response to conversation
		conversation = append(conversation, "Assistant: "+response)

		// Limit conversation history if it gets too long
		if len(conversation) > 20 {
			log.Printf("DEBUG: Trimming conversation history from %d to 20 entries", len(conversation))
			conversation = conversation[len(conversation)-20:]
		}
	}
}
