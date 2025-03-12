#!/usr/bin/env python3
import argparse
import json
import numpy as np
from transformers import CodeGenForCausalLM, CodeGenTokenizer
import torch

def load_embeddings(file_path):
    with open(file_path, 'r') as f:
        return json.load(f)

def generate_query_embedding(query_text, model_name="Salesforce/codegen-350M-mono"):
    # Load query from file if needed
    if query_text.endswith('.txt'):
        with open(query_text, 'r') as f:
            query_text = f.read()
            
    # Load model and tokenizer
    tokenizer = CodeGenTokenizer.from_pretrained(model_name)
    model = CodeGenForCausalLM.from_pretrained(model_name, output_hidden_states=True)
    model.eval()
    
    # Check if GPU is available
    device = torch.device("cuda" if torch.cuda.is_available() else "cpu")
    model.to(device)
    
    # Generate embedding
    inputs = tokenizer(query_text, return_tensors="pt", truncation=True, 
                      max_length=512, padding="max_length")
    inputs = {k: v.to(device) for k, v in inputs.items()}
    
    with torch.no_grad():
        outputs = model(**inputs, output_hidden_states=True)
    
    # Use the mean of the last hidden layer as the embedding
    embedding = outputs.hidden_states[-1].mean(dim=1).squeeze().cpu().numpy()
    return embedding

def cosine_similarity(a, b):
    return np.dot(a, b) / (np.linalg.norm(a) * np.linalg.norm(b))

def find_relevant_files(query_embedding, embeddings, limit):
    scores = []
    for item in embeddings:
        file_embedding = np.array(item["embedding"])
        score = cosine_similarity(query_embedding, file_embedding)
        scores.append((item["file_path"], score))
    
    # Sort by score (descending)
    scores.sort(key=lambda x: x[1], reverse=True)
    
    # Return top n results
    return scores[:limit]

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Find relevant files for a query")
    parser.add_argument("--query", required=True, help="Query text or file path")
    parser.add_argument("--embeddings", required=True, help="Path to embeddings JSON file")
    parser.add_argument("--limit", type=int, default=5, help="Number of results to return")
    parser.add_argument("--model", default="Salesforce/codegen-350M-mono", help="CodeGen model to use")
    args = parser.parse_args()
    
    # Load embeddings
    embeddings = load_embeddings(args.embeddings)
    
    # Generate query embedding
    query_embedding = generate_query_embedding(args.query, args.model)
    
    # Find relevant files
    results = find_relevant_files(query_embedding, embeddings, args.limit)
    
    # Print results (one file path per line)
    for file_path, score in results:
        print(file_path)