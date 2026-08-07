package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestA2AAgentCard(t *testing.T) {
	rec := sendRequest(NewRouter(newTestState()), http.MethodGet, "/a2a/agent.json", "")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, contentTypeJSON, rec.Header().Get("Content-Type"))

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, appName, body[fieldName])
	assert.Equal(t, appDescription, body[fieldDescription])
	assert.Equal(t, appVersion, body[fieldVersion])

	url, ok := body[fieldURL].(string)
	require.True(t, ok)
	assert.True(t, strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://"))
	assert.Contains(t, url, "/a2a")
}

func TestA2AAgentCardForwardedProto(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/a2a/agent.json", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	rec := httptest.NewRecorder()
	NewRouter(newTestState()).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	url, ok := body[fieldURL].(string)
	require.True(t, ok)
	assert.True(t, strings.HasPrefix(url, "https://"))
}

func TestA2ADispatchHealth(t *testing.T) {
	rec := sendRequest(NewRouter(newTestState()), http.MethodPost, "/a2a/tasks", `{"tool":"health","arguments":{}}`)
	assert.Equal(t, http.StatusOK, rec.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.NotEmpty(t, body[fieldID])
	assert.Equal(t, a2aStatusCompleted, body[fieldStatus])

	result, ok := body[fieldResult].(map[string]any)
	require.True(t, ok)
	assert.NotNil(t, result[fieldOutput])
	assert.Nil(t, result[fieldError])
}

func TestA2ADispatchUnknownTool(t *testing.T) {
	rec := sendRequest(NewRouter(newTestState()), http.MethodPost, "/a2a/tasks", `{"tool":"nope","arguments":{}}`)
	assert.Equal(t, http.StatusOK, rec.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, a2aStatusFailed, body[fieldStatus])

	result, ok := body[fieldResult].(map[string]any)
	require.True(t, ok)
	assert.Nil(t, result[fieldOutput])
	assert.NotNil(t, result[fieldError])
}

func TestA2ADispatchMalformedJSON(t *testing.T) {
	rec := sendRequest(NewRouter(newTestState()), http.MethodPost, "/a2a/tasks", `{`)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, contentTypeJSON, rec.Header().Get("Content-Type"))
	assertErrorResponse(t, rec, "BAD_REQUEST")
}

func TestA2ADispatchMissingTool(t *testing.T) {
	rec := sendRequest(NewRouter(newTestState()), http.MethodPost, "/a2a/tasks", `{"arguments":{}}`)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assertErrorResponse(t, rec, "BAD_REQUEST")
}
