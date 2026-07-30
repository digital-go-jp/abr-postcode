package data

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

type OutputFile struct {
	Path    string
	Headers []string
}

// Convert processes the ABR CSV data and splits it into three normalized CSV files.
func Convert(inputPath string, dataDir string) error {
	outputs := []OutputFile{
		{
			Path:    filepath.Join(dataDir, "city.csv"),
			Headers: []string{"lg_code", "pref", "county", "city", "ward"},
		},
		{
			Path:    filepath.Join(dataDir, "town.csv"),
			Headers: []string{"lg_code", "machiaza_id", "kyoto_st", "oaza_cho", "chome", "koaza", "machiaza_dist"},
		},
		{
			Path:    filepath.Join(dataDir, "post_code_mapping.csv"),
			Headers: []string{"post_code", "lg_code", "machiaza_id"},
		},
	}

	return processCSV(inputPath, outputs)
}

func processCSV(inputPath string, outputs []OutputFile) error {
	in, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file %s: %w", inputPath, err)
	}
	defer in.Close()

	reader := csv.NewReader(in)
	reader.ReuseRecord = true

	idxMap, err := readColumnIndex(reader)
	if err != nil {
		return err
	}
	if err := requireColumns(idxMap, outputs, inputPath); err != nil {
		return err
	}
	deletedAt := idxMap["dlt_date"]

	writers, err := newOutputWriters(outputs, idxMap)
	defer writers.discard()
	if err != nil {
		return err
	}

	imported := 0

	for {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read CSV row: %w", err)
		}

		// Skip deleted records. Half-width normalization cannot empty a field,
		// so the raw value answers this.
		if record[deletedAt] != "" {
			continue
		}
		imported++

		if err := writers.writeRow(record); err != nil {
			return err
		}
	}

	// An input that yields nothing would otherwise be published and served as
	// an empty dataset, which is indistinguishable from every address simply
	// not being found.
	if imported == 0 {
		return fmt.Errorf("input CSV %s holds no importable records", inputPath)
	}

	if err := writers.commit(); err != nil {
		return err
	}

	slog.Info("Data conversion completed", "records", imported)
	return nil
}

// outputWriter accumulates one normalized CSV in a temporary file alongside its
// destination. The column positions and the row buffers are resolved once, so
// converting a row costs no allocation beyond the values it keeps.
type outputWriter struct {
	OutputFile
	tmpPath string
	file    *os.File
	csv     *csv.Writer
	columns []int    // position in the input record, per header
	values  []string // reused row buffer
	key     []byte   // reused deduplication key
	seen    map[string]struct{}
}

type outputWriters []*outputWriter

// newOutputWriters opens a temporary file per output and writes its header row.
// Every writer it managed to open is returned even on failure, so the caller's
// deferred discard cleans up a partial set.
func newOutputWriters(outputs []OutputFile, idxMap map[string]int) (outputWriters, error) {
	var writers outputWriters

	for _, out := range outputs {
		tmpPath := out.Path + ".tmp"
		file, err := os.Create(tmpPath)
		if err != nil {
			return writers, fmt.Errorf("failed to create output file %s: %w", tmpPath, err)
		}

		w := &outputWriter{
			OutputFile: out,
			tmpPath:    tmpPath,
			file:       file,
			csv:        csv.NewWriter(file),
			columns:    make([]int, len(out.Headers)),
			values:     make([]string, len(out.Headers)),
			seen:       make(map[string]struct{}),
		}
		for i, h := range out.Headers {
			w.columns[i] = idxMap[h]
		}
		writers = append(writers, w)

		// Truncating the destination used to keep its mode. Renaming a fresh
		// file over it would not, so carry the mode across when there is one
		// to carry.
		if info, err := os.Stat(out.Path); err == nil {
			if err := file.Chmod(info.Mode().Perm()); err != nil {
				return writers, fmt.Errorf("failed to set the mode of %s: %w", tmpPath, err)
			}
		}

		if err := w.csv.Write(out.Headers); err != nil {
			return writers, fmt.Errorf("failed to write headers to %s: %w", out.Path, err)
		}
	}

	return writers, nil
}

// writeRow projects the input record onto each output's columns and writes the
// result unless that exact row was written before. NUL joins the values so a
// comma inside a field cannot make two distinct rows collide.
func (ws outputWriters) writeRow(record []string) error {
	for _, w := range ws {
		w.key = w.key[:0]
		for i, column := range w.columns {
			value := toHalfWidth(record[column])
			w.values[i] = value

			if i > 0 {
				w.key = append(w.key, 0)
			}
			w.key = append(w.key, value...)
		}

		// Looking the key up as a string does not copy it; only storing does.
		if _, exists := w.seen[string(w.key)]; exists {
			continue
		}
		w.seen[string(w.key)] = struct{}{}

		if err := w.csv.Write(w.values); err != nil {
			return fmt.Errorf("failed to write row to %s: %w", w.Path, err)
		}
	}
	return nil
}

// commit flushes and closes every output before renaming any of them, so a
// conversion that fails while writing leaves the previous CSVs untouched
// rather than truncated.
//
// The renames are per file, so a failure between them can leave one CSV
// renewed and another not. The caller does not record the new data timestamp
// when commit fails, which makes the next run import the whole set again.
func (ws outputWriters) commit() error {
	for _, w := range ws {
		w.csv.Flush()
		if err := w.csv.Error(); err != nil {
			return fmt.Errorf("failed to flush %s: %w", w.Path, err)
		}
		if err := w.file.Close(); err != nil {
			return fmt.Errorf("failed to close %s: %w", w.tmpPath, err)
		}
	}

	for _, w := range ws {
		if err := os.Rename(w.tmpPath, w.Path); err != nil {
			return fmt.Errorf("failed to publish %s: %w", w.Path, err)
		}
		w.tmpPath = ""
	}

	return nil
}

// discard closes and removes any temporary file that was not committed. It
// runs after both success and failure, so it discards the errors it gets:
// commit has already closed the files it published, and a file that was never
// created has nothing to remove.
func (ws outputWriters) discard() {
	for _, w := range ws {
		_ = w.file.Close()
		if w.tmpPath != "" {
			_ = os.Remove(w.tmpPath)
		}
	}
}

// readColumnIndex reads the header row and maps each column name to its
// position. The BOM the distribution CSV starts with is stripped so the first
// column name matches.
func readColumnIndex(reader *csv.Reader) (map[string]int, error) {
	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV headers: %w", err)
	}
	headers[0] = strings.TrimPrefix(headers[0], "\ufeff")

	idxMap := make(map[string]int, len(headers))
	for i, h := range headers {
		idxMap[h] = i
	}
	return idxMap, nil
}

// requireColumns reports the columns the outputs read that the input lacks.
// Missing columns would otherwise map to empty strings and silently produce
// broken output, for example a missing dlt_date disabling the deleted-record
// filter.
func requireColumns(idxMap map[string]int, outputs []OutputFile, inputPath string) error {
	required := []string{"dlt_date"}
	for _, out := range outputs {
		required = append(required, out.Headers...)
	}

	var missing []string
	for _, h := range required {
		if _, ok := idxMap[h]; !ok {
			missing = append(missing, h)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("required columns %s not found in input CSV %s", strings.Join(missing, ", "), inputPath)
	}
	return nil
}

// toHalfWidth converts full-width digits to half-width
func toHalfWidth(s string) string {
	return strings.Map(func(r rune) rune {
		if r >= '０' && r <= '９' {
			return r - '０' + '0'
		}
		return r
	}, s)
}
