package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestMetadata(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name              string
		version, modified string
		wantVersion       string
		wantModified      string
	}{
		{
			name:         "both set",
			version:      "1.2.3",
			modified:     "2026-04-24T09:18:35.000Z",
			wantVersion:  "1.2.3",
			wantModified: "2026-04-24T09:18:35.000Z",
		},
		{
			name:         "both empty headers are omitted",
			version:      "",
			modified:     "",
			wantVersion:  "",
			wantModified: "",
		},
		{
			name:         "version only",
			version:      "1.0.0",
			modified:     "",
			wantVersion:  "1.0.0",
			wantModified: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.Use(Metadata(tt.version, tt.modified))
			r.GET("/", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
			r.ServeHTTP(w, req)

			if got := w.Header().Get("X-App-Version"); got != tt.wantVersion {
				t.Errorf("X-App-Version = %q, want %q", got, tt.wantVersion)
			}
			if got := w.Header().Get("X-Data-Modified"); got != tt.wantModified {
				t.Errorf("X-Data-Modified = %q, want %q", got, tt.wantModified)
			}
		})
	}
}
