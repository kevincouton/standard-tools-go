package api

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/kevincouton/standard-tools-go/internal/agent"
	"github.com/kevincouton/standard-tools-go/internal/core"
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
	assertErrorResponse(t, rec.Body.Bytes(), "BAD_REQUEST")
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
	assertErrorResponse(t, rec.Body.Bytes(), "BAD_REQUEST")
}

func TestFetchOhlcvInvalidDate(t *testing.T) {
	rec := sendRequest(NewRouter(newTestState()), http.MethodGet, "/api/v1/market-data/TEST?start=notadate&end=2024-01-05", "")
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assertErrorResponse(t, rec.Body.Bytes(), "BAD_REQUEST")
}

func TestFetchOhlcvMissingDate(t *testing.T) {
	rec := sendRequest(NewRouter(newTestState()), http.MethodGet, "/api/v1/market-data/TEST", "")
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assertErrorResponse(t, rec.Body.Bytes(), "BAD_REQUEST")
}

func TestFetchOhlcvInvalidInterval(t *testing.T) {
	rec := sendRequest(NewRouter(newTestState()), http.MethodGet, "/api/v1/market-data/TEST?start=2024-01-01&end=2024-01-05&interval=hourly", "")
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assertErrorResponse(t, rec.Body.Bytes(), "BAD_REQUEST")
}

func TestFetchOhlcvIntervalCaseInsensitive(t *testing.T) {
	rec := sendRequest(NewRouter(newTestState()), http.MethodGet, "/api/v1/market-data/TEST?start=2024-01-01&end=2024-01-05&interval=WEEKLY", "")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NotEmpty(t, rec.Body.Bytes())

	// The synthetic provider emits one bar per calendar day regardless of the
	// requested interval, so we cannot assert aggregation here. We verify the
	// response is valid JSON that decodes into OHLCV bars.
	var series []core.OHLCV
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &series))
	assert.NotEmpty(t, series)
}

func TestDomainErrorMapping(t *testing.T) {
	state := newTestState()
	router := NewRouter(state)

	cases := []struct {
		name       string
		method     string
		path       string
		body       string
		wantStatus int
		wantCode   string
	}{
		{
			name:       "invalid command",
			method:     http.MethodPost,
			path:       "/api/v1/agent/dispatch",
			body:       `{"tool":"nope","arguments":{}}`,
			wantStatus: http.StatusBadRequest,
			wantCode:   "BAD_REQUEST",
		},
		{
			name:       "invalid ticker",
			method:     http.MethodGet,
			path:       "/api/v1/market-data/%20?start=2024-01-01&end=2024-01-05",
			wantStatus: http.StatusBadRequest,
			wantCode:   "BAD_REQUEST",
		},
		{
			name:       "invalid date range",
			method:     http.MethodGet,
			path:       "/api/v1/market-data/TEST?start=2024-01-05&end=2024-01-01",
			wantStatus: http.StatusBadRequest,
			wantCode:   "BAD_REQUEST",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := sendRequest(router, tc.method, tc.path, tc.body)
			assert.Equal(t, tc.wantStatus, rec.Code)
			assertErrorResponse(t, rec.Body.Bytes(), tc.wantCode)
		})
	}
}

func TestEndToEndServer(t *testing.T) {
	state := newTestState()

	httpLis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer httpLis.Close()

	grpcLis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer grpcLis.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- serve(ctx, state, httpLis, grpcLis)
	}()

	baseURL := "http://" + httpLis.Addr().String()
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

	grpcCtx, grpcCancel := context.WithTimeout(ctx, 5*time.Second)
	defer grpcCancel()
	conn, err := grpc.DialContext(grpcCtx, grpcLis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()
	client := grpc_health_v1.NewHealthClient(conn)
	check, err := client.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
	require.NoError(t, err)
	assert.Equal(t, grpc_health_v1.HealthCheckResponse_SERVING, check.Status)

	cancel()
	select {
	case err := <-serveErr:
		assert.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("Serve did not shut down")
	}
}

func assertErrorResponse(t *testing.T, body []byte, wantCode string) {
	t.Helper()
	var er errorResponse
	require.NoError(t, json.Unmarshal(body, &er))
	assert.NotEmpty(t, er.Error)
	assert.Equal(t, wantCode, er.Code)
}

func TestErrorCode(t *testing.T) {
	assert.Equal(t, "BAD_REQUEST", errorCode(http.StatusBadRequest))
	assert.Equal(t, "NOT_FOUND", errorCode(http.StatusNotFound))
	assert.Equal(t, "BAD_GATEWAY", errorCode(http.StatusBadGateway))
	assert.Equal(t, "SERVICE_UNAVAILABLE", errorCode(http.StatusServiceUnavailable))
	assert.Equal(t, "INTERNAL_SERVER_ERROR", errorCode(http.StatusInternalServerError))
	assert.Equal(t, "UNKNOWN", errorCode(http.StatusTeapot))
}

