package mcp

import (
	"PRism/config"
	"PRism/github"
	"PRism/llm"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

func RunMCPServer() {
	decoder := json.NewDecoder(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	// Handle initial manifest request
	var req MCPRequest
	if err := decoder.Decode(&req); err != nil {
		log.Fatalf("Failed to decode request: %v", err)
	}

	if req.Method == "mcp.manifest" {
		manifest := MCPManifest{
			Schema:  "mcp-0.7.1",
			Name:    "prism",
			Version: "1.0.0",
			Tools: []MCPTool{
				{
					Name:        "search_repo",
					Description: "Search repository for relevant code files and information",
					Parameters: map[string]MCPParam{
						"query": {
							Type:        "string",
							Description: "The search query to find relevant code",
						},
					},
				},
				{
					Name:        "get_pr_details",
					Description: "Get details about a specific PR or the current PR",
					Parameters: map[string]MCPParam{
						"pr_number": {
							Type:        "integer",
							Description: "PR number to fetch details for (0 for current PR)",
						},
					},
				},
			},
		}

		resp := MCPResponse{
			ID:     req.ID,
			Result: map[string]any{"manifest": manifest},
		}

		if err := encoder.Encode(resp); err != nil {
			log.Fatalf("Failed to encode response: %v", err)
		}
	}

	// Process tool calls
	for decoder.More() {
		if err := decoder.Decode(&req); err != nil {
			log.Fatalf("Failed to decode request: %v", err)
			continue
		}

		var resp MCPResponse
		resp.ID = req.ID

		switch req.Method {
		case "tool.search_repo":
			query, ok := req.Params["query"].(string)
			if !ok {
				resp.Error = &MCPError{
					Code:    400,
					Message: "Invalid query parameter",
				}
			} else {
				result, err := handleSearchRepo(query)
				if err != nil {
					resp.Error = &MCPError{
						Code:    500,
						Message: err.Error(),
					}
				} else {
					resp.Result = result
				}
			}

		case "tool.get_pr_details":
			prNum, ok := req.Params["pr_number"].(float64)
			if !ok {
				resp.Error = &MCPError{
					Code:    400,
					Message: "Invalid PR number parameter",
				}
			} else {
				result, err := handleGetPRDetails(int(prNum))
				if err != nil {
					resp.Error = &MCPError{
						Code:    500,
						Message: err.Error(),
					}
				} else {
					resp.Result = result
				}
			}

		default:
			resp.Error = &MCPError{
				Code:    404,
				Message: fmt.Sprintf("Unknown method: %s", req.Method),
			}
		}

		if err := encoder.Encode(resp); err != nil {
			log.Fatalf("Failed to encode response: %v", err)
		}
	}
}

func handleSearchRepo(query string) (map[string]any, error) {
	cfg := config.LoadConfig()

	// Generate repo embeddings for context
	repoURL := fmt.Sprintf("https://github.com/%s/%s", cfg.RepoOwner, cfg.RepoName)
	embeddings, err := llm.GenerateCodeEmbeddingsFromGitHub(cfg, repoURL)
	if err != nil {
		return nil, fmt.Errorf("error generating code embeddings: %v", err)
	}

	relevantFiles, err := llm.FindRelevantFiles(query, embeddings, cfg, 3)
	if err != nil {
		return nil, fmt.Errorf("error finding relevant files: %v", err)
	}

	fileContents := make(map[string]string)
	for _, filePath := range relevantFiles {
		for _, emb := range embeddings {
			if emb.FilePath == filePath {
				fileContents[filePath] = emb.Content
				break
			}
		}
	}

	return map[string]any{
		"relevant_files": relevantFiles,
		"file_contents":  fileContents,
	}, nil
}

func handleGetPRDetails(prNumber int) (map[string]any, error) {
	cfg := config.LoadConfig()
	cfg.PRNumber = prNumber

	ctx := context.Background()
	githubClient := github.InitializeGithubClient(cfg, ctx)

	prDetails, err := github.FetchPRDetails(githubClient, cfg)
	if err != nil {
		return nil, fmt.Errorf("error fetching PR details: %v", err)
	}

	return map[string]any{
		"pr_number": prDetails["number"],
		"title":     prDetails["title"],
		"body":      prDetails["body"],
		"diff":      prDetails["diff"],
		"author":    prDetails["author"],
		"branch":    prDetails["branch"],
	}, nil
}
