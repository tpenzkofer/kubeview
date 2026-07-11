package ui

import "testing"

func TestLogSeverity(t *testing.T) {
	cases := []struct {
		line string
		want severity
	}{
		{`{"level":"error","msg":"db down"}`, sevError},
		{"panic: runtime error", sevError},
		{"connection failed after 3 tries", sevError},
		{"level=warn deprecated flag used", sevWarn},
		{`{"level":"warn"}`, sevWarn},
		{"GET /healthz 200 ok", sevNone},
		{"2026-07-11 listening on :8080", sevNone}, // must not trip on the digits
		{"user 404 not found", sevNone},            // no longer a false warn
	}
	for _, c := range cases {
		if got := logSeverity(c.line); got != c.want {
			t.Errorf("logSeverity(%q) = %d, want %d", c.line, got, c.want)
		}
	}
}
