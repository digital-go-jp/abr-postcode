package service

import (
	"os"
	"path/filepath"
	"testing"
)

func writeCSV(t *testing.T, dir, name, content string) string {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create %s: %v", name, err)
	}
	return path
}

func TestForEachRow(t *testing.T) {
	tests := []struct {
		name       string
		csvContent string
		wantRows   [][2]string
	}{
		{
			name:       "valid CSV",
			csvContent: "name,value\nfoo,bar\nbaz,qux\n",
			wantRows:   [][2]string{{"foo", "bar"}, {"baz", "qux"}},
		},
		{
			name:       "empty CSV with header only",
			csvContent: "name,value\n",
		},
		{
			name:       "missing column reads as empty",
			csvContent: "name\nfoo\n",
			wantRows:   [][2]string{{"foo", ""}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeCSV(t, t.TempDir(), "in.csv", tt.csvContent)

			var got [][2]string
			if err := forEachRow(path, func(row *csvReader) {
				got = append(got, [2]string{row.value("name"), row.value("value")})
			}); err != nil {
				t.Fatalf("forEachRow: %v", err)
			}

			if len(got) != len(tt.wantRows) {
				t.Fatalf("read %d rows, want %d: %v", len(got), len(tt.wantRows), got)
			}
			for i, want := range tt.wantRows {
				if got[i] != want {
					t.Errorf("row %d = %v, want %v", i, got[i], want)
				}
			}
		})
	}
}

// TestForEachRowEmptyFile covers a file with no header row at all, which
// leaves no column names to read by.
func TestForEachRowEmptyFile(t *testing.T) {
	path := writeCSV(t, t.TempDir(), "empty.csv", "")

	err := forEachRow(path, func(*csvReader) {})
	if err == nil {
		t.Error("forEachRow: expected an error for a file without a header, got nil")
	}
}

func TestForEachRowFileNotFound(t *testing.T) {
	err := forEachRow("/nonexistent/path/file.csv", func(*csvReader) {})
	if err == nil {
		t.Error("forEachRow: expected an error for a nonexistent file, got nil")
	}
}

// TestForEachRowMalformedRow verifies that a row with the wrong field count
// fails instead of being silently skipped or truncated.
func TestForEachRowMalformedRow(t *testing.T) {
	path := writeCSV(t, t.TempDir(), "in.csv", "name,value\nfoo,bar\nbroken\n")

	err := forEachRow(path, func(*csvReader) {})
	if err == nil {
		t.Error("forEachRow: expected an error for a malformed row, got nil")
	}
}

func TestLoadAddressData(t *testing.T) {
	dir := t.TempDir()
	writeCSV(t, dir, "city.csv",
		"lg_code,pref,county,city,ward\n"+
			"011061,北海道,,札幌市,南区\n")
	writeCSV(t, dir, "town.csv",
		"lg_code,machiaza_id,kyoto_st,oaza_cho,chome,koaza,machiaza_dist\n"+
			"011061,0056000,,藤野,,,\n")
	writeCSV(t, dir, "post_code_mapping.csv",
		"post_code,lg_code,machiaza_id\n"+
			"0612271,011061,0056000\n"+
			"0050840,011061,0056000\n")

	data, err := LoadAddressData(dir)
	if err != nil {
		t.Fatalf("LoadAddressData: %v", err)
	}

	city, ok := data.Cities["011061"]
	if !ok {
		t.Fatal("city 011061 not found")
	}
	if city.Pref != "北海道" || city.Ward != "南区" {
		t.Errorf("city = %+v, want pref 北海道 and ward 南区", city)
	}

	town, ok := data.Towns["0110610056000"]
	if !ok {
		t.Fatal("town 0110610056000 not found")
	}
	if town.OazaCho != "藤野" {
		t.Errorf("town = %+v, want oaza_cho 藤野", town)
	}

	if got := len(data.PostCodeMappings["0050840"]); got != 1 {
		t.Errorf("post code 0050840 has %d mappings, want 1", got)
	}

	// The reverse index is sorted regardless of the order in the file.
	postCodes := data.TownToPostCodes["0110610056000"]
	want := []string{"0050840", "0612271"}
	if len(postCodes) != len(want) {
		t.Fatalf("post codes = %v, want %v", postCodes, want)
	}
	for i, w := range want {
		if postCodes[i] != w {
			t.Errorf("post codes = %v, want %v", postCodes, want)
			break
		}
	}
}

func TestLoadAddressDataMissingFile(t *testing.T) {
	if _, err := LoadAddressData(t.TempDir()); err == nil {
		t.Error("LoadAddressData: expected an error when the CSVs are absent, got nil")
	}
}
