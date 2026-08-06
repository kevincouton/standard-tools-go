package agent

import "encoding/json"

type ToolDefinition struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

type ToolCall struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type ToolResult struct {
	Output json.RawMessage `json:"output"`
}

func OkResult(output json.RawMessage) ToolResult {
	return ToolResult{Output: output}
}
