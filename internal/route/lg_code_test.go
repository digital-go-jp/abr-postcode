package route

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"abr-postcode/internal/model"

	"github.com/gin-gonic/gin"
)

func TestRegisterLgCodeRoutes(t *testing.T) {
	data := newTestAddressData()

	tests := []struct {
		name       string
		lgCode     string
		wantStatus int
		wantBody   any
	}{
		{
			name:       "valid lg_code",
			lgCode:     "011061",
			wantStatus: http.StatusOK,
			wantBody: model.City{
				LgCode: "011061",
				Pref:   "北海道",
				County: "",
				City:   "札幌市",
				Ward:   "南区",
			},
		},
		{
			name:       "valid lg_code in a county",
			lgCode:     "013030",
			wantStatus: http.StatusOK,
			wantBody: model.City{
				LgCode: "013030",
				Pref:   "北海道",
				County: "石狩郡",
				City:   "当別町",
				Ward:   "",
			},
		},
		{
			name:       "lg_code not found",
			lgCode:     "999999",
			wantStatus: http.StatusNotFound,
			wantBody:   map[string]string{"error": "lg_code not found"},
		},
		{
			name:       "invalid lg_code format - too short",
			lgCode:     "01106",
			wantStatus: http.StatusBadRequest,
			wantBody:   map[string]string{"error": "Invalid lg_code format"},
		},
		{
			name:       "invalid lg_code format - too long",
			lgCode:     "0110611",
			wantStatus: http.StatusBadRequest,
			wantBody:   map[string]string{"error": "Invalid lg_code format"},
		},
		{
			name:       "invalid lg_code format - contains letters",
			lgCode:     "01106a",
			wantStatus: http.StatusBadRequest,
			wantBody:   map[string]string{"error": "Invalid lg_code format"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			RegisterLgCodeRoutes(router, data)

			w := httptest.NewRecorder()
			req, _ := http.NewRequestWithContext(t.Context(), "GET", "/lg_code/"+tt.lgCode, nil)
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
			}

			var got any
			if tt.wantStatus == http.StatusOK {
				var city model.City
				if err := json.Unmarshal(w.Body.Bytes(), &city); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				got = city
			} else {
				var errResp map[string]string
				if err := json.Unmarshal(w.Body.Bytes(), &errResp); err != nil {
					t.Fatalf("failed to unmarshal error response: %v", err)
				}
				got = errResp
			}

			wantJSON, _ := json.Marshal(tt.wantBody)
			gotJSON, _ := json.Marshal(got)
			if string(wantJSON) != string(gotJSON) {
				t.Errorf("expected body %s, got %s", wantJSON, gotJSON)
			}
		})
	}
}
