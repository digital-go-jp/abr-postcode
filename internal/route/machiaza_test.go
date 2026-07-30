package route

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRegisterMachiazaRoutes(t *testing.T) {
	data := newTestAddressData()

	tests := []struct {
		name       string
		lgCode     string
		machiazaID string
		wantStatus int
		checkBody  func(t *testing.T, body []byte)
	}{
		{
			name:       "valid machiaza - 011061/0056000",
			lgCode:     "011061",
			machiazaID: "0056000",
			wantStatus: http.StatusOK,
			checkBody: func(t *testing.T, body []byte) {
				var result map[string]any
				if err := json.Unmarshal(body, &result); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if result["lg_code"] != "011061" {
					t.Errorf("expected lg_code 011061, got %v", result["lg_code"])
				}
				if result["machiaza_id"] != "0056000" {
					t.Errorf("expected machiaza_id 0056000, got %v", result["machiaza_id"])
				}
				if result["oaza_cho"] != "藤野" {
					t.Errorf("expected oaza_cho 藤野, got %v", result["oaza_cho"])
				}
				postCodes := result["post_codes"].([]any)
				if len(postCodes) != 2 || postCodes[0] != "0050840" || postCodes[1] != "0612271" {
					t.Errorf("expected post_codes [0050840, 0612271], got %v", postCodes)
				}
			},
		},
		{
			name:       "valid machiaza - 342131/0065001",
			lgCode:     "342131",
			machiazaID: "0065001",
			wantStatus: http.StatusOK,
			checkBody: func(t *testing.T, body []byte) {
				var result map[string]any
				if err := json.Unmarshal(body, &result); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if result["lg_code"] != "342131" {
					t.Errorf("expected lg_code 342131, got %v", result["lg_code"])
				}
				if result["pref"] != "広島県" {
					t.Errorf("expected pref 広島県, got %v", result["pref"])
				}
				if result["city"] != "廿日市市" {
					t.Errorf("expected city 廿日市市, got %v", result["city"])
				}
			},
		},
		{
			name:       "valid machiaza with kyoto_st - 261025/0087101",
			lgCode:     "261025",
			machiazaID: "0087101",
			wantStatus: http.StatusOK,
			checkBody: func(t *testing.T, body []byte) {
				var result map[string]any
				if err := json.Unmarshal(body, &result); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if result["lg_code"] != "261025" {
					t.Errorf("expected lg_code 261025, got %v", result["lg_code"])
				}
				if result["pref"] != "京都府" {
					t.Errorf("expected pref 京都府, got %v", result["pref"])
				}
				if result["city"] != "京都市" {
					t.Errorf("expected city 京都市, got %v", result["city"])
				}
				if result["ward"] != "上京区" {
					t.Errorf("expected ward 上京区, got %v", result["ward"])
				}
				if result["kyoto_st"] != "上御霊鳥居前通鞍馬口下る" {
					t.Errorf("expected kyoto_st 上御霊鳥居前通鞍馬口下る, got %v", result["kyoto_st"])
				}
				if result["oaza_cho"] != "上御霊竪町" {
					t.Errorf("expected oaza_cho 上御霊竪町, got %v", result["oaza_cho"])
				}
				postCodes := result["post_codes"].([]any)
				if len(postCodes) != 1 || postCodes[0] != "6020896" {
					t.Errorf("expected post_codes [6020896], got %v", postCodes)
				}
			},
		},
		{
			name:       "valid machiaza in a county",
			lgCode:     "013030",
			machiazaID: "0001000",
			wantStatus: http.StatusOK,
			checkBody: func(t *testing.T, body []byte) {
				var result map[string]any
				if err := json.Unmarshal(body, &result); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if result["county"] != "石狩郡" {
					t.Errorf("expected county 石狩郡, got %v", result["county"])
				}
				if result["city"] != "当別町" {
					t.Errorf("expected city 当別町, got %v", result["city"])
				}
				if result["ward"] != "" {
					t.Errorf("expected empty ward, got %v", result["ward"])
				}
			},
		},
		{
			name:       "invalid lg_code format - too short",
			lgCode:     "34213",
			machiazaID: "0069000",
			wantStatus: http.StatusBadRequest,
			checkBody:  wantErrorBody("Invalid lg_code format"),
		},
		{
			name:       "invalid machiaza_id format - too short",
			lgCode:     "342131",
			machiazaID: "006900",
			wantStatus: http.StatusBadRequest,
			checkBody:  wantErrorBody("Invalid machiaza_id format"),
		},
		{
			name:       "machiaza not found",
			lgCode:     "999999",
			machiazaID: "9999999",
			wantStatus: http.StatusNotFound,
			checkBody:  wantErrorBody("machiaza not found"),
		},
		{
			// The town exists but names a city that does not, so the join
			// fails. That is missing data, not a malformed request.
			name:       "machiaza whose city is unknown",
			lgCode:     "999998",
			machiazaID: "0001001",
			wantStatus: http.StatusNotFound,
			checkBody:  wantErrorBody("machiaza not found"),
		},
		{
			name:       "valid machiaza with no post codes returns empty array",
			lgCode:     "131016",
			machiazaID: "9999998",
			wantStatus: http.StatusOK,
			checkBody: func(t *testing.T, body []byte) {
				var result map[string]any
				if err := json.Unmarshal(body, &result); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				postCodes, ok := result["post_codes"].([]any)
				if !ok {
					t.Fatalf("post_codes should be an array, got %T: %v", result["post_codes"], result["post_codes"])
				}
				if len(postCodes) != 0 {
					t.Errorf("expected empty post_codes, got %v", postCodes)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			RegisterMachiazaRoutes(router, data)

			w := httptest.NewRecorder()
			req, _ := http.NewRequestWithContext(t.Context(), "GET", "/machiaza/"+tt.lgCode+"/"+tt.machiazaID, nil)
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
			}

			tt.checkBody(t, w.Body.Bytes())
		})
	}
}
