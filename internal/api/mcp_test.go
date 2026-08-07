package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMcpCapabilities(t *testing.T) {
	rec := sendRequest(NewRouter(newTestState()), http.MethodGet, "/mcp/capabilities", "")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, contentTypeJSON, rec.Header().Get("Content-Type"))

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, mcpProtocolVersion, body[fieldProtocolVersion])
	assert.NotNil(t, body[fieldCapabilities])

	serverInfo, ok := body[fieldServerInfo].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, appName, serverInfo[fieldName])
	assert.Equal(t, appVersion, serverInfo[fieldVersion])
}

func TestMcpListTools(t *testing.T) {
	rec := sendRequest(NewRouter(newTestState()), http.MethodPost, "/mcp/tools/list", "")
	assert.Equal(t, http.StatusOK, rec.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	tools, ok := body[fieldTools].([]any)
	require.True(t, ok)
	assert.NotEmpty(t, tools)
}

func TestMcpCallToolHealth(t *testing.T) {
	rec := sendRequest(NewRouter(newTestState()), http.MethodPost, "/mcp/tools/call", `{"name":"health","arguments":{}}`)
	assert.Equal(t, http.StatusOK, rec.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	content, ok := body[fieldContent].([]any)
	require.True(t, ok)
	require.NotEmpty(t, content)

	first, ok := content[0].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, mcpContentTypeText, first[fieldType])
	text, ok := first[fieldText].(string)
	require.True(t, ok)
	assert.NotEmpty(t, text)
}

func TestMcpCallToolUnknownTool(t *testing.T) {
	rec := sendRequest(NewRouter(newTestState()), http.MethodPost, "/mcp/tools/call", `{"name":"nope","arguments":{}}`)
	assert.Equal(t, http.StatusOK, rec.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	content, ok := body[fieldContent].([]any)
	require.True(t, ok)
	require.NotEmpty(t, content)

	first, ok := content[0].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, mcpContentTypeText, first[fieldType])
	text, ok := first[fieldText].(string)
	require.True(t, ok)
	assert.True(t, strings.HasPrefix(text, mcpErrorPrefix))
}

func TestMcpCallToolMalformedJSON(t *testing.T) {
	rec := sendRequest(NewRouter(newTestState()), http.MethodPost, "/mcp/tools/call", `{`)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, contentTypeJSON, rec.Header().Get("Content-Type"))
	assertErrorResponse(t, rec, "BAD_REQUEST")
}

func TestMcpCallToolMissingName(t *testing.T) {
	rec := sendRequest(NewRouter(newTestState()), http.MethodPost, "/mcp/tools/call", `{"arguments":{}}`)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assertErrorResponse(t, rec, "BAD_REQUEST")
}
