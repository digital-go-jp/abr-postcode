package logging

import (
	"log/slog"
	"os"
	"strings"
)

// New builds a slog.Logger writing to stderr.
//
//   - level: DEBUG|INFO|WARN|ERROR, case-insensitive; anything else is INFO
//   - format: "json" for JSON output, otherwise text
func New(level, format string) *slog.Logger {
	return slog.New(newHandler(parseLevel(level), format))
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// newHandler creates a slog.Handler according to format.
func newHandler(level slog.Leveler, format string) slog.Handler {
	opts := &slog.HandlerOptions{Level: level}
	if strings.ToLower(strings.TrimSpace(format)) == "json" {
		return slog.NewJSONHandler(os.Stderr, opts)
	}
	return slog.NewTextHandler(os.Stderr, opts)
}
