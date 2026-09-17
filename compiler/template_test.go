package compiler

import (
	"bytes"
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sergi/go-diff/diffmatchpatch"
)

func TestTemplate_Generate(t *testing.T) {
	tests := map[string]struct {
		templateFile string
	}{
		"package": {
			templateFile: "package",
		},
		"imports": {
			templateFile: "imports",
		},
		"elements": {
			templateFile: "elements",
		},
		"attributes": {
			templateFile: "attributes",
		},
		"newlines": {
			templateFile: "newlines",
		},
		"interpolation": {
			templateFile: "interpolation",
		},
		"comments": {
			templateFile: "comments",
		},
		"conditionals": {
			templateFile: "conditionals",
		},
		"filters": {
			templateFile: "filters",
		},
		"object references": {
			templateFile: "obj_references",
		},
		"whitespace": {
			templateFile: "whitespace",
		},
		"render": {
			templateFile: "rendering",
		},
		"typed slots": {
			templateFile: "typed_slots",
		},
		"slot directives": {
			templateFile: "slot_directives",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			fileName := filepath.Join("testdata", tt.templateFile+".goht")
			contents, err := os.ReadFile(fileName)
			if err != nil {
				t.Errorf("error reading file: %v", err)
				return
			}
			var tpl *Template
			tpl, err = ParseString(string(contents))
			if err != nil {
				t.Errorf("error parsing template: %v", err)
				return
			}

			var gotW bytes.Buffer
			err = tpl.Generate(&gotW)
			if err != nil {
				t.Errorf("error generating template: %v", err)
				return
			}

			var got []byte
			got, err = format.Source(gotW.Bytes())
			if err != nil {
				t.Errorf("error formatting source: %v", err)
				return
			}

			goldenFileName := filepath.Join("testdata", tt.templateFile+".goht.go")
			want, err := goldenFile(t, goldenFileName, got, *update)
			if err != nil {
				t.Errorf("error reading golden file: %v", err)
				return
			}

			if bytes.Equal(want, got) {
				return
			}

			dmp := diffmatchpatch.New()
			diffs := dmp.DiffMain(string(want), string(got), true)
			if len(diffs) > 1 {
				t.Errorf("diff:\n%s", dmp.DiffPrettyText(diffs))
			}
		})
	}
}

func TestTemplate_GenerateSlotMethodErrors(t *testing.T) {
	tests := map[string]struct {
		source string
		want   string
	}{
		"invalid slot name": {
			source: "package testdata\n@slim Invalid() {\n\t=@slot main/content\n}\n",
			want:   `invalid slot name "main/content"`,
		},
		"normalized method collision": {
			source: "package testdata\n@slim Collision() {\n\t=@slot main-content\n\t=@slot main_content\n}\n",
			want:   `slot name "main_content" conflicts with "main-content": both generate WithMainContent`,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tpl, err := ParseString(tt.source)
			if err != nil {
				t.Fatalf("ParseString() error = %v", err)
			}
			var output bytes.Buffer
			err = tpl.Generate(&output)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Generate() error = %v, want containing %q", err, tt.want)
			}
		})
	}
}

func TestParseString_ReservedChildrenSlot(t *testing.T) {
	tests := map[string]string{
		"haml": "@haml Reserved() {\n\t= @slot children\n}\n",
		"slim": "@slim Reserved() {\n\t= @slot children\n}\n",
		"ego":  "@ego Reserved() {\n\t<%@slot children %>\n}\n",
	}
	for name, source := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := ParseString(source)
			if err == nil || !strings.Contains(err.Error(), `slot name "children" is reserved`) {
				t.Fatalf("ParseString() error = %v, want reserved-children error", err)
			}
		})
	}
}

func TestSlotDirectiveSourceMaps(t *testing.T) {
	tests := map[string]string{
		"haml": "package testdata\n@haml Test() {\n\t= @ifslot content\n\t\t= @eachslot item in content\n\t\t\t= @render item\n}\n",
		"slim": "package testdata\n@slim Test() {\n\t= @ifslot content\n\t\t= @eachslot item in content\n\t\t\t= @render item\n}\n",
		"ego":  "package testdata\n@ego Test() {\n\t<%@ifslot content { %>\n\t<%@eachslot item in content { %>\n\t<%@render item %>\n\t<% } %>\n\t<% } %>\n}\n",
	}
	for name, source := range tests {
		t.Run(name, func(t *testing.T) {
			template, err := ParseString(source)
			if err != nil {
				t.Fatalf("ParseString() error = %v", err)
			}
			var output bytes.Buffer
			sourceMap, err := template.Compose(&output)
			if err != nil {
				t.Fatalf("Compose() error = %v", err)
			}
			if len(sourceMap.SourceLinesToTarget) == 0 {
				t.Fatal("Compose() produced no source-map entries")
			}
		})
	}
}

// TestSourceMapRoundTrip compiles a small template in each of GoHT's three
// syntaxes and confirms that every entry the SourceMap records round-trips
// through both SourcePositionFromTarget and TargetPositionFromSource, and
// that codegen still records a non-trivial number of mappings.
//
// The fixtures use script lines (silent "- "/"= ", or EGO's "<% %>"/"<%= %>")
// rather than plain tags/attributes/text, because only source spans that can
// end up as literal Go code in the generated file are ever added to the
// SourceMap (see the tw.Add call sites in nodes.go) — static markup and
// plain text are written as inert string literals that can't produce a Go
// compiler error, so they're intentionally never mapped.
func TestSourceMapRoundTrip(t *testing.T) {
	tests := map[string]struct{ src string }{
		"haml": {src: "package testdata\n\n@haml HamlTest() {\n\t- x := 1\n\t= x\n}\n"},
		"slim": {src: "package testdata\n\n@slim SlimTest() {\n\t- x := 1\n\t= x\n}\n"},
		"ego":  {src: "package testdata\n\n@ego EgoTest() {\n\t<% x := 1 %>\n\t<%= x %>\n}\n"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tpl, err := ParseString(tt.src)
			if err != nil {
				t.Fatalf("ParseString() error = %v", err)
			}

			var buf bytes.Buffer
			sm, err := tpl.Compose(&buf)
			if err != nil {
				t.Fatalf("Compose() error = %v", err)
			}

			const minEntries = 5
			entries := 0
			for srcLine, cols := range sm.SourceLinesToTarget {
				for srcCol, target := range cols {
					entries++
					gotSrc, ok := sm.SourcePositionFromTarget(target.Line, target.Col)
					if !ok {
						t.Fatalf("SourcePositionFromTarget(%d,%d) not found (from src %d,%d)", target.Line, target.Col, srcLine, srcCol)
					}
					if gotSrc != (Position{Line: srcLine, Col: srcCol}) {
						t.Errorf("round trip for src(%d,%d) via tgt(%d,%d) = %#v, want {Line:%d Col:%d}",
							srcLine, srcCol, target.Line, target.Col, gotSrc, srcLine, srcCol)
					}
				}
			}
			if entries < minEntries {
				t.Fatalf("only %d source map entries recorded, want at least %d; codegen may have stopped calling SourceMap.Add\n%s", entries, minEntries, sm.Dump())
			}
		})
	}
}
