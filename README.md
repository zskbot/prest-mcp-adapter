# pREST MCP Adapter

Official stdio adapter for [pREST](https://github.com/prest/prest)’s HTTP MCP endpoint.

This adapter lets MCP clients that require stdio connect to a running pREST instance exposing an HTTP MCP endpoint at `/_mcp`.

```text
MCP client / AI IDE
        ↓ stdio (NDJSON JSON-RPC)
prest-mcp
        ↓ HTTP POST
http://localhost:3000/_mcp
        ↓
pREST
```

The adapter is a **tiny transport bridge**. It does not introspect schemas, generate SQL, or implement MCP tools. Those live in pREST.

## Requirements

- A running pREST instance with the `/_mcp` route available
- Go 1.26+ (to build from source)

## Environment variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `PREST_MCP_URL` | yes | — | URL of the pREST HTTP MCP endpoint, e.g. `http://localhost:3000/_mcp` |
| `PREST_MCP_TOKEN` | no | — | Bearer token when pREST auth is enabled |
| `PREST_MCP_TIMEOUT_MS` | no | `30000` | HTTP timeout in milliseconds |

## Install

### From source

```bash
go install github.com/prest/prest-mcp-adapter/cmd/prest-mcp@latest
```

### Docker / OCI

```bash
docker pull ghcr.io/prest/prest-mcp-adapter:0.1.0
```

## Local usage

```bash
PREST_MCP_URL=http://localhost:3000/_mcp prest-mcp
```

### With token

```bash
PREST_MCP_URL=https://api.example.com/_mcp \
PREST_MCP_TOKEN=secret \
prest-mcp
```

### Docker usage

```bash
docker run --rm -i \
  -e PREST_MCP_URL=http://host.docker.internal:3000/_mcp \
  ghcr.io/prest/prest-mcp-adapter:0.1.0
```

## MCP client configuration

### Cursor

Add to your MCP settings (e.g. `.cursor/mcp.json`):

```json
{
  "mcpServers": {
    "prest": {
      "command": "prest-mcp",
      "env": {
        "PREST_MCP_URL": "http://localhost:3000/_mcp"
      }
    }
  }
}
```

### Claude Desktop

Add to `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "prest": {
      "command": "prest-mcp",
      "env": {
        "PREST_MCP_URL": "http://localhost:3000/_mcp"
      }
    }
  }
}
```

### Generic MCP client

Any client that can launch a stdio MCP server:

```bash
PREST_MCP_URL=http://localhost:3000/_mcp prest-mcp
```

Protocol messages are newline-delimited JSON-RPC on stdin/stdout. Logs go only to stderr.

## Security recommendation

Use a read-only PostgreSQL user for AI/MCP workflows unless write access is explicitly required.

Do not expose unauthenticated MCP endpoints publicly.

Prefer local or private-network pREST MCP endpoints.

When pREST auth is enabled, set `PREST_MCP_TOKEN` to a valid JWT.

## Development

```bash
make test       # go test -race ./... and overall coverage >= 80%
make coverage   # coverage gate only
make vet
make build      # writes bin/prest-mcp
```

## License

MIT — see [LICENSE](LICENSE).
