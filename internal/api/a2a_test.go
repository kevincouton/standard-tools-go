package api

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestA2AAgentCard(t *testing.T) {
	rec := sendRequest(NewRouter(newTestState()), http.MethodGet, "/a2a/agent.json", "")
	assert.Equal(t, http.StatusOK, rec.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "standard-tools-go", body["name"])
	assert.Equal(t, "Quantitative finance toolkit agent", body["description"])
	assert.Equal(t, "0.1.0", body["version"])
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
	require.Contains(t, []int{http.StatusOK, http.StatusBadRequest}, rec.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "failed", body["status"])

	result, ok := body["result"].(map[string]any)
	require.True(t, ok)
	assert.Nil(t, result["output"])
	assert.NotNil(t, result["error"])
}
