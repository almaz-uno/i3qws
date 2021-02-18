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

const focusURL = "http://localhost/focus/"

// focusCmd represents the focus command
var focusCmd = &cobra.Command{
	Use:   "focus",
	Short: "Focus specified window in saved history",
	Long: `Focus window by number in the window queue. Number starts from 0 (current focused window), 1 - previous focused window and so on.

Negative number mean posion from the tail of the queue: -1 - the last window, -2 - one from the tail and so on.`,
	Run: func(cmd *cobra.Command, args []string) {
		doMain(func(ctx context.Context) error {
			return doFocus(ctx, args[0])
		})
	},
	Args: cobra.ExactArgs(1),
}

func init() {
	rootCmd.AddCommand(focusCmd)
}

func doFocus(ctx context.Context, num string) error {
	str, err := getURL(ctx, focusURL+num)
	if err == nil {
		fmt.Fprintln(os.Stdout, str)
		return nil
	}
	return err
}
