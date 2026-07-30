package config

import (
	"bufio"
	"cmp"
	"log/slog"
	"os"
	"strings"

	"abr-postcode/internal/data"
)

const (
	DefaultDataDir         = "data/"
	DefaultPort            = "8080"
	DefaultCORSAllowOrigin = "*"
	DefaultLogLevel        = "INFO"
	DefaultLogFormat       = "json"
)

type Config struct {
	DCATFeedURL      string
	DataURL          string
	DataDir          string
	Port             string
	CORSAllowOrigins []string
	LogLevel         string
	LogFormat        string
}

// Load reads the configuration from environment variables. Applying a .env
// file is the caller's responsibility; see LoadDotEnv.
func Load() *Config {
	return &Config{
		DCATFeedURL:      cmp.Or(os.Getenv("ABRP_FEED_URL"), data.DefaultDCATFeedURL),
		DataURL:          cmp.Or(os.Getenv("ABRP_DATA_URL"), data.DefaultDataURL),
		DataDir:          cmp.Or(os.Getenv("ABRP_DATA_DIR"), DefaultDataDir),
		Port:             cmp.Or(os.Getenv("PORT"), DefaultPort),
		CORSAllowOrigins: splitOrigins(cmp.Or(os.Getenv("CORS_ALLOW_ORIGIN"), DefaultCORSAllowOrigin)),
		LogLevel:         cmp.Or(os.Getenv("LOG_LEVEL"), DefaultLogLevel),
		LogFormat:        cmp.Or(os.Getenv("LOG_FORMAT"), DefaultLogFormat),
	}
}

// splitOrigins turns the configured value into the origins the CORS middleware
// matches against. Commas separate them, so more than one frontend can be
// allowed; surrounding spaces and empty entries are ignored. A value that
// leaves nothing behind falls back to the default rather than disabling every
// origin, which the middleware rejects outright.
func splitOrigins(value string) []string {
	var origins []string
	for origin := range strings.SplitSeq(value, ",") {
		if origin = strings.TrimSpace(origin); origin != "" {
			origins = append(origins, origin)
		}
	}
	if len(origins) == 0 {
		return []string{DefaultCORSAllowOrigin}
	}
	return origins
}

// LoadDotEnv reads a .env file and sets environment variables.
func LoadDotEnv() {
	file, err := os.Open(".env")
	if err != nil {
		return // .env file not required
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
			(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
			value = value[1 : len(value)-1]
		}

		// Only set if not already set by environment variable
		if os.Getenv(key) == "" {
			if err := os.Setenv(key, value); err != nil {
				slog.Warn("failed to set env", "key", key, "error", err)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		slog.Warn("failed to read .env", "error", err)
	}
}
