import os
import torch
import numpy as np
import tempfile
import subprocess
from transformers import CodeGenForCausalLM, CodeGenTokenizer
from typing import List, Dict, Tuple
import argparse
import json
import logging

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s',
    datefmt='%Y-%m-%d %H:%M:%S'
)
logger = logging.getLogger(__name__)

class CodeGenEmbedder:
    def __init__(self, model_name: str = "Salesforce/codegen-350M-mono"):
        logger.info(f"Loading CodeGen model: {model_name}")
        self.tokenizer = CodeGenTokenizer.from_pretrained(model_name)
        
        # Fix: Set pad token to eos token
        self.tokenizer.pad_token = self.tokenizer.eos_token
        
        self.model = CodeGenForCausalLM.from_pretrained(model_name, output_hidden_states=True)
        self.model.eval()  # Set to evaluation mode
        
        # Check if GPU is available
        self.device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
        self.model.to(self.device)
        logger.info(f"Using device: {self.device}")
        

    def generate_embedding(self, code_snippet: str) -> np.ndarray:
        """Generate embedding for a code snippet using the hidden states of the CodeGen model."""
        inputs = self.tokenizer(code_snippet, return_tensors="pt", truncation=True, 
                               max_length=512, padding="max_length")
        
        # Move inputs to the same device as the model
        inputs = {k: v.to(self.device) for k, v in inputs.items()}
        
        with torch.no_grad():
            outputs = self.model(**inputs, output_hidden_states=True)
        
        # Use the mean of the last hidden layer as the embedding
        embedding = outputs.hidden_states[-1].mean(dim=1).squeeze().cpu().numpy()
        return embedding

def clone_github_repo(repo_url: str) -> str:
    """Clone a GitHub repository and return the path to the cloned repo."""
    temp_dir = tempfile.mkdtemp()
    logger.info(f"Cloning repository {repo_url} to {temp_dir}")
    
    try:
        subprocess.run(
            ["git", "clone", repo_url, temp_dir],
            check=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE
        )
        return temp_dir
    except subprocess.CalledProcessError as e:
        logger.error(f"Error cloning repository: {e}")
        logger.error(f"Stderr: {e.stderr.decode()}")
        raise

def is_code_file(file_path: str) -> bool:
    """Check if a file is a code file based on its extension."""
    code_extensions = {
        ".py", ".js", ".java", ".c", ".cpp", ".h", ".cs", ".go", 
        ".rb", ".php", ".ts", ".jsx", ".tsx", ".swift", ".kt", ".rs"
    }
    return os.path.splitext(file_path)[1].lower() in code_extensions

def should_skip_dir(dir_name: str) -> bool:
    """Check if a directory should be skipped."""
    skip_dirs = {".git", "node_modules", "dist", "build", "target", "venv", "__pycache__"}
    return dir_name in skip_dirs

def load_codebase(directory: str) -> List[Dict[str, str]]:
    """Load code files from a directory and return a list of file paths and contents."""
    code_files = []
    
    for root, dirs, files in os.walk(directory):
        # Skip directories that should be ignored
        dirs[:] = [d for d in dirs if not should_skip_dir(d)]
        
        for file in files:
            file_path = os.path.join(root, file)
            rel_path = os.path.relpath(file_path, directory)
            
            if is_code_file(file_path):
                try:
                    with open(file_path, "r", encoding="utf-8", errors="ignore") as f:
                        content = f.read()
                    
                    code_files.append({
                        "file_path": rel_path,
                        "content": content
                    })
                except Exception as e:
                    logger.error(f"Error reading file {file_path}: {e}")
    
    return code_files

def process_repo(repo_url: str, output_file: str, model_name: str) -> None:
    """Process a GitHub repository and generate embeddings for its code files."""
    # Clone the repository
    repo_dir = clone_github_repo(repo_url)
    
    try:
        # Load the codebase
        logger.info("Loading code files...")
        code_files = load_codebase(repo_dir)
        logger.info(f"Found {len(code_files)} code files")
        
        # Initialize the embedder
        embedder = CodeGenEmbedder(model_name)
        
        # Generate embeddings
        logger.info("Generating embeddings...")
        result = []
        
        for i, file_info in enumerate(code_files):
            logger.info(f"Processing file {i+1}/{len(code_files)}: {file_info['file_path']}")
            embedding = embedder.generate_embedding(file_info["content"])
            
            result.append({
                "file_path": file_info["file_path"],
                "content": file_info["content"],
                "embedding": embedding.tolist()
            })
        
        # Save the embeddings
        logger.info(f"Saving embeddings to {output_file}")
        with open(output_file, "w") as f:
            json.dump(result, f)
            
        logger.info("Done!")
        
    finally:
        # Clean up the cloned repository
        logger.info(f"Cleaning up temporary directory: {repo_dir}")
        # Uncomment to remove the directory:
        # import shutil
        # shutil.rmtree(repo_dir)

def find_relevant_files(query: str, embeddings_file: str, top_n: int = 5, model_name: str = "Salesforce/codegen-350M-mono") -> List[Tuple[str, float]]:
    """Find the most relevant files for a query using cosine similarity."""
    # Load embeddings
    with open(embeddings_file, "r") as f:
        embeddings_data = json.load(f)
    
    # Initialize the embedder
    embedder = CodeGenEmbedder(model_name)
    
    # Generate embedding for the query
    query_embedding = embedder.generate_embedding(query)
    
    # Calculate cosine similarity
    scores = []
    for file_data in embeddings_data:
        file_embedding = np.array(file_data["embedding"])
        similarity = cosine_similarity(query_embedding, file_embedding)
        scores.append((file_data["file_path"], similarity))
    
    # Sort by similarity (highest first)
    scores.sort(key=lambda x: x[1], reverse=True)
    
    # Return top N results
    return scores[:top_n]

def cosine_similarity(a: np.ndarray, b: np.ndarray) -> float:
    """Calculate the cosine similarity between two vectors."""
    return np.dot(a, b) / (np.linalg.norm(a) * np.linalg.norm(b))

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Generate code embeddings for a GitHub repository")
    parser.add_argument("repo_url", help="URL of the GitHub repository")
    parser.add_argument("--output", "-o", default="embeddings.json", help="Output file for embeddings")
    parser.add_argument("--model", "-m", default="Salesforce/codegen-350M-mono", 
                        help="CodeGen model to use")
    args = parser.parse_args()
    
    process_repo(args.repo_url, args.output, args.model)