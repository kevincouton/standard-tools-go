package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/kevincouton/standard-tools-go/internal/agent"
	"github.com/kevincouton/standard-tools-go/internal/marketdata"
	"github.com/stretchr/testify/assert"
)

func newTestState() *AppState {
	svc := marketdata.NewService("synthetic", marketdata.NewInMemoryCache())
	svc.Register(&marketdata.SyntheticProvider{})
	return &AppState{
		Dispatcher: agent.NewDispatcher(svc),
		MarketData: svc,
	}
}

func sendJSON(r *chi.Mux, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestHealth(t *testing.T) {
	rec := sendJSON(NewRouter(newTestState()), http.MethodGet, "/health", "")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "ok", rec.Body.String())
}

func TestListTools(t *testing.T) {
	rec := sendJSON(NewRouter(newTestState()), http.MethodGet, "/api/v1/agent/tools", "")
	assert.Equal(t, http.StatusOK, rec.Code)
	var tools []agent.ToolDefinition
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &tools))
	assert.NotEmpty(t, tools)
}

func TestDispatchHealth(t *testing.T) {
	rec := sendJSON(NewRouter(newTestState()), http.MethodPost, "/api/v1/agent/dispatch", `{"tool":"health","arguments":{}}`)
	assert.Equal(t, http.StatusOK, rec.Code)
	var result agent.ToolResult
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &result))
	assert.Nil(t, result.Error)
}

func TestFetchOhlcv(t *testing.T) {
	rec := sendJSON(NewRouter(newTestState()), http.MethodGet, "/api/v1/market-data/TEST?start=2024-01-01&end=2024-01-05&interval=daily", "")
	assert.Equal(t, http.StatusOK, rec.Code)
	var series []map[string]any
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &series))
	assert.Len(t, series, 5)
}
