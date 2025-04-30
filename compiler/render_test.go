package compiler

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"

	"github.com/sergi/go-diff/diffmatchpatch"

	"github.com/stackus/goht"
	"github.com/stackus/goht/compiler/testdata"
)

func TestRender(t *testing.T) {
	tests := map[string]struct {
		template goht.Template
		htmlFile string
	}{
		"package": {
			template: testdata.PackageTest(),
			htmlFile: "package",
		},
		"imports": {
			template: testdata.ImportsTest(),
			htmlFile: "imports",
		},
		"elements": {
			template: testdata.ElementsTest(),
			htmlFile: "elements",
		},
		"attributes": {
			template: testdata.AttributesTest(),
			htmlFile: "attributes",
		},
		"newlines": {
			template: testdata.NewlinesTest(),
			htmlFile: "newlines",
		},
		"interpolation": {
			template: testdata.InterpolationTest(),
			htmlFile: "interpolation",
		},
		"comments": {
			template: testdata.CommentsTest(),
			htmlFile: "comments",
		},
		"conditionals.true": {
			template: testdata.ConditionalsTest(true),
			htmlFile: "conditionals.true",
		},
		"conditionals.false": {
			template: testdata.ConditionalsTest(false),
			htmlFile: "conditionals.false",
		},
		"filters": {
			template: testdata.FiltersTest(),
			htmlFile: "filters",
		},
		"object references": {
			template: testdata.ObjectReferencesTest(),
			htmlFile: "obj_references",
		},
		"whitespace": {
			template: testdata.WhitespaceTest(),
			htmlFile: "whitespace",
		},
		"render": {
			template: testdata.RenderTest(),
			htmlFile: "rendering",
		},
		"without children": {
			template: testdata.ChildrenTest("passed-in"),
			htmlFile: "without_children",
		},
		"nesting": {
			template: testdata.NestedRenderTest(),
			htmlFile: "nesting",
		},
		"slim template": {
			template: testdata.SlimTemplate(),
			htmlFile: "slim_template",
		},
		"ego template": {
			template: testdata.EgoTemplate(),
			htmlFile: "ego_template",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var gotW bytes.Buffer
			err := tt.template.Render(context.Background(), &gotW)
			if err != nil {
				t.Errorf("error generating template: %v", err)
				return
			}

			got := gotW.Bytes()
			goldenFileName := filepath.Join("testdata", tt.htmlFile+".html")
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
