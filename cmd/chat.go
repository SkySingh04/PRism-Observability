// cmd/chat.go
package cmd

import (
	"github.com/spf13/cobra"
)

var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Interactive chat about your repository",
	Long: `Start an interactive chat session with Claude AI about your repository.
You can ask questions about code, PRs, best practices, and more.`,
	Run: func(cmd *cobra.Command, args []string) {
		// runChat()
	},
}

func init() {
	rootCmd.AddCommand(chatCmd)

	chatCmd.Flags().Bool("with-context", true, "Include repository context in the conversation")
	chatCmd.Flags().Int("pr-context", 0, "Specific PR to use as context (defaults to current PR if set)")
}

// func runChat() {
// 	cfg := loadConfig()

// 	fmt.Println("Starting PRism chat session. Type 'exit' or 'quit' to end the session.")
// 	fmt.Printf("Connected to repository: %s/%s\n", cfg.RepoOwner, cfg.RepoName)

// 	// Initialize context
// 	ctx := context.Background()
// 	githubClient := github.InitializeGithubClient(cfg, ctx)

// 	// Get PR details if available for context
// 	var prContext string
// 	if cfg.PRNumber > 0 {
// 		prDetails, err := github.FetchPRDetails(githubClient, cfg)
// 		if err != nil {
// 			fmt.Printf("Warning: Could not fetch PR details: %v\n", err)
// 		} else {
// 			prContext = fmt.Sprintf("PR #%d: %s\n%s\n",
// 				cfg.PRNumber,
// 				prDetails.Title,
// 				prDetails.Diff)
// 		}
// 	}

// 	// Start chat loop
// 	reader := bufio.NewReader(os.Stdin)
// 	conversation := []string{}

// 	if prContext != "" {
// 		conversation = append(conversation,
// 			"System: The following is context about a GitHub PR being discussed:",
// 			prContext)
// 	}

// 	for {
// 		fmt.Print("\nYou: ")
// 		input, _ := reader.ReadString('\n')
// 		input = strings.TrimSpace(input)

// 		if input == "exit" || input == "quit" {
// 			fmt.Println("Exiting chat session. Goodbye!")
// 			break
// 		}

// 		// Add user message to conversation
// 		conversation = append(conversation, "User: "+input)

// 		// Build complete prompt with conversation history
// 		prompt := strings.Join(conversation, "\n")

// 		// Call Claude API with conversation
// 		response, err := llm.SimpleClaudeChat(prompt, cfg)
// 		if err != nil {
// 			fmt.Printf("Error: %v\n", err)
// 			continue
// 		}

// 		fmt.Println("\nPRism: " + response)

// 		// Add assistant response to conversation
// 		conversation = append(conversation, "Assistant: "+response)

// 		// Limit conversation history if it gets too long
// 		if len(conversation) > 20 {
// 			conversation = conversation[len(conversation)-20:]
// 		}
// 	}
// }
