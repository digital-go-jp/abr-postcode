package version

import (
	"strings"
	"testing"
)

func TestString(t *testing.T) {
	got := String()
	for _, want := range []string{"Version:", Version, "Commit:", Commit} {
		if !strings.Contains(got, want) {
			t.Errorf("String() = %q, should contain %q", got, want)
		}
	}
	if lines := strings.Split(got, "\n"); len(lines) != 2 {
		t.Errorf("String() should have 2 lines, got %d", len(lines))
	}
}
