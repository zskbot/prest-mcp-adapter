package logging_test

import (
	"testing"

	"github.com/prest/prest-mcp-adapter/internal/logging"
)

func TestNew_returnsLogger(t *testing.T) {
	t.Parallel()

	log := logging.New()
	if log == nil {
		t.Fatal("expected non-nil logger")
	}
	// Must not panic; output goes to stderr only.
	log.Info("coverage")
}
