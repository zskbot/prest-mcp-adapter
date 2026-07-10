package transport

import (
	"bufio"
	"fmt"
	"io"
)

// MaxStdioMessageBytes is the maximum size of a single NDJSON line on stdin.
// Large tools/list or tools/call payloads from pREST need headroom beyond the
// default bufio.Scanner limit (64 KiB).
const MaxStdioMessageBytes = 10 * 1024 * 1024 // 10 MiB

// Stdio reads and writes newline-delimited JSON-RPC messages.
// Protocol messages go to out; callers must never write logs to out.
type Stdio struct {
	in  *bufio.Scanner
	out io.Writer
}

// NewStdio creates a stdio transport reading from in and writing to out.
func NewStdio(in io.Reader, out io.Writer) *Stdio {
	sc := bufio.NewScanner(in)
	sc.Buffer(make([]byte, 0, 64*1024), MaxStdioMessageBytes)
	return &Stdio{in: sc, out: out}
}

// ReadMessage reads the next newline-delimited message from stdin.
// Returns io.EOF when the input stream ends.
func (s *Stdio) ReadMessage() ([]byte, error) {
	if !s.in.Scan() {
		if err := s.in.Err(); err != nil {
			return nil, fmt.Errorf("read stdin: %w", err)
		}
		return nil, io.EOF
	}
	line := s.in.Bytes()
	// Copy because Scanner reuses its buffer.
	msg := make([]byte, len(line))
	copy(msg, line)
	return msg, nil
}

// WriteMessage writes a single JSON-RPC message followed by a newline.
func (s *Stdio) WriteMessage(msg []byte) error {
	if _, err := s.out.Write(msg); err != nil {
		return fmt.Errorf("write stdout: %w", err)
	}
	if len(msg) == 0 || msg[len(msg)-1] != '\n' {
		if _, err := s.out.Write([]byte{'\n'}); err != nil {
			return fmt.Errorf("write stdout newline: %w", err)
		}
	}
	return nil
}
