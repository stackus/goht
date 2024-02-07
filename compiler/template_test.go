package compiler

import (
	"bytes"
	"go/format"
	"os"
	"path/filepath"
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
