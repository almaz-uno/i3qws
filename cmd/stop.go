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

const stopURL = "http://localhost/stop/"

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stops application",
	Long:  `Stops application`,
	Run: func(cmd *cobra.Command, args []string) {
		doMain(func(ctx context.Context) error {
			str, err := getURL(ctx, stopURL)
			if err == nil {
				fmt.Fprintln(os.Stdout, str)
				return nil
			}
			return err
		})
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
