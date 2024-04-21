/*
Copyright © 2021 Maxim Kovrov
*/
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.i3wm.org/i3/v4"
)

type (
	window struct {
		ID               i3.NodeID           `json:"id"`
		Name             string              `json:"name"` // window: title, container: internal name
		Type             string              `json:"type"`
		Urgent           bool                `json:"urgent"` // urgency hint set
		Marks            []string            `json:"marks"`
		Window           int64               `json:"window"` // X11 window ID of the client window
		WindowProperties i3.WindowProperties `json:"window_properties"`
	}
)

const (
	classColumnWidthSett = "width-class"
	titleColumnWidthSett = "width-title"

	// scratchName = "♺"
	scratchName = ""
)

// rofiCmd represents the rofi command
var rofiCmd = &cobra.Command{
	Use:   "rofi",
	Short: "Output windows list for rofi switcher or switch window by user's choice was done in rofi",
	Long: `This command ready to using in pipe. For example, you can show window list in rofi for choosing:
    bindsym $mod+Tab exec --no-startup-id bash -c 'i3qws rofi | rofi -dmenu -p window | i3qws rofi'
If standart input is not a named pipe, command formats and output windows list.
If stdin is a named pipe, stdin will be read to get number window to switch.
`,
}

func init() {
	rootCmd.AddCommand(rofiCmd)

	rofiCmd.PersistentFlags().Uint(classColumnWidthSett, 20, "window class column width")
	rofiCmd.PersistentFlags().Uint(titleColumnWidthSett, 80, "window title column width")

	if err := viper.BindPFlags(rofiCmd.PersistentFlags()); err != nil {
		panic("unable to bind flags " + err.Error())
	}
}
