package route

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRegisterHealthRoutes(t *testing.T) {
	router := gin.New()
	RegisterHealthRoutes(router, "0.0.9", "2026-06-22T07:51:06.000Z")

	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(t.Context(), "GET", "/health", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	if got := w.Header().Get("X-App-Version"); got != "0.0.9" {
		t.Errorf("X-App-Version = %q, want 0.0.9", got)
	}
	if got := w.Header().Get("X-Data-Modified"); got != "2026-06-22T07:51:06.000Z" {
		t.Errorf("X-Data-Modified = %q, want 2026-06-22T07:51:06.000Z", got)
	}

	var result map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if result["status"] != "ok" {
		t.Errorf("expected status 'ok', got %v", result["status"])
	}
}
