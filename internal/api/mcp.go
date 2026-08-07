package api

import (
	"encoding/json"
	"net/http"

	"github.com/kevincouton/standard-tools-go/internal/agent"
)

func mcpCapabilities(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]any{
			"tools": map[string]any{},
		},
		"serverInfo": map[string]string{
			"name":    "standard-tools-go",
			"version": "0.1.0",
		},
	})
}

func mcpListTools(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"tools": agent.ListTools(),
	})
}

func mcpCallTool(state *AppState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Name      string          `json:"name"`
			Arguments json.RawMessage `json:"arguments"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		if req.Arguments == nil {
			req.Arguments = json.RawMessage("{}")
		}

		result, err := state.Dispatcher.Dispatch(r.Context(), agent.ToolCall{Name: req.Name, Arguments: req.Arguments})
		if err != nil {
			writeJSON(w, http.StatusOK, map[string]any{
				"content": []map[string]string{
					{"type": "text", "text": err.Error()},
				},
			})
			return
		}

		encoded, err := json.Marshal(result.Output)
		if err != nil {
			writeJSON(w, http.StatusOK, map[string]any{
				"content": []map[string]string{
					{"type": "text", "text": err.Error()},
				},
			})
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"content": []map[string]string{
				{"type": "text", "text": string(encoded)},
			},
		})
	}
}
