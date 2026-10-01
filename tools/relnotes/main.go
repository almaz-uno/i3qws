// Command relnotes prints the text of a GitHub release from the release notes:
// relnotes <file.adoc> <version> <base of links> (specs/001-releases).
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/cured-plumbum/i3qws/internal/relnotes"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Fprintln(os.Stderr, "usage: relnotes <RELEASE-NOTES.adoc> <X.Y.Z> <https://…/blob/<tag>/>")
		os.Exit(2)
	}
	b, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	s, err := relnotes.Extract(string(b), strings.TrimPrefix(os.Args[2], "v"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Print(relnotes.Markdown(s, os.Args[3]))
}
