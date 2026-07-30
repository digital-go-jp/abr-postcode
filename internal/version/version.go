// Package version provides build version information injected via ldflags.
package version

import "fmt"

// Build-time version information injected via ldflags:
//
//	go build -ldflags "-X abr-postcode/internal/version.Version=x.y.z -X abr-postcode/internal/version.Commit=abc123"
var (
	Version = "dev"  // semantic version (e.g. "1.0.0")
	Commit  = "none" // git commit hash
)

// String returns formatted version information.
func String() string {
	return fmt.Sprintf("Version: %s\nCommit: %s", Version, Commit)
}
