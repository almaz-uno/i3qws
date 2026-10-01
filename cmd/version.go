/*
Copyright © 2026 Maxim Kovrov
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// version is set by make build from git describe (specs/001-releases); a
// plain go build or go install leaves "dev"
var version = "dev"

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints the version",
	Long:  `Prints the version of i3qws: the release tag it was built from, or the commit.`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("i3qws " + version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
