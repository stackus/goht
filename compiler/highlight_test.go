package compiler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func semanticFixtures(t *testing.T) []string {
	t.Helper()
	files, err := filepath.Glob("testdata/semantic/*.goht")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("no semantic fixtures found")
	}
	return files
}

// formatHighlight renders tokens as `line:col-line:col kind "text"` using
// zero-based lines and byte columns.
func formatHighlight(src []byte, tokens []HighlightToken) string {
	position := func(offset int) (int, int) {
		line := strings.Count(string(src[:offset]), "\n")
		col := offset - (strings.LastIndex(string(src[:offset]), "\n") + 1)
		return line, col
	}
	var b strings.Builder
	for _, tok := range tokens {
		sl, sc := position(tok.Start)
		el, ec := position(tok.End)
		fmt.Fprintf(&b, "%d:%d-%d:%d %s %q\n", sl, sc, el, ec, tok.Kind, src[tok.Start:tok.End])
	}
	return b.String()
}

func TestHighlightGolden(t *testing.T) {
	for _, file := range semanticFixtures(t) {
		t.Run(filepath.Base(file), func(t *testing.T) {
			src, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			got := formatHighlight(src, Highlight(src))
			want, err := goldenFile(t, strings.TrimSuffix(file, ".goht")+".tokens", []byte(got), *update)
			if err != nil {
				t.Fatal(err)
			}
			if got != string(want) {
				t.Errorf("tokens mismatch\n got:\n%s\nwant:\n%s", got, want)
			}
		})
	}
}

func TestHighlightInvariants(t *testing.T) {
	for _, file := range semanticFixtures(t) {
		t.Run(filepath.Base(file), func(t *testing.T) {
			src, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			tokens := Highlight(src)
			if len(tokens) == 0 {
				t.Fatal("expected tokens")
			}
			prevEnd := 0
			for _, tok := range tokens {
				if tok.Start < prevEnd || tok.Start >= tok.End || tok.End > len(src) {
					t.Fatalf("bad token %v after %d (len %d)", tok, prevEnd, len(src))
				}
				if text := string(src[tok.Start:tok.End]); strings.TrimSpace(text) != text {
					t.Errorf("token %v has surrounding whitespace: %q", tok, text)
				}
				prevEnd = tok.End
			}
		})
	}
}

func TestHighlightSkipsGoCode(t *testing.T) {
	src := []byte("package main\n\n@slim A(user User) {\n\tp= user.Name\n\tp #{user.Age}\n}\n")
	for _, tok := range Highlight(src) {
		if text := string(src[tok.Start:tok.End]); strings.Contains(text, "user") {
			t.Errorf("Go code classified as %s: %q", tok.Kind, text)
		}
	}
}
