package api

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/kevincouton/standard-tools-go/internal/agent"
	"github.com/kevincouton/standard-tools-go/internal/marketdata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func newTestState() *AppState {
	svc := marketdata.NewService("synthetic", marketdata.NewInMemoryCache())
	svc.Register(&marketdata.SyntheticProvider{})
	return &AppState{
		Dispatcher: agent.NewDispatcher(svc),
		MarketData: svc,
	}
}

func sendRequest(r chi.Router, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestHealth(t *testing.T) {
	rec := sendRequest(NewRouter(newTestState()), http.MethodGet, "/health", "")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"status":"ok"}`, rec.Body.String())
}

func TestListTools(t *testing.T) {
	rec := sendRequest(NewRouter(newTestState()), http.MethodGet, "/api/v1/agent/tools", "")
	assert.Equal(t, http.StatusOK, rec.Code)
	var tools []agent.ToolDefinition
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &tools))
	assert.NotEmpty(t, tools)
}

func TestDispatchHealth(t *testing.T) {
	rec := sendRequest(NewRouter(newTestState()), http.MethodPost, "/api/v1/agent/dispatch", `{"tool":"health","arguments":{}}`)
	assert.Equal(t, http.StatusOK, rec.Code)
	var result agent.ToolResult
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &result))
	assert.Nil(t, result.Error)
}

func TestDispatchUnknownTool(t *testing.T) {
	rec := sendRequest(NewRouter(newTestState()), http.MethodPost, "/api/v1/agent/dispatch", `{"tool":"nope","arguments":{}}`)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestFetchOhlcv(t *testing.T) {
	rec := sendRequest(NewRouter(newTestState()), http.MethodGet, "/api/v1/market-data/TEST?start=2024-01-01&end=2024-01-05&interval=daily", "")
	assert.Equal(t, http.StatusOK, rec.Code)
	var series []map[string]any
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &series))
	assert.Len(t, series, 5)
}

func TestFetchOhlcvInvalidTicker(t *testing.T) {
	rec := sendRequest(NewRouter(newTestState()), http.MethodGet, "/api/v1/market-data/%20?start=2024-01-01&end=2024-01-05", "")
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestFetchOhlcvInvalidDate(t *testing.T) {
	rec := sendRequest(NewRouter(newTestState()), http.MethodGet, "/api/v1/market-data/TEST?start=notadate&end=2024-01-05", "")
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestFetchOhlcvMissingDate(t *testing.T) {
	rec := sendRequest(NewRouter(newTestState()), http.MethodGet, "/api/v1/market-data/TEST", "")
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func freePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

func TestEndToEndServer(t *testing.T) {
	state := newTestState()
	httpPort := freePort(t)
	grpcPort := freePort(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- Serve(ctx, state, httpPort, grpcPort)
	}()

	baseURL := "http://127.0.0.1:" + strconv.Itoa(httpPort)
	require.Eventually(t, func() bool {
		resp, err := http.Get(baseURL + "/health")
		if err != nil {
			return false
		}
		_ = resp.Body.Close()
		return resp.StatusCode == http.StatusOK
	}, 2*time.Second, 50*time.Millisecond, "HTTP server did not become ready")

	resp, err := http.Get(baseURL + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	conn, err := grpc.Dial("127.0.0.1:"+strconv.Itoa(grpcPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()
	client := grpc_health_v1.NewHealthClient(conn)
	check, err := client.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
	require.NoError(t, err)
	assert.Equal(t, grpc_health_v1.HealthCheckResponse_SERVING, check.Status)

	cancel()
	select {
	case <-serveErr:
	case <-time.After(2 * time.Second):
		t.Fatal("Serve did not shut down")
	}
}
