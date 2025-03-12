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
	log.Println("MCP server starting...")
	log.Println("Waiting for initial manifest request...")

	decoder := json.NewDecoder(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	// Handle initial manifest request
	var req MCPRequest
	if err := decoder.Decode(&req); err != nil {
		log.Printf("Failed to decode initial request: %v", err)
		log.Println("Check if Cursor is sending the expected MCP format")
		log.Fatalf("Exiting due to decode error: %v", err)
	}

	log.Printf("Received request: Method=%s, ID=%s", req.Method, req.ID)

	if req.Method == "mcp.manifest" {
		log.Println("Processing manifest request")
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

		log.Println("Sending manifest response")
		if err := encoder.Encode(resp); err != nil {
			log.Fatalf("Failed to encode response: %v", err)
		}
		log.Println("Manifest response sent successfully")
	} else {
		log.Printf("ERROR: Expected mcp.manifest method but got %s", req.Method)
		resp := MCPResponse{
			ID: req.ID,
			Error: &MCPError{
				Code:    400,
				Message: fmt.Sprintf("Expected mcp.manifest but got %s", req.Method),
			},
		}
		encoder.Encode(resp)
		log.Fatalf("Exiting due to unexpected method")
	}

	log.Println("Entering tool processing loop...")
	// Process tool calls
	for {
		log.Println("Waiting for next request...")
		if err := decoder.Decode(&req); err != nil {
			log.Printf("Failed to decode request or connection closed: %v", err)
			break
		}

		log.Printf("Received tool request: Method=%s, ID=%s", req.Method, req.ID)

		var resp MCPResponse
		resp.ID = req.ID

		switch req.Method {
		case "tool.search_repo":
			query, ok := req.Params["query"].(string)
			if !ok {
				log.Println("ERROR: Invalid query parameter")
				resp.Error = &MCPError{
					Code:    400,
					Message: "Invalid query parameter",
				}
			} else {
				log.Printf("Processing search_repo: query=%s", query)
				result, err := handleSearchRepo(query)
				if err != nil {
					log.Printf("ERROR in search_repo: %v", err)
					resp.Error = &MCPError{
						Code:    500,
						Message: err.Error(),
					}
				} else {
					log.Printf("Search found %d relevant files", len(result["relevant_files"].([]string)))
					resp.Result = result
				}
			}

		case "tool.get_pr_details":
			prNum, ok := req.Params["pr_number"].(float64)
			if !ok {
				log.Println("ERROR: Invalid PR number parameter")
				resp.Error = &MCPError{
					Code:    400,
					Message: "Invalid PR number parameter",
				}
			} else {
				log.Printf("Processing get_pr_details: pr_number=%d", int(prNum))
				result, err := handleGetPRDetails(int(prNum))
				if err != nil {
					log.Printf("ERROR in get_pr_details: %v", err)
					resp.Error = &MCPError{
						Code:    500,
						Message: err.Error(),
					}
				} else {
					log.Printf("PR details retrieved successfully: #%v", result["pr_number"])
					resp.Result = result
				}
			}

		default:
			log.Printf("ERROR: Unknown method: %s", req.Method)
			resp.Error = &MCPError{
				Code:    404,
				Message: fmt.Sprintf("Unknown method: %s", req.Method),
			}
		}

		log.Println("Sending response")
		if err := encoder.Encode(resp); err != nil {
			log.Printf("Failed to encode response: %v", err)
			break
		}
		log.Println("Response sent successfully")
	}

	log.Println("MCP server exiting")
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
