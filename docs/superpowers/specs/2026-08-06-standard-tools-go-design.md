# Standard-Tools Go Port — Design

## Goal

Port the Standard-Tools quantitative finance toolkit to Go, mirroring the Rust port's scope: 42+ agent-callable tools exposed via REST, gRPC, A2A, and MCP, with hash-chained audit records and PostgreSQL persistence.

## Approach

Idiomatic Go monolith:
- One Go module (`github.com/kevincouton/standard-tools-go`) with internal domain packages.
- `cmd/server` for the HTTP/gRPC binary and `cmd/cli` for the CLI.
- No workspaces or micro-modules — keep it simple, testable, and reviewable.

## Repository Layout

```
standard-tools-go/
├── cmd/
│   ├── server/main.go
│   └── cli/main.go
├── internal/
│   ├── api/          # REST, gRPC, A2A, MCP wiring
│   ├── core/         # errors, decimal, OHLCV, ticker, date-range
│   ├── marketdata/   # provider interface, Yahoo Finance adapter, cache
│   ├── indicators/   # technical indicators
│   ├── metrics/      # risk and return metrics
│   ├── analysis/     # regression, cointegration, Hurst, PCA, options
│   ├── backtest/     # strategy backtesting
│   ├── portfolio/    # mean-variance, risk-parity, Black-Litterman
│   ├── screener/     # fundamental + indicator screening
│   ├── agent/        # tool registry and dispatcher
│   ├── audit/        # hash-chained records + storage backends
│   └── orders/       # order domain and persistence
├── proto/            # shared .proto files
├── scripts/          # local CI helpers
├── .github/workflows/ci.yml
├── .mise.toml
├── Dockerfile
├── Dockerfile.native
├── docker-compose.yml
├── go.mod
├── go.sum
└── README.md
```

## Technology Stack

- **Go:** 1.23+
- **REST:** `github.com/go-chi/chi/v5`
- **gRPC:** `google.golang.org/grpc` + `google.golang.org/protobuf`
- **PostgreSQL:** `github.com/jackc/pgx/v5`
- **Math:** `gonum.org/v1/gonum/...`
- **Decimal:** `github.com/shopspring/decimal`
- **Config:** `github.com/knadh/koanf`
- **Logging:** `log/slog`
- **Testing:** `testify`, `net/http/httptest`, `google.golang.org/grpc/test/bufconn`

## API Surface

### REST
- `GET /health`
- `GET /api/v1/agent/tools`
- `POST /api/v1/agent/dispatch`
- `GET /api/v1/market-data/:ticker`
- `POST /api/v1/indicators/:indicator`
- `GET /api/v1/metrics/:metric`
- `POST /api/v1/analysis/:method`
- `POST /api/v1/backtest/:strategy`
- `POST /api/v1/portfolio/mean-variance`
- `POST /api/v1/portfolio/risk-parity`
- `POST /api/v1/portfolio/black-litterman`
- `POST /api/v1/screen`
- `POST /api/v1/audit/verify`
- `POST /api/v1/orders` / `GET /api/v1/orders`
- `GET|POST|DELETE /api/v1/orders/:id`

### gRPC
- `standard_tools.health.Health/Check`
- `standard_tools.agent.Agent/ListTools`
- `standard_tools.agent.Agent/Dispatch`

### A2A
- `POST /a2a/tasks/send`
- `POST /a2a/tasks/get`
- `POST /a2a/tasks/cancel`

### MCP
- `POST /mcp/tools/list`
- `POST /mcp/tools/call`

## Testing Strategy

- **Unit tests:** per package, in-memory dependencies.
- **Integration tests:** domain packages wired with test doubles.
- **E2E tests:** `internal/api/e2e_test.go` using `httptest` and `bufconn`, with a synthetic market-data provider and no external network calls.

## Native / Docker

- `Dockerfile`: `golang:1.23-bookworm` builder + `debian:bookworm-slim` runtime.
- `Dockerfile.native`: `CGO_ENABLED=0` static binary on `scratch`/`gcr.io/distroless/static`.
- `docker-compose.yml`: PostgreSQL + service.
- `.mise.toml` tasks: `build`, `test`, `test-integration`, `fmt`, `vet`, `lint`, `image`, `image-native`, `act`.

## Success Criteria

- `go test ./...` passes.
- `go vet ./...` passes.
- Static analysis (golangci-lint) passes.
- CI runs tests, lint, and builds both Docker images.
- E2E tests cover `/api/v1/agent/tools`, `/api/v1/agent/dispatch`, `/a2a/tasks/send`, `/mcp/tools/call`, and audit verification.
