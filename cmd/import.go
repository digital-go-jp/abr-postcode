package cmd

import (
	"archive/zip"
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"abr-postcode/internal/config"
	"abr-postcode/internal/data"

	"github.com/spf13/cobra"
)

var (
	dryRun bool
)

// UpdateAvailableError reports that a dry-run found pending changes. It is a
// result rather than a failure, so it exits 1 the way diff(1) reports
// differences, leaving 2 to mean the check itself could not be completed.
type UpdateAvailableError struct{}

func (UpdateAvailableError) Error() string { return "update available" }

func (UpdateAvailableError) ExitCode() int { return 1 }

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import/update ABR data",
	Long:  "Check DCAT feed for updates and import ABR data if needed",
	RunE: func(cmd *cobra.Command, args []string) error {
		return importData(config.Load(), dryRun)
	},
}

func init() {
	importCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Check for updates without importing")
}

func importData(cfg *config.Config, dryRun bool) error {
	dataDir := cfg.DataDir

	// Check for updates
	remote, err := data.FetchRemoteModified(cfg.DCATFeedURL)
	if err != nil {
		slog.Error("Failed to fetch DCAT feed", "error", err)
		return err
	}

	local := data.GetLocalModified(dataDir)

	slog.Info("Data update check",
		"local_modified", cmp.Or(local, "(none)"),
		"remote_modified", remote,
	)

	updateAvailable := local != remote

	if dryRun {
		result := map[string]any{
			"update_available": updateAvailable,
			"local_modified":   local,
			"remote_modified":  remote,
		}
		output, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to encode dry-run result: %w", err)
		}
		fmt.Println(string(output))
		if updateAvailable {
			return UpdateAvailableError{}
		}
		return nil
	}

	if !updateAvailable {
		slog.Info("Data is up to date, skipping import")
		return nil
	}

	slog.Info("Update available, downloading and converting data")
	if err := downloadAndConvert(cfg.DataURL, dataDir); err != nil {
		slog.Error("Failed to import data", "error", err)
		return err
	}

	// Record the data's timestamp (already fetched above) so the next run can
	// detect updates. Without it every later run re-imports the same data and
	// nothing reports why, so this fails rather than continuing.
	if err := data.SaveDataModified(dataDir, remote); err != nil {
		slog.Error("Failed to record the data timestamp", "error", err)
		return fmt.Errorf("failed to record the data timestamp: %w", err)
	}

	slog.Info("Data import completed", "remote_modified", remote)
	return nil
}

// downloadAndConvert downloads the ABR data and converts it to CSV
func downloadAndConvert(dataURL, dataDir string) error {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	slog.Info("Downloading ABR data", "url", dataURL)
	zipPath := filepath.Join(dataDir, "abr_post_code.zip")
	if err := downloadFile(dataURL, zipPath); err != nil {
		return fmt.Errorf("failed to download data: %w", err)
	}

	slog.Info("Extracting zip file")
	if err := unzipFile(zipPath, dataDir); err != nil {
		return fmt.Errorf("failed to unzip data: %w", err)
	}

	// The distribution zip (built by tm-postcode) contains abr_post_code.csv
	// and metadata.txt at the archive root.
	slog.Info("Converting data")
	csvPath := filepath.Join(dataDir, "abr_post_code.csv")
	if err := data.Convert(csvPath, dataDir); err != nil {
		return fmt.Errorf("failed to convert data: %w", err)
	}

	// Remove the downloaded zip and extracted files now that the normalized
	// CSVs have been written.
	for _, p := range []string{zipPath, csvPath, filepath.Join(dataDir, "metadata.txt")} {
		_ = os.Remove(p)
	}

	return nil
}

// downloadFile downloads a file from URL and saves to disk atomically via temp file.
// Only the connection setup phases are bounded; the body transfer itself is not, so a
// slow-but-progressing download of a multi-MB zip is not cut off.
func downloadFile(url, dstPath string) error {
	client := &http.Client{
		Transport: &http.Transport{
			DialContext:           (&net.Dialer{Timeout: 10 * time.Second}).DialContext,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 30 * time.Second,
		},
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	tmpPath := dstPath + ".tmp"
	out, err := os.Create(tmpPath)
	if err != nil {
		return err
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		_ = out.Close()
		_ = os.Remove(tmpPath)
		return err
	}

	if err := out.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}

	return os.Rename(tmpPath, dstPath)
}

// unzipFile extracts the flat distribution zip (files at the archive root) into
// dataDir. Extraction is confined to dataDir via os.Root, which rejects entries
// that escape the tree through ".." or symlinks (Zip Slip).
func unzipFile(zipPath, dataDir string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %w", err)
	}
	defer reader.Close()

	root, err := os.OpenRoot(dataDir)
	if err != nil {
		return fmt.Errorf("failed to open data dir %s: %w", dataDir, err)
	}
	defer root.Close()

	for _, file := range reader.File {
		if file.FileInfo().IsDir() {
			continue
		}
		if err := extractZipEntry(root, file); err != nil {
			return err
		}
	}

	return nil
}

// extractZipEntry writes a single zip entry into root.
func extractZipEntry(root *os.Root, file *zip.File) error {
	rc, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to read file from zip: %w", err)
	}
	defer rc.Close()

	out, err := root.OpenFile(file.Name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", file.Name, err)
	}
	defer out.Close()

	if _, err := io.Copy(out, rc); err != nil {
		return fmt.Errorf("failed to extract file %s: %w", file.Name, err)
	}
	return nil
}
