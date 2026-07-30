package cmd

import (
	"errors"
	"fmt"
	"testing"
)

// codedError carries an arbitrary exit code, so a test can tell whether
// exitCodeFor reads the reported value or assumes the dry-run code.
type codedError struct{ code int }

func (codedError) Error() string { return "coded" }

func (e codedError) ExitCode() int { return e.code }

// TestExitCodeFor pins the exit code contract the data update workflow
// branches on: 0 means up to date, 1 means an update is pending, and anything
// else means the check itself could not be completed.
func TestExitCodeFor(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{
			name: "no error",
			err:  nil,
			want: 0,
		},
		{
			name: "update available",
			err:  UpdateAvailableError{},
			want: 1,
		},
		{
			name: "wrapped update available",
			err:  fmt.Errorf("check for updates: %w", UpdateAvailableError{}),
			want: 1,
		},
		{
			name: "plain error",
			err:  errors.New("feed unreachable"),
			want: 2,
		},
		{
			name: "coded error reports its own code",
			err:  codedError{code: 3},
			want: 3,
		},
		{
			name: "wrapped coded error reports its own code",
			err:  fmt.Errorf("run command: %w", codedError{code: 3}),
			want: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := exitCodeFor(tt.err); got != tt.want {
				t.Errorf("exitCodeFor(%v) = %d, want %d", tt.err, got, tt.want)
			}
		})
	}
}
