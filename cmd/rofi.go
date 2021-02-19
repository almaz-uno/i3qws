/*
Copyright © 2021 Maxim Kovrov

*/
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.i3wm.org/i3/v4"
)

type (
	window struct {
		ID               int64               `json:"id"`
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
)

// rofiCmd represents the rofi command
var rofiCmd = &cobra.Command{
	Use:   "rofi",
	Short: "Output windows list for rofi switcher or switch window by user's choice in done in rofi",
	Long: `This command ready to using in pipe. For example, you can show window list in rofi for choosing:
    bindsym $mod+Tab exec --no-startup-id bash -c 'i3qws rofi | rofi -dmenu -p window | i3qws rofi'
If standart input is not a named pipe, command formats and output windows list.
If stdin is a named pipe, stdin will be read to get number window to switch.
`,
	Run: func(cmd *cobra.Command, args []string) {
		doMain(func(ctx context.Context) error {
			fi, err := os.Stdin.Stat()
			if err != nil {
				return err
			}
			if fi.Mode()&os.ModeNamedPipe == 0 {
				return outputMenu(ctx)
			} else {
				bb, err := ioutil.ReadAll(os.Stdin)
				if err != nil {
					return err
				}
				a := strings.TrimSpace(string(bb))
				num := strings.SplitN(a, " ", 2)[0]
				return doFocus(ctx, num)
			}
		})
	},
}

func init() {
	rootCmd.AddCommand(rofiCmd)

	rofiCmd.PersistentFlags().Uint(classColumnWidthSett, 20, "window class column width")
	rofiCmd.PersistentFlags().Uint(titleColumnWidthSett, 80, "window title column width")

	if err := viper.BindPFlags(rofiCmd.PersistentFlags()); err != nil {
		panic("unable to bind flags " + err.Error())
	}
}

func spaceAlign(orig string, max uint) string {
	if max == 0 {
		return orig
	}
	rr := []rune(orig)
	if len(rr) < int(max) {
		rr = append(rr, []rune(strings.Repeat(" ", int(max)-len(rr)))...)
	}
	rr = rr[0 : max-1]
	return string(rr)
}

func outputMenu(ctx context.Context) error {
	str, err := getURL(ctx, listURL)
	if err != nil {
		return err
	}

	var ww []window
	err = json.Unmarshal([]byte(str), &ww)
	if err != nil {
		return err
	}

	for i, w := range ww {
		marks := strings.Builder{}
		for _, m := range w.Marks {
			marks.WriteString("[")
			marks.WriteString(m)
			marks.WriteString("]")
		}
		fmt.Fprintf(os.Stdout, "%3d %s %s %s\n", i,
			spaceAlign(w.WindowProperties.Class, viper.GetUint(classColumnWidthSett)),
			spaceAlign(w.Name, viper.GetUint(titleColumnWidthSett)),
			marks.String())
	}

	return err
}
