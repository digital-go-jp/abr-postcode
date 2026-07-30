package middleware

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// runWithAccessLog serves one request through the middleware and returns the
// record it wrote.
func runWithAccessLog(t *testing.T, handler gin.HandlerFunc) map[string]any {
	t.Helper()

	gin.SetMode(gin.TestMode)

	var logged bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&logged, nil)))
	t.Cleanup(func() { slog.SetDefault(prev) })

	r := gin.New()
	r.Use(AccessLog())
	r.GET("/things/:id", handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/things/42", nil)
	r.ServeHTTP(w, req)

	var record map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(logged.Bytes()), &record); err != nil {
		t.Fatalf("access log is not JSON: %v\noutput: %s", err, logged.String())
	}
	return record
}

func TestAccessLogRecordsTheRequest(t *testing.T) {
	record := runWithAccessLog(t, func(c *gin.Context) {
		c.Status(http.StatusTeapot)
	})

	want := map[string]any{
		"msg":    "request",
		"method": http.MethodGet,
		"path":   "/things/42",
		"status": float64(http.StatusTeapot),
	}
	for key, wantValue := range want {
		if record[key] != wantValue {
			t.Errorf("%s = %v, want %v", key, record[key], wantValue)
		}
	}

	if _, ok := record["duration_ms"]; !ok {
		t.Errorf("access log is missing duration_ms: %v", record)
	}
	if _, ok := record["error"]; ok {
		t.Errorf("access log reports an error for a clean request: %v", record)
	}
}

// TestAccessLogRecordsHandlerErrors covers errors a handler attaches to the
// context, which would otherwise leave no trace once the response is written.
func TestAccessLogRecordsHandlerErrors(t *testing.T) {
	record := runWithAccessLog(t, func(c *gin.Context) {
		_ = c.Error(errors.New("upstream refused"))
		c.Status(http.StatusBadGateway)
	})

	got, ok := record["error"].(string)
	if !ok {
		t.Fatalf("access log is missing error: %v", record)
	}
	if !bytes.Contains([]byte(got), []byte("upstream refused")) {
		t.Errorf("error = %q, want it to mention the handler error", got)
	}
}
