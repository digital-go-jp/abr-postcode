package logging

import (
	"log/slog"
	"testing"
)

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected slog.Level
	}{
		{
			input:    "debug",
			expected: slog.LevelDebug,
		},
		{
			input:    "DEBUG",
			expected: slog.LevelDebug,
		},
		{
			input:    " WARN ",
			expected: slog.LevelWarn,
		},
		{
			input:    "error",
			expected: slog.LevelError,
		},
		{
			input:    "INFO",
			expected: slog.LevelInfo,
		},
		{
			input:    "bogus",
			expected: slog.LevelInfo,
		},
		{
			input:    "",
			expected: slog.LevelInfo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseLevel(tt.input)
			if got != tt.expected {
				t.Errorf("parseLevel(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

// TestNewHandlerSelection pins which handler each LOG_FORMAT value maps to,
// including the empty string that config passes through when the variable is
// set but blank.
func TestNewHandlerSelection(t *testing.T) {
	tests := []struct {
		name       string
		format     string
		expectJSON bool
	}{
		{name: "config default", format: "json", expectJSON: true},
		{name: "explicitly empty", format: "", expectJSON: false},
		{name: "text", format: "text", expectJSON: false},
		{name: "unrecognised", format: "logfmt", expectJSON: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := New("INFO", tt.format)
			_, isJSON := logger.Handler().(*slog.JSONHandler)
			if isJSON != tt.expectJSON {
				t.Errorf("New(%q) handler = %T, expectJSON = %v", tt.format, logger.Handler(), tt.expectJSON)
			}
		})
	}
}

func TestNewHandler(t *testing.T) {
	tests := []struct {
		name       string
		format     string
		expectJSON bool
	}{
		{
			name:       "json format",
			format:     "json",
			expectJSON: true,
		},
		{
			name:       "JSON with spaces and uppercase",
			format:     " JSON ",
			expectJSON: true,
		},
		{
			name:       "text format",
			format:     "text",
			expectJSON: false,
		},
		{
			name:       "empty format defaults to text",
			format:     "",
			expectJSON: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandler(slog.LevelInfo, tt.format)

			if tt.expectJSON {
				if _, ok := h.(*slog.JSONHandler); !ok {
					t.Errorf("newHandler(%q) returned %T, want *slog.JSONHandler", tt.format, h)
				}
			} else {
				if _, ok := h.(*slog.TextHandler); !ok {
					t.Errorf("newHandler(%q) returned %T, want *slog.TextHandler", tt.format, h)
				}
			}
		})
	}
}
