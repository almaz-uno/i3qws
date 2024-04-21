/*
Copyright © 2024 Maxim Kovrov
*/
package cmd

import (
	"context"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// selectCmd represents the select command
var selectCmd = &cobra.Command{
	Use:   "select",
	Short: "Focus the chosen window",
	Long:  `Focus the chosen window`,
	Run: func(cmd *cobra.Command, args []string) {
		doMain(func(ctx context.Context) error {
			bb, err := io.ReadAll(os.Stdin)
			if err != nil {
				return err
			}
			a := strings.TrimSpace(string(bb))
			num := strings.SplitN(a, " ", 2)[0]
			return doFocus(ctx, num)
		})
	},
}

func init() {
	rofiCmd.AddCommand(selectCmd)
}
