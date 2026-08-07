# Standard-Tools Go Porting Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` (recommended) or `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Port the Python *Standard Quant Tools* SDK (`/Users/kevincouton/Repo/Standard-Tools`) into the existing Go project (`github.com/kevincouton/standard-tools-go`), achieving functional parity for the 42+ agent tools and adding production-grade REST/gRPC/A2A/MCP APIs, PostgreSQL persistence, hash-chained audit, native/container packaging, and local CI via `act` + `podman`.

**Architecture:** One Go module with `internal/` packages aligned to domain boundaries (`core`, `marketdata`, `indicators`, `metrics`, `analysis`, `backtest`, `portfolio`, `screener`, `agent`, `audit`, `storage`, `config`, `api`, `cli`). Python's pandas/numpy vectorized kernels become explicit slice-based Go kernels using `gonum` + plain loops. Agent tools are strongly-typed functions registered in `internal/agent`. APIs are thin, stateless adapters over the dispatcher. Audit records are hash-chained and persisted to PostgreSQL. The CLI uses `cobra`.

**Tech Stack:** Go 1.25+, Chi, grpc-go, pgx/v5, shopspring/decimal, gonum, testify, koanf, cobra, podman, act.

---

## 1. Porting Style & Philosophy

### 1.1 Keep it idiomatic Go
- Prefer explicit error returns over exceptions. Map Python's exception hierarchy to typed sentinel errors in `internal/core/errors.go`.
- Use value objects (`Ticker`, `DateRange`, `OHLCV`) with constructors that validate.
- Use interfaces for extension points (`marketdata.Provider`, `audit.Storage`, `config.Loader`).
- Avoid reflection and `interface{}` where possible; agent tool inputs become Go structs.

### 1.2 Replace pandas/numpy with Go-native structures
- OHLCV data is `[]core.OHLCV` (already implemented).
- Time series of floats are `[]float64` aligned with a `[]time.Time` index or a `Series` struct.
- Cross-sectional tables are `[]TableRow` structs or `map[string][]float64` keyed by column.
- Use `gonum` for linear algebra (PCA, regression), optimization, and statistics.
- Use `shopspring/decimal` for money/prices; use `float64` for indicator math when the Python code uses floats.

### 1.3 Preserve behavior, not line-for-line code
- Match outputs (rounding, NaN handling, lag semantics) rather than translating Python idioms literally.
- Where Python uses `shift(1)` for signal lag, the Go engine must apply the same one-bar delay.
- Record rounding rules from Python outputs in test expectations.

### 1.4 Concurrency
- Replace Python `asyncio` / `ProcessPoolExecutor` with goroutines + `errgroup`.
- Keep provider caches thread-safe (already done for the in-memory cache).
- Use worker pools for CPU-bound screens and grid searches to bound resource usage.

### 1.5 Configuration
- Single source of truth: `koanf` loading env vars + optional `.env` file via `joho/godotenv`.
- Env-var names stay close to Python (`SQT_CACHE_DIR`, `SQT_AUDIT_*`, `SQT_POLYGON_API_KEY`, etc.).

### 1.6 Testing parity
- Port the Python unit tests to Go table-driven tests.
- Maintain a *parity suite* that compares Go outputs against recorded Python outputs for a fixed set of inputs.
- Integration tests require PostgreSQL and optionally Polygon/Yahoo network; gate them with build tags.

---

## 2. Cross-Cutting Decisions

| Concern | Decision | Rationale |
|---------|----------|-----------|
| **Numeric type** | Prices/volumes use `decimal.Decimal`; indicator math uses `float64`. | Matches Python's mixed use of `float` for math and implicit decimal for display. |
| **OHLCV representation** | `[]core.OHLCV` sorted ascending by date. | Simple, cache-friendly, easy to marshal. |
| **Missing data** | Use NaN `math.NaN()` for missing floats; omit bars for missing dates. | Matches pandas NaN semantics; document clearly. |
| **Date/time** | All internal times are UTC `time.Time`; daily bars truncate to midnight UTC. | Avoids timezone bugs; reproduce Python's UTC normalization. |
| **Provider errors** | All provider errors wrap `core.ErrProviderNotAvailable` or `core.ErrDataQuality`. | Allows unified HTTP mapping. |
| **Audit storage** | PostgreSQL is the primary store; local JSONL remains an optional backend for offline use. | User requested PostgreSQL; hash-chain logic is storage-agnostic. |
| **API protocols** | REST for humans, gRPC for internal services, A2A for agent-agent tasks, MCP for model context. | Covers all user-requested surfaces. |
| **Container builds** | Two images: `Dockerfile` (distroless, dynamic libc) and `Dockerfile.static` (scratch, statically linked). | "Native" in Go means a static binary; the classic image supports CGO if needed later. |
| **Local CI** | `act` runs the GitHub Actions workflow inside `podman`. | User requested mise + podman local CI. |

---

## 3. Phase Roadmap

### Phase 2 — Foundation: Data Layer, Persistence, Audit, Config, CLI

**Goal:** Make the project a runnable, configurable, auditable service with extended data providers and a real CLI.

**Files to create:**
- `internal/config/config.go`
- `internal/storage/migrations/*.sql`
- `internal/storage/postgres.go`
- `internal/audit/record.go`, `writer.go`, `verifier.go`, `replay.go`
- `internal/marketdata/metadata.go`, `info.go`, `retry.go`, `factory.go`
- `internal/marketdata/polygon.go`, `yahoo.go` (real HTTP client)
- `internal/cli/root.go`, `cmd/cli/main.go` (replace stub)
- `internal/api/a2a.go`, `mcp.go`
- `Dockerfile`, `Dockerfile.static`, `scripts/run-act-local.sh`
- `.mise.toml` updates

**Files to modify:**
- `internal/core/errors.go` — add provider/audit/storage errors
- `internal/core/value_objects.go` — add `Interval`, `Metadata`, `TickerInfo`, `FinancialRatios`
- `internal/marketdata/provider.go` — extend interface
- `internal/marketdata/service.go` — add metadata/financials methods
- `internal/api/rest.go` — add A2A/MCP routes
- `internal/api/server.go` — add config-driven startup
- `cmd/server/main.go` — load config, connect DB
- `.github/workflows/ci.yml` — add Postgres service for integration tests

#### Task 2.1: Configuration package

- [ ] **Step 1: Add `github.com/knadh/koanf` and `github.com/joho/godotenv` dependencies**

Run:
```bash
cd /Users/kevincouton/Repo/standard-tools-go
go get github.com/knadh/koanf/v2 github.com/knadh/koanf/parsers/toml github.com/knadh/koanf/providers/env github.com/knadh/koanf/providers/file github.com/joho/godotenv
```

- [ ] **Step 2: Define config schema and loader**

Create `internal/config/config.go`:

```go
package config

import (
	"fmt"
	"strings"

	"github.com/joho/godotenv"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type Config struct {
	HTTPPort    int    `koanf:"http_port"`
	GRPCPort    int    `koanf:"grpc_port"`
	LogLevel    string `koanf:"log_level"`
	DatabaseURL string `koanf:"database_url"`
	CacheDir    string `koanf:"cache_dir"`
	AuditDir    string `koanf:"audit_dir"`
	Polygon     PolygonConfig `koanf:"polygon"`
}

type PolygonConfig struct {
	APIKey string `koanf:"api_key"`
}

func Load(paths ...string) (*Config, error) {
	_ = godotenv.Load(".env")
	k := koanf.New(".")
	for _, p := range paths {
		_ = k.Load(file.Provider(p), toml.Parser())
	}
	_ = k.Load(env.Provider("SQT_", ".", func(s string) string {
		key := strings.ToLower(strings.TrimPrefix(s, "SQT_"))
		// Double underscore in env vars denotes nesting: SQT_POLYGON__API_KEY -> polygon.api_key
		return strings.ReplaceAll(key, "__", ".")
	}), nil)
	var cfg Config
	if err := k.Unmarshal("", &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	setDefaults(&cfg)
	return &cfg, nil
}

func setDefaults(cfg *Config) {
	if cfg.HTTPPort == 0 {
		cfg.HTTPPort = 8080
	}
	if cfg.GRPCPort == 0 {
		cfg.GRPCPort = 50051
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}
}
```

- [ ] **Step 3: Write config test**

Create `internal/config/config_test.go`:

```go
package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadDefaults(t *testing.T) {
	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, 8080, cfg.HTTPPort)
	assert.Equal(t, 50051, cfg.GRPCPort)
}

func TestLoadEnv(t *testing.T) {
	t.Setenv("SQT_HTTP_PORT", "9090")
	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, 9090, cfg.HTTPPort)
}
```

- [ ] **Step 4: Run tests**

```bash
go test ./internal/config -v
```
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/config go.mod go.sum
git commit -m "feat(config): add koanf-based configuration loader"
```

#### Task 2.2: PostgreSQL storage and migrations

- [ ] **Step 1: Add `pgx/v5` and `golang-migrate`**

```bash
go get github.com/jackc/pgx/v5
```

- [ ] **Step 2: Create initial migration**

Create `internal/storage/migrations/000001_create_audit_table.up.sql`:

```sql
CREATE TABLE IF NOT EXISTS audit_records (
    id BIGSERIAL PRIMARY KEY,
    request_id UUID NOT NULL UNIQUE,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    tool_name TEXT NOT NULL,
    input_hash TEXT NOT NULL,
    output_hash TEXT NOT NULL,
    status TEXT NOT NULL,
    error TEXT,
    git_commit_sha TEXT,
    package_version TEXT,
    random_seed BIGINT,
    prev_record_hash TEXT,
    record_hash TEXT NOT NULL,
    raw JSONB NOT NULL
);

CREATE INDEX idx_audit_recorded_at ON audit_records(recorded_at);
CREATE INDEX idx_audit_tool_name ON audit_records(tool_name);
```

Create `internal/storage/migrations/000001_create_audit_table.down.sql`:

```sql
DROP TABLE IF EXISTS audit_records;
```

- [ ] **Step 3: Add postgres connection helper**

Create `internal/storage/postgres.go`:

```go
package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return pool, nil
}
```

- [ ] **Step 4: Add storage tests**

Create `internal/storage/postgres_test.go`:

```go
//go:build integration

package storage

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewPool(t *testing.T) {
	ctx := context.Background()
	url := os.Getenv("SQT_DATABASE_URL")
	if url == "" {
		t.Skip("SQT_DATABASE_URL not set")
	}
	pool, err := NewPool(ctx, url)
	require.NoError(t, err)
	defer pool.Close()
}
```

- [ ] **Step 5: Commit**

```bash
git add internal/storage go.mod go.sum
git commit -m "feat(storage): add PostgreSQL pool and audit migration"
```

#### Task 2.3: Hash-chained audit in Go

- [ ] **Step 1: Define audit model and hashing**

Create `internal/audit/record.go`:

```go
package audit

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"
)

type DecisionRecord struct {
	RequestID      string    `json:"request_id"`
	RecordedAt     time.Time `json:"recorded_at"`
	ToolName       string    `json:"tool_name"`
	Input          any       `json:"input"`
	InputHash      string    `json:"input_hash"`
	Output         any       `json:"output"`
	OutputHash     string    `json:"output_hash"`
	Status         string    `json:"status"`
	Error          string    `json:"error,omitempty"`
	GitCommitSHA   string    `json:"git_commit_sha"`
	PackageVersion string    `json:"package_version"`
	RandomSeed     int64     `json:"random_seed"`
	PrevRecordHash string    `json:"prev_record_hash"`
	RecordHash     string    `json:"record_hash"`
}

func hashBytes(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

func HashRecord(r DecisionRecord) (string, error) {
	r.RecordHash = ""
	b, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return hashBytes(b), nil
}
```

- [ ] **Step 2: Define storage interface and PostgreSQL writer**

Create `internal/audit/writer.go`:

```go
package audit

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage interface {
	Append(ctx context.Context, r DecisionRecord) error
	Latest(ctx context.Context) (DecisionRecord, error)
}

type PostgresStorage struct {
	pool *pgxpool.Pool
}

func NewPostgresStorage(pool *pgxpool.Pool) *PostgresStorage {
	return &PostgresStorage{pool: pool}
}

func (p *PostgresStorage) Append(ctx context.Context, r DecisionRecord) error {
	inputBytes, _ := json.Marshal(r.Input)
	outputBytes, _ := json.Marshal(r.Output)
	r.InputHash = hashBytes(inputBytes)
	r.OutputHash = hashBytes(outputBytes)
	recordHash, err := HashRecord(r)
	if err != nil {
		return fmt.Errorf("hash record: %w", err)
	}
	r.RecordHash = recordHash

	_, err = p.pool.Exec(ctx, `
		INSERT INTO audit_records
		(request_id, recorded_at, tool_name, input_hash, output_hash, status, error,
		 git_commit_sha, package_version, random_seed, prev_record_hash, record_hash, raw)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`, r.RequestID, r.RecordedAt, r.ToolName, r.InputHash, r.OutputHash, r.Status, r.Error,
		r.GitCommitSHA, r.PackageVersion, r.RandomSeed, r.PrevRecordHash, r.RecordHash, inputBytes)
	return err
}

func (p *PostgresStorage) Latest(ctx context.Context) (DecisionRecord, error) {
	var r DecisionRecord
	var raw []byte
	err := p.pool.QueryRow(ctx, `
		SELECT request_id, recorded_at, tool_name, input_hash, output_hash, status, error,
		       git_commit_sha, package_version, random_seed, prev_record_hash, record_hash, raw
		FROM audit_records
		ORDER BY id DESC
		LIMIT 1
	`).Scan(&r.RequestID, &r.RecordedAt, &r.ToolName, &r.InputHash, &r.OutputHash, &r.Status, &r.Error,
		&r.GitCommitSHA, &r.PackageVersion, &r.RandomSeed, &r.PrevRecordHash, &r.RecordHash, &raw)
	if err == pgx.ErrNoRows {
		return DecisionRecord{}, nil
	}
	if err != nil {
		return DecisionRecord{}, err
	}
	_ = json.Unmarshal(raw, &r.Input)
	return r, nil
}
```

- [ ] **Step 3: Add verifier**

Create `internal/audit/verifier.go`:

```go
package audit

import (
	"context"
	"fmt"
)

type Verifier struct {
	storage Storage
}

func NewVerifier(s Storage) *Verifier {
	return &Verifier{storage: s}
}

func (v *Verifier) VerifyChain(ctx context.Context) error {
	latest, err := v.storage.Latest(ctx)
	if err != nil {
		return err
	}
	if latest.RequestID == "" {
		return nil
	}
	expected, err := HashRecord(latest)
	if err != nil {
		return err
	}
	if expected != latest.RecordHash {
		return fmt.Errorf("latest record hash mismatch")
	}
	return nil
}
```

- [ ] **Step 4: Add unit tests using in-memory storage**

Create `internal/audit/audit_test.go`:

```go
package audit

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type memStorage struct {
	records []DecisionRecord
}

func (m *memStorage) Append(_ context.Context, r DecisionRecord) error {
	m.records = append(m.records, r)
	return nil
}

func (m *memStorage) Latest(_ context.Context) (DecisionRecord, error) {
	if len(m.records) == 0 {
		return DecisionRecord{}, nil
	}
	return m.records[len(m.records)-1], nil
}

func TestHashRecordIsStable(t *testing.T) {
	r := DecisionRecord{RequestID: uuid.NewString(), ToolName: "health", Status: "ok"}
	h1, err := HashRecord(r)
	require.NoError(t, err)
	h2, err := HashRecord(r)
	require.NoError(t, err)
	assert.Equal(t, h1, h2)
}

func TestVerifier(t *testing.T) {
	store := &memStorage{}
	v := NewVerifier(store)
	require.NoError(t, v.VerifyChain(context.Background()))
}
```

Run `go get github.com/google/uuid`.

- [ ] **Step 5: Commit**

```bash
git add internal/audit go.mod go.sum
git commit -m "feat(audit): add hash-chained audit model, Postgres storage, and verifier"
```

#### Task 2.4: Extend market data providers

- [ ] **Step 1: Extend provider interface**

Modify `internal/marketdata/provider.go`:

```go
type Provider interface {
	Name() string
	Fetch(ctx context.Context, ticker core.Ticker, interval core.BarInterval, rng core.DateRange) ([]core.OHLCV, error)
	FetchAsync(ctx context.Context, ticker core.Ticker, interval core.BarInterval, rng core.DateRange) (<-chan FetchResult, error)
	GetTickerInfo(ctx context.Context, ticker core.Ticker) (core.TickerInfo, error)
	GetFinancialRatios(ctx context.Context, ticker core.Ticker) (core.FinancialRatios, error)
	GetMetadata(ctx context.Context) (core.DataSetMetadata, error)
}

type FetchResult struct {
	Series []core.OHLCV
	Err    error
}
```

- [ ] **Step 2: Add metadata value objects**

Modify `internal/core/value_objects.go`:

```go
type TickerInfo struct {
	Symbol     string `json:"symbol"`
	Name       string `json:"name"`
	Sector     string `json:"sector"`
	Industry   string `json:"industry"`
	Employees  int64  `json:"employees"`
	City       string `json:"city"`
	Country    string `json:"country"`
	Website    string `json:"website"`
}

type FinancialRatios struct {
	Symbol         string          `json:"symbol"`
	ForwardPE      decimal.Decimal `json:"forward_pe"`
	TrailingPE     decimal.Decimal `json:"trailing_pe"`
	PriceToBook    decimal.Decimal `json:"price_to_book"`
	DebtToEquity   decimal.Decimal `json:"debt_to_equity"`
	ROE            decimal.Decimal `json:"roe"`
	ProfitMargins  decimal.Decimal `json:"profit_margins"`
	DividendYield  decimal.Decimal `json:"dividend_yield"`
	MarketCap      int64           `json:"market_cap"`
}

type DataSetMetadata struct {
	Provider         string    `json:"provider"`
	Adjusted         bool      `json:"adjusted"`
	SurvivorshipFree bool      `json:"survivorship_free"`
	PointInTime      bool      `json:"point_in_time"`
	Frequency        string    `json:"frequency"`
	Timezone         string    `json:"timezone"`
	RetrievedAt      time.Time `json:"retrieved_at"`
}
```

- [ ] **Step 3: Implement Yahoo provider**

Create `internal/marketdata/yahoo.go` with a real HTTP client (use `net/http`, parse Yahoo Finance chart API). Return `core.ErrProviderNotAvailable` for unimplemented endpoints like `GetFinancialRatios`.

- [ ] **Step 4: Implement Polygon provider**

Create `internal/marketdata/polygon.go` with REST client for `api.polygon.io`.

- [ ] **Step 5: Add factory**

Create `internal/marketdata/factory.go`:

```go
package marketdata

import (
	"fmt"
	"os"

	"github.com/kevincouton/standard-tools-go/internal/core"
)

func NewProvider(name string) (Provider, error) {
	switch name {
	case "synthetic":
		return &SyntheticProvider{}, nil
	case "yahoo":
		return &YahooProvider{}, nil
	case "polygon":
		return NewPolygonProvider(os.Getenv("SQT_POLYGON_API_KEY")), nil
	default:
		return nil, fmt.Errorf("%w: unknown provider %s", core.ErrProviderNotAvailable, name)
	}
}
```

- [ ] **Step 6: Commit**

```bash
git add internal/marketdata internal/core go.mod go.sum
git commit -m "feat(marketdata): extend provider interface and add Yahoo/Polygon implementations"
```

#### Task 2.5: CLI using cobra

- [ ] **Step 1: Add cobra**

```bash
go get github.com/spf13/cobra
```

- [ ] **Step 2: Implement audit commands**

Replace `cmd/cli/main.go` stub with `internal/cli/root.go` containing `sqt replay`, `verify`, `report`, `keygen`, `anchor` commands wired to audit storage.

- [ ] **Step 3: Commit**

```bash
git add cmd/cli internal/cli go.mod go.sum
git commit -m "feat(cli): add cobra-based sqt CLI with audit commands"
```

#### Task 2.6: A2A and MCP skeleton

- [ ] **Step 1: Add A2A Agent Card and Task endpoints**

Create `internal/api/a2a.go`:

```go
package api

import "net/http"

func a2aAgentCard(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"name":        "standard-tools-go",
		"description": "Quantitative finance toolkit agent",
		"version":     "0.1.0",
		"capabilities": map[string]any{
			"streaming": false,
			"pushNotifications": false,
		},
		"skills": []map[string]any{},
	})
}
```

Register `GET /a2a/agent.json` and `POST /a2a/tasks` in `internal/api/rest.go`.

- [ ] **Step 2: Add MCP capability declaration**

Create `internal/api/mcp.go`:

```go
package api

import "net/http"

func mcpCapabilities(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"protocolVersion": "2024-11-05",
		"capabilities":    map[string]any{"tools": map[string]any{}},
		"serverInfo":      map[string]any{"name": "standard-tools-go", "version": "0.1.0"},
	})
}
```

Register `GET /mcp/capabilities` and `POST /mcp/tools/call`.

- [ ] **Step 3: Commit**

```bash
git add internal/api go.mod go.sum
git commit -m "feat(api): add A2A and MCP endpoint skeletons"
```

#### Task 2.7: Dockerfiles and local CI

- [ ] **Step 1: Add classic Dockerfile**

Create `Dockerfile`:

```dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o server ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /app/server /server
EXPOSE 8080 50051
ENTRYPOINT ["/server"]
```

- [ ] **Step 2: Add static Dockerfile**

Create `Dockerfile.static`:

```dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o server ./cmd/server

FROM scratch
COPY --from=builder /app/server /server
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
EXPOSE 8080 50051
ENTRYPOINT ["/server"]
```

- [ ] **Step 3: Add act runner script**

Create `scripts/run-act-local.sh`:

```bash
#!/usr/bin/env bash
set -euo pipefail
act --container-backend podman --pull=false -W .github/workflows/ci.yml "$@"
```

```bash
chmod +x scripts/run-act-local.sh
```

- [ ] **Step 4: Update `.mise.toml`**

Uncomment image/image-native/act tasks and ensure podman/act versions are pinned.

- [ ] **Step 5: Update CI**

Add a Postgres service container to `.github/workflows/ci.yml` for integration tests.

- [ ] **Step 6: Commit**

```bash
git add Dockerfile Dockerfile.static scripts/run-act-local.sh .mise.toml .github/workflows/ci.yml
git commit -m "feat(ci): add classic and static Dockerfiles, enable act+podman local CI"
```

---

### Phase 3 — Indicators & Metrics

**Goal:** Port the 14 technical indicators and risk/return metrics.

**Files to create:**
- `internal/indicators/trend.go`, `momentum.go`, `volatility.go`, `volume.go`
- `internal/metrics/return.go`, `risk.go`, `diagnostics.go`, `volatility.go`
- Corresponding `_test.go` files with parity expectations.

**Key tasks:**
- Define `type Series struct { Dates []time.Time; Values []float64 }` in `internal/core/series.go`.
- Implement SMA, EMA, RSI, MACD, Bollinger Bands, ATR, ADX, Parabolic SAR, MFI, OBV, VWAP.
- Implement Sharpe, Sortino, max drawdown, Calmar, win rate, profit factor, trade expectancy, VaR, CVaR.
- Add `@validate_series` equivalent: a helper `validateLength(inputs ...Series) error`.
- Build a parity test harness that reads JSON fixtures generated from Python and asserts Go outputs match within tolerance.

---

### Phase 4 — Analysis

**Goal:** Port regression, cointegration, PCA, Hurst, correlation, options BSM.

**Files to create:**
- `internal/analysis/regression.go`
- `internal/analysis/cointegration.go`
- `internal/analysis/pca.go`
- `internal/analysis/hurst.go`
- `internal/analysis/options.go`
- `internal/analysis/correlation.go`

**Key tasks:**
- Use `gonum/mat` for OLS and PCA.
- Implement Engle-Granger cointegration with ADF-style critical values (or port Python logic).
- Implement DFA and R/S Hurst estimators.
- Implement Black-Scholes-Merton price + Greeks and Newton-Raphson implied vol.
- Provide parity fixtures.

---

### Phase 5 — Backtest Engine

**Goal:** Port the vectorized backtest engine, built-in strategies, grid search, walk-forward, pair-trade, signal-panel, portfolio simulation, robustness, Monte Carlo, stress test, liquidity.

**Files to create:**
- `internal/backtest/engine.go`
- `internal/backtest/strategy.go`
- `internal/backtest/strategies.go`
- `internal/backtest/grid.go`
- `internal/backtest/walkforward.go`
- `internal/backtest/pairs.go`
- `internal/backtest/panel.go`
- `internal/backtest/portfolio_engine.go`
- `internal/backtest/costs.go`
- `internal/backtest/constraints.go`
- `internal/backtest/sizing.go`
- `internal/backtest/robustness.go`
- `internal/backtest/monte_carlo.go`
- `internal/backtest/stress.go`
- `internal/backtest/liquidity.go`
- `internal/backtest/artifacts.go`

**Key tasks:**
- Define `BacktestInput`, `BacktestResult`, `Trade`, `SignalType` in `internal/agent/models.go`.
- Preserve signal lag semantics (signals shifted by one bar).
- Support long/short direction, position sizing (ATR-based, Kelly), costs, constraints.
- Run portfolio simulation with shared cash and rebalancing.
- Add `run_backtest_compact` with artifact URIs.

---

### Phase 6 — Portfolio & Screener

**Goal:** Port portfolio metrics and optimization; async stock screener.

**Files to create:**
- `internal/portfolio/portfolio.go`
- `internal/portfolio/optimize.go`
- `internal/screener/screener.go`

**Key tasks:**
- Portfolio metrics: returns, volatility, Sharpe, diversification ratio, risk attribution (MCR, PCA, factor).
- Optimization: mean-variance, risk parity, Black-Litterman.
- Screener: universe filter with fundamental/technical predicates; worker pool for fetching.

---

### Phase 7 — API Surfaces & Packaging

**Goal:** Expose all agent tools via REST, gRPC, A2A, and MCP; produce native/static images.

**Files to create/modify:**
- `internal/api/rest.go` — add routes for every tool category.
- `internal/api/grpc.go` — define proto service for tool dispatch.
- `internal/api/a2a.go`, `mcp.go` — full protocol handlers.
- `proto/standard_tools.proto`
- `cmd/server/main.go` — wire DB, audit, config.

**Key tasks:**
- Generate gRPC code from proto.
- Implement streaming where protocols require it.
- Add OpenAPI spec generation.
- Build and push both classic and static images in CI.
- Run full e2e tests against the container.

---

## 4. Verification Strategy

1. **Unit tests:** every package has table-driven tests with edge cases.
2. **Parity suite:** Python generates `tests/fixtures/parity/*.json`; Go tests load and assert within `1e-9` absolute tolerance (or documented tolerance where rounding differs).
3. **Integration tests:** gated by `//go:build integration`; require PostgreSQL and optionally Polygon.
4. **E2E tests:** spin up the server container and call REST/gRPC/A2A/MCP endpoints.
5. **CI:** `go test ./...`, `go vet ./...`, `gofmt -l .`, container builds, `act --container-backend podman` smoke test.
6. **Audit integrity:** run `sqt verify` after a workload and assert chain is unbroken.

---

## 5. Risks & Open Questions

1. **C++ extension:** Python's `_sqt_core` accelerates Hurst/backtest kernels. Options:
   - Rewrite hot paths in Go (preferred for portability).
   - Keep C++ and bind via CGO (adds build complexity).
2. **Bloomberg `blpapi`:** no official Go SDK. Defer until requested; stub with `ErrProviderNotAvailable`.
3. **Pandas semantics:** some behaviors (NaN propagation, groupby, resample) need careful replication. Use parity fixtures aggressively.
4. **A2A/MCP specs:** both protocols are evolving; pin to stable versions and version the endpoints.
5. **Database migrations in tests:** use `testcontainers` or a single ephemeral Postgres per CI run. For local dev, document `docker run postgres`.

---

## 6. Spec Coverage Check

| Standard-Tools Capability | Target Phase |
|---------------------------|--------------|
| Extended data providers (Yahoo, Polygon, Bloomberg stub) | Phase 2 |
| Ticker info & financial ratios | Phase 2 |
| Hash-chained audit + CLI | Phase 2 |
| REST/gRPC/A2A/MCP skeleton | Phase 2 |
| Docker classic + static images | Phase 2 |
| 14 technical indicators | Phase 3 |
| Risk/return/diagnostics metrics | Phase 3 |
| Regression, cointegration, PCA, Hurst, options | Phase 4 |
| Backtest engine + strategies | Phase 5 |
| Grid search, walk-forward, pairs, panel, portfolio simulation | Phase 5 |
| Robustness, Monte Carlo, stress, liquidity | Phase 5 |
| Portfolio optimization | Phase 6 |
| Stock screener | Phase 6 |
| Full A2A/MCP tool exposure | Phase 7 |
| OpenAPI + native image CI | Phase 7 |
