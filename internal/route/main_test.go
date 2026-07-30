package route

import (
	"encoding/json"
	"os"
	"testing"

	"abr-postcode/internal/model"
	"abr-postcode/internal/service"

	"github.com/gin-gonic/gin"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

// wantErrorBody builds a body check asserting the error envelope carries message.
func wantErrorBody(message string) func(*testing.T, []byte) {
	return func(t *testing.T, body []byte) {
		t.Helper()
		var got map[string]string
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}
		if got["error"] != message {
			t.Errorf("expected error %q, got %q", message, got["error"])
		}
	}
}

// newTestAddressData builds a small registry for the handler tests. Every row
// is one that appears in the ABR data, except where a comment says otherwise:
// those rows cannot occur there, and exist only so the guards against them
// stay covered.
func newTestAddressData() *service.AddressData {
	return service.BuildIndexes(
		[]model.City{
			{LgCode: "011061", Pref: "北海道", County: "", City: "札幌市", Ward: "南区"},
			// A town in a county. A city carries either a county or a ward,
			// never both.
			{LgCode: "013030", Pref: "北海道", County: "石狩郡", City: "当別町", Ward: ""},
			{LgCode: "261025", Pref: "京都府", County: "", City: "京都市", Ward: "上京区"},
			{LgCode: "342131", Pref: "広島県", County: "", City: "廿日市市", Ward: ""},
			{LgCode: "131016", Pref: "東京都", County: "", City: "千代田区", Ward: ""},
		},
		[]model.Town{
			// Carries two postal codes, which /machiaza returns sorted.
			{LgCode: "011061", MachiazaID: "0056000", OazaCho: "藤野"},
			{LgCode: "013030", MachiazaID: "0001000", OazaCho: "字青山"},
			// Reached by a Kyoto street name.
			{LgCode: "261025", MachiazaID: "0087101", KyotoSt: "上御霊鳥居前通鞍馬口下る", OazaCho: "上御霊竪町"},
			// Two towns sharing one postal code, in a city that has no ward.
			{LgCode: "342131", MachiazaID: "0065001", OazaCho: "宮島口", Chome: "一丁目"},
			{LgCode: "342131", MachiazaID: "0065002", OazaCho: "宮島口", Chome: "二丁目"},
			// Not in the ABR data: every town there carries at least one postal
			// code. Covers the empty post_codes response.
			{LgCode: "131016", MachiazaID: "9999998", OazaCho: "郵便番号のない町"},
			// Not in the ABR data: LgCode 999998 has no City, so the join fails.
			{LgCode: "999998", MachiazaID: "0001001", OazaCho: "存在しない市の町"},
		},
		[]model.PostCodeMapping{
			{PostCode: "0050840", LgCode: "011061", MachiazaID: "0056000"},
			{PostCode: "0612271", LgCode: "011061", MachiazaID: "0056000"},
			{PostCode: "0610201", LgCode: "013030", MachiazaID: "0001000"},
			{PostCode: "6020896", LgCode: "261025", MachiazaID: "0087101"},
			{PostCode: "7390411", LgCode: "342131", MachiazaID: "0065001"},
			{PostCode: "7390411", LgCode: "342131", MachiazaID: "0065002"},
			// Not in the ABR data: names a town that does not exist, so
			// /post_code/9990000 finds the mapping but joins nothing.
			{PostCode: "9990000", LgCode: "011061", MachiazaID: "9999999"},
		},
	)
}
