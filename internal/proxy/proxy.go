// Package proxy forwards MCP JSON-RPC messages to pREST's HTTP MCP endpoint.
// It is intentionally dumb: it does not reinterpret tool schemas or results.
package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/prest/prest-mcp-adapter/internal/transport"
)

const (
	jsonRPCVersion = "2.0"
	errParse       = -32700
	errServer      = -32000
)

// Forwarder posts raw MCP payloads to pREST.
type Forwarder interface {
	Post(ctx context.Context, body []byte) (transport.Response, error)
}

// Writer writes a single protocol message (typically to stdout).
type Writer interface {
	WriteMessage(msg []byte) error
}

// Proxy bridges stdio MCP messages to the HTTP MCP endpoint.
type Proxy struct {
	forwarder Forwarder
	log       *slog.Logger
}

// New creates a Proxy that uses forwarder for HTTP and log for diagnostics.
func New(forwarder Forwarder, log *slog.Logger) *Proxy {
	if log == nil {
		log = slog.Default()
	}
	return &Proxy{forwarder: forwarder, log: log}
}

type envelope struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Method  string          `json:"method"`
}

// Handle processes one inbound NDJSON line and optionally writes a response.
func (p *Proxy) Handle(ctx context.Context, raw []byte, out Writer) error {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return nil
	}

	var env envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		p.log.Warn("invalid JSON-RPC on stdin", "error", err)
		return out.WriteMessage(mustMarshal(jsonRPCError(nil, errParse, "parse error")))
	}

	isNotification := env.ID == nil || string(env.ID) == "null"

	resp, err := p.forwarder.Post(ctx, raw)
	if err != nil {
		p.log.Error("pREST MCP request failed", "error", err, "method", env.Method)
		if isNotification {
			return nil
		}
		return out.WriteMessage(mustMarshal(jsonRPCError(env.ID, errServer, fmt.Sprintf("pREST unreachable: %v", err))))
	}

	if isNotification {
		return nil
	}

	body := bytes.TrimSpace(resp.Body)
	if len(body) == 0 {
		msg := fmt.Sprintf("empty response from pREST (HTTP %d)", resp.StatusCode)
		p.log.Error(msg)
		return out.WriteMessage(mustMarshal(jsonRPCError(env.ID, errServer, msg)))
	}
	if !json.Valid(body) || body[0] != '{' {
		msg := fmt.Sprintf("non-JSON response from pREST (HTTP %d)", resp.StatusCode)
		p.log.Error(msg, "body_prefix", truncate(body, 200))
		return out.WriteMessage(mustMarshal(jsonRPCError(env.ID, errServer, msg)))
	}

	// Pass through pREST's JSON-RPC body unchanged (including HTTP 4xx error payloads).
	return out.WriteMessage(body)
}

type rpcErrorBody struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type rpcErrorResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Error   rpcErrorBody    `json:"error"`
}

func jsonRPCError(id json.RawMessage, code int, message string) rpcErrorResponse {
	if id == nil {
		id = json.RawMessage("null")
	}
	return rpcErrorResponse{
		JSONRPC: jsonRPCVersion,
		ID:      id,
		Error: rpcErrorBody{
			Code:    code,
			Message: message,
		},
	}
}

func mustMarshal(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		// Should be unreachable for our fixed error shapes.
		return []byte(`{"jsonrpc":"2.0","id":null,"error":{"code":-32603,"message":"internal marshal error"}}`)
	}
	return b
}

func truncate(b []byte, n int) string {
	if len(b) <= n {
		return string(b)
	}
	return string(b[:n]) + "…"
}
