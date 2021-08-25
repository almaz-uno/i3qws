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
	"unicode/utf8"

	"github.com/sirupsen/logrus"
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
	rr = rr[0:max]
	return string(rr)
}

func outputMenu(ctx context.Context) error {
	tree, err := i3.GetTree()
	if err != nil {
		return err
	}

	mapping := make(map[i3.NodeID]string)
	fillFromNode(mapping, tree.Root)
	maxLenWsp := 0
	for k, v := range mapping {
		if v == "__i3_scratch" {
			mapping[k] = scratchName
			v = scratchName
		}

		l := utf8.RuneCountInString(v)
		if l > maxLenWsp {
			maxLenWsp = l
		}
	}

	str, err := getURL(ctx, listURL)
	if err != nil {
		return err
	}

	var ww []window
	err = json.Unmarshal([]byte(str), &ww)
	if err != nil {
		return err
	}

	titleMaxWidth := 0
	for _, w := range ww {
		n := strings.TrimSpace(w.Name)
		l := utf8.RuneCountInString(n)
		if l > titleMaxWidth {
			titleMaxWidth = l
		}
	}

	if titleMaxWidth > 0 && titleMaxWidth > viper.GetInt(titleColumnWidthSett) {
		titleMaxWidth = viper.GetInt(titleColumnWidthSett)
	}

	for i, w := range ww {
		marks := strings.Builder{}
		for _, m := range w.Marks {
			marks.WriteString("[")
			marks.WriteString(m)
			marks.WriteString("]")
		}

		wsp := mapping[w.ID]

		fmt.Fprintf(os.Stdout, "%3d  %s  %s %s %s\n", i,
			spaceAlign(wsp, uint(maxLenWsp)),
			spaceAlign(w.WindowProperties.Class, viper.GetUint(classColumnWidthSett)),
			spaceAlign(strings.TrimSpace(w.Name), uint(titleMaxWidth)),
			marks.String())
	}

	return err
}

// getWspWindowMap returns map windowID ⇒ workspace name
func fillFromNode(mapping map[i3.NodeID]string, node *i3.Node) {
	doInWorkspace(mapping, "", node)
}

func doInWorkspace(mapping map[i3.NodeID]string, workspace string, node *i3.Node) {
	switch node.Type {
	case i3.WorkspaceNode:
		if workspace != "" {
			logrus.Warnf("Workspace '%s' put in another '%s'. We shall using nested instead.", node.Name, workspace)
		}
		workspace = node.Name
	case i3.Con:
		mapping[node.ID] = workspace
	}

	for _, n := range node.Nodes {
		doInWorkspace(mapping, workspace, n)
	}

	for _, n := range node.FloatingNodes {
		doInWorkspace(mapping, workspace, n)
	}
}
