package data

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const inputHeader = "lg_code,machiaza_id,pref,county,city,ward,kyoto_st,oaza_cho,chome,koaza,machiaza_dist,post_code,add_date,dlt_date"

// newInput writes input to a fresh directory and reports both paths, for tests
// that have to prepare the directory before converting.
func newInput(t *testing.T, input string) (dir, src string) {
	t.Helper()
	dir = t.TempDir()
	src = filepath.Join(dir, "in.csv")
	if err := os.WriteFile(src, []byte(input), 0644); err != nil {
		t.Fatalf("write input: %v", err)
	}
	return dir, src
}

func runConvert(t *testing.T, input string) (string, error) {
	t.Helper()
	dir, src := newInput(t, input)
	return dir, Convert(src, dir)
}

func readOutput(t *testing.T, dir, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return string(b)
}

func TestConvert(t *testing.T) {
	input := inputHeader + "\n" +
		// normal row (full-width digits normalized to half-width)
		"131016,0001001,東京都,,千代田区,,,内幸町,１丁目,,,1000011,2023-01-01,\n" +
		// deleted row (dlt_date set) is skipped
		"131016,0001002,東京都,,千代田区,,,内幸町,２丁目,,,1000012,2023-01-01,2024-01-01\n" +
		// same town, another post code: deduplicated in town.csv
		"131016,0001001,東京都,,千代田区,,,内幸町,１丁目,,,1000099,2023-01-01,\n"

	dir, err := runConvert(t, input)
	if err != nil {
		t.Fatalf("Convert: %v", err)
	}

	city := readOutput(t, dir, "city.csv")
	if want := "131016,東京都,,千代田区,"; !strings.Contains(city, want) {
		t.Errorf("city.csv missing %q, got:\n%s", want, city)
	}

	town := readOutput(t, dir, "town.csv")
	if want := "131016,0001001,,内幸町,1丁目,,"; !strings.Contains(town, want) {
		t.Errorf("town.csv missing half-width converted row %q, got:\n%s", want, town)
	}
	if strings.Contains(town, "0001002") {
		t.Errorf("town.csv contains deleted record:\n%s", town)
	}
	if got := strings.Count(town, "0001001"); got != 1 {
		t.Errorf("town.csv row not deduplicated: %d occurrences:\n%s", got, town)
	}

	mapping := readOutput(t, dir, "post_code_mapping.csv")
	for _, want := range []string{"1000011,131016,0001001", "1000099,131016,0001001"} {
		if !strings.Contains(mapping, want) {
			t.Errorf("post_code_mapping.csv missing %q, got:\n%s", want, mapping)
		}
	}
	if strings.Contains(mapping, "1000012") {
		t.Errorf("post_code_mapping.csv contains deleted record:\n%s", mapping)
	}
}

func TestConvertDedupsPostCodeMappingIgnoringAddDate(t *testing.T) {
	// add_date is not part of the output, so two input rows differing only by
	// add_date collapse to a single post_code_mapping row.
	input := inputHeader + "\n" +
		"131016,0001001,東京都,,千代田区,,,内幸町,,,,1000011,2023-01-01,\n" +
		"131016,0001001,東京都,,千代田区,,,内幸町,,,,1000011,2024-06-01,\n"

	dir, err := runConvert(t, input)
	if err != nil {
		t.Fatalf("Convert: %v", err)
	}
	mapping := readOutput(t, dir, "post_code_mapping.csv")
	if got := strings.Count(mapping, "1000011,131016,0001001"); got != 1 {
		t.Errorf("expected one post_code_mapping row, got %d:\n%s", got, mapping)
	}
}

func TestConvertBOMHeader(t *testing.T) {
	input := "\ufeff" + inputHeader + "\n" +
		"131016,0001001,東京都,,千代田区,,,内幸町,,,,1000011,2023-01-01,\n"

	dir, err := runConvert(t, input)
	if err != nil {
		t.Fatalf("Convert with BOM: %v", err)
	}
	town := readOutput(t, dir, "town.csv")
	if want := "131016,0001001,,内幸町,,,"; !strings.Contains(town, want) {
		t.Errorf("town.csv missing %q (BOM broke header mapping?), got:\n%s", want, town)
	}
}

func TestConvertMalformedRow(t *testing.T) {
	input := inputHeader + "\n" +
		"131016,0001001,東京都,,千代田区,,,内幸町,,,,1000011,2023-01-01,\n" +
		"broken,row\n" +
		"131016,0001002,東京都,,千代田区,,,大手町,,,,1000004,2023-01-01,\n"

	if _, err := runConvert(t, input); err == nil {
		t.Fatal("Convert must fail on malformed row instead of silently truncating")
	}
}

// TestConvertLeavesNothingBehindOnFailure covers the atomicity of the output:
// a conversion that fails part-way must not publish a truncated CSV, and must
// not leave its temporary files behind either.
func TestConvertLeavesNothingBehindOnFailure(t *testing.T) {
	input := inputHeader + "\n" +
		"131016,0001001,東京都,,千代田区,,,内幸町,,,,1000011,2023-01-01,\n" +
		"broken,row\n"

	dir, err := runConvert(t, input)
	if err == nil {
		t.Fatal("Convert must fail on a malformed row")
	}

	for _, name := range []string{"city.csv", "town.csv", "post_code_mapping.csv"} {
		for _, path := range []string{name, name + ".tmp"} {
			if _, err := os.Stat(filepath.Join(dir, path)); !os.IsNotExist(err) {
				t.Errorf("%s must not exist after a failed conversion", path)
			}
		}
	}
}

// TestConvertReportsUncreatableOutput covers a temporary file that cannot be
// opened. Blocking the second output means the first one is already open when
// the failure happens, so it also covers the partial set being cleaned up.
func TestConvertReportsUncreatableOutput(t *testing.T) {
	dir, src := newInput(t, inputHeader+"\n"+
		"131016,0001001,東京都,,千代田区,,,内幸町,,,,1000011,2023-01-01,\n")

	// A directory where a temporary file belongs makes os.Create fail.
	if err := os.Mkdir(filepath.Join(dir, "town.csv.tmp"), 0o755); err != nil {
		t.Fatalf("create blocking directory: %v", err)
	}

	if err := Convert(src, dir); err == nil {
		t.Fatal("Convert must fail when an output cannot be created")
	}

	// city.csv.tmp was opened before the failure and has to be removed.
	for _, name := range []string{"city.csv", "city.csv.tmp", "town.csv", "post_code_mapping.csv", "post_code_mapping.csv.tmp"} {
		if _, err := os.Stat(filepath.Join(dir, name)); !os.IsNotExist(err) {
			t.Errorf("%s must not exist after a failed conversion", name)
		}
	}
}

// TestConvertReportsUnpublishableOutput covers a rename that cannot complete
// while republishing over an existing set. city.csv precedes town.csv, so the
// failure leaves the first file renewed and the rest stale: the
// mixed-generation case the per-file renames cannot avoid.
func TestConvertReportsUnpublishableOutput(t *testing.T) {
	dir, src := newInput(t, inputHeader+"\n"+
		"131016,0001001,東京都,,千代田区,,,内幸町,,,,1000011,2023-01-01,\n")

	const stale = "stale\n"
	if err := os.WriteFile(filepath.Join(dir, "city.csv"), []byte(stale), 0644); err != nil {
		t.Fatalf("seed city.csv: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "post_code_mapping.csv"), []byte(stale), 0644); err != nil {
		t.Fatalf("seed post_code_mapping.csv: %v", err)
	}

	// A non-empty directory cannot be replaced by a rename.
	blocked := filepath.Join(dir, "town.csv")
	if err := os.Mkdir(blocked, 0o755); err != nil {
		t.Fatalf("create blocking directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(blocked, "keep"), []byte("x"), 0644); err != nil {
		t.Fatalf("populate blocking directory: %v", err)
	}

	if err := Convert(src, dir); err == nil {
		t.Fatal("Convert must fail when an output cannot be published")
	}

	for _, name := range []string{"city.csv.tmp", "town.csv.tmp", "post_code_mapping.csv.tmp"} {
		if _, err := os.Stat(filepath.Join(dir, name)); !os.IsNotExist(err) {
			t.Errorf("%s must not be left behind", name)
		}
	}

	if got := readOutput(t, dir, "city.csv"); got == stale {
		t.Error("city.csv precedes town.csv and was renamed before the failure, want it renewed")
	}
	if got := readOutput(t, dir, "post_code_mapping.csv"); got != stale {
		t.Errorf("post_code_mapping.csv follows town.csv, want it left stale, got:\n%s", got)
	}
}

// TestConvertKeepsTheExistingMode covers republishing over CSVs whose mode was
// set outside the importer. Truncating in place used to preserve it, and
// renaming a replacement into position has to do the same.
func TestConvertKeepsTheExistingMode(t *testing.T) {
	dir, src := newInput(t, inputHeader+"\n"+
		"131016,0001001,東京都,,千代田区,,,内幸町,,,,1000011,2023-01-01,\n")

	const mode os.FileMode = 0o640
	for _, name := range []string{"city.csv", "town.csv", "post_code_mapping.csv"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("stale\n"), mode); err != nil {
			t.Fatalf("seed %s: %v", name, err)
		}
	}

	if err := Convert(src, dir); err != nil {
		t.Fatalf("Convert: %v", err)
	}

	for _, name := range []string{"city.csv", "town.csv", "post_code_mapping.csv"} {
		info, err := os.Stat(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("stat %s: %v", name, err)
		}
		if got := info.Mode().Perm(); got != mode {
			t.Errorf("%s mode = %v, want %v", name, got, mode)
		}
	}
}

func TestConvertRejectsEmptyInput(t *testing.T) {
	// An empty dataset would otherwise be published and served as a working
	// API where every address is simply not found.
	cases := map[string]string{
		"header only":     inputHeader + "\n",
		"all rows delete": inputHeader + "\n131016,0001001,東京都,,千代田区,,,内幸町,,,,1000011,2023-01-01,2024-01-01\n",
	}

	for name, input := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := runConvert(t, input); err == nil {
				t.Fatal("Convert must fail when the input holds no importable records")
			}
		})
	}
}

func TestConvertMissingColumn(t *testing.T) {
	// A missing dlt_date would silently disable the deleted-record filter,
	// so a missing required column must be an explicit error.
	input := "lg_code,machiaza_id,pref\n131016,0001001,東京都\n"

	_, err := runConvert(t, input)
	if err == nil {
		t.Fatal("Convert must fail when required columns are missing")
	}
	if !strings.Contains(err.Error(), "post_code") {
		t.Errorf("error should name a missing column, got: %v", err)
	}
}
