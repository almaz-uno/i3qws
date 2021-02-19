/*
Copyright © 2021 Maxim Kovrov

*/
package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const listURL = "http://localhost/list/"

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "Returns the windows list as a JSON array",
	Long:    `Returns collected list of windows as a JSON string`,
	Run: func(cmd *cobra.Command, args []string) {
		doMain(func(ctx context.Context) error {
			str, err := getURL(ctx, listURL)
			if err == nil {
				fmt.Fprintln(os.Stdout, str)
				return nil
			}
			return err
		})
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
