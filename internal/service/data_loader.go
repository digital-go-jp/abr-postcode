package service

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
)

// csvReader walks a CSV file row by row, exposing each row by column name.
// Rows are streamed rather than collected so a full copy of the input never
// has to exist alongside the indexes built from it.
type csvReader struct {
	path   string
	file   *os.File
	reader *csv.Reader
	column map[string]int
	record []string
}

func openCSV(path string) (*csvReader, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}

	reader := csv.NewReader(file)
	reader.ReuseRecord = true

	headers, err := reader.Read()
	if err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("parse CSV %s: %w", path, err)
	}

	column := make(map[string]int, len(headers))
	for i, h := range headers {
		column[h] = i
	}

	return &csvReader{path: path, file: file, reader: reader, column: column}, nil
}

// next advances to the following row, reporting false at end of input.
func (r *csvReader) next() (bool, error) {
	record, err := r.reader.Read()
	if errors.Is(err, io.EOF) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("parse CSV %s: %w", r.path, err)
	}
	r.record = record
	return true, nil
}

// value returns the current row's column, or "" when the column is absent.
// A row with the wrong number of fields fails at next, so a resolved column is
// always in range.
func (r *csvReader) value(name string) string {
	if i, ok := r.column[name]; ok {
		return r.record[i]
	}
	return ""
}

// close releases the file. Read failures surface from next, so a close error
// on a file opened for reading carries nothing to act on.
func (r *csvReader) close() {
	r.file.Close()
}

// forEachRow streams path, calling fn once per row.
func forEachRow(path string, fn func(row *csvReader)) error {
	r, err := openCSV(path)
	if err != nil {
		return err
	}
	defer r.close()

	for {
		ok, err := r.next()
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
		fn(r)
	}
}
