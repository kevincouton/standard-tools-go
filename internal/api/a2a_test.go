package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestA2AAgentCard(t *testing.T) {
	rec := sendRequest(NewRouter(newTestState()), http.MethodGet, "/a2a/agent.json", "")
	assert.Equal(t, http.StatusOK, rec.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, appName, body["name"])
	assert.Equal(t, appDescription, body["description"])
	assert.Equal(t, appVersion, body["version"])

	url, ok := body["url"].(string)
	require.True(t, ok)
	assert.True(t, strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://"))
	assert.Contains(t, url, "/a2a")
}

func TestA2ADispatchHealth(t *testing.T) {
	rec := sendRequest(NewRouter(newTestState()), http.MethodPost, "/a2a/tasks", `{"tool":"health","arguments":{}}`)
	assert.Equal(t, http.StatusOK, rec.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.NotEmpty(t, body["id"])
	assert.Equal(t, "completed", body["status"])

	result, ok := body["result"].(map[string]any)
	require.True(t, ok)
	assert.NotNil(t, result["output"])
	assert.Nil(t, result["error"])
}

func TestA2ADispatchUnknownTool(t *testing.T) {
	rec := sendRequest(NewRouter(newTestState()), http.MethodPost, "/a2a/tasks", `{"tool":"nope","arguments":{}}`)
	assert.Equal(t, http.StatusOK, rec.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "failed", body["status"])

	result, ok := body["result"].(map[string]any)
	require.True(t, ok)
	assert.Nil(t, result["output"])
	assert.NotNil(t, result["error"])
}

func TestA2ADispatchMalformedJSON(t *testing.T) {
	rec := sendRequest(NewRouter(newTestState()), http.MethodPost, "/a2a/tasks", `{`)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assertErrorResponse(t, rec.Body.Bytes(), "BAD_REQUEST")
}

func TestA2ADispatchMissingTool(t *testing.T) {
	rec := sendRequest(NewRouter(newTestState()), http.MethodPost, "/a2a/tasks", `{"arguments":{}}`)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assertErrorResponse(t, rec.Body.Bytes(), "BAD_REQUEST")
}
