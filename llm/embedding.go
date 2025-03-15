package llm

import (
	"tracepr/config"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GenerateCodeEmbeddingsFromGitHub clones a GitHub repo and creates embeddings using CodeGen
func GenerateCodeEmbeddingsFromGitHub(cfg config.Config, repoURL string) ([]config.CodeEmbedding, error) {
	log.Printf("Generating code embeddings for GitHub repo: %s", repoURL)

	// Create temporary directory for output
	tempFile, err := os.CreateTemp("", "embeddings-*.json")
	if err != nil {
		log.Printf("Error creating temp file: %v", err)
		return nil, fmt.Errorf("failed to create temp file: %v", err)
	}
	tempFile.Close()
	outputFile := tempFile.Name()
	log.Printf("Created temporary file: %s", outputFile)

	// Clean up temp file when done
	defer os.Remove(outputFile)

	// Get the path to the Python script
	scriptPath := filepath.Join("./codegen", "codegen_embeddings.py")
	log.Printf("Using Python script at: %s", scriptPath)

	// Prepare the command
	cmd := exec.Command("python3", scriptPath, repoURL, "--output", outputFile)

	// Capture stdout and stderr
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run the Python script
	log.Printf("Running Python script to generate embeddings for %s", repoURL)
	if err := cmd.Run(); err != nil {
		log.Printf("Error running Python script: %v\nStderr: %s", err, stderr.String())
		return nil, fmt.Errorf("failed to run Python script: %v\nStderr: %s", err, stderr.String())
	}

	// Print the output
	log.Printf("Python script output: %s", stdout.String())

	// Load the embeddings from the JSON file
	log.Printf("Loading embeddings from JSON file: %s", outputFile)
	embeddings, err := loadEmbeddingsFromJSON(outputFile)
	if err != nil {
		log.Printf("Error loading embeddings: %v", err)
		return nil, fmt.Errorf("failed to load embeddings: %v", err)
	}

	log.Printf("Successfully generated %d embeddings", len(embeddings))
	return embeddings, nil
}

// loadEmbeddingsFromJSON loads embeddings from a JSON file
func loadEmbeddingsFromJSON(filePath string) ([]config.CodeEmbedding, error) {
	log.Printf("Loading embeddings from file: %s", filePath)

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("Error opening embeddings file: %v", err)
		return nil, fmt.Errorf("failed to open embeddings file: %v", err)
	}
	defer file.Close()

	// Read the file
	data, err := io.ReadAll(file)
	if err != nil {
		log.Printf("Error reading embeddings file: %v", err)
		return nil, fmt.Errorf("failed to read embeddings file: %v", err)
	}

	// Parse the JSON
	var embeddings []config.CodeEmbedding
	if err := json.Unmarshal(data, &embeddings); err != nil {
		log.Printf("Error parsing embeddings JSON: %v", err)
		return nil, fmt.Errorf("failed to parse embeddings JSON: %v", err)
	}

	log.Printf("Successfully loaded %d embeddings from file", len(embeddings))
	return embeddings, nil
}

// FindRelevantFiles finds the most relevant files for a query using cosine similarity
func FindRelevantFiles(query string, embeddings []config.CodeEmbedding, cfg config.Config, limit int) ([]string, error) {
	log.Printf("Finding relevant files for query with limit %d", limit)

	// Create temporary file for query
	tempFile, err := os.CreateTemp("", "query-*.txt")
	if err != nil {
		log.Printf("Error creating query temp file: %v", err)
		return nil, fmt.Errorf("failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Write query to file
	if _, err := tempFile.WriteString(query); err != nil {
		log.Printf("Error writing query to temp file: %v", err)
		return nil, fmt.Errorf("failed to write query to temp file: %v", err)
	}
	tempFile.Close()
	log.Printf("Created query temp file: %s", tempFile.Name())

	// Create temporary file for embeddings
	embFile, err := os.CreateTemp("", "embeddings-*.json")
	if err != nil {
		log.Printf("Error creating embeddings temp file: %v", err)
		return nil, fmt.Errorf("failed to create temp file: %v", err)
	}
	defer os.Remove(embFile.Name())

	// Write embeddings to file
	embData, err := json.Marshal(embeddings)
	if err != nil {
		log.Printf("Error marshaling embeddings: %v", err)
		return nil, fmt.Errorf("failed to marshal embeddings: %v", err)
	}
	if err := os.WriteFile(embFile.Name(), embData, 0644); err != nil {
		log.Printf("Error writing embeddings to temp file: %v", err)
		return nil, fmt.Errorf("failed to write embeddings to temp file: %v", err)
	}
	log.Printf("Created embeddings temp file: %s", embFile.Name())

	// Get the path to the Python script
	scriptPath := filepath.Join("./codegen", "codegen_query.py")
	log.Printf("Using query script at: %s", scriptPath)

	// Prepare the command
	cmd := exec.Command("python3", scriptPath,
		"--query", tempFile.Name(),
		"--embeddings", embFile.Name(),
		"--limit", fmt.Sprintf("%d", limit))

	// Capture stdout and stderr
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run the Python script
	log.Printf("Running query Python script")
	if err := cmd.Run(); err != nil {
		log.Printf("Error running query Python script: %v\nStderr: %s", err, stderr.String())
		return nil, fmt.Errorf("failed to run query Python script: %v\nStderr: %s", err, stderr.String())
	}

	// Parse the result (assuming the script outputs one filename per line)
	var result []string
	for _, line := range strings.Split(stdout.String(), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			result = append(result, line)
		}
	}

	log.Printf("Found %d relevant files", len(result))
	return result, nil
}
