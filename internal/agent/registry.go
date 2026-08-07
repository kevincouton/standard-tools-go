package agent

import "encoding/json"

const (
	ToolHealth     = "health"
	ToolListTools  = "list_tools"
	ToolFetchOhlcv = "fetch_ohlcv"
)

func ListTools() []ToolDefinition {
	return []ToolDefinition{
		{
			Name:        ToolHealth,
			Description: "Return agent health status.",
			Parameters:  json.RawMessage(`{"type":"object","properties":{}}`),
		},
		{
			Name:        ToolListTools,
			Description: "List all registered tool names.",
			Parameters:  json.RawMessage(`{"type":"object","properties":{}}`),
		},
		{
			Name:        ToolFetchOhlcv,
			Description: "Fetch OHLCV bars for a single ticker.",
			Parameters:  json.RawMessage(`{"type":"object","properties":{"ticker":{"type":"string"},"start":{"type":"string","format":"date"},"end":{"type":"string","format":"date"},"interval":{"type":"string","enum":["daily","weekly","monthly"]},"provider":{"type":"string"}},"required":["ticker","start","end"]}`),
		},
	}
}

func FindTool(name string) (*ToolDefinition, bool) {
	for _, t := range ListTools() {
		if t.Name == name {
			found := t
			return &found, true
		}
	}
	return nil, false
}
