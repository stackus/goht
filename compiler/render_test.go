package compiler

import (
	"bytes"
	"context"
	"io"
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
		"html_attributes": {
			template: testdata.HtmlAttributesTest(),
			htmlFile: "html_attributes",
		},
		"haml complex names": {
			template: testdata.HamlComplexNamesTest(),
			htmlFile: "haml_complex_names",
		},
		"slim complex names": {
			template: testdata.SlimComplexNamesTest(),
			htmlFile: "slim_complex_names",
		},
		"newlines": {
			template: testdata.NewlinesTest(),
			htmlFile: "newlines",
		},
		"interpolation": {
			template: testdata.InterpolationTest(),
			htmlFile: "interpolation",
		},
		"slim interpolation": {
			template: testdata.SlimInterpolationTest(),
			htmlFile: "slim_interpolation",
		},
		"comments": {
			template: testdata.CommentsTest(),
			htmlFile: "comments",
		},
		"slim comments": {
			template: testdata.SlimCommentsTest(),
			htmlFile: "slim_comments",
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
		"slim filters": {
			template: testdata.SlimFiltersTest(),
			htmlFile: "slim_filters",
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
		"slim render": {
			template: testdata.SlimRenderTest("child"),
			htmlFile: "slim_rendering",
		},
		"ego render": {
			template: testdata.EgoRenderTest(),
			htmlFile: "ego_rendering",
		},
		"render with slot": {
			template: testdata.RenderWithSlotTest().WithContent(templateText("slotted content")),
			htmlFile: "render_with_slot",
		},
		"slots": {
			template: testdata.SlotTest().
				WithFirst(templateText("first")).
				WithSecond(goht.Fragment{templateText("second one"), templateText("second two")}).
				WithThird(templateText("discarded")).
				WithThird(templateText("third")),
			htmlFile: "slots",
		},
		"slots with defaults": {
			template: testdata.SlotWithDefaultTest(),
			htmlFile: "slots_with_defaults",
		},
		"slim render with slot": {
			template: testdata.SlimRenderWithSlotTest().WithContent(templateText("slotted content")),
			htmlFile: "slim_render_with_slot",
		},
		"ego render with slot": {
			template: testdata.EgoRenderWithSlotTest().WithContent(templateText("slotted content")),
			htmlFile: "ego_render_with_slot",
		},
		"slim slots": {
			template: testdata.SlimSlotTest().
				WithFirst(templateText("first")).
				WithSecond(goht.Fragment{templateText("second one"), templateText("second two")}).
				WithThird(templateText("third")),
			htmlFile: "slim_slots",
		},
		"ego slots": {
			template: testdata.EgoSlotTest().
				WithFirst(templateText("first")).
				WithSecond(goht.Fragment{templateText("second one"), templateText("second two")}).
				WithThird(templateText("third")),
			htmlFile: "ego_slots",
		},
		"ego slots without content": {
			template: testdata.EgoSlotTest(),
			htmlFile: "ego_slots_without_content",
		},
		"slim slots with defaults": {
			template: testdata.SlimSlotWithDefaultTest(),
			htmlFile: "slim_slots_with_defaults",
		},
		"ego slots with defaults": {
			template: testdata.EgoSlotWithDefaultTest(),
			htmlFile: "ego_slots_with_defaults",
		},
		"ego slots with defaults replaced": {
			template: testdata.EgoSlotWithDefaultTest().
				WithFirst(templateText("first")).
				WithSecond(templateText("second")).
				WithThird(templateText("third")).
				WithFourth(templateText("fourth")),
			htmlFile: "ego_slots_with_defaults_replaced",
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

func templateText(value string) goht.TemplateFunc {
	return func(_ context.Context, w io.Writer, _ goht.Slots) error {
		_, err := io.WriteString(w, value)
		return err
	}
}
