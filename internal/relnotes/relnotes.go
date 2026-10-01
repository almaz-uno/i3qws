// Package relnotes extracts the section of a release from RELEASE-NOTES.adoc
// and turns it into Markdown for the GitHub release
// (specs/001-releases; after qws, its specification 004).
package relnotes

import (
	"fmt"
	"regexp"
	"strings"
)

// Extract returns the text of the release version ("0.1.0", without "v")
// without its heading: for X.Y.0 the chapter "== X.Y.0 …" up to its first
// subsection or the next chapter — patches are releases of their own; for a
// patch the subsection "=== X.Y.Z …" up to the next heading.
func Extract(adoc, version string) (string, error) {
	level := "== "
	if !strings.HasSuffix(version, ".0") {
		level = "=== "
	}
	lines := strings.Split(adoc, "\n")
	start := -1
	for i, l := range lines {
		if l == level+version || strings.HasPrefix(l, level+version+" ") {
			start = i + 1
			break
		}
	}
	if start < 0 {
		return "", fmt.Errorf("no section %q in the release notes", strings.TrimSpace(level+version))
	}
	end := len(lines)
	for i := start; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "== ") || strings.HasPrefix(lines[i], "=== ") {
			end = i
			break
		}
	}
	return strings.TrimSpace(strings.Join(lines[start:end], "\n")) + "\n", nil
}

var (
	linkRe   = regexp.MustCompile(`link:([^\s\[]+)\[([^\]]*)\]`)
	labelRe  = regexp.MustCompile(`(?m)^(\S[^\n]*)::$`)
	strongRe = regexp.MustCompile(`(^|[\s(])\*([^*\s][^*\n]*[^*\s]|[^*\s])\*`)
)

// Markdown translates the markup of the notes: link:path[text] becomes a link
// to a file of the repository (base + path; the text may break over a line),
// "Label::" and *bold* become Markdown bold; the rest stays as it is — lists
// and `code` are the same in both.
func Markdown(s, base string) string {
	s = linkRe.ReplaceAllStringFunc(s, func(m string) string {
		p := linkRe.FindStringSubmatch(m)
		text := strings.Join(strings.Fields(p[2]), " ")
		url := p[1]
		if !strings.Contains(url, "://") {
			url = base + url
		}
		return "[" + text + "](" + url + ")"
	})
	s = labelRe.ReplaceAllString(s, "**$1**")
	return strongRe.ReplaceAllString(s, "$1**$2**")
}
