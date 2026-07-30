package cmd

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"abr-postcode/internal/config"
	"abr-postcode/internal/data"
)

// buildDistributionZip builds a zip with the same layout as tm-postcode's
// output: abr_post_code.csv and metadata.txt at the archive root.
func buildDistributionZip(t *testing.T) []byte {
	t.Helper()

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	csvBody := "lg_code,machiaza_id,pref,county,city,ward,kyoto_st,oaza_cho,chome,koaza,machiaza_dist,post_code,add_date,dlt_date\r\n" +
		"131016,0001001,東京都,,千代田区,,,内幸町,一丁目,,,1000011,2026-01-01,\r\n"
	f, err := zw.Create("abr_post_code.csv")
	if err != nil {
		t.Fatalf("create csv entry: %v", err)
	}
	if _, err := f.Write([]byte(csvBody)); err != nil {
		t.Fatalf("write csv entry: %v", err)
	}

	m, err := zw.Create("metadata.txt")
	if err != nil {
		t.Fatalf("create metadata entry: %v", err)
	}
	if _, err := m.Write([]byte("郵便番号データ版: 2605\r\n")); err != nil {
		t.Fatalf("write metadata entry: %v", err)
	}

	if err := zw.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	return buf.Bytes()
}

// TestUnzipFileRejectsPathTraversal verifies that a malicious entry escaping
// dataDir via ".." is rejected and no file is written outside the tree.
func TestUnzipFileRejectsPathTraversal(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	f, err := zw.Create("../escape.txt")
	if err != nil {
		t.Fatalf("create escaping entry: %v", err)
	}
	if _, err := f.Write([]byte("owned")); err != nil {
		t.Fatalf("write escaping entry: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}

	parent := t.TempDir()
	dataDir := filepath.Join(parent, "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		t.Fatalf("mkdir dataDir: %v", err)
	}
	zipPath := filepath.Join(dataDir, "malicious.zip")
	if err := os.WriteFile(zipPath, buf.Bytes(), 0644); err != nil {
		t.Fatalf("write malicious zip: %v", err)
	}

	if err := unzipFile(zipPath, dataDir); err == nil {
		t.Fatal("expected unzipFile to reject path traversal, got nil error")
	}
	if _, err := os.Stat(filepath.Join(parent, "escape.txt")); !os.IsNotExist(err) {
		t.Errorf("escaping file was written outside dataDir")
	}
}

func TestDownloadAndConvert(t *testing.T) {
	srv := newDataServer(t, buildDistributionZip(t))

	dir := t.TempDir()
	if err := downloadAndConvert(srv.URL, dir); err != nil {
		t.Fatalf("downloadAndConvert: %v", err)
	}

	for _, name := range []string{"city.csv", "town.csv", "post_code_mapping.csv"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Errorf("expected output %s: %v", name, err)
		}
	}

	town, err := os.ReadFile(filepath.Join(dir, "town.csv"))
	if err != nil {
		t.Fatalf("read town.csv: %v", err)
	}
	if !strings.Contains(string(town), "131016,0001001,,内幸町,一丁目,,") {
		t.Errorf("town.csv missing expected row, got:\n%s", town)
	}

	// Downloaded and extracted artifacts must be cleaned up after conversion.
	for _, name := range []string{"abr_post_code.zip", "abr_post_code.csv", "metadata.txt"} {
		if _, err := os.Stat(filepath.Join(dir, name)); !os.IsNotExist(err) {
			t.Errorf("expected %s to be removed after conversion", name)
		}
	}
}

// newDataServer serves the distribution zip.
func newDataServer(t *testing.T, zipBytes []byte) *httptest.Server {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if _, err := w.Write(zipBytes); err != nil {
			t.Errorf("serve zip: %v", err)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

// newFeedServer serves a DCAT feed carrying the ABR postcode dataset with the
// given modification timestamp.
func newFeedServer(t *testing.T, modified string) *httptest.Server {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, err := fmt.Fprintf(w, `{"dataset":[{"identifier":"urn:uuid:%s","modified":%q}]}`,
			data.ABRPostCodeItemID, modified)
		if err != nil {
			t.Errorf("serve feed: %v", err)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

// captureStdout runs fn with os.Stdout redirected and returns what it wrote.
// It replaces os.Stdout process-wide, so callers must not run in parallel with
// anything that writes to stdout.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("open pipe: %v", err)
	}
	orig := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = orig }()

	type captured struct {
		out string
		err error
	}
	done := make(chan captured, 1)
	go func() {
		var buf bytes.Buffer
		_, err := io.Copy(&buf, r)
		done <- captured{out: buf.String(), err: err}
	}()

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("close pipe writer: %v", err)
	}
	got := <-done
	if got.err != nil {
		t.Fatalf("read captured stdout: %v", got.err)
	}
	out := got.out
	if err := r.Close(); err != nil {
		t.Fatalf("close pipe reader: %v", err)
	}
	return out
}

// TestImportDataDryRun pins the dry-run result: the exit code the data update
// workflow branches on, and the three keys reported on stdout.
func TestImportDataDryRun(t *testing.T) {
	const remote = "2099-01-01T00:00:00.000Z"

	tests := []struct {
		name          string
		localModified string
		wantAvailable bool
		wantExitCode  int
	}{
		{
			name:          "update available",
			localModified: "",
			wantAvailable: true,
			wantExitCode:  1,
		},
		{
			name:          "up to date",
			localModified: remote,
			wantAvailable: false,
			wantExitCode:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dataDir := t.TempDir()
			if tt.localModified != "" {
				if err := data.SaveDataModified(dataDir, tt.localModified); err != nil {
					t.Fatalf("SaveDataModified: %v", err)
				}
			}
			srv := newFeedServer(t, remote)

			t.Chdir(t.TempDir())
			t.Setenv("ABRP_FEED_URL", srv.URL)
			t.Setenv("ABRP_DATA_DIR", dataDir)
			var err error
			out := captureStdout(t, func() { err = importData(config.Load(), true) })

			if got := exitCodeFor(err); got != tt.wantExitCode {
				t.Errorf("exit code = %d, want %d (err = %v)", got, tt.wantExitCode, err)
			}

			var result map[string]any
			if err := json.Unmarshal([]byte(out), &result); err != nil {
				t.Fatalf("dry-run output is not JSON: %v\noutput: %s", err, out)
			}

			want := map[string]any{
				"local_modified":   tt.localModified,
				"remote_modified":  remote,
				"update_available": tt.wantAvailable,
			}
			if len(result) != len(want) {
				t.Errorf("dry-run output has %d keys, want %d: %v", len(result), len(want), result)
			}
			for key, wantValue := range want {
				got, ok := result[key]
				if !ok {
					t.Errorf("dry-run output missing key %q: %v", key, result)
					continue
				}
				if got != wantValue {
					t.Errorf("%s = %v, want %v", key, got, wantValue)
				}
			}
		})
	}
}

// TestImportDataFailsWhenTimestampCannotBeRecorded covers the case where the
// CSVs are converted but the timestamp cannot be written. Reporting success
// would leave every later run re-importing the same data unnoticed.
func TestImportDataFailsWhenTimestampCannotBeRecorded(t *testing.T) {
	dataSrv := newDataServer(t, buildDistributionZip(t))
	feedSrv := newFeedServer(t, "2099-01-01T00:00:00.000Z")

	// A directory in place of data_modified.txt makes the write fail without
	// touching the permissions the test process runs with.
	dataDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dataDir, "data_modified.txt"), 0o755); err != nil {
		t.Fatalf("create blocking directory: %v", err)
	}

	t.Chdir(t.TempDir())
	t.Setenv("ABRP_FEED_URL", feedSrv.URL)
	t.Setenv("ABRP_DATA_URL", dataSrv.URL)
	t.Setenv("ABRP_DATA_DIR", dataDir)
	err := importData(config.Load(), false)
	if err == nil {
		t.Fatal("importData: got nil error when the timestamp could not be recorded")
	}
	if got := exitCodeFor(err); got != 2 {
		t.Errorf("exit code = %d, want 2 (err = %v)", got, err)
	}
}

// TestImportDataDryRunReportsFetchFailure verifies that a failed check exits 2
// rather than 1, so a crash cannot be read as "an update is pending".
func TestImportDataDryRunReportsFetchFailure(t *testing.T) {
	srv := newFeedServer(t, "2099-01-01T00:00:00.000Z")
	unreachable := srv.URL
	srv.Close()

	t.Chdir(t.TempDir())
	t.Setenv("ABRP_FEED_URL", unreachable)
	t.Setenv("ABRP_DATA_DIR", t.TempDir())
	err := importData(config.Load(), true)
	if err == nil {
		t.Fatal("importData: got nil error for an unreachable feed")
	}
	if got := exitCodeFor(err); got != 2 {
		t.Errorf("exit code = %d, want 2 (err = %v)", got, err)
	}
}
