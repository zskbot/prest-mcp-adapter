// Command prest-mcp is the official stdio adapter for pREST's HTTP MCP endpoint.
//
// It reads MCP JSON-RPC messages from stdin, forwards them to PREST_MCP_URL,
// and writes responses to stdout. Logs go only to stderr.
package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/prest/prest-mcp-adapter/internal/config"
	"github.com/prest/prest-mcp-adapter/internal/logging"
	"github.com/prest/prest-mcp-adapter/internal/proxy"
	"github.com/prest/prest-mcp-adapter/internal/transport"
)

func main() {
	os.Exit(runMain(os.Stdin, os.Stdout, os.Stderr))
}

func runMain(in io.Reader, out, errOut io.Writer) int {
	log := logging.New()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(errOut, "prest-mcp: %v\n", err)
		return 1
	}

	client := &http.Client{Timeout: cfg.Timeout}
	forwarder := transport.NewHTTPForwarder(client, cfg.URL, cfg.Token)
	p := proxy.New(forwarder, log)
	stdio := transport.NewStdio(in, out)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Info("prest-mcp started", "url", cfg.URL, "timeout", cfg.Timeout.String())

	if err := run(ctx, p, stdio); err != nil && err != io.EOF && err != context.Canceled {
		fmt.Fprintf(errOut, "prest-mcp: %v\n", err)
		return 1
	}
	return 0
}

func run(ctx context.Context, p *proxy.Proxy, stdio *transport.Stdio) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		msg, err := stdio.ReadMessage()
		if err != nil {
			return err
		}
		if err := p.Handle(ctx, msg, stdio); err != nil {
			return err
		}
	}
}
