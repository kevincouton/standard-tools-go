# standard-tools-go

Go port of the Standard-Tools quantitative finance toolkit.

## Stack

- Go 1.25+
- Chi (REST)
- grpc-go (gRPC)
- pgx (PostgreSQL) — planned for Phase 2
- Gonum (math) — planned for Phase 2

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
