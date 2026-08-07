package api

import (
	"errors"
	"net/http"

	"github.com/kevincouton/standard-tools-go/internal/agent"
)

type mcpContentItem struct {
	Type string `json:"type"`
	Text any    `json:"text"`
}

func mcpTextError(err error) []mcpContentItem {
	return []mcpContentItem{
		{Type: mcpContentTypeText, Text: mcpErrorPrefix + err.Error()},
	}
}

func mcpCapabilities(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		fieldProtocolVersion: mcpProtocolVersion,
		fieldCapabilities: map[string]any{
			fieldTools: map[string]any{},
		},
		fieldServerInfo: map[string]string{
			fieldName:    appName,
			fieldVersion: appVersion,
		},
	})
}

func mcpListTools(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		fieldTools: agent.ListTools(),
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
				fieldContent: mcpTextError(err),
			})
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{
			fieldContent: []mcpContentItem{
				{Type: mcpContentTypeText, Text: result.Output},
			},
		})
	}
}
