package examples

import (
	"bytes"
	"context"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/stackus/errors"

	"github.com/stackus/goht"
	"github.com/stackus/goht/examples/attributes"
	"github.com/stackus/goht/examples/commands"
	"github.com/stackus/goht/examples/comments"
	"github.com/stackus/goht/examples/doctype"
	"github.com/stackus/goht/examples/filters"
	"github.com/stackus/goht/examples/formatting"
	example "github.com/stackus/goht/examples/go"
	"github.com/stackus/goht/examples/hello"
	"github.com/stackus/goht/examples/indents"
	"github.com/stackus/goht/examples/tags"
	unescape "github.com/stackus/goht/examples/unescaping"
)

var (
	update = flag.Bool("update", false, "update the generated golden files")
)

func goldenFile(t *testing.T, fileName string, got []byte, update bool) ([]byte, error) {
	t.Helper()

	want, err := os.ReadFile(fileName)
	if err != nil {
		if !update || !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
	}

	// If the update flag is set, write the golden file when either the file does not exist or the contents do not match.
	if update && (!bytes.Equal(want, got) || err != nil) {
		err := os.WriteFile(fileName, got, 0644)
		if err != nil {
			return nil, err
		}

		return got, nil
	}

	return want, nil
}

func TestExamples(t *testing.T) {
	tests := map[string]struct {
		template goht.Template
		htmlFile string
	}{
		"attributes_attributesCmd": {
			template: attributes.AttributesCmd(),
			htmlFile: "attributes_attributesCmd",
		},
		"attributes_classes": {
			template: attributes.Classes(),
			htmlFile: "attributes_classes",
		},
		"attributes_staticAttrs": {
			template: attributes.StaticAttrs(),
			htmlFile: "attributes_staticAttrs",
		},
		"attributes_dynamicAttrs": {
			template: attributes.DynamicAttrs(),
			htmlFile: "attributes_dynamicAttrs",
		},
		"attributes_multilineAttrs": {
			template: attributes.MultilineAttrs(),
			htmlFile: "attributes_multilineAttrs",
		},
		"attributes_whitespaceAttrs": {
			template: attributes.WhitespaceAttrs(),
			htmlFile: "attributes_whitespaceAttrs",
		},
		"attributes_formattedValue": {
			template: attributes.FormattedValue(),
			htmlFile: "attributes_formattedValue",
		},
		"attributes_simpleNames": {
			template: attributes.SimpleNames(),
			htmlFile: "attributes_simpleNames",
		},
		"attributes_complexNames": {
			template: attributes.ComplexNames(),
			htmlFile: "attributes_complexNames",
		},
		"attributes_conditionalAttrs": {
			template: attributes.ConditionalAttrs(),
			htmlFile: "attributes_conditionalAttrs",
		},
		"commands_childrenExample": {
			template: commands.ChildrenExample(),
			htmlFile: "commands_childrenExample",
		},
		"commands_renderExample": {
			template: commands.RenderExample(),
			htmlFile: "commands_renderExample",
		},
		"commands_renderWithChildrenExample": {
			template: commands.RenderWithChildrenExample(),
			htmlFile: "commands_renderWithChildrenExample",
		},
		"comments_htmlComments": {
			template: comments.HtmlComments(),
			htmlFile: "comments_htmlComments",
		},
		"comments_htmlCommentsNested": {
			template: comments.HtmlCommentsNested(),
			htmlFile: "comments_htmlCommentsNested",
		},
		"comments_rubyStyle": {
			template: comments.RubyStyle(),
			htmlFile: "comments_rubyStyle",
		},
		"comments_rubyStyleNested": {
			template: comments.RubyStyleNested(),
			htmlFile: "comments_rubyStyleNested",
		},
		"doctype_doctype": {
			template: doctype.Doctype(),
			htmlFile: "doctype_doctype",
		},
		"filters_css": {
			template: filters.Css(),
			htmlFile: "filters_css",
		},
		"filters_javascript": {
			template: filters.JavaScript(),
			htmlFile: "filters_javascript",
		},
		"filters_plain": {
			template: filters.Plain(),
			htmlFile: "filters_plain",
		},
		"filters_escaped": {
			template: filters.Escaped(),
			htmlFile: "filters_escaped",
		},
		"filters_preserve": {
			template: filters.Preserve(),
			htmlFile: "filters_preserve",
		},
		"formatting_intExample": {
			template: formatting.IntExample(),
			htmlFile: "formatting_intExample",
		},
		"formatting_floatExample": {
			template: formatting.FloatExample(),
			htmlFile: "formatting_floatExample",
		},
		"formatting_boolExample": {
			template: formatting.BoolExample(),
			htmlFile: "formatting_boolExample",
		},
		"formatting_stringExample": {
			template: formatting.StringExample(),
			htmlFile: "formatting_stringExample",
		},
		"example_executeCode": {
			template: example.ExecuteCode(),
			htmlFile: "example_executeCode",
		},
		"example_renderCode": {
			template: example.RenderCode(),
			htmlFile: "example_renderCode",
		},
		"example_doc": {
			template: example.Doc(),
			htmlFile: "example_doc",
		},
		"example_importExample": {
			template: example.ImportExample(),
			htmlFile: "example_importExample",
		},
		"example_conditional": {
			template: example.Conditional(),
			htmlFile: "example_conditional",
		},
		"example_shorthandConditional": {
			template: example.ShorthandConditional(),
			htmlFile: "example_shorthandConditional",
		},
		"example_shorthandSwitch": {
			template: example.ShorthandSwitch(),
			htmlFile: "example_shorthandSwitch",
		},
		"example_interpolateCode": {
			template: example.InterpolateCode(),
			htmlFile: "example_interpolateCode",
		},
		"example_noInterpolation": {
			template: example.NoInterpolation(),
			htmlFile: "example_noInterpolation",
		},
		"example_escapeInterpolation": {
			template: example.EscapeInterpolation(),
			htmlFile: "example_escapeInterpolation",
		},
		"example_ignoreInterpolation": {
			template: example.IgnoreInterpolation(),
			htmlFile: "example_ignoreInterpolation",
		},
		"example_packageExample": {
			template: example.PackageExample(),
			htmlFile: "example_packageExample",
		},
		"example_userDetails": {
			template: func() goht.Template {
				user := example.User{
					Name: "John",
					Age:  30,
				}
				return user.Details()
			}(),
			htmlFile: "example_userDetails",
		},
		"indents_usingSpaces": {
			template: indents.UsingSpaces(),
			htmlFile: "indents_usingSpaces",
		},
		"indents_usingTabs": {
			template: indents.UsingTabs(),
			htmlFile: "indents_usingTabs",
		},
		"tags_specifyTag": {
			template: tags.SpecifyTag(),
			htmlFile: "tags_specifyTag",
		},
		"tags_defaultToDivs": {
			template: tags.DefaultToDivs(),
			htmlFile: "tags_defaultToDivs",
		},
		"tabs_combined": {
			template: tags.Combined(),
			htmlFile: "tags_combined",
		},
		"tags_multipleClasses": {
			template: tags.MultipleClasses(),
			htmlFile: "tags_multipleClasses",
		},
		"tags_objectRefs": {
			template: func() goht.Template {
				foo := tags.Foo{
					ID: "foo",
				}
				return tags.ObjectRefs(foo)
			}(),
			htmlFile: "tags_objectRefs",
		},
		"tags_prefixedObjectRefs": {
			template: func() goht.Template {
				foo := tags.Foo{
					ID: "foo",
				}
				return tags.PrefixedObjectRefs(foo)
			}(),
			htmlFile: "tags_prefixedObjectRefs",
		},
		"tags_selfClosing": {
			template: tags.SelfClosing(),
			htmlFile: "tags_selfClosing",
		},
		"tags_alsoSelfClosing": {
			template: tags.AlsoSelfClosing(),
			htmlFile: "tags_alsoSelfClosing",
		},
		"tags_whitespace": {
			template: tags.Whitespace(),
			htmlFile: "tags_whitespace",
		},
		"tags_removeWhitespace": {
			template: tags.RemoveWhitespace(),
			htmlFile: "tags_removeWhitespace",
		},
		"hello_world": {
			template: hello.World(),
			htmlFile: "hello_world",
		},
		"unescape_unescapeCode": {
			template: unescape.UnescapeCode(),
			htmlFile: "unescape_unescapeCode",
		},
		"unescape_unescapeInterpolation": {
			template: unescape.UnescapeInterpolation(),
			htmlFile: "unescape_unescapeInterpolation",
		},
		"unescape_unescapeText": {
			template: unescape.UnescapeText(),
			htmlFile: "unescape_unescapeText",
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
