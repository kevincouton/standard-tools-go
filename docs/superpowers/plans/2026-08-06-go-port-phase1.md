# Standard-Tools Go Port — Phase 1 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Bootstrap the Go repository, implement shared core value objects/errors, a synthetic Yahoo-compatible market-data provider, and a minimal HTTP/gRPC API skeleton that can list and dispatch tools end-to-end.

**Architecture:** One Go module with internal packages. `internal/core` holds errors and value objects. `internal/marketdata` defines a provider interface and ships an in-memory cache plus a Yahoo Finance adapter backed by a synthetic test double. `internal/agent` registers a small initial tool set and dispatches calls. `internal/api` wires Chi REST routes and a grpc-go health service. E2E tests exercise the routes without network calls.

**Tech Stack:** Go 1.23+, `github.com/go-chi/chi/v5`, `google.golang.org/grpc`, `github.com/shopspring/decimal`, `github.com/stretchr/testify`.

---

## File Map

| File | Responsibility |
|------|----------------|
| `go.mod` / `go.sum` | Module definition and dependencies |
| `cmd/server/main.go` | HTTP/gRPC server entrypoint |
| `cmd/cli/main.go` | CLI entrypoint (`server`, `audit verify`) |
| `internal/core/errors.go` | Shared error types (`InvalidCommand`, `NotFound`, `DataQuality`, `ProviderNotAvailable`, `Internal`) |
| `internal/core/value_objects.go` | `Ticker`, `DateRange`, `BarInterval`, `Ohlcv` |
| `internal/marketdata/provider.go` | `Provider` interface |
| `internal/marketdata/cache.go` | In-memory cache interface/impl |
| `internal/marketdata/service.go` | `Service` orchestrates providers + cache |
| `internal/marketdata/yahoo.go` | Yahoo Finance provider (production) |
| `internal/marketdata/synthetic.go` | Deterministic synthetic provider (tests) |
| `internal/agent/tool.go` | `ToolDefinition`, `ToolCall`, `ToolResult` |
| `internal/agent/registry.go` | Static tool registry: `health`, `list_tools`, `fetch_ohlcv` |
| `internal/agent/dispatcher.go` | Validates tool name and dispatches known tools |
| `internal/api/server.go` | Starts HTTP + gRPC servers |
| `internal/api/rest.go` | Chi REST router: `/health`, `/api/v1/agent/tools`, `/api/v1/agent/dispatch`, `/api/v1/market-data/:ticker` |
| `internal/api/grpc.go` | gRPC health service |
| `internal/api/state.go` | Shared `AppState` |
| `internal/api/e2e_test.go` | End-to-end tests for REST + gRPC |
| `.mise.toml` | Local task runner definitions |
| `.github/workflows/ci.yml` | GitHub Actions CI |
| `README.md` | Quick start and stack overview |

---

### Task 0: Bootstrap Go Module

**Files:**
- Create: `go.mod`
- Create: `.mise.toml`
- Create: `.github/workflows/ci.yml`
- Create: `README.md`

- [ ] **Step 1: Initialize module**

```bash
cd /Users/kevincouton/Repo/standard-tools-go
go mod init github.com/kevincouton/standard-tools-go
```

- [ ] **Step 2: Add `.mise.toml`**

```toml
[tools]
go = "1.23"
act = "0.2.75"
podman = "latest"

[tasks]
build = "go build ./cmd/server && go build ./cmd/cli"
test = "go test ./... 2>&1 | tee test-output.log && ./scripts/visual-test-report.sh test-output.log test-report.html"
test-integration = "go test ./... -tags integration"
fmt = "go fmt ./..."
vet = "go vet ./..."
lint = "golangci-lint run ./..."
run = "go run ./cmd/server"
image = "podman build -f Dockerfile -t standard-tools-go:latest ."
image-native = "podman build -f Dockerfile.native -t standard-tools-go:native ."
act = "./scripts/run-act-local.sh"
```

- [ ] **Step 3: Add CI workflow**

Create `.github/workflows/ci.yml`:

```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  quality:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.23"
      - run: go fmt ./...
      - run: go vet ./...
      - run: go test ./...

  build-images:
    runs-on: ubuntu-latest
    needs: quality
    steps:
      - uses: actions/checkout@v4
      - run: |
          sudo apt-get update
          sudo apt-get install -y podman
      - run: podman build -f Dockerfile -t standard-tools-go:latest .
      - run: podman build -f Dockerfile.native -t standard-tools-go:native .
```

- [ ] **Step 4: Add README**

Create `README.md`:

```markdown
# standard-tools-go

Go port of the Standard-Tools quantitative finance toolkit.

## Stack

- Go 1.23+
- Chi (REST)
- grpc-go (gRPC)
- pgx (PostgreSQL)
- Gonum (math)

## Quick Start

```bash
mise install
mise run build
mise run test
```

## Endpoints

- REST: `/api/v1/*`
- gRPC: `standard_tools.health`
- A2A: `/a2a/*` (planned)
- MCP: `/mcp/*` (planned)
```

- [ ] **Step 5: Commit**

```bash
git add go.mod .mise.toml .github/workflows/ci.yml README.md
git commit -m "chore: bootstrap Go module, mise, CI, README"
```

---

### Task 1: Core Errors and Value Objects

**Files:**
- Create: `internal/core/errors.go`
- Create: `internal/core/value_objects.go`
- Create: `internal/core/value_objects_test.go`

- [ ] **Step 1: Write failing test for `Ticker`**

Create `internal/core/value_objects_test.go`:

```go
package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTicker(t *testing.T) {
	ticker, err := NewTicker("AAPL")
	assert.NoError(t, err)
	assert.Equal(t, "AAPL", ticker.Symbol)
}

func TestNewTickerRejectsEmpty(t *testing.T) {
	_, err := NewTicker("")
	assert.ErrorIs(t, err, ErrInvalidCommand)
}
```

- [ ] **Step 2: Run tests and confirm failure**

```bash
go test ./internal/core -v
```

Expected: `undefined: NewTicker`, `undefined: ErrInvalidCommand`.

- [ ] **Step 3: Implement errors and value objects**

Create `internal/core/errors.go`:

```go
package core

import "errors"

var (
	ErrInvalidCommand        = errors.New("invalid command")
	ErrNotFound              = errors.New("not found")
	ErrDataQuality           = errors.New("data quality")
	ErrProviderNotAvailable  = errors.New("provider not available")
	ErrInternal              = errors.New("internal error")
)
```

Create `internal/core/value_objects.go`:

```go
package core

import (
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

type Ticker struct {
	Symbol   string
	Exchange string
}

func NewTicker(symbol string) (Ticker, error) {
	s := strings.TrimSpace(symbol)
	if s == "" {
		return Ticker{}, ErrInvalidCommand
	}
	return Ticker{Symbol: s}, nil
}

type BarInterval int

const (
	Daily BarInterval = iota
	Weekly
	Monthly
)

func (b BarInterval) String() string {
	switch b {
	case Daily:
		return "daily"
	case Weekly:
		return "weekly"
	case Monthly:
		return "monthly"
	default:
		return "daily"
	}
}

type DateRange struct {
	Start time.Time
	End   time.Time
}

func NewDateRange(start, end time.Time) (DateRange, error) {
	if end.Before(start) {
		return DateRange{}, ErrInvalidCommand
	}
	return DateRange{Start: start, End: end}, nil
}

type Ohlcv struct {
	Date   time.Time
	Open   decimal.Decimal
	High   decimal.Decimal
	Low    decimal.Decimal
	Close  decimal.Decimal
	Volume int64
}
```

- [ ] **Step 4: Run tests and confirm pass**

```bash
go test ./internal/core -v
```

Expected: `PASS`.

- [ ] **Step 5: Commit**

```bash
git add internal/core/
git commit -m "feat(core): add errors, ticker, date-range, OHLCV"
```

---

### Task 2: Market Data Provider and Service

**Files:**
- Create: `internal/marketdata/provider.go`
- Create: `internal/marketdata/cache.go`
- Create: `internal/marketdata/service.go`
- Create: `internal/marketdata/synthetic.go`
- Create: `internal/marketdata/yahoo.go`
- Create: `internal/marketdata/service_test.go`

- [ ] **Step 1: Write interface and cache**

Create `internal/marketdata/provider.go`:

```go
package marketdata

import (
	"context"

	"github.com/kevincouton/standard-tools-go/internal/core"
)

type Provider interface {
	Name() string
	Fetch(ctx context.Context, ticker core.Ticker, interval core.BarInterval, rng core.DateRange) ([]core.Ohlcv, error)
}
```

Create `internal/marketdata/cache.go`:

```go
package marketdata

import (
	"context"

	"github.com/kevincouton/standard-tools-go/internal/core"
)

type Cache interface {
	Get(ctx context.Context, key string) ([]core.Ohlcv, bool)
	Put(ctx context.Context, key string, series []core.Ohlcv)
}

type InMemoryCache struct {
	data map[string][]core.Ohlcv
}

func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{data: make(map[string][]core.Ohlcv)}
}

func (c *InMemoryCache) Get(ctx context.Context, key string) ([]core.Ohlcv, bool) {
	series, ok := c.data[key]
	return series, ok
}

func (c *InMemoryCache) Put(ctx context.Context, key string, series []core.Ohlcv) {
	c.data[key] = series
}
```

- [ ] **Step 2: Write synthetic provider**

Create `internal/marketdata/synthetic.go`:

```go
package marketdata

import (
	"context"
	"time"

	"github.com/kevincouton/standard-tools-go/internal/core"
	"github.com/shopspring/decimal"
)

type SyntheticProvider struct{}

func (s *SyntheticProvider) Name() string { return "synthetic" }

func (s *SyntheticProvider) Fetch(ctx context.Context, ticker core.Ticker, interval core.BarInterval, rng core.DateRange) ([]core.Ohlcv, error) {
	var bars []core.Ohlcv
	price := decimal.NewFromInt(100)
	for d := rng.Start; !d.After(rng.End); d = d.AddDate(0, 0, 1) {
		open := price
		close := price.Add(decimal.NewFromInt(1))
		high := open.Max(close).Add(decimal.NewFromFloat(0.5))
		low := open.Min(close).Sub(decimal.NewFromFloat(0.5))
		bars = append(bars, core.Ohlcv{
			Date:   d,
			Open:   open,
			High:   high,
			Low:    low,
			Close:  close,
			Volume: 1_000_000,
		})
		price = close
	}
	return bars, nil
}
```

- [ ] **Step 3: Write service and stub Yahoo provider**

Create `internal/marketdata/service.go`:

```go
package marketdata

import (
	"context"
	"fmt"

	"github.com/kevincouton/standard-tools-go/internal/core"
)

type Service struct {
	defaultProvider string
	providers       map[string]Provider
	cache           Cache
}

func NewService(defaultProvider string, cache Cache) *Service {
	return &Service{
		defaultProvider: defaultProvider,
		providers:       make(map[string]Provider),
		cache:           cache,
	}
}

func (s *Service) Register(provider Provider) {
	s.providers[provider.Name()] = provider
}

func (s *Service) Fetch(ctx context.Context, ticker core.Ticker, interval core.BarInterval, rng core.DateRange, providerName *string) ([]core.Ohlcv, error) {
	name := s.defaultProvider
	if providerName != nil && *providerName != "" {
		name = *providerName
	}
	provider, ok := s.providers[name]
	if !ok {
		return nil, fmt.Errorf("%w: provider %s not registered", core.ErrProviderNotAvailable, name)
	}
	key := fmt.Sprintf("%s:%s:%s:%s", name, ticker.Symbol, interval.String(), rng.Start.Format("2006-01-02"))
	if cached, hit := s.cache.Get(ctx, key); hit {
		return cached, nil
	}
	series, err := provider.Fetch(ctx, ticker, interval, rng)
	if err != nil {
		return nil, err
	}
	s.cache.Put(ctx, key, series)
	return series, nil
}
```

Create `internal/marketdata/yahoo.go`:

```go
package marketdata

import (
	"context"
	"fmt"

	"github.com/kevincouton/standard-tools-go/internal/core"
)

type YahooProvider struct{}

func (y *YahooProvider) Name() string { return "yahoo" }

func (y *YahooProvider) Fetch(ctx context.Context, ticker core.Ticker, interval core.BarInterval, rng core.DateRange) ([]core.Ohlcv, error) {
	return nil, fmt.Errorf("%w: yahoo fetch not yet implemented", core.ErrProviderNotAvailable)
}
```

- [ ] **Step 4: Write service test**

Create `internal/marketdata/service_test.go`:

```go
package marketdata

import (
	"context"
	"testing"
	"time"

	"github.com/kevincouton/standard-tools-go/internal/core"
	"github.com/stretchr/testify/assert"
)

func TestServiceFetchesSyntheticData(t *testing.T) {
	svc := NewService("synthetic", NewInMemoryCache())
	svc.Register(&SyntheticProvider{})
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC)
	rng, _ := core.NewDateRange(start, end)
	ticker, _ := core.NewTicker("TEST")
	series, err := svc.Fetch(context.Background(), ticker, core.Daily, rng, nil)
	assert.NoError(t, err)
	assert.Len(t, series, 5)
}
```

- [ ] **Step 5: Run tests and commit**

```bash
go test ./internal/marketdata -v
```

Expected: `PASS`.

```bash
git add internal/marketdata/
git commit -m "feat(marketdata): add provider interface, cache, synthetic and yahoo providers"
```

---

### Task 3: Agent Tool Registry and Dispatcher

**Files:**
- Create: `internal/agent/tool.go`
- Create: `internal/agent/registry.go`
- Create: `internal/agent/dispatcher.go`
- Create: `internal/agent/dispatcher_test.go`

- [ ] **Step 1: Define tool types**

Create `internal/agent/tool.go`:

```go
package agent

import "encoding/json"

type ToolDefinition struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

type ToolCall struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type ToolResult struct {
	Output json.RawMessage `json:"output"`
	Error  *string         `json:"error,omitempty"`
}

func OkResult(output json.RawMessage) ToolResult {
	return ToolResult{Output: output}
}

func ErrResult(msg string) ToolResult {
	return ToolResult{Error: &msg}
}
```

- [ ] **Step 2: Build registry**

Create `internal/agent/registry.go`:

```go
package agent

import "encoding/json"

func ListTools() []ToolDefinition {
	return []ToolDefinition{
		{
			Name:        "health",
			Description: "Return agent health status.",
			Parameters:  json.RawMessage(`{"type":"object","properties":{}}`),
		},
		{
			Name:        "list_tools",
			Description: "List all registered tool names.",
			Parameters:  json.RawMessage(`{"type":"object","properties":{}}`),
		},
		{
			Name:        "fetch_ohlcv",
			Description: "Fetch OHLCV bars for a single ticker.",
			Parameters:  json.RawMessage(`{"type":"object","properties":{"ticker":{"type":"string"},"start":{"type":"string","format":"date"},"end":{"type":"string","format":"date"},"interval":{"type":"string","enum":["daily","weekly","monthly"]},"provider":{"type":"string"}},"required":["ticker","start","end"]}`),
		},
	}
}

func FindTool(name string) (*ToolDefinition, bool) {
	for _, t := range ListTools() {
		if t.Name == name {
			return &t, true
		}
	}
	return nil, false
}
```

- [ ] **Step 3: Build dispatcher**

Create `internal/agent/dispatcher.go`:

```go
package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kevincouton/standard-tools-go/internal/core"
	"github.com/kevincouton/standard-tools-go/internal/marketdata"
)

type Dispatcher struct {
	marketData *marketdata.Service
}

func NewDispatcher(marketData *marketdata.Service) *Dispatcher {
	return &Dispatcher{marketData: marketData}
}

func (d *Dispatcher) Dispatch(ctx context.Context, call ToolCall) (ToolResult, error) {
	if _, ok := FindTool(call.Name); !ok {
		return ToolResult{}, fmt.Errorf("%w: unknown tool %s", core.ErrInvalidCommand, call.Name)
	}
	return d.dispatchKnown(ctx, call)
}

func (d *Dispatcher) dispatchKnown(ctx context.Context, call ToolCall) (ToolResult, error) {
	switch call.Name {
	case "health":
		return OkResult(json.RawMessage(`{"status":"ok"}`)), nil
	case "list_tools":
		names := make([]string, 0, len(ListTools()))
		for _, t := range ListTools() {
			names = append(names, t.Name)
		}
		out, _ := json.Marshal(names)
		return OkResult(out), nil
	case "fetch_ohlcv":
		return d.fetchOhlcv(ctx, call.Arguments)
	default:
		return ToolResult{}, fmt.Errorf("%w: unknown tool %s", core.ErrInvalidCommand, call.Name)
	}
}

func (d *Dispatcher) fetchOhlcv(ctx context.Context, args json.RawMessage) (ToolResult, error) {
	var payload struct {
		Ticker   string  `json:"ticker"`
		Start    string  `json:"start"`
		End      string  `json:"end"`
		Interval string  `json:"interval"`
		Provider *string `json:"provider"`
	}
	if err := json.Unmarshal(args, &payload); err != nil {
		return ToolResult{}, fmt.Errorf("%w: invalid arguments: %v", core.ErrInvalidCommand, err)
	}
	ticker, err := core.NewTicker(payload.Ticker)
	if err != nil {
		return ToolResult{}, err
	}
	start, err := time.Parse("2006-01-02", payload.Start)
	if err != nil {
		return ToolResult{}, fmt.Errorf("%w: invalid start date", core.ErrInvalidCommand)
	}
	end, err := time.Parse("2006-01-02", payload.End)
	if err != nil {
		return ToolResult{}, fmt.Errorf("%w: invalid end date", core.ErrInvalidCommand)
	}
	rng, err := core.NewDateRange(start, end)
	if err != nil {
		return ToolResult{}, err
	}
	interval := core.Daily
	switch payload.Interval {
	case "weekly":
		interval = core.Weekly
	case "monthly":
		interval = core.Monthly
	}
	series, err := d.marketData.Fetch(ctx, ticker, interval, rng, payload.Provider)
	if err != nil {
		return ErrResult(err.Error()), nil
	}
	out, _ := json.Marshal(series)
	return OkResult(out), nil
}
```

`dispatcher.go` imports:

```go
import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kevincouton/standard-tools-go/internal/core"
	"github.com/kevincouton/standard-tools-go/internal/marketdata"
)
```

- [ ] **Step 4: Write dispatcher test**

Create `internal/agent/dispatcher_test.go`:

```go
package agent

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/kevincouton/standard-tools-go/internal/marketdata"
	"github.com/stretchr/testify/assert"
)

func TestDispatchHealth(t *testing.T) {
	svc := marketdata.NewService("synthetic", marketdata.NewInMemoryCache())
	svc.Register(&marketdata.SyntheticProvider{})
	d := NewDispatcher(svc)
	res, err := d.Dispatch(context.Background(), ToolCall{Name: "health", Arguments: json.RawMessage(`{}`)})
	assert.NoError(t, err)
	assert.Nil(t, res.Error)
	assert.JSONEq(t, `{"status":"ok"}`, string(res.Output))
}

func TestDispatchUnknownTool(t *testing.T) {
	svc := marketdata.NewService("synthetic", marketdata.NewInMemoryCache())
	d := NewDispatcher(svc)
	_, err := d.Dispatch(context.Background(), ToolCall{Name: "nope"})
	assert.Error(t, err)
}
```

- [ ] **Step 5: Run tests and commit**

```bash
go test ./internal/agent -v
```

Expected: `PASS`.

```bash
git add internal/agent/
git commit -m "feat(agent): add tool registry and dispatcher"
```

---

### Task 4: REST and gRPC API Skeleton

**Files:**
- Create: `internal/api/state.go`
- Create: `internal/api/rest.go`
- Create: `internal/api/grpc.go`
- Create: `internal/api/server.go`
- Create: `cmd/server/main.go`
- Create: `cmd/cli/main.go`
- Create: `internal/api/e2e_test.go`

- [ ] **Step 1: Add shared state**

Create `internal/api/state.go`:

```go
package api

import "github.com/kevincouton/standard-tools-go/internal/agent"

type AppState struct {
	Dispatcher *agent.Dispatcher
}
```

- [ ] **Step 2: Add REST router**

Create `internal/api/rest.go`:

```go
package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/kevincouton/standard-tools-go/internal/agent"
	"github.com/kevincouton/standard-tools-go/internal/core"
)

func NewRouter(state *AppState) *chi.Mux {
	r := chi.NewRouter()
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	r.Get("/api/v1/agent/tools", listTools)
	r.Post("/api/v1/agent/dispatch", dispatchTool(state))
	r.Get("/api/v1/market-data/:ticker", fetchOhlcv(state))
	return r
}

func listTools(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, agent.ListTools())
}

func dispatchTool(state *AppState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Tool      string          `json:"tool"`
			Arguments json.RawMessage `json:"arguments"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		result, err := state.Dispatcher.Dispatch(r.Context(), agent.ToolCall{Name: req.Tool, Arguments: req.Arguments})
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusOK, result)
	}
}

func fetchOhlcv(state *AppState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ticker, err := core.NewTicker(chi.URLParam(r, "ticker"))
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		query := r.URL.Query()
		start, err := time.Parse("2006-01-02", query.Get("start"))
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		end, err := time.Parse("2006-01-02", query.Get("end"))
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		rng, err := core.NewDateRange(start, end)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		interval := core.Daily
		switch query.Get("interval") {
		case "weekly":
			interval = core.Weekly
		case "monthly":
			interval = core.Monthly
		}
		provider := query.Get("provider")
		var p *string
		if provider != "" {
			p = &provider
		}
		series, err := state.MarketData.Fetch(r.Context(), ticker, interval, rng, p)
		if err != nil {
			writeError(w, http.StatusServiceUnavailable, err)
			return
		}
		writeJSON(w, http.StatusOK, series)
	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}
```

Wait, `state.Dispatcher.MarketDataService()` doesn't exist. Add a method to Dispatcher or pass service separately. Better: add `MarketData *marketdata.Service` to AppState. Then use `state.MarketData.Fetch`. Also `fetchOhlcv` duplicates dispatcher logic. Simplify: use the dispatcher with tool `fetch_ohlcv`? But the REST endpoint builds arguments. Let's keep AppState with MarketData and Dispatcher.

Revise `internal/api/state.go`:

```go
package api

import (
	"github.com/kevincouton/standard-tools-go/internal/agent"
	"github.com/kevincouton/standard-tools-go/internal/marketdata"
)

type AppState struct {
	Dispatcher *agent.Dispatcher
	MarketData *marketdata.Service
}
```

And in `fetchOhlcv` use `state.MarketData.Fetch`.

- [ ] **Step 3: Add gRPC health service**

Create `proto/health.proto`:

```protobuf
syntax = "proto3";
package standard_tools.health;
option go_package = "github.com/kevincouton/standard-tools-go/proto/health";

service Health {
  rpc Check (HealthCheckRequest) returns (HealthCheckResponse);
}

message HealthCheckRequest {}

message HealthCheckResponse {
  string status = 1;
}
```

Create `internal/api/grpc.go`:

```go
package api

import (
	"context"

	pb "github.com/kevincouton/standard-tools-go/proto/health"
)

type HealthServer struct {
	pb.UnimplementedHealthServer
}

func (s *HealthServer) Check(ctx context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	return &pb.HealthCheckResponse{Status: "ok"}, nil
}
```

- [ ] **Step 4: Add server wiring**

Create `internal/api/server.go`:

```go
package api

import (
	"fmt"
	"net"
	"net/http"

	pb "github.com/kevincouton/standard-tools-go/proto/health"
	"google.golang.org/grpc"
)

func Serve(state *AppState, httpPort, grpcPort int) error {
	errCh := make(chan error, 2)

	go func() {
		r := NewRouter(state)
		errCh <- http.ListenAndServe(fmt.Sprintf(":%d", httpPort), r)
	}()

	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
		if err != nil {
			errCh <- err
			return
		}
		s := grpc.NewServer()
		pb.RegisterHealthServer(s, &HealthServer{})
		errCh <- s.Serve(lis)
	}()

	return <-errCh
}
```

- [ ] **Step 5: Add cmd/server and cmd/cli**

Create `cmd/server/main.go`:

```go
package main

import (
	"log/slog"
	"os"

	"github.com/kevincouton/standard-tools-go/internal/agent"
	"github.com/kevincouton/standard-tools-go/internal/api"
	"github.com/kevincouton/standard-tools-go/internal/marketdata"
)

func main() {

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	cache := marketdata.NewInMemoryCache()
	svc := marketdata.NewService("synthetic", cache)
	svc.Register(&marketdata.SyntheticProvider{})
	svc.Register(&marketdata.YahooProvider{})

	state := &api.AppState{
		Dispatcher: agent.NewDispatcher(svc),
		MarketData: svc,
	}

	slog.Info("starting server", "http", 8080, "grpc", 50051)
	if err := api.Serve(state, 8080, 50051); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}
```

Create `cmd/cli/main.go`:

```go
package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: standard-tools <server|audit>")
		os.Exit(1)
	}
	switch os.Args[1] {
	case "server":
		// Delegate to server binary; kept simple for now.
		fmt.Println("use cmd/server to run the server")
	case "audit":
		fmt.Println("audit verify not yet implemented")
	default:
		fmt.Printf("unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}
```

- [ ] **Step 6: Add E2E test**

Create `internal/api/e2e_test.go`:

```go
package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

func TestHealth(t *testing.T) {
	r := NewRouter(newTestState())
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "ok", rec.Body.String())
}

func TestListTools(t *testing.T) {
	r := NewRouter(newTestState())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/agent/tools", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	var tools []agent.ToolDefinition
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &tools))
	assert.NotEmpty(t, tools)
}

func TestDispatchHealth(t *testing.T) {
	r := NewRouter(newTestState())
	body := `{"tool":"health","arguments":{}}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/agent/dispatch", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	var result agent.ToolResult
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &result))
	assert.Nil(t, result.Error)
}

func TestFetchOhlcv(t *testing.T) {
	r := NewRouter(newTestState())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/market-data/TEST?start=2024-01-01&end=2024-01-05&interval=daily", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	var series []map[string]any
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &series))
	assert.Len(t, series, 5)
}
```

Add missing `strings` import.

- [ ] **Step 7: Generate gRPC code and tidy**

```bash
cd /Users/kevincouton/Repo/standard-tools-go
go get google.golang.org/protobuf/cmd/protoc-gen-go google.golang.org/grpc/cmd/protoc-gen-go-grpc
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/health.proto
go mod tidy
```

- [ ] **Step 8: Run all tests**

```bash
go test ./... -v
```

Expected: all tests pass.

- [ ] **Step 9: Commit**

```bash
git add internal/api cmd proto go.mod go.sum
git commit -m "feat(api): add REST, gRPC health, and E2E tests"
```

---

## Spec Coverage Check

| Spec Requirement | Task |
|------------------|------|
| Go module + mise/CI | Task 0 |
| Core errors and value objects | Task 1 |
| Market data provider + cache + synthetic | Task 2 |
| Agent registry + dispatcher | Task 3 |
| REST health/tools/dispatch/market-data | Task 4 |
| gRPC health service | Task 4 |
| E2E tests | Task 4 |

A2A, MCP, PostgreSQL audit, orders, indicators/metrics/analysis/backtest/portfolio/screener are deferred to Phase 2.
