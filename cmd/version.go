package cmd

import (
	"fmt"

	"abr-postcode/internal/version"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:                   "version",
	Short:                 "Show version information",
	DisableFlagsInUseLine: true,
	RunE: func(_ *cobra.Command, _ []string) error {
		fmt.Println(version.String())
		return nil
	},
}
