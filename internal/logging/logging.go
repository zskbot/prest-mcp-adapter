// Package logging configures slog to write only to stderr.
// MCP clients treat stdout as the protocol channel; logs must never go there.
package logging

import (
	"log/slog"
	"os"
)

// New returns a JSON slog logger that writes exclusively to stderr.
func New() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}
