package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/kevincouton/standard-tools-go/internal/agent"
)

func mcpCapabilities(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"protocolVersion": mcpProtocolVersion,
		"capabilities": map[string]any{
			"tools": map[string]any{},
		},
		"serverInfo": map[string]string{
			"name":    appName,
			"version": appVersion,
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
		req, err := decodeToolCall(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		if req.Name == "" {
			writeError(w, http.StatusBadRequest, errors.New("name is required"))
			return
		}

		result, err := state.Dispatcher.Dispatch(r.Context(), agent.ToolCall{Name: req.Name, Arguments: req.Arguments})
		if err != nil {
			writeJSON(w, http.StatusOK, map[string]any{
				"content": []map[string]string{
					{"type": "text", "text": "error: " + err.Error()},
				},
			})
			return
		}

		encoded, err := json.Marshal(result.Output)
		if err != nil {
			writeJSON(w, http.StatusOK, map[string]any{
				"content": []map[string]string{
					{"type": "text", "text": "error: " + err.Error()},
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
