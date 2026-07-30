package data

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Where the ABR postcode dataset lives. Keeping the endpoints together lets
// the download URL be built from the item id rather than restating it.
const (
	// ABRPostCodeItemID identifies the ABR postcode dataset, both in the DCAT
	// feed and in the download URL.
	ABRPostCodeItemID = "1098a4f61dab4f52997839d04245132a"

	// DefaultDCATFeedURL is the feed that reports the dataset's last change.
	DefaultDCATFeedURL = "https://dataset.address-br.digital.go.jp/api/feed/dcat-us/1.1.json"

	// DefaultDataURL is where the dataset itself is distributed.
	DefaultDataURL = "https://www.arcgis.com/sharing/rest/content/items/" + ABRPostCodeItemID + "/data"
)

// FetchRemoteModified queries the DCAT feed for the ABR postcode dataset's
// last-modified timestamp.
func FetchRemoteModified(feedURL string) (string, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, feedURL, nil)
	if err != nil {
		return "", fmt.Errorf("fetch DCAT feed: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch DCAT feed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("DCAT feed: unexpected status %d", resp.StatusCode)
	}

	var feed struct {
		Dataset []struct {
			Identifier string `json:"identifier"`
			Modified   string `json:"modified"`
		} `json:"dataset"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&feed); err != nil {
		return "", fmt.Errorf("decode DCAT feed: %w", err)
	}

	for _, ds := range feed.Dataset {
		if strings.Contains(ds.Identifier, ABRPostCodeItemID) {
			return ds.Modified, nil
		}
	}
	return "", fmt.Errorf("dataset %s not found in DCAT feed", ABRPostCodeItemID)
}

// GetLocalModified reads the locally cached ABR data modification timestamp.
// An absent file reports an empty timestamp, which the next check reads as an
// update being available. A file that exists but cannot be read reports the
// same, so it is logged to keep it distinguishable.
func GetLocalModified(dataDir string) string {
	path := filepath.Join(dataDir, "data_modified.txt")

	b, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			slog.Warn("Failed to read data_modified.txt", "path", path, "error", err)
		}
		return ""
	}
	return strings.TrimSpace(string(b))
}

// SaveDataModified records the data's last-modified timestamp in data_modified.txt.
func SaveDataModified(dataDir, modified string) error {
	return os.WriteFile(filepath.Join(dataDir, "data_modified.txt"), []byte(modified+"\n"), 0o644)
}
