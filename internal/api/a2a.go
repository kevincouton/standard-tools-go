package api

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/kevincouton/standard-tools-go/internal/agent"
)

func requestScheme(r *http.Request) string {
	if s := r.Header.Get("X-Forwarded-Proto"); s != "" {
		return s
	}
	if r.URL.Scheme != "" {
		return r.URL.Scheme
	}
	return "http"
}

func a2aAgentCard(w http.ResponseWriter, r *http.Request) {
	scheme := requestScheme(r)

	writeJSON(w, http.StatusOK, map[string]any{
		fieldName:        appName,
		fieldDescription: appDescription,
		fieldVersion:     appVersion,
		fieldURL:         scheme + "://" + r.Host + "/a2a",
		fieldCapabilities: map[string]bool{
			fieldStreaming:         false,
			fieldPushNotifications: false,
		},
		fieldSkills: []any{},
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
				fieldID:     id,
				fieldStatus: a2aStatusFailed,
				fieldResult: map[string]any{
					fieldOutput: nil,
					fieldError:  err.Error(),
				},
			})
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{
			fieldID:     id,
			fieldStatus: a2aStatusCompleted,
			fieldResult: map[string]any{
				fieldOutput: result.Output,
				fieldError:  nil,
			},
		})
	}
}
