package config_test

import (
	"testing"
	"time"

	"github.com/prest/prest-mcp-adapter/internal/config"
)

func TestLoad_requiresURL(t *testing.T) {
	t.Setenv("PREST_MCP_URL", "")
	t.Setenv("PREST_MCP_TOKEN", "")
	t.Setenv("PREST_MCP_TIMEOUT_MS", "")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when PREST_MCP_URL is missing")
	}
}

func TestLoad_rejectsBadScheme(t *testing.T) {
	t.Setenv("PREST_MCP_URL", "ftp://localhost:3000/_mcp")
	t.Setenv("PREST_MCP_TIMEOUT_MS", "")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for non-http(s) scheme")
	}
}

func TestLoad_rejectsMissingHost(t *testing.T) {
	t.Setenv("PREST_MCP_URL", "http:///path")
	t.Setenv("PREST_MCP_TIMEOUT_MS", "")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for URL without host")
	}
}

func TestLoad_defaultsAndNormalize(t *testing.T) {
	t.Setenv("PREST_MCP_URL", "http://localhost:3000/_mcp/")
	t.Setenv("PREST_MCP_TOKEN", " secret ")
	t.Setenv("PREST_MCP_TIMEOUT_MS", "")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.URL != "http://localhost:3000/_mcp" {
		t.Fatalf("URL = %q, want trimmed trailing slash", cfg.URL)
	}
	if cfg.Token != "secret" {
		t.Fatalf("Token = %q, want trimmed", cfg.Token)
	}
	if cfg.Timeout != 30*time.Second {
		t.Fatalf("Timeout = %v, want 30s", cfg.Timeout)
	}
}

func TestLoad_customTimeout(t *testing.T) {
	t.Setenv("PREST_MCP_URL", "https://api.example.com/_mcp")
	t.Setenv("PREST_MCP_TIMEOUT_MS", "5000")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Timeout != 5*time.Second {
		t.Fatalf("Timeout = %v, want 5s", cfg.Timeout)
	}
}

func TestLoad_rejectsNonPositiveTimeout(t *testing.T) {
	t.Setenv("PREST_MCP_URL", "http://localhost:3000/_mcp")
	t.Setenv("PREST_MCP_TIMEOUT_MS", "0")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for non-positive timeout")
	}
}

func TestLoad_rejectsNonNumericTimeout(t *testing.T) {
	t.Setenv("PREST_MCP_URL", "http://localhost:3000/_mcp")
	t.Setenv("PREST_MCP_TIMEOUT_MS", "abc")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for non-numeric timeout")
	}
}

func TestLoad_rejectsInvalidURL(t *testing.T) {
	t.Setenv("PREST_MCP_URL", "http://[::1")
	t.Setenv("PREST_MCP_TIMEOUT_MS", "")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
}
