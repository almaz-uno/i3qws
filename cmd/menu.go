/*
Copyright © 2024 Maxim Kovrov
*/
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.i3wm.org/i3/v4"
)

// menuCmd represents the menu command
var menuCmd = &cobra.Command{
	Use:   "menu",
	Short: "Output menu for Rofi",
	Long:  `Output menu for Rofi`,
	Run: func(cmd *cobra.Command, args []string) {
		doMain(outputMenu)
	},
}

func init() {
	rofiCmd.AddCommand(menuCmd)
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
