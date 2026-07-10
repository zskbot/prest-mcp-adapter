package main

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prest/prest-mcp-adapter/internal/proxy"
	"github.com/prest/prest-mcp-adapter/internal/transport"
)

func TestRunMain_missingURL(t *testing.T) {
	t.Setenv("PREST_MCP_URL", "")
	t.Setenv("PREST_MCP_TOKEN", "")
	t.Setenv("PREST_MCP_TIMEOUT_MS", "")

	var errOut bytes.Buffer
	code := runMain(strings.NewReader(""), io.Discard, &errOut)
	if code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	if !strings.Contains(errOut.String(), "PREST_MCP_URL is required") {
		t.Fatalf("stderr = %q", errOut.String())
	}
}

func TestRunMain_happyPathEOF(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{}}`))
	}))
	t.Cleanup(srv.Close)

	t.Setenv("PREST_MCP_URL", srv.URL)
	t.Setenv("PREST_MCP_TOKEN", "")
	t.Setenv("PREST_MCP_TIMEOUT_MS", "5000")

	in := strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize"}` + "\n")
	var out, errOut bytes.Buffer
	code := runMain(in, &out, &errOut)
	if code != 0 {
		t.Fatalf("exit = %d, stderr=%q", code, errOut.String())
	}
	if !strings.Contains(out.String(), `"result"`) {
		t.Fatalf("stdout = %q", out.String())
	}
}

func TestRun_contextCanceled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	p := proxy.New(&staticForwarder{}, nil)
	stdio := transport.NewStdio(strings.NewReader(""), io.Discard)
	err := run(ctx, p, stdio)
	if err != context.Canceled {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
}

func TestRun_handleError(t *testing.T) {
	t.Parallel()

	p := proxy.New(&staticForwarder{}, nil)
	stdio := transport.NewStdio(strings.NewReader("{not-json\n"), &errWriter{})
	err := run(context.Background(), p, stdio)
	if err == nil {
		t.Fatal("expected write error from Handle")
	}
}

func TestRun_eof(t *testing.T) {
	t.Parallel()

	p := proxy.New(&staticForwarder{}, nil)
	stdio := transport.NewStdio(strings.NewReader(""), io.Discard)
	err := run(context.Background(), p, stdio)
	if err != io.EOF {
		t.Fatalf("err = %v, want EOF", err)
	}
}

type staticForwarder struct{}

func (staticForwarder) Post(context.Context, []byte) (transport.Response, error) {
	return transport.Response{StatusCode: 200, Body: []byte(`{"jsonrpc":"2.0","id":1,"result":{}}`)}, nil
}

type errWriter struct{}

func (errWriter) Write([]byte) (int, error) {
	return 0, io.ErrClosedPipe
}

func TestRunMain_proxyWriteFailure(t *testing.T) {
	// Unreachable pREST still writes a JSON-RPC error to stdout; use a
	// writer that fails to exercise the non-EOF error path in runMain.
	t.Setenv("PREST_MCP_URL", "http://127.0.0.1:1")
	t.Setenv("PREST_MCP_TIMEOUT_MS", "100")

	in := strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize"}` + "\n")
	var errOut bytes.Buffer
	code := runMain(in, errWriter{}, &errOut)
	if code != 1 {
		t.Fatalf("exit = %d, want 1; stderr=%q", code, errOut.String())
	}
}
