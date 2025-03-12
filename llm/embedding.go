package llm

import (
	"PRism/config"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CodeEmbedding represents an embedding for a code file
type CodeEmbedding struct {
	FilePath  string    `json:"file_path"`
	Content   string    `json:"content"`
	Embedding []float32 `json:"embedding"`
}

// GenerateCodeEmbeddingsFromGitHub clones a GitHub repo and creates embeddings using CodeGen
func GenerateCodeEmbeddingsFromGitHub(cfg config.Config, repoURL string) ([]CodeEmbedding, error) {
	// Create temporary directory for output
	tempFile, err := os.CreateTemp("", "embeddings-*.json")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %v", err)
	}
	tempFile.Close()
	outputFile := tempFile.Name()

	// Clean up temp file when done
	defer os.Remove(outputFile)

	// Get the path to the Python script
	scriptPath := filepath.Join("./codegen", "codegen_embeddings.py")

	// Prepare the command
	cmd := exec.Command("python3", scriptPath, repoURL, "--output", outputFile)

	// Capture stdout and stderr
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run the Python script
	fmt.Printf("Running Python script to generate embeddings for %s\n", repoURL)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to run Python script: %v\nStderr: %s", err, stderr.String())
	}

	// Print the output
	fmt.Println(stdout.String())

	// Load the embeddings from the JSON file
	embeddings, err := loadEmbeddingsFromJSON(outputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load embeddings: %v", err)
	}

	return embeddings, nil
}

// loadEmbeddingsFromJSON loads embeddings from a JSON file
func loadEmbeddingsFromJSON(filePath string) ([]CodeEmbedding, error) {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open embeddings file: %v", err)
	}
	defer file.Close()

	// Read the file
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read embeddings file: %v", err)
	}

	// Parse the JSON
	var embeddings []CodeEmbedding
	if err := json.Unmarshal(data, &embeddings); err != nil {
		return nil, fmt.Errorf("failed to parse embeddings JSON: %v", err)
	}

	return embeddings, nil
}

// FindRelevantFiles finds the most relevant files for a query using cosine similarity
func FindRelevantFiles(query string, embeddings []CodeEmbedding, cfg config.Config, limit int) ([]string, error) {
	// Create temporary file for query
	tempFile, err := os.CreateTemp("", "query-*.txt")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Write query to file
	if _, err := tempFile.WriteString(query); err != nil {
		return nil, fmt.Errorf("failed to write query to temp file: %v", err)
	}
	tempFile.Close()

	// Create temporary file for embeddings
	embFile, err := os.CreateTemp("", "embeddings-*.json")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %v", err)
	}
	defer os.Remove(embFile.Name())

	// Write embeddings to file
	embData, err := json.Marshal(embeddings)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal embeddings: %v", err)
	}
	if err := os.WriteFile(embFile.Name(), embData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write embeddings to temp file: %v", err)
	}

	// Get the path to the Python script
	scriptPath := filepath.Join("./codegen", "codegen_query.py")

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
	if err := cmd.Run(); err != nil {
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

	return result, nil
}

// cosineSimilarity calculates similarity between two vectors
func cosineSimilarity(a, b []float32) float32 {
	var dotProduct, magnitudeA, magnitudeB float32

	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		magnitudeA += a[i] * a[i]
		magnitudeB += b[i] * b[i]
	}

	magnitudeA = float32(math.Sqrt(float64(magnitudeA)))
	magnitudeB = float32(math.Sqrt(float64(magnitudeB)))

	if magnitudeA == 0 || magnitudeB == 0 {
		return 0
	}

	return dotProduct / (magnitudeA * magnitudeB)
}
