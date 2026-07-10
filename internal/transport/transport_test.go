package transport_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prest/prest-mcp-adapter/internal/transport"
)

func TestStdio_roundTrip(t *testing.T) {
	t.Parallel()

	in := strings.NewReader("{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"initialize\"}\n")
	var out bytes.Buffer
	s := transport.NewStdio(in, &out)

	msg, err := s.ReadMessage()
	if err != nil {
		t.Fatalf("ReadMessage: %v", err)
	}
	if !bytes.Contains(msg, []byte(`"initialize"`)) {
		t.Fatalf("unexpected message: %s", msg)
	}

	if err := s.WriteMessage([]byte(`{"jsonrpc":"2.0","id":1,"result":{}}`)); err != nil {
		t.Fatalf("WriteMessage: %v", err)
	}
	got := out.String()
	if !strings.HasSuffix(got, "\n") {
		t.Fatalf("expected trailing newline, got %q", got)
	}
	if strings.Count(got, "\n") != 1 {
		t.Fatalf("expected single line, got %q", got)
	}

	_, err = s.ReadMessage()
	if err != io.EOF {
		t.Fatalf("expected EOF, got %v", err)
	}
}

func TestStdio_writeDoesNotDuplicateNewline(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	s := transport.NewStdio(strings.NewReader(""), &out)
	if err := s.WriteMessage([]byte("{\"ok\":true}\n")); err != nil {
		t.Fatalf("WriteMessage: %v", err)
	}
	if out.String() != "{\"ok\":true}\n" {
		t.Fatalf("got %q", out.String())
	}
}

func TestStdio_writeError(t *testing.T) {
	t.Parallel()

	s := transport.NewStdio(strings.NewReader(""), errWriter{})
	if err := s.WriteMessage([]byte(`{"ok":true}`)); err == nil {
		t.Fatal("expected write error")
	}
}

func TestStdio_readError(t *testing.T) {
	t.Parallel()

	s := transport.NewStdio(errReader{}, io.Discard)
	_, err := s.ReadMessage()
	if err == nil {
		t.Fatal("expected read error")
	}
}

type errWriter struct{}

func (errWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}

type errReader struct{}

func (errReader) Read([]byte) (int, error) {
	return 0, errors.New("read failed")
}

func TestHTTPForwarder_postsWithHeaders(t *testing.T) {
	t.Parallel()

	var gotAuth, gotCT, gotAccept string
	var gotBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotCT = r.Header.Get("Content-Type")
		gotAccept = r.Header.Get("Accept")
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{}}`))
	}))
	t.Cleanup(srv.Close)

	f := transport.NewHTTPForwarder(srv.Client(), srv.URL, "tok")
	resp, err := f.Post(context.Background(), []byte(`{"jsonrpc":"2.0","id":1,"method":"initialize"}`))
	if err != nil {
		t.Fatalf("Post: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	if gotAuth != "Bearer tok" {
		t.Fatalf("Authorization = %q", gotAuth)
	}
	if gotCT != "application/json" {
		t.Fatalf("Content-Type = %q", gotCT)
	}
	if gotAccept != "application/json" {
		t.Fatalf("Accept = %q", gotAccept)
	}
	if !bytes.Contains(gotBody, []byte(`"initialize"`)) {
		t.Fatalf("body = %s", gotBody)
	}
	if !bytes.Contains(resp.Body, []byte(`"result"`)) {
		t.Fatalf("response body = %s", resp.Body)
	}
}

func TestHTTPForwarder_noTokenOmitsAuth(t *testing.T) {
	t.Parallel()

	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(srv.Close)

	f := transport.NewHTTPForwarder(srv.Client(), srv.URL, "")
	if _, err := f.Post(context.Background(), []byte(`{}`)); err != nil {
		t.Fatalf("Post: %v", err)
	}
	if gotAuth != "" {
		t.Fatalf("Authorization = %q, want empty", gotAuth)
	}
}

func TestHTTPForwarder_networkError(t *testing.T) {
	t.Parallel()

	f := transport.NewHTTPForwarder(&http.Client{}, "http://127.0.0.1:1", "")
	_, err := f.Post(context.Background(), []byte(`{}`))
	if err == nil {
		t.Fatal("expected network error")
	}
}

func TestHTTPForwarder_invalidURL(t *testing.T) {
	t.Parallel()

	f := transport.NewHTTPForwarder(&http.Client{}, "://bad", "")
	_, err := f.Post(context.Background(), []byte(`{}`))
	if err == nil {
		t.Fatal("expected create request error")
	}
}
