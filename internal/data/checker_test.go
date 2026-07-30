package data

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSaveDataModifiedWritesValue(t *testing.T) {
	dir := t.TempDir()
	const ts = "2026-06-17T07:38:19.000Z"

	if err := SaveDataModified(dir, ts); err != nil {
		t.Fatalf("SaveDataModified: %v", err)
	}

	if got := GetLocalModified(dir); got != ts {
		t.Errorf("GetLocalModified = %q, want %q", got, ts)
	}
}

// TestGetLocalModifiedMissingFile covers the not-yet-imported case, which
// reports an empty timestamp so the next check sees an update as available.
func TestGetLocalModifiedMissingFile(t *testing.T) {
	if got := GetLocalModified(t.TempDir()); got != "" {
		t.Errorf("GetLocalModified = %q, want empty for a missing file", got)
	}
}

// TestGetLocalModifiedUnreadableFile covers a timestamp that exists but cannot
// be read. It reports the same empty value as an absent one, so the caller
// still sees an update as available.
func TestGetLocalModifiedUnreadableFile(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "data_modified.txt"), 0o755); err != nil {
		t.Fatalf("create blocking directory: %v", err)
	}

	if got := GetLocalModified(dir); got != "" {
		t.Errorf("GetLocalModified = %q, want empty for an unreadable file", got)
	}
}

func TestFetchRemoteModified(t *testing.T) {
	tests := []struct {
		name         string
		serverFunc   func(http.ResponseWriter, *http.Request)
		closeServer  bool
		wantModified string
		wantErr      string
	}{
		{
			name: "success",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = fmt.Fprintf(w, `{
					"dataset": [
						{
							"identifier": "urn:uuid:1098a4f61dab4f52997839d04245132a",
							"modified": "2026-06-17T07:38:19.000Z"
						}
					]
				}`)
			},
			wantModified: "2026-06-17T07:38:19.000Z",
		},
		{
			name: "non-200 status",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			wantErr: "unexpected status",
		},
		{
			name: "invalid json",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				_, _ = fmt.Fprint(w, "not json")
			},
			wantErr: "decode",
		},
		{
			name: "identifier not found",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = fmt.Fprintf(w, `{
					"dataset": [
						{
							"identifier": "urn:uuid:different-id",
							"modified": "2026-06-17T07:38:19.000Z"
						}
					]
				}`)
			},
			wantErr: "not found",
		},
		{
			name:        "connection failed",
			closeServer: true,
			wantErr:     "fetch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var server *httptest.Server
			if tt.closeServer {
				server = httptest.NewServer(http.HandlerFunc(tt.serverFunc))
				server.Close()
			} else {
				server = httptest.NewServer(http.HandlerFunc(tt.serverFunc))
				defer server.Close()
			}

			got, err := FetchRemoteModified(server.URL)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("FetchRemoteModified: got nil error, want error containing %q", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("FetchRemoteModified error = %v, want substring %q", err, tt.wantErr)
				}
			} else {
				if err != nil {
					t.Fatalf("FetchRemoteModified: %v", err)
				}
				if got != tt.wantModified {
					t.Errorf("FetchRemoteModified = %q, want %q", got, tt.wantModified)
				}
			}
		})
	}
}
