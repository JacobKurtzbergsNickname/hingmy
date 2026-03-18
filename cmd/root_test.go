package cmd

import (
	"strings"
	"testing"
)

func TestPadLine(t *testing.T) {
	tests := []struct {
		input string
		width int
	}{
		{"hello", 10},
		{"", 5},
		{"exactly", 7},
	}
	for _, tc := range tests {
		got := padLine(tc.input, tc.width)
		if len(got) != tc.width {
			t.Errorf("padLine(%q, %d): len = %d, want %d", tc.input, tc.width, len(got), tc.width)
		}
		if !strings.HasPrefix(got, tc.input) {
			t.Errorf("padLine(%q, %d) = %q: should start with original string", tc.input, tc.width, got)
		}
	}
}

func TestPadMessage_ShortMessage(t *testing.T) {
	msg := "hello"
	width := 20
	got := padMessage(msg, width)
	if len(got) != width {
		t.Errorf("padMessage(%q, %d): len = %d, want %d", msg, width, len(got), width)
	}
	if !strings.HasPrefix(got, msg) {
		t.Errorf("padMessage(%q, %d) = %q: should start with original message", msg, width, got)
	}
}

func TestPadMessage_ExactLength(t *testing.T) {
	msg := "exactly20characters!"
	got := padMessage(msg, len(msg))
	if got != msg {
		t.Errorf("padMessage exact length: got %q, want %q", got, msg)
	}
}

func TestPadMessage_LongMessage(t *testing.T) {
	msg := "This is a long message that should be wrapped across multiple lines for testing"
	width := 20
	got := padMessage(msg, width)
	lines := strings.Split(got, "\n")
	if len(lines) < 2 {
		t.Errorf("padMessage long: expected multiple lines, got %d", len(lines))
	}
	for i, line := range lines {
		if len(line) != width {
			t.Errorf("padMessage long: line %d has len %d, want %d (line: %q)", i, len(line), width, line)
		}
	}
}
