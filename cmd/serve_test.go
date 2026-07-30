package cmd

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"abr-postcode/internal/service"

	"github.com/gin-gonic/gin"
)

func newTestRouter(t *testing.T, corsAllowOrigins ...string) *gin.Engine {
	t.Helper()

	if len(corsAllowOrigins) == 0 {
		corsAllowOrigins = []string{"*"}
	}
	gin.SetMode(gin.TestMode)
	return newRouterWith(corsAllowOrigins)
}

func newRouterWith(corsAllowOrigins []string) *gin.Engine {
	return newRouter(
		service.BuildIndexes(nil, nil, nil),
		corsAllowOrigins,
		"1.2.3",
		"2026-07-01T05:20:43.000Z",
	)
}

// TestNewRouterKeepsRecovery guards the middleware that turns a panicking
// handler into a 500 instead of taking the process down, and checks that such
// a request still reaches the access log.
func TestNewRouterKeepsRecovery(t *testing.T) {
	var logged bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&logged, nil)))
	t.Cleanup(func() { slog.SetDefault(prev) })

	r := newTestRouter(t)
	r.GET("/panic", func(*gin.Context) {
		panic("boom")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/panic", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}

	var record map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(logged.Bytes()), &record); err != nil {
		t.Fatalf("panicking request was not logged: %v\noutput: %s", err, logged.String())
	}
	if record["path"] != "/panic" {
		t.Errorf("logged path = %v, want /panic", record["path"])
	}
	if record["status"] != float64(http.StatusInternalServerError) {
		t.Errorf("logged status = %v, want %d", record["status"], http.StatusInternalServerError)
	}
}

func TestNewRouterAppliesCORS(t *testing.T) {
	r := newTestRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/lg_code/131016", nil)
	// httptest defaults the host to example.com, so the origin has to differ
	// from it for the request to count as cross-origin.
	req.Header.Set("Origin", "https://caller.test")
	r.ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, "*")
	}
	if got := w.Header().Get("Access-Control-Expose-Headers"); got != "Content-Length" {
		t.Errorf("Access-Control-Expose-Headers = %q, want %q", got, "Content-Length")
	}
}

// TestNewRouterTreatsNoOriginsAsEveryOrigin covers a caller that passes no
// origins at all. The middleware rejects a configuration allowing none, so
// this must not reach it.
func TestNewRouterTreatsNoOriginsAsEveryOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := newRouterWith(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/lg_code/131016", nil)
	req.Header.Set("Origin", "https://caller.test")
	r.ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, "*")
	}
}

// TestNewRouterAllowsSeveralOrigins covers the comma-separated configuration:
// each listed origin is echoed back, and one that is not listed is refused.
func TestNewRouterAllowsSeveralOrigins(t *testing.T) {
	r := newTestRouter(t, "https://a.example", "https://b.example")

	tests := []struct {
		origin  string
		allowed bool
	}{
		{origin: "https://a.example", allowed: true},
		{origin: "https://b.example", allowed: true},
		{origin: "https://c.example", allowed: false},
	}

	for _, tt := range tests {
		t.Run(tt.origin, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/lg_code/131016", nil)
			req.Header.Set("Origin", tt.origin)
			r.ServeHTTP(w, req)

			got := w.Header().Get("Access-Control-Allow-Origin")
			if tt.allowed {
				if got != tt.origin {
					t.Errorf("Access-Control-Allow-Origin = %q, want the request origin %q", got, tt.origin)
				}
				if vary := w.Header().Get("Vary"); vary != "Origin" {
					t.Errorf("Vary = %q, want Origin so caches do not mix the two", vary)
				}
				return
			}
			if got != "" {
				t.Errorf("Access-Control-Allow-Origin = %q, want it absent for an origin that is not listed", got)
			}
		})
	}
}

func TestNewRouterCompressesWhenAsked(t *testing.T) {
	r := newTestRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/health", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	r.ServeHTTP(w, req)

	if got := w.Header().Get("Content-Encoding"); got != "gzip" {
		t.Errorf("Content-Encoding = %q, want %q", got, "gzip")
	}
}

// TestNewRouterScopesMetadataHeaders pins the metadata headers to /health.
func TestNewRouterScopesMetadataHeaders(t *testing.T) {
	r := newTestRouter(t)

	tests := []struct {
		path        string
		wantHeaders bool
	}{
		{path: "/health", wantHeaders: true},
		{path: "/lg_code/131016", wantHeaders: false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, tt.path, nil)
			r.ServeHTTP(w, req)

			gotVersion := w.Header().Get("X-App-Version")
			if tt.wantHeaders && gotVersion != "1.2.3" {
				t.Errorf("X-App-Version = %q, want %q", gotVersion, "1.2.3")
			}
			if !tt.wantHeaders && gotVersion != "" {
				t.Errorf("X-App-Version = %q, want it absent", gotVersion)
			}
		})
	}
}
