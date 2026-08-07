package api

import (
	"bytes"
	"encoding/json"
	"net/http"
)

// toolCallRequest is the shared JSON shape for A2A and MCP tool invocations.
type toolCallRequest struct {
	Tool      string          `json:"tool"`
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

// decodeToolCall decodes a tool invocation request body and ensures the
// arguments field is at least an empty object.
func decodeToolCall(r *http.Request) (toolCallRequest, error) {
	var req toolCallRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return req, err
	}
	if len(req.Arguments) == 0 || bytes.Equal(req.Arguments, []byte("null")) {
		req.Arguments = json.RawMessage("{}")
	}
	return req, nil
}
