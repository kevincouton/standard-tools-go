package agent

import "encoding/json"

func ListTools() []ToolDefinition {
	return []ToolDefinition{
		{
			Name:        "health",
			Description: "Return agent health status.",
			Parameters:  json.RawMessage(`{"type":"object","properties":{}}`),
		},
		{
			Name:        "list_tools",
			Description: "List all registered tool names.",
			Parameters:  json.RawMessage(`{"type":"object","properties":{}}`),
		},
		{
			Name:        "fetch_ohlcv",
			Description: "Fetch OHLCV bars for a single ticker.",
			Parameters:  json.RawMessage(`{"type":"object","properties":{"ticker":{"type":"string"},"start":{"type":"string","format":"date"},"end":{"type":"string","format":"date"},"interval":{"type":"string","enum":["daily","weekly","monthly"]},"provider":{"type":"string"}},"required":["ticker","start","end"]}`),
		},
	}
}

func FindTool(name string) (*ToolDefinition, bool) {
	for _, t := range ListTools() {
		if t.Name == name {
			return &t, true
		}
	}
	return nil, false
}
