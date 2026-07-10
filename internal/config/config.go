// Package config loads adapter settings from environment variables.
package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	envURL       = "PREST_MCP_URL"
	envToken     = "PREST_MCP_TOKEN"
	envTimeoutMS = "PREST_MCP_TIMEOUT_MS"

	defaultTimeoutMS = 30000
)

// Config holds runtime settings for the stdio→HTTP adapter.
type Config struct {
	URL     string
	Token   string
	Timeout time.Duration
}

// Load reads and validates configuration from the environment.
func Load() (Config, error) {
	rawURL := strings.TrimSpace(os.Getenv(envURL))
	if rawURL == "" {
		return Config{}, fmt.Errorf("%s is required", envURL)
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return Config{}, fmt.Errorf("%s must be a valid http:// or https:// URL: %w", envURL, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return Config{}, fmt.Errorf("%s must be a valid http:// or https:// URL", envURL)
	}
	if u.Host == "" {
		return Config{}, fmt.Errorf("%s must be a valid http:// or https:// URL", envURL)
	}

	timeoutMS := defaultTimeoutMS
	if raw := strings.TrimSpace(os.Getenv(envTimeoutMS)); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil {
			return Config{}, fmt.Errorf("%s must be a positive integer: %w", envTimeoutMS, err)
		}
		if n <= 0 {
			return Config{}, fmt.Errorf("%s must be a positive integer", envTimeoutMS)
		}
		timeoutMS = n
	}

	return Config{
		URL:     strings.TrimRight(rawURL, "/"),
		Token:   strings.TrimSpace(os.Getenv(envToken)),
		Timeout: time.Duration(timeoutMS) * time.Millisecond,
	}, nil
}
