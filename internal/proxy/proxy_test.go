package proxy_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prest/prest-mcp-adapter/internal/proxy"
	"github.com/prest/prest-mcp-adapter/internal/transport"
)

type bufWriter struct {
	buf bytes.Buffer
}

func (w *bufWriter) WriteMessage(msg []byte) error {
	w.buf.Write(msg)
	if len(msg) == 0 || msg[len(msg)-1] != '\n' {
		w.buf.WriteByte('\n')
	}
	return nil
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

type stubForwarder struct {
	resp transport.Response
	err  error
	n    int
}

func (s *stubForwarder) Post(context.Context, []byte) (transport.Response, error) {
	s.n++
	return s.resp, s.err
}

func TestProxy_initializePassthrough(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if !bytes.Contains(body, []byte(`"initialize"`)) {
			t.Errorf("unexpected body: %s", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"serverInfo":{"name":"prest"}}}`))
	}))
	t.Cleanup(srv.Close)

	p := proxy.New(transport.NewHTTPForwarder(srv.Client(), srv.URL, ""), discardLogger())
	out := &bufWriter{}
	err := p.Handle(context.Background(), []byte(`{"jsonrpc":"2.0","id":1,"method":"initialize"}`), out)
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if !strings.Contains(out.buf.String(), `"serverInfo"`) {
		t.Fatalf("response = %s", out.buf.String())
	}
}

func TestProxy_toolsListPassthrough(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":2,"result":{"tools":[{"name":"prest.list_databases"}]}}`))
	}))
	t.Cleanup(srv.Close)

	p := proxy.New(transport.NewHTTPForwarder(srv.Client(), srv.URL, ""), discardLogger())
	out := &bufWriter{}
	err := p.Handle(context.Background(), []byte(`{"jsonrpc":"2.0","id":2,"method":"tools/list"}`), out)
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if !strings.Contains(out.buf.String(), `prest.list_databases`) {
		t.Fatalf("response = %s", out.buf.String())
	}
}

func TestProxy_httpErrorBodyPassthrough(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"error":{"code":401,"message":"unauthorized"}}`))
	}))
	t.Cleanup(srv.Close)

	p := proxy.New(transport.NewHTTPForwarder(srv.Client(), srv.URL, ""), discardLogger())
	out := &bufWriter{}
	err := p.Handle(context.Background(), []byte(`{"jsonrpc":"2.0","id":1,"method":"initialize"}`), out)
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if !strings.Contains(out.buf.String(), `"unauthorized"`) {
		t.Fatalf("response = %s", out.buf.String())
	}
	if !strings.Contains(out.buf.String(), `"code":401`) {
		t.Fatalf("expected passthrough error code, got %s", out.buf.String())
	}
}

func TestProxy_unreachableSynthesizesError(t *testing.T) {
	t.Parallel()

	f := transport.NewHTTPForwarder(&http.Client{}, "http://127.0.0.1:1", "")
	p := proxy.New(f, discardLogger())
	out := &bufWriter{}
	err := p.Handle(context.Background(), []byte(`{"jsonrpc":"2.0","id":9,"method":"initialize"}`), out)
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}

	var resp struct {
		Error *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
		ID json.RawMessage `json:"id"`
	}
	if err := json.Unmarshal(bytes.TrimSpace(out.buf.Bytes()), &resp); err != nil {
		t.Fatalf("unmarshal: %v; body=%s", err, out.buf.String())
	}
	if resp.Error == nil || resp.Error.Code != -32000 {
		t.Fatalf("expected -32000 error, got %+v", resp.Error)
	}
	if string(resp.ID) != "9" {
		t.Fatalf("id = %s", resp.ID)
	}
	if !strings.Contains(resp.Error.Message, "pREST unreachable") {
		t.Fatalf("message = %q", resp.Error.Message)
	}
}

func TestProxy_notificationNoStdout(t *testing.T) {
	t.Parallel()

	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(srv.Close)

	p := proxy.New(transport.NewHTTPForwarder(srv.Client(), srv.URL, ""), discardLogger())
	out := &bufWriter{}
	err := p.Handle(context.Background(), []byte(`{"jsonrpc":"2.0","method":"notifications/initialized"}`), out)
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if !called {
		t.Fatal("expected HTTP forward for notification")
	}
	if out.buf.Len() != 0 {
		t.Fatalf("expected no stdout for notification, got %q", out.buf.String())
	}
}

func TestProxy_nullIDTreatedAsNotification(t *testing.T) {
	t.Parallel()

	f := &stubForwarder{resp: transport.Response{StatusCode: 200, Body: []byte(`{"ok":true}`)}}
	p := proxy.New(f, discardLogger())
	out := &bufWriter{}
	err := p.Handle(context.Background(), []byte(`{"jsonrpc":"2.0","id":null,"method":"notifications/initialized"}`), out)
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if f.n != 1 {
		t.Fatalf("Post calls = %d", f.n)
	}
	if out.buf.Len() != 0 {
		t.Fatalf("expected no stdout, got %q", out.buf.String())
	}
}

func TestProxy_notificationForwardErrorSilent(t *testing.T) {
	t.Parallel()

	f := &stubForwarder{err: errors.New("boom")}
	p := proxy.New(f, discardLogger())
	out := &bufWriter{}
	err := p.Handle(context.Background(), []byte(`{"jsonrpc":"2.0","method":"notifications/cancelled"}`), out)
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if out.buf.Len() != 0 {
		t.Fatalf("expected no stdout, got %q", out.buf.String())
	}
}

func TestProxy_emptyLineIgnored(t *testing.T) {
	t.Parallel()

	f := &stubForwarder{}
	p := proxy.New(f, nil) // also covers New with nil logger
	out := &bufWriter{}
	if err := p.Handle(context.Background(), []byte("   \n"), out); err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if f.n != 0 {
		t.Fatal("empty line should not POST")
	}
	if out.buf.Len() != 0 {
		t.Fatalf("expected no stdout, got %q", out.buf.String())
	}
}

func TestProxy_emptyHTTPBody(t *testing.T) {
	t.Parallel()

	f := &stubForwarder{resp: transport.Response{StatusCode: 502, Body: []byte("  ")}}
	p := proxy.New(f, discardLogger())
	out := &bufWriter{}
	err := p.Handle(context.Background(), []byte(`{"jsonrpc":"2.0","id":4,"method":"initialize"}`), out)
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if !strings.Contains(out.buf.String(), `empty response`) {
		t.Fatalf("response = %s", out.buf.String())
	}
}

func TestProxy_invalidJSON(t *testing.T) {
	t.Parallel()

	p := proxy.New(&stubForwarder{err: errors.New("should not be called")}, discardLogger())
	out := &bufWriter{}
	err := p.Handle(context.Background(), []byte(`{not-json`), out)
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if !strings.Contains(out.buf.String(), `"code":-32700`) {
		t.Fatalf("expected parse error, got %s", out.buf.String())
	}
}

func TestProxy_nonJSONHTTPBody(t *testing.T) {
	t.Parallel()

	long := bytes.Repeat([]byte("x"), 250)
	f := &stubForwarder{resp: transport.Response{StatusCode: 502, Body: long}}
	p := proxy.New(f, discardLogger())
	out := &bufWriter{}
	err := p.Handle(context.Background(), []byte(`{"jsonrpc":"2.0","id":3,"method":"initialize"}`), out)
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if !strings.Contains(out.buf.String(), `"code":-32000`) {
		t.Fatalf("expected server error, got %s", out.buf.String())
	}
	if !strings.Contains(out.buf.String(), `non-JSON`) {
		t.Fatalf("expected non-JSON message, got %s", out.buf.String())
	}
}

func TestProxy_jsonArrayRejected(t *testing.T) {
	t.Parallel()

	f := &stubForwarder{resp: transport.Response{StatusCode: 200, Body: []byte(`[{"jsonrpc":"2.0"}]`)}}
	p := proxy.New(f, discardLogger())
	out := &bufWriter{}
	err := p.Handle(context.Background(), []byte(`{"jsonrpc":"2.0","id":5,"method":"initialize"}`), out)
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if !strings.Contains(out.buf.String(), `non-JSON`) {
		t.Fatalf("response = %s", out.buf.String())
	}
}
