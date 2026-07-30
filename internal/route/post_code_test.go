package route

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRegisterPostcodeRoutes(t *testing.T) {
	data := newTestAddressData()

	tests := []struct {
		name       string
		postCode   string
		wantStatus int
		checkBody  func(t *testing.T, body []byte)
	}{
		{
			name:       "valid post_code - 0050840",
			postCode:   "0050840",
			wantStatus: http.StatusOK,
			checkBody: func(t *testing.T, body []byte) {
				var result []map[string]any
				if err := json.Unmarshal(body, &result); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if len(result) != 1 {
					t.Errorf("expected 1 result, got %d", len(result))
				}
				if result[0]["post_code"] != "0050840" {
					t.Errorf("expected post_code 0050840, got %v", result[0]["post_code"])
				}
			},
		},
		{
			name:       "valid post_code - 0612271",
			postCode:   "0612271",
			wantStatus: http.StatusOK,
			checkBody: func(t *testing.T, body []byte) {
				var result []map[string]any
				if err := json.Unmarshal(body, &result); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if len(result) != 1 {
					t.Errorf("expected 1 result, got %d", len(result))
				}
			},
		},
		{
			name:       "valid post_code - 7390411",
			postCode:   "7390411",
			wantStatus: http.StatusOK,
			checkBody: func(t *testing.T, body []byte) {
				var result []map[string]any
				if err := json.Unmarshal(body, &result); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if len(result) != 2 {
					t.Errorf("expected 2 results, got %d", len(result))
				}
				if result[0]["pref"] != "広島県" {
					t.Errorf("expected pref 広島県, got %v", result[0]["pref"])
				}
			},
		},
		{
			name:       "valid post_code in a county",
			postCode:   "0610201",
			wantStatus: http.StatusOK,
			checkBody: func(t *testing.T, body []byte) {
				var result []map[string]any
				if err := json.Unmarshal(body, &result); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if len(result) != 1 {
					t.Fatalf("expected 1 result, got %d", len(result))
				}
				if result[0]["county"] != "石狩郡" {
					t.Errorf("expected county 石狩郡, got %v", result[0]["county"])
				}
				if result[0]["ward"] != "" {
					t.Errorf("expected empty ward, got %v", result[0]["ward"])
				}
			},
		},
		{
			name:       "post_code not found",
			postCode:   "0000000",
			wantStatus: http.StatusNotFound,
			checkBody:  wantErrorBody("post_code not found"),
		},
		{
			// The mapping exists but names no town, so nothing joins. That
			// leaves no address to return, same as an unknown post_code.
			name:       "post_code maps only to an unknown town",
			postCode:   "9990000",
			wantStatus: http.StatusNotFound,
			checkBody:  wantErrorBody("post_code not found"),
		},
		{
			name:       "invalid post_code - too short",
			postCode:   "000000",
			wantStatus: http.StatusBadRequest,
			checkBody:  wantErrorBody("Invalid post_code format"),
		},
		{
			name:       "invalid post_code - too short 404",
			postCode:   "404",
			wantStatus: http.StatusBadRequest,
			checkBody:  wantErrorBody("Invalid post_code format"),
		},
		{
			name:       "invalid post_code - contains letters",
			postCode:   "a",
			wantStatus: http.StatusBadRequest,
			checkBody:  wantErrorBody("Invalid post_code format"),
		},
		{
			// Right length, but not all digits.
			name:       "invalid post_code - letter among digits",
			postCode:   "1234a67",
			wantStatus: http.StatusBadRequest,
			checkBody:  wantErrorBody("Invalid post_code format"),
		},
		{
			// Seven full-width digits must not pass as a seven-digit code.
			name:       "invalid post_code - full-width digits",
			postCode:   "１２３４５６７",
			wantStatus: http.StatusBadRequest,
			checkBody:  wantErrorBody("Invalid post_code format"),
		},
		{
			name:       "invalid post_code - contains Japanese",
			postCode:   "あ",
			wantStatus: http.StatusBadRequest,
			checkBody:  wantErrorBody("Invalid post_code format"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			RegisterPostcodeRoutes(router, data)

			w := httptest.NewRecorder()
			req, _ := http.NewRequestWithContext(t.Context(), "GET", "/post_code/"+tt.postCode, nil)
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
			}

			tt.checkBody(t, w.Body.Bytes())
		})
	}
}
