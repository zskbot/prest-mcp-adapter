# prest-mcp-adapter — agent / Copilot instructions

## What this repo is

`prest-mcp` is a **stdio → HTTP transport bridge** for [pREST](https://github.com/prest/prest)’s `/_mcp` endpoint.

It is **not** a second MCP server. Do not add schema introspection, SQL generation, tool definitions, or pREST business logic here.

```text
MCP client  --stdio NDJSON-->  prest-mcp  --HTTP POST-->  pREST /_mcp
```

## Layout

| Path | Role |
|------|------|
| `cmd/prest-mcp/` | Entrypoint: load config, wire HTTP client, run stdio loop |
| `internal/config/` | Env: `PREST_MCP_URL`, `PREST_MCP_TOKEN`, `PREST_MCP_TIMEOUT_MS` |
| `internal/transport/` | NDJSON stdio + HTTP forwarder |
| `internal/proxy/` | Dumb request/response bridging + JSON-RPC error synthesis |
| `internal/logging/` | `slog` → **stderr only** |

## Hard rules

1. **stdout = protocol only** (newline-delimited JSON-RPC). Never write logs to stdout.
2. **stderr = logs/diagnostics** via `slog`.
3. **Fail closed** on missing/invalid `PREST_MCP_URL`.
4. **Pass through** pREST JSON response bodies unchanged when they are JSON objects (including HTTP 4xx JSON-RPC errors).
5. **Synthesize** JSON-RPC errors only for parse failures, network errors, empty/non-JSON HTTP bodies.
6. **Notifications** (no `id`): forward to HTTP; write nothing to stdout.
7. Go version: match `go.mod` (currently **1.26**). Format with `gofmt`. Wrap errors with `%w`.

## Validate

```sh
go build ./...
go vet ./...
gofmt -l .
make test       # unit tests + overall coverage >= 80%
make coverage   # COVERAGE_MIN=80 scripts/check-coverage.sh
```

Overall statement coverage must stay at **≥80%** (`make coverage`).

## Out of scope (v0.1+)

`prest-mcp doctor`, Homebrew, native `prestd mcp --stdio`, inventing MCP tool wrappers, Content-Length framing (use NDJSON).
