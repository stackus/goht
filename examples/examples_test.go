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
		// Create the directory if it doesn't exist
		dir := filepath.Dir(fileName)
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return nil, err
		}
		err = os.WriteFile(fileName, got, 0644)
		if err != nil {
			return nil, err
		}

		return got, nil
	}

	return want, nil
}

func TestHamlExamples(t *testing.T) {
	tests := map[string]struct {
		template goht.Template
		htmlFile string
	}{
		"attributes_attributesCmd": {
			template: attributes.HamlAttributesCmd(),
			htmlFile: "attributes_attributesCmd",
		},
		"attributes_classes": {
			template: attributes.HamlClasses(),
			htmlFile: "attributes_classes",
		},
		"attributes_staticAttrs": {
			template: attributes.HamlStaticAttrs(),
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
			template: commands.HamlChildrenExample(),
			htmlFile: "commands_childrenExample",
		},
		"commands_renderExample": {
			template: commands.HamlRenderExample(),
			htmlFile: "commands_renderExample",
		},
		"commands_renderWithChildrenExample": {
			template: commands.HamlRenderWithChildrenExample(),
			htmlFile: "commands_renderWithChildrenExample",
		},
		"comments_htmlComments": {
			template: comments.HamlHtmlComments(),
			htmlFile: "comments_htmlComments",
		},
		"comments_htmlCommentsNested": {
			template: comments.HamlHtmlCommentsNested(),
			htmlFile: "comments_htmlCommentsNested",
		},
		"comments_rubyStyle": {
			template: comments.HamlRubyStyle(),
			htmlFile: "comments_rubyStyle",
		},
		"comments_rubyStyleNested": {
			template: comments.HamlRubyStyleNested(),
			htmlFile: "comments_rubyStyleNested",
		},
		"doctype_doctype": {
			template: doctype.HamlDoctype(),
			htmlFile: "doctype_doctype",
		},
		"filters_css": {
			template: filters.HamlCss(),
			htmlFile: "filters_css",
		},
		"filters_javascript": {
			template: filters.HamlJavaScript(),
			htmlFile: "filters_javascript",
		},
		"filters_plain": {
			template: filters.HamlPlain(),
			htmlFile: "filters_plain",
		},
		"filters_escaped": {
			template: filters.HamlEscaped(),
			htmlFile: "filters_escaped",
		},
		"filters_preserve": {
			template: filters.HamlPreserve(),
			htmlFile: "filters_preserve",
		},
		"formatting_intExample": {
			template: formatting.HamlIntExample(),
			htmlFile: "formatting_intExample",
		},
		"formatting_floatExample": {
			template: formatting.HamlFloatExample(),
			htmlFile: "formatting_floatExample",
		},
		"formatting_boolExample": {
			template: formatting.HamlBoolExample(),
			htmlFile: "formatting_boolExample",
		},
		"formatting_stringExample": {
			template: formatting.HamlStringExample(),
			htmlFile: "formatting_stringExample",
		},
		"example_executeCode": {
			template: example.HamlExecuteCode(),
			htmlFile: "example_executeCode",
		},
		"example_renderCode": {
			template: example.HamlRenderCode(),
			htmlFile: "example_renderCode",
		},
		"example_doc": {
			template: example.HamlDoc(),
			htmlFile: "example_doc",
		},
		"example_importExample": {
			template: example.HamlImportExample(),
			htmlFile: "example_importExample",
		},
		"example_conditional": {
			template: example.HamlConditional(),
			htmlFile: "example_conditional",
		},
		"example_shorthandConditional": {
			template: example.HamlShorthandConditional(),
			htmlFile: "example_shorthandConditional",
		},
		"example_shorthandSwitch": {
			template: example.HamlShorthandSwitch(),
			htmlFile: "example_shorthandSwitch",
		},
		"example_interpolateCode": {
			template: example.HamlInterpolateCode(),
			htmlFile: "example_interpolateCode",
		},
		"example_noInterpolation": {
			template: example.HamlNoInterpolation(),
			htmlFile: "example_noInterpolation",
		},
		"example_escapeInterpolation": {
			template: example.HamlEscapeInterpolation(),
			htmlFile: "example_escapeInterpolation",
		},
		"example_ignoreInterpolation": {
			template: example.HamlIgnoreInterpolation(),
			htmlFile: "example_ignoreInterpolation",
		},
		"example_userDetails": {
			template: func() goht.Template {
				user := example.User{
					Name: "John",
					Age:  30,
				}
				return user.HamlDetails()
			}(),
			htmlFile: "example_userDetails",
		},
		"indents_usingTabs": {
			template: indents.HamlUsingTabs(),
			htmlFile: "indents_usingTabs",
		},
		"tags_specifyTag": {
			template: tags.HamlSpecifyTag(),
			htmlFile: "tags_specifyTag",
		},
		"tags_defaultToDivs": {
			template: tags.HamlDefaultToDivs(),
			htmlFile: "tags_defaultToDivs",
		},
		"tabs_combined": {
			template: tags.HamlCombined(),
			htmlFile: "tags_combined",
		},
		"tags_multipleClasses": {
			template: tags.HamlMultipleClasses(),
			htmlFile: "tags_multipleClasses",
		},
		"tags_objectRefs": {
			template: func() goht.Template {
				foo := tags.Foo{
					ID: "foo",
				}
				return tags.HamlObjectRefs(foo)
			}(),
			htmlFile: "tags_objectRefs",
		},
		"tags_prefixedObjectRefs": {
			template: func() goht.Template {
				foo := tags.Foo{
					ID: "foo",
				}
				return tags.HamlPrefixedObjectRefs(foo)
			}(),
			htmlFile: "tags_prefixedObjectRefs",
		},
		"tags_selfClosing": {
			template: tags.HamlSelfClosing(),
			htmlFile: "tags_selfClosing",
		},
		"tags_alsoSelfClosing": {
			template: tags.HamlAlsoSelfClosing(),
			htmlFile: "tags_alsoSelfClosing",
		},
		"tags_whitespace": {
			template: tags.HamlWhitespace(),
			htmlFile: "tags_whitespace",
		},
		"tags_removeWhitespace": {
			template: tags.HamlRemoveWhitespace(),
			htmlFile: "tags_removeWhitespace",
		},
		"hello_world": {
			template: hello.World(),
			htmlFile: "hello_world",
		},
		"unescape_unescapeCode": {
			template: unescape.HamlUnescapeCode(),
			htmlFile: "unescape_unescapeCode",
		},
		"unescape_unescapeInterpolation": {
			template: unescape.HamlUnescapeInterpolation(),
			htmlFile: "unescape_unescapeInterpolation",
		},
		"unescape_unescapeText": {
			template: unescape.HamlUnescapeText(),
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
			goldenFileName := filepath.Join("testdata", "haml", tt.htmlFile+".html")
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

func TestSlimExamples(t *testing.T) {
	tests := map[string]struct {
		template goht.Template
		htmlFile string
	}{
		"attributes_attributesCmd": {
			template: attributes.SlimAttributesCmd(),
			htmlFile: "attributes_attributesCmd",
		},
		"attributes_classes": {
			template: attributes.SlimClasses(),
			htmlFile: "attributes_classes",
		},
		"attributes_staticAttrs": {
			template: attributes.SlimStaticAttrs(),
			htmlFile: "attributes_staticAttrs",
		},
		"attributes_dynamicAttrs": {
			template: attributes.SlimDynamicAttrs(),
			htmlFile: "attributes_dynamicAttrs",
		},
		"attributes_multilineAttrs": {
			template: attributes.SlimMultilineAttrs(),
			htmlFile: "attributes_multilineAttrs",
		},
		"attributes_whitespaceAttrs": {
			template: attributes.SlimWhitespaceAttrs(),
			htmlFile: "attributes_whitespaceAttrs",
		},
		"attributes_formattedValue": {
			template: attributes.SlimFormattedValue(),
			htmlFile: "attributes_formattedValue",
		},
		"attributes_simpleNames": {
			template: attributes.SlimSimpleNames(),
			htmlFile: "attributes_simpleNames",
		},
		"attributes_complexNames": {
			template: attributes.SlimComplexNames(),
			htmlFile: "attributes_complexNames",
		},
		"attributes_conditionalAttrs": {
			template: attributes.SlimConditionalAttrs(),
			htmlFile: "attributes_conditionalAttrs",
		},
		"commands_childrenExample": {
			template: commands.SlimChildrenExample(),
			htmlFile: "commands_childrenExample",
		},
		"commands_renderExample": {
			template: commands.SlimRenderExample(),
			htmlFile: "commands_renderExample",
		},
		"commands_renderWithChildrenExample": {
			template: commands.SlimRenderWithChildrenExample(),
			htmlFile: "commands_renderWithChildrenExample",
		},
		"comments_htmlComments": {
			template: comments.SlimHtmlComments(),
			htmlFile: "comments_htmlComments",
		},
		"comments_htmlCommentsNested": {
			template: comments.SlimHtmlCommentsNested(),
			htmlFile: "comments_htmlCommentsNested",
		},
		"comments_rubyStyle": {
			template: comments.SlimRubyStyle(),
			htmlFile: "comments_rubyStyle",
		},
		"comments_rubyStyleNested": {
			template: comments.SlimRubyStyleNested(),
			htmlFile: "comments_rubyStyleNested",
		},
		"doctype_doctype": {
			template: doctype.SlimDoctype(),
			htmlFile: "doctype_doctype",
		},
		"filters_css": {
			template: filters.SlimCss(),
			htmlFile: "filters_css",
		},
		"filters_javascript": {
			template: filters.SlimJavaScript(),
			htmlFile: "filters_javascript",
		},
		// "filters_plain": {
		// 	template: filters.SlimPlain(),
		// 	htmlFile: "filters_plain",
		// },
		// "filters_escaped": {
		// 	template: filters.SlimEscaped(),
		// 	htmlFile: "filters_escaped",
		// },
		// "filters_preserve": {
		// 	template: filters.SlimPreserve(),
		// 	htmlFile: "filters_preserve",
		// },
		"formatting_intExample": {
			template: formatting.SlimIntExample(),
			htmlFile: "formatting_intExample",
		},
		"formatting_floatExample": {
			template: formatting.SlimFloatExample(),
			htmlFile: "formatting_floatExample",
		},
		"formatting_boolExample": {
			template: formatting.SlimBoolExample(),
			htmlFile: "formatting_boolExample",
		},
		"formatting_stringExample": {
			template: formatting.SlimStringExample(),
			htmlFile: "formatting_stringExample",
		},
		"example_executeCode": {
			template: example.SlimExecuteCode(),
			htmlFile: "example_executeCode",
		},
		"example_renderCode": {
			template: example.SlimRenderCode(),
			htmlFile: "example_renderCode",
		},
		"example_doc": {
			template: example.SlimDoc(),
			htmlFile: "example_doc",
		},
		"example_importExample": {
			template: example.SlimImportExample(),
			htmlFile: "example_importExample",
		},
		"example_conditional": {
			template: example.SlimConditional(),
			htmlFile: "example_conditional",
		},
		"example_shorthandConditional": {
			template: example.SlimShorthandConditional(),
			htmlFile: "example_shorthandConditional",
		},
		"example_shorthandSwitch": {
			template: example.SlimShorthandSwitch(),
			htmlFile: "example_shorthandSwitch",
		},
		"example_interpolateCode": {
			template: example.SlimInterpolateCode(),
			htmlFile: "example_interpolateCode",
		},
		"example_noInterpolation": {
			template: example.SlimNoInterpolation(),
			htmlFile: "example_noInterpolation",
		},
		"example_userDetails": {
			template: func() goht.Template {
				user := example.User{
					Name: "John",
					Age:  30,
				}
				return user.SlimDetails()
			}(),
			htmlFile: "example_userDetails",
		},
		"indents_usingTabs": {
			template: indents.SlimUsingTabs(),
			htmlFile: "indents_usingTabs",
		},
		"tags_inlineTags": {
			template: tags.SlimInlineTags(),
			htmlFile: "tags_inlineTags",
		},
		"tags_specifyTag": {
			template: tags.SlimSpecifyTag(),
			htmlFile: "tags_specifyTag",
		},
		"tags_defaultToDivs": {
			template: tags.SlimDefaultToDivs(),
			htmlFile: "tags_defaultToDivs",
		},
		"tabs_combined": {
			template: tags.SlimCombined(),
			htmlFile: "tags_combined",
		},
		"tags_multipleClasses": {
			template: tags.SlimMultipleClasses(),
			htmlFile: "tags_multipleClasses",
		},
		"tags_selfClosing": {
			template: tags.SlimSelfClosing(),
			htmlFile: "tags_selfClosing",
		},
		"tags_alsoSelfClosing": {
			template: tags.SlimAlsoSelfClosing(),
			htmlFile: "tags_alsoSelfClosing",
		},
		"tags_whitespace": {
			template: tags.SlimWhitespace(),
			htmlFile: "tags_whitespace",
		},
		"tags_addWhitespace": {
			template: tags.SlimAddWhitespace(),
			htmlFile: "tags_addWhitespace",
		},
		"hello_world": {
			template: hello.SlimWorld(),
			htmlFile: "hello_world",
		},
		"unescape_unescapeCode": {
			template: unescape.SlimUnescapeCode(),
			htmlFile: "unescape_unescapeCode",
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
			goldenFileName := filepath.Join("testdata", "slim", tt.htmlFile+".html")
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
