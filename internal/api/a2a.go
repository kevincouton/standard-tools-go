package api

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/kevincouton/standard-tools-go/internal/agent"
)

func a2aAgentCard(w http.ResponseWriter, r *http.Request) {
	scheme := r.URL.Scheme
	if scheme == "" {
		scheme = "http"
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"name":        appName,
		"description": appDescription,
		"version":     appVersion,
		"url":         scheme + "://" + r.Host + "/a2a",
		"capabilities": map[string]bool{
			"streaming":         false,
			"pushNotifications": false,
		},
		"skills": []any{},
	})
}

func a2aDispatchTask(state *AppState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := decodeToolCall(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		if req.Tool == "" {
			writeError(w, http.StatusBadRequest, errors.New("tool is required"))
			return
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
