package relnotes

import (
	"os"
	"strings"
	"testing"
)

// Criterion K1 of specs/001-releases

// TestExtractNotes takes the section of the first release from the real notes
func TestExtractNotes(t *testing.T) {
	b, err := os.ReadFile("../../RELEASE-NOTES.adoc")
	if err != nil {
		t.Fatal(err)
	}
	s, err := Extract(string(b), "0.1.0")
	if err != nil {
		t.Fatal(err)
	}
	if s == "\n" || strings.Contains(s, "== ") {
		t.Fatalf("0.1.0:\n%s", s)
	}
	if _, err := Extract(string(b), "9.9.9"); err == nil {
		t.Fatal("no error for a missing section")
	}
}

// TestExtractLevels separates a chapter from its patches
func TestExtractLevels(t *testing.T) {
	adoc := "= Notes\n\n== 1.0.0 — 2026-10-05\n\nOne.\n\n=== 1.0.1 — 2026-10-06\n\nOne one.\n\n" +
		"== 0.3.0 — 2026-10-02\n\nThree.\n\n=== 0.3.1 — 2026-10-03\n\nThree one.\n\n" +
		"=== 0.3.2 — 2026-10-04\n\nThree two.\n\n== 0.2.0 — 2026-10-01\n\nTwo.\n"
	for version, want := range map[string]string{
		"1.0.0": "One.\n", "1.0.1": "One one.\n",
		"0.3.0": "Three.\n", "0.3.1": "Three one.\n", "0.3.2": "Three two.\n", "0.2.0": "Two.\n",
	} {
		got, err := Extract(adoc, version)
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Errorf("%s: %q, want %q", version, got, want)
		}
	}
}

func TestMarkdown(t *testing.T) {
	in := "Faster — link:specs/001-releases/spec.adoc[specification\n001]; see link:https://example.org/x[x].\n\n" +
		"*Compatibility.* The same keys.\n\nFixed::\n* `i3qws rofi select` focused twice; `*.yaml` untouched.\n"
	got := Markdown(in, "https://github.com/almaz-uno/i3qws/blob/v0.1.0/")
	for _, want := range []string{
		"[specification 001](https://github.com/almaz-uno/i3qws/blob/v0.1.0/specs/001-releases/spec.adoc)",
		"[x](https://example.org/x)",
		"**Compatibility.** The same keys.",
		"**Fixed**\n* `i3qws rofi select` focused twice; `*.yaml` untouched.",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("no %q in\n%s", want, got)
		}
	}
}
