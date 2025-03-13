package config

type MCPRequest struct {
	ID     string         `json:"id"`
	Method string         `json:"method"`
	Params map[string]any `json:"params"`
}

type MCPResponse struct {
	ID     string         `json:"id"`
	Result map[string]any `json:"result,omitempty"`
	Error  *MCPError      `json:"error,omitempty"`
}

type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type MCPManifest struct {
	Schema  string    `json:"schema"`
	Name    string    `json:"name"`
	Version string    `json:"version"`
	Tools   []MCPTool `json:"tools"`
}

type MCPTool struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Parameters  map[string]MCPParam `json:"parameters"`
}

type MCPParam struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}
