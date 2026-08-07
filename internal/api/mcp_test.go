package api

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMcpCapabilities(t *testing.T) {
	rec := sendRequest(NewRouter(newTestState()), http.MethodGet, "/mcp/capabilities", "")
	assert.Equal(t, http.StatusOK, rec.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, mcpProtocolVersion, body["protocolVersion"])
	assert.NotNil(t, body["capabilities"])

	serverInfo, ok := body["serverInfo"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, appName, serverInfo["name"])
	assert.Equal(t, appVersion, serverInfo["version"])
}

func TestMcpListTools(t *testing.T) {
	rec := sendRequest(NewRouter(newTestState()), http.MethodPost, "/mcp/tools/list", "")
	assert.Equal(t, http.StatusOK, rec.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	tools, ok := body["tools"].([]any)
	require.True(t, ok)
	assert.NotEmpty(t, tools)
}

func TestMcpCallToolHealth(t *testing.T) {
	rec := sendRequest(NewRouter(newTestState()), http.MethodPost, "/mcp/tools/call", `{"name":"health","arguments":{}}`)
	assert.Equal(t, http.StatusOK, rec.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	content, ok := body["content"].([]any)
	require.True(t, ok)
	require.NotEmpty(t, content)

	first, ok := content[0].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "text", first["type"])
	assert.NotEmpty(t, first["text"])
}

func TestMcpCallToolUnknownTool(t *testing.T) {
	rec := sendRequest(NewRouter(newTestState()), http.MethodPost, "/mcp/tools/call", `{"name":"nope","arguments":{}}`)
	assert.Equal(t, http.StatusOK, rec.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	content, ok := body["content"].([]any)
	require.True(t, ok)
	require.NotEmpty(t, content)

	first, ok := content[0].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "text", first["type"])
	assert.Contains(t, first["text"], "error")
}

func TestMcpCallToolMalformedJSON(t *testing.T) {
	rec := sendRequest(NewRouter(newTestState()), http.MethodPost, "/mcp/tools/call", `{`)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assertErrorResponse(t, rec.Body.Bytes(), "BAD_REQUEST")
}

func TestMcpCallToolMissingName(t *testing.T) {
	rec := sendRequest(NewRouter(newTestState()), http.MethodPost, "/mcp/tools/call", `{"arguments":{}}`)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assertErrorResponse(t, rec.Body.Bytes(), "BAD_REQUEST")
}
