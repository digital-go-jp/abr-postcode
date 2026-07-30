package config

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestLoadReadsABRPDataDir(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("ABRP_DATA_DIR", "/custom/dir/")
	if got := Load().DataDir; got != "/custom/dir/" {
		t.Errorf("DataDir = %q, want %q", got, "/custom/dir/")
	}
}

func TestLoadDataDirDefault(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("ABRP_DATA_DIR", "")
	if got := Load().DataDir; got != "data/" {
		t.Errorf("DataDir = %q, want default %q", got, "data/")
	}
}

func TestLoadReadsABRPFeedURL(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("ABRP_FEED_URL", "https://custom.example.com/feed.json")
	if got := Load().DCATFeedURL; got != "https://custom.example.com/feed.json" {
		t.Errorf("DCATFeedURL = %q, want %q", got, "https://custom.example.com/feed.json")
	}
}

func TestLoadABRPFeedURLDefault(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("ABRP_FEED_URL", "")
	if got := Load().DCATFeedURL; got != "https://dataset.address-br.digital.go.jp/api/feed/dcat-us/1.1.json" {
		t.Errorf("DCATFeedURL = %q, want default %q", got, "https://dataset.address-br.digital.go.jp/api/feed/dcat-us/1.1.json")
	}
}

func TestLoadReadsABRPDataURL(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("ABRP_DATA_URL", "https://custom.example.com/data.zip")
	if got := Load().DataURL; got != "https://custom.example.com/data.zip" {
		t.Errorf("DataURL = %q, want %q", got, "https://custom.example.com/data.zip")
	}
}

func TestLoadABRPDataURLDefault(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("ABRP_DATA_URL", "")
	if got := Load().DataURL; got != "https://www.arcgis.com/sharing/rest/content/items/1098a4f61dab4f52997839d04245132a/data" {
		t.Errorf("DataURL = %q, want default %q", got, "https://www.arcgis.com/sharing/rest/content/items/1098a4f61dab4f52997839d04245132a/data")
	}
}

func TestLoadReadsPort(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("PORT", "9000")
	if got := Load().Port; got != "9000" {
		t.Errorf("Port = %q, want %q", got, "9000")
	}
}

func TestLoadPortDefault(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("PORT", "")
	if got := Load().Port; got != "8080" {
		t.Errorf("Port = %q, want default %q", got, "8080")
	}
}

// TestLoadCORSAllowOrigins covers the comma-separated form, which lets more
// than one frontend be allowed, and the values that must not leave the list
// empty: the middleware rejects a configuration allowing no origin at all.
func TestLoadCORSAllowOrigins(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  []string
	}{
		{
			name:  "unset falls back to the default",
			value: "",
			want:  []string{"*"},
		},
		{
			name:  "one origin",
			value: "https://example.com",
			want:  []string{"https://example.com"},
		},
		{
			name:  "several origins",
			value: "https://a.example,https://b.example",
			want:  []string{"https://a.example", "https://b.example"},
		},
		{
			name:  "spaces and empty entries are ignored",
			value: " https://a.example , , https://b.example ",
			want:  []string{"https://a.example", "https://b.example"},
		},
		{
			name:  "a value that leaves nothing falls back to the default",
			value: " , ",
			want:  []string{"*"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Chdir(t.TempDir())
			t.Setenv("CORS_ALLOW_ORIGIN", tt.value)

			got := Load().CORSAllowOrigins
			if !slices.Equal(got, tt.want) {
				t.Errorf("CORSAllowOrigins = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLoadLogSettings(t *testing.T) {
	tests := []struct {
		name       string
		level      string
		format     string
		wantLevel  string
		wantFormat string
	}{
		{
			name:       "empty falls back to the defaults",
			wantLevel:  "INFO",
			wantFormat: "json",
		},
		{
			name:       "explicit values are kept",
			level:      "DEBUG",
			format:     "text",
			wantLevel:  "DEBUG",
			wantFormat: "text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Chdir(t.TempDir())
			t.Setenv("LOG_LEVEL", tt.level)
			t.Setenv("LOG_FORMAT", tt.format)

			cfg := Load()
			if cfg.LogLevel != tt.wantLevel {
				t.Errorf("LogLevel = %q, want %q", cfg.LogLevel, tt.wantLevel)
			}
			if cfg.LogFormat != tt.wantFormat {
				t.Errorf("LogFormat = %q, want %q", cfg.LogFormat, tt.wantFormat)
			}
		})
	}
}

func TestLoadDotEnvBasicFormat(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	envPath := filepath.Join(tmpDir, ".env")
	err := os.WriteFile(envPath, []byte("ABRP_TEST_BASIC=value123\n"), 0o644)
	if err != nil {
		t.Fatalf("failed to write .env: %v", err)
	}

	t.Setenv("ABRP_TEST_BASIC", "")
	LoadDotEnv()
	if got := os.Getenv("ABRP_TEST_BASIC"); got != "value123" {
		t.Errorf("ABRP_TEST_BASIC = %q, want %q", got, "value123")
	}
}

func TestLoadDotEnvDoubleQuotes(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	envPath := filepath.Join(tmpDir, ".env")
	err := os.WriteFile(envPath, []byte(`ABRP_TEST_QUOTED="quoted value"`+"\n"), 0o644)
	if err != nil {
		t.Fatalf("failed to write .env: %v", err)
	}

	t.Setenv("ABRP_TEST_QUOTED", "")
	LoadDotEnv()
	if got := os.Getenv("ABRP_TEST_QUOTED"); got != "quoted value" {
		t.Errorf("ABRP_TEST_QUOTED = %q, want %q", got, "quoted value")
	}
}

func TestLoadDotEnvSingleQuotes(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	envPath := filepath.Join(tmpDir, ".env")
	err := os.WriteFile(envPath, []byte("ABRP_TEST_SINGLE='single quoted'\n"), 0o644)
	if err != nil {
		t.Fatalf("failed to write .env: %v", err)
	}

	t.Setenv("ABRP_TEST_SINGLE", "")
	LoadDotEnv()
	if got := os.Getenv("ABRP_TEST_SINGLE"); got != "single quoted" {
		t.Errorf("ABRP_TEST_SINGLE = %q, want %q", got, "single quoted")
	}
}

func TestLoadDotEnvSkipsComments(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	envPath := filepath.Join(tmpDir, ".env")
	err := os.WriteFile(envPath, []byte("# this is a comment\nABRP_TEST_COMMENT=value\n# another comment\n"), 0o644)
	if err != nil {
		t.Fatalf("failed to write .env: %v", err)
	}

	t.Setenv("ABRP_TEST_COMMENT", "")
	LoadDotEnv()
	if got := os.Getenv("ABRP_TEST_COMMENT"); got != "value" {
		t.Errorf("ABRP_TEST_COMMENT = %q, want %q", got, "value")
	}
}

func TestLoadDotEnvSkipsEmptyLines(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	envPath := filepath.Join(tmpDir, ".env")
	err := os.WriteFile(envPath, []byte("\n\nABRP_TEST_EMPTY=value\n\n\n"), 0o644)
	if err != nil {
		t.Fatalf("failed to write .env: %v", err)
	}

	t.Setenv("ABRP_TEST_EMPTY", "")
	LoadDotEnv()
	if got := os.Getenv("ABRP_TEST_EMPTY"); got != "value" {
		t.Errorf("ABRP_TEST_EMPTY = %q, want %q", got, "value")
	}
}

func TestLoadDotEnvIgnoresInvalidLines(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	envPath := filepath.Join(tmpDir, ".env")
	err := os.WriteFile(envPath, []byte("no_equals_sign\nABRP_TEST_INVALID=value\nanother_bad_line\n"), 0o644)
	if err != nil {
		t.Fatalf("failed to write .env: %v", err)
	}

	t.Setenv("ABRP_TEST_INVALID", "")
	t.Setenv("no_equals_sign", "")
	t.Setenv("another_bad_line", "")
	LoadDotEnv()
	if got := os.Getenv("ABRP_TEST_INVALID"); got != "value" {
		t.Errorf("ABRP_TEST_INVALID = %q, want %q", got, "value")
	}
	if got := os.Getenv("no_equals_sign"); got != "" {
		t.Errorf("no_equals_sign = %q, want empty (invalid line should not be set)", got)
	}
	if got := os.Getenv("another_bad_line"); got != "" {
		t.Errorf("another_bad_line = %q, want empty (invalid line should not be set)", got)
	}
}

func TestLoadDotEnvValueContainsEquals(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	envPath := filepath.Join(tmpDir, ".env")
	err := os.WriteFile(envPath, []byte("ABRP_TEST_EQ=a=b\n"), 0o644)
	if err != nil {
		t.Fatalf("failed to write .env: %v", err)
	}

	t.Setenv("ABRP_TEST_EQ", "")
	LoadDotEnv()
	if got := os.Getenv("ABRP_TEST_EQ"); got != "a=b" {
		t.Errorf("ABRP_TEST_EQ = %q, want %q", got, "a=b")
	}
}

func TestLoadDotEnvDoesNotOverwriteExistingEnv(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	envPath := filepath.Join(tmpDir, ".env")
	err := os.WriteFile(envPath, []byte("ABRP_TEST_EXISTING=from_file\n"), 0o644)
	if err != nil {
		t.Fatalf("failed to write .env: %v", err)
	}

	t.Setenv("ABRP_TEST_EXISTING", "from_env")
	LoadDotEnv()
	if got := os.Getenv("ABRP_TEST_EXISTING"); got != "from_env" {
		t.Errorf("ABRP_TEST_EXISTING = %q, want %q (should not be overwritten)", got, "from_env")
	}
}

func TestLoadDotEnvMissingFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	// No .env file created, LoadDotEnv should not error
	LoadDotEnv()
	// If we get here without panicking, test passes
}

func TestLoadDotEnvWithWhitespace(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	envPath := filepath.Join(tmpDir, ".env")
	err := os.WriteFile(envPath, []byte("  ABRP_TEST_WHITESPACE  =  spaced value  \n"), 0o644)
	if err != nil {
		t.Fatalf("failed to write .env: %v", err)
	}

	t.Setenv("ABRP_TEST_WHITESPACE", "")
	LoadDotEnv()
	if got := os.Getenv("ABRP_TEST_WHITESPACE"); got != "spaced value" {
		t.Errorf("ABRP_TEST_WHITESPACE = %q, want %q", got, "spaced value")
	}
}
