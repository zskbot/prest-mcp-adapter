package transport

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
)

// HTTPForwarder POSTs MCP JSON-RPC payloads to the pREST HTTP MCP endpoint.
type HTTPForwarder struct {
	client *http.Client
	url    string
	token  string
}

// NewHTTPForwarder creates a forwarder that posts to url using client.
// If token is non-empty, Authorization: Bearer <token> is set on each request.
func NewHTTPForwarder(client *http.Client, url, token string) *HTTPForwarder {
	return &HTTPForwarder{
		client: client,
		url:    url,
		token:  token,
	}
}

// Response is the raw HTTP result from pREST.
type Response struct {
	StatusCode int
	Body       []byte
}

// Post sends body to the configured MCP URL and returns the raw response.
func (f *HTTPForwarder) Post(ctx context.Context, body []byte) (Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, f.url, bytes.NewReader(body))
	if err != nil {
		return Response{}, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if f.token != "" {
		req.Header.Set("Authorization", "Bearer "+f.token)
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return Response{}, fmt.Errorf("POST %s: %w", f.url, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{}, fmt.Errorf("read response from %s: %w", f.url, err)
	}

	return Response{
		StatusCode: resp.StatusCode,
		Body:       respBody,
	}, nil
}
