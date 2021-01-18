/*
Copyright © 2021 Maxim Kovrov

*/
package cmd

import (
	"context"

	"github.com/spf13/cobra"
)

const listURL = "http://localhost/list/"

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		doMain(func(ctx context.Context) error {
			return getURL(ctx, listURL)
		})
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
