package api

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/kevincouton/standard-tools-go/internal/agent"
)

func a2aAgentCard(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"name":        "standard-tools-go",
		"description": "Quantitative finance toolkit agent",
		"version":     "0.1.0",
		"url":         "http://localhost:8080/a2a",
		"capabilities": map[string]bool{
			"streaming":         false,
			"pushNotifications": false,
		},
		"skills": []any{},
	})
}

func a2aDispatchTask(state *AppState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Tool      string          `json:"tool"`
			Arguments json.RawMessage `json:"arguments"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		if req.Arguments == nil {
			req.Arguments = json.RawMessage("{}")
		}

		id := uuid.NewString()
		result, err := state.Dispatcher.Dispatch(r.Context(), agent.ToolCall{Name: req.Tool, Arguments: req.Arguments})
		if err != nil {
			writeJSON(w, http.StatusOK, map[string]any{
				"id":     id,
				"status": "failed",
				"result": map[string]any{
					"output": nil,
					"error":  err.Error(),
				},
			})
			return
		}

		var output any = result.Output
		writeJSON(w, http.StatusOK, map[string]any{
			"id":     id,
			"status": "completed",
			"result": map[string]any{
				"output": output,
				"error":  nil,
			},
		})
	}
}
