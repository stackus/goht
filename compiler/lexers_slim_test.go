package compiler

import (
	"testing"
)

func Test_SlimText(t *testing.T) {
	tests := map[string]struct {
		input string
		want  []token
	}{
		"simple": {
			input: "@slim test() {\n\t|foobar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tPlainText, lit: "foobar"},
				{typ: tEOF, lit: ""},
			},
		},
		"simple with space": {
			input: "@slim test() {\n\t| foobar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tPlainText, lit: "foobar"},
				{typ: tEOF, lit: ""},
			},
		},
		"simple with spaces": {
			input: "@slim test() {\n\t|   foobar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tPlainText, lit: "  foobar"},
				{typ: tEOF, lit: ""},
			},
		},
		"simple with tab": {
			input: "@slim test() {\n\t|\tfoobar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tPlainText, lit: "foobar"},
				{typ: tEOF, lit: ""},
			},
		},
		"simple with tabs": {
			input: "@slim test() {\n\t|\t\tfoobar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tPlainText, lit: "\tfoobar"},
				{typ: tEOF, lit: ""},
			},
		},
		"multiple lines": {
			input: "@slim test() {\n\t|foobar\n\t|baz",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tPlainText, lit: "foobar"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: "\t"},
				{typ: tPlainText, lit: "baz"},
				{typ: tEOF, lit: ""},
			},
		},
		"multiple indented lines": {
			input: "@slim test() {\n\t|foobar\n\t\tbaz",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tPlainText, lit: "foobar"},
				{typ: tNewLine, lit: "\n"},
				{typ: tPlainText, lit: "baz"},
				{typ: tEOF, lit: ""},
			},
		},
		"multiple indented lines with space": {
			input: "@slim test() {\n\t| foobar\n\t\t baz",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tPlainText, lit: "foobar"},
				{typ: tNewLine, lit: "\n"},
				{typ: tPlainText, lit: "baz"},
				{typ: tEOF, lit: ""},
			},
		},
		"multiple indented lines with tab": {
			input: "@slim test() {\n\t| foobar\n\t\t\tbaz",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tPlainText, lit: "foobar"},
				{typ: tNewLine, lit: "\n"},
				{typ: tPlainText, lit: "baz"},
				{typ: tEOF, lit: ""},
			},
		},
		"multiple indented lines with spaces": {
			input: "@slim test() {\n\t| foobar\n\t\t   baz",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tPlainText, lit: "foobar"},
				{typ: tNewLine, lit: "\n"},
				{typ: tPlainText, lit: "  baz"},
				{typ: tEOF, lit: ""},
			},
		},
		"multiple indented lines with tabs and spaces": {
			input: "@slim test() {\n\t| foobar\n\t\t\t baz",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tPlainText, lit: "foobar"},
				{typ: tNewLine, lit: "\n"},
				{typ: tPlainText, lit: " baz"},
				{typ: tEOF, lit: ""},
			},
		},
		"text with dynamic value": {
			input: "@slim test() {\n\t|foo #{bar} baz",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tPlainText, lit: "foo "},
				{typ: tDynamicText, lit: "bar"},
				{typ: tPlainText, lit: " baz"},
				{typ: tEOF, lit: ""},
			},
		},
		"text and close": {
			input: "@slim test() {\n\t|foo bar\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tPlainText, lit: "foo bar"},
				{typ: tNewLine, lit: "\n"},
				{typ: tTemplateEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := newLexer([]byte(tt.input))
			for _, want := range tt.want {
				got := l.nextToken()
				if got.typ != want.typ || got.lit != want.lit {
					t.Errorf("want %v, got %v", want, got)
				}
			}
		})
	}
}

func Test_SlimTag(t *testing.T) {
	tests := map[string]struct {
		input string
		want  []token
	}{
		"simple": {
			input: "@slim test() {\n\tfoo\n",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tNewLine, lit: "\n"},
				{typ: tEOF, lit: ""},
			},
		},
		"multiple tags": {
			input: "@slim test() {\n\tfoo\n\tbar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"tag and id": {
			input: "@slim test() {\n\tfoo#bar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tId, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"tag and class": {
			input: "@slim test() {\n\tfoo.bar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tClass, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"tag and classes": {
			input: "@slim test() {\n\tfoo.bar.baz",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tClass, lit: "bar"},
				{typ: tClass, lit: "baz"},
				{typ: tEOF, lit: ""},
			},
		},
		"tag and attribute": {
			input: "@slim test() {\n\tfoo{id:\"bar\"}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tAttrName, lit: "id"},
				{typ: tAttrOperator, lit: ":"},
				{typ: tAttrEscapedValue, lit: "\"bar\""},
				{typ: tEOF, lit: ""},
			},
		},
		"tag and text": {
			input: "@slim test() {\n\tfoo bar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tPlainText, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"tag and text and close": {
			input: "@slim test() {\n\tfoo bar\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tPlainText, lit: "bar"},
				{typ: tNewLine, lit: "\n"},
				{typ: tTemplateEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"tag and multiple text lines": {
			input: "@slim test() {\n\tfoo bar\n\t\tbaz",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tPlainText, lit: "bar"},
				{typ: tNewLine, lit: "\n"},
				{typ: tPlainText, lit: "baz"},
				{typ: tEOF, lit: ""},
			},
		},
		"tag and interpolation": {
			input: "@slim test() {\n\tfoo #{bar} baz",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tDynamicText, lit: "bar"},
				{typ: tPlainText, lit: " baz"},
				{typ: tEOF, lit: ""},
			},
		},
		"tag and multiline interpolation": {
			input: "@slim test() {\n\tfoo bar\n\t\t#{baz} qux",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tPlainText, lit: "bar"},
				{typ: tNewLine, lit: "\n"},
				{typ: tDynamicText, lit: "baz"},
				{typ: tPlainText, lit: " qux"},
				{typ: tEOF, lit: ""},
			},
		},
		"tag and output code": {
			input: "@slim test() {\n\tfoo= bar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tScript, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"inlined tag": {
			input: "@slim test() {\n\tfoo: a.bar\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tTag, lit: "a"},
				{typ: tClass, lit: "bar"},
				{typ: tNewLine, lit: "\n"},
				{typ: tTemplateEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := newLexer([]byte(tt.input))
			for _, want := range tt.want {
				got := l.nextToken()
				if got.typ != want.typ || got.lit != want.lit {
					t.Errorf("want %v, got %v", want, got)
				}
			}
		})
	}
}

func Test_SlimId(t *testing.T) {
	tests := map[string]struct {
		input string
		want  []token
	}{
		"simple": {
			input: "@slim test() {\n\t#bar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tId, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"multiple ids": {
			input: "@slim test() {\n\t#foo\n\t#bar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tId, lit: "foo"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: "\t"},
				{typ: tId, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"with underscore": {
			input: "@slim test() {\n\t#foo_bar\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tId, lit: "foo_bar"},
				{typ: tNewLine, lit: "\n"},
				{typ: tTemplateEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"with hyphen": {
			input: "@slim test() {\n\t#foo-bar\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tId, lit: "foo-bar"},
				{typ: tNewLine, lit: "\n"},
				{typ: tTemplateEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"id and class": {
			input: "@slim test() {\n\t#foo.bar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tId, lit: "foo"},
				{typ: tClass, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"id and attribute": {
			input: "@slim test() {\n\t#foo{id:\"bar\"}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tId, lit: "foo"},
				{typ: tAttrName, lit: "id"},
				{typ: tAttrOperator, lit: ":"},
				{typ: tAttrEscapedValue, lit: "\"bar\""},
				{typ: tEOF, lit: ""},
			},
		},
		"id and text": {
			input: "@slim test() {\n\t#foo bar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tId, lit: "foo"},
				{typ: tPlainText, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"id and interpolation": {
			input: "@slim test() {\n\t#foo #{bar} baz",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tId, lit: "foo"},
				{typ: tDynamicText, lit: "bar"},
				{typ: tPlainText, lit: " baz"},
				{typ: tEOF, lit: ""},
			},
		},
		"id and output code": {
			input: "@slim test() {\n\t#foo= bar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tId, lit: "foo"},
				{typ: tScript, lit: "bar"},
			},
		},
		"id and id again": {
			input: "@slim test() {\n\t#foo#bar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tId, lit: "foo"},
				{typ: tId, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := newLexer([]byte(tt.input))
			for _, want := range tt.want {
				got := l.nextToken()
				if got.typ != want.typ || got.lit != want.lit {
					t.Errorf("want %v, got %v", want, got)
				}
			}
		})
	}
}

func Test_SlimClass(t *testing.T) {
	tests := map[string]struct {
		input string
		want  []token
	}{
		"simple": {
			input: "@slim test() {\n\t.bar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tClass, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"multiple classes": {
			input: "@slim test() {\n\t.foo\n\t.bar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tClass, lit: "foo"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: "\t"},
				{typ: tClass, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"class and id": {
			input: "@slim test() {\n\t.foo#bar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tClass, lit: "foo"},
				{typ: tId, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"class and attribute": {
			input: "@slim test() {\n\t.foo{id:\"bar\"}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tClass, lit: "foo"},
				{typ: tAttrName, lit: "id"},
				{typ: tAttrOperator, lit: ":"},
				{typ: tAttrEscapedValue, lit: "\"bar\""},
				{typ: tEOF, lit: ""},
			},
		},
		"class and text": {
			input: "@slim test() {\n\t.foo bar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tClass, lit: "foo"},
				{typ: tPlainText, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"class and interpolation": {
			input: "@slim test() {\n\t.foo #{bar} baz",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tClass, lit: "foo"},
				{typ: tDynamicText, lit: "bar"},
				{typ: tPlainText, lit: " baz"},
				{typ: tEOF, lit: ""},
			},
		},
		"class and output code": {
			input: "@slim test() {\n\t.foo= bar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tClass, lit: "foo"},
				{typ: tScript, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := newLexer([]byte(tt.input))
			for _, want := range tt.want {
				got := l.nextToken()
				if got.typ != want.typ || got.lit != want.lit {
					t.Errorf("want %v, got %v", want, got)
				}
			}
		})
	}
}

func Test_SlimAttributes(t *testing.T) {
	tests := map[string]struct {
		input string
		want  []token
	}{
		"simple": {
			input: "@slim test() {\n\tfoo{id:\"bar\"}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tAttrName, lit: "id"},
				{typ: tAttrOperator, lit: ":"},
				{typ: tAttrEscapedValue, lit: "\"bar\""},
				{typ: tEOF, lit: ""},
			},
		},
		"no tag": {
			input: "@slim test() {\n\t{id:\"bar\"}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tAttrName, lit: "id"},
				{typ: tAttrOperator, lit: ":"},
				{typ: tAttrEscapedValue, lit: "\"bar\""},
				{typ: tEOF, lit: ""},
			},
		},
		"names with dashes": {
			input: "@slim test() {\n\tfoo{data-foo:\"bar\"}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tAttrName, lit: "data-foo"},
				{typ: tAttrOperator, lit: ":"},
				{typ: tAttrEscapedValue, lit: "\"bar\""},
				{typ: tEOF, lit: ""},
			},
		},
		"names with underscores": {
			input: "@slim test() {\n\tfoo{data_foo:\"bar\"}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tAttrName, lit: "data_foo"},
				{typ: tAttrOperator, lit: ":"},
				{typ: tAttrEscapedValue, lit: "\"bar\""},
				{typ: tEOF, lit: ""},
			},
		},
		"names with numbers": {
			input: "@slim test() {\n\tfoo{data1:\"bar\"}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tAttrName, lit: "data1"},
				{typ: tAttrOperator, lit: ":"},
				{typ: tAttrEscapedValue, lit: "\"bar\""},
				{typ: tEOF, lit: ""},
			},
		},
		"names with colons": {
			input: "@slim test() {\n\tfoo{\":x-data\":\"bar\",`x-on:click`:#{onClick}}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tAttrName, lit: ":x-data"},
				{typ: tAttrOperator, lit: ":"},
				{typ: tAttrEscapedValue, lit: "\"bar\""},
				{typ: tAttrName, lit: "x-on:click"},
				{typ: tAttrOperator, lit: ":"},
				{typ: tAttrDynamicValue, lit: "onClick"},
				{typ: tEOF, lit: ""},
			},
		},
		"names with dots": {
			input: "@slim test() {\n\tfoo{data.foo:\"bar\",x.on.click:#{onClick}}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tAttrName, lit: "data.foo"},
				{typ: tAttrOperator, lit: ":"},
				{typ: tAttrEscapedValue, lit: "\"bar\""},
				{typ: tAttrName, lit: "x.on.click"},
				{typ: tAttrOperator, lit: ":"},
				{typ: tAttrDynamicValue, lit: "onClick"},
				{typ: tEOF, lit: ""},
			},
		},
		"names with at signs": {
			input: "@slim test() {\n\tfoo{\"@data\":\"bar\",`x@on@click`:#{onClick}}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tAttrName, lit: "@data"},
				{typ: tAttrOperator, lit: ":"},
				{typ: tAttrEscapedValue, lit: "\"bar\""},
				{typ: tAttrName, lit: "x@on@click"},
				{typ: tAttrOperator, lit: ":"},
				{typ: tAttrDynamicValue, lit: "onClick"},
			},
		},
		"several": {
			input: "@slim test() {\n\tfoo{id:\"bar\", class: `baz` , title : \"qux\"}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tAttrName, lit: "id"},
				{typ: tAttrOperator, lit: ":"},
				{typ: tAttrEscapedValue, lit: "\"bar\""},
				{typ: tAttrName, lit: "class"},
				{typ: tAttrOperator, lit: ":"},
				{typ: tAttrEscapedValue, lit: "`baz`"},
				{typ: tAttrName, lit: "title"},
				{typ: tAttrOperator, lit: ":"},
				{typ: tAttrEscapedValue, lit: "\"qux\""},
				{typ: tEOF, lit: ""},
			},
		},
		"several on multiple lines": {
			input: "@slim test() {\n\tfoo{\n\tid:\"bar\",\n\tclass: `baz` ,\n\ttitle : \"qux\"\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tAttrName, lit: "id"},
				{typ: tAttrOperator, lit: ":"},
				{typ: tAttrEscapedValue, lit: "\"bar\""},
				{typ: tAttrName, lit: "class"},
				{typ: tAttrOperator, lit: ":"},
				{typ: tAttrEscapedValue, lit: "`baz`"},
				{typ: tAttrName, lit: "title"},
				{typ: tAttrOperator, lit: ":"},
				{typ: tAttrEscapedValue, lit: "\"qux\""},
				{typ: tEOF, lit: ""},
			},
		},
		"static value with escaped quotes": {
			input: "@slim test() {\n\tfoo{id:\"bar\\\"baz\"}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tAttrName, lit: "id"},
				{typ: tAttrOperator, lit: ":"},
				{typ: tAttrEscapedValue, lit: "\"bar\\\"baz\""},
				{typ: tEOF, lit: ""},
			},
		},
		"dynamic value": {
			input: "@slim test() {\n\tfoo{id:#{bar}}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tAttrName, lit: "id"},
				{typ: tAttrOperator, lit: ":"},
				{typ: tAttrDynamicValue, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"dynamic value with escaped curly": {
			input: "@slim test() {\n\tfoo{id:#{\"big}\"}, class: #{\"ba\"+'}'}}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tAttrName, lit: "id"},
				{typ: tAttrOperator, lit: ":"},
				{typ: tAttrDynamicValue, lit: "\"big}\""},
				{typ: tAttrName, lit: "class"},
				{typ: tAttrOperator, lit: ":"},
				{typ: tAttrDynamicValue, lit: "\"ba\"+'}'"},
				{typ: tEOF, lit: ""},
			},
		},
		"dynamic values": {
			input: "@slim test() {\n\tfoo{id:#{bar}, class: #{baz}}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tAttrName, lit: "id"},
				{typ: tAttrOperator, lit: ":"},
				{typ: tAttrDynamicValue, lit: "bar"},
				{typ: tAttrName, lit: "class"},
				{typ: tAttrOperator, lit: ":"},
				{typ: tAttrDynamicValue, lit: "baz"},
				{typ: tEOF, lit: ""},
			},
		},
		"boolean attribute": {
			input: "@slim test() {\n\tfoo{bar}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tAttrName, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"boolean operator": {
			input: "@slim test() {\n\tfoo{bar?#{isBar}, baz ? #{isBaz}}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tAttrName, lit: "bar"},
				{typ: tAttrOperator, lit: "?"},
				{typ: tAttrDynamicValue, lit: "isBar"},
				{typ: tAttrName, lit: "baz"},
				{typ: tAttrOperator, lit: "?"},
				{typ: tAttrDynamicValue, lit: "isBaz"},
				{typ: tEOF, lit: ""},
			},
		},
		"attributes command": {
			input: "@slim test() {\n\tfoo{@attributes:#{listA, \"}}\", listB}}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tAttributesCommand, lit: "listA, \"}}\", listB"},
				{typ: tEOF, lit: ""},
			},
		},
		"missing delimiter": {
			input: "@slim test() {\n\tfoo{id\"bar\", class: \"baz\"}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tAttrName, lit: "id"},
				{typ: tError, lit: "unexpected character: '\"'"},
				{typ: tEOF, lit: ""},
			},
		},
		"missing separator": {
			input: "@slim test() {\n\tfoo{id:\"bar\" class: \"baz\"}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tAttrName, lit: "id"},
				{typ: tAttrOperator, lit: ":"},
				{typ: tAttrEscapedValue, lit: "\"bar\""},
				{typ: tError, lit: "unexpected character: c"},
				{typ: tEOF, lit: ""},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := newLexer([]byte(tt.input))
			for _, want := range tt.want {
				got := l.nextToken()
				if got.typ != want.typ || got.lit != want.lit {
					t.Errorf("want %v, got %v", want, got)
				}
			}
		})
	}
}

func Test_SlimWhitespaceAddition(t *testing.T) {
	tests := map[string]struct {
		input string
		want  []token
	}{
		"add after": {
			input: "@slim test() {\n\tp>\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "p"},
				{typ: tAddWhitespaceAfter, lit: ""},
				{typ: tNewLine, lit: "\n"},
				{typ: tTemplateEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"add before": {
			input: "@slim test() {\n\tp<\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "p"},
				{typ: tAddWhitespaceBefore, lit: ""},
				{typ: tNewLine, lit: "\n"},
				{typ: tTemplateEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"add both": {
			input: "@slim test() {\n\tp<>\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "p"},
				{typ: tAddWhitespaceBefore, lit: ""},
				{typ: tAddWhitespaceAfter, lit: ""},
				{typ: tNewLine, lit: "\n"},
				{typ: tTemplateEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := newLexer([]byte(tt.input))
			for _, want := range tt.want {
				got := l.nextToken()
				if got.typ != want.typ || got.lit != want.lit {
					t.Errorf("want %v, got %v", want, got)
				}
			}
		})
	}
}

func TestSlimDoctype(t *testing.T) {
	tests := map[string]struct {
		input string
		want  []token
	}{
		"simple": {
			input: "@slim test() {\n\tdoctype",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tDoctype, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"with type": {
			input: "@slim test() {\n\tdoctype Strict",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tDoctype, lit: "Strict"},
				{typ: tEOF, lit: ""},
			},
		},
		"with content": {
			input: "@slim test() {\n\tdoctype 5\n\thtml\n\t\ttitle foo\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tDoctype, lit: "5"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "html"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: "\t\t"},
				{typ: tTag, lit: "title"},
				{typ: tPlainText, lit: "foo"},
				{typ: tNewLine, lit: "\n"},
				{typ: tTemplateEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := newLexer([]byte(tt.input))
			for _, want := range tt.want {
				got := l.nextToken()
				if got.typ != want.typ || got.lit != want.lit {
					t.Errorf("want %v, got %v", want, got)
				}
			}
		})
	}
}

func Test_SlimVoidTags(t *testing.T) {
	tests := map[string]struct {
		input string
		want  []token
	}{
		"simple": {
			input: "@slim test() {\n\tfoo/",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tVoidTag, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"line is ignored": {
			input: "@slim test() {\n\tfoo/ bar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tError, lit: "self-closing tags can't have content"},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := newLexer([]byte(tt.input))
			for _, want := range tt.want {
				got := l.nextToken()
				if got.typ != want.typ || got.lit != want.lit {
					t.Errorf("want %v, got %v", want, got)
				}
			}
		})
	}
}

func Test_SlimComment(t *testing.T) {
	tests := map[string]struct {
		input string
		want  []token
	}{
		"simple": {
			input: "@slim test() {\n\t/ foo",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tRubyComment, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"html comment": {
			input: "@slim test() {\n\t/! foo",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tComment, lit: "foo"},
				{typ: tEOF, lit: ""},
			},
		},
		"simple with content": {
			input: "@slim test() {\n\t/\n\tp bar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tRubyComment, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "p"},
				{typ: tPlainText, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"html comment with content": {
			input: "@slim test() {\n\t/! foo\n\tp bar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tComment, lit: "foo"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "p"},
				{typ: tPlainText, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"multiline comment": {
			input: "@slim test() {\n\t/foo\n\t\tbar\n\tp baz",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tRubyComment, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "p"},
				{typ: tPlainText, lit: "baz"},
				{typ: tEOF, lit: ""},
			},
		},
		"multiline html comment": {
			input: "@slim test() {\n\t/! foo\n\t\tbar\n\tp baz",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tComment, lit: "foo"},
				{typ: tNewLine, lit: "\n"},
				{typ: tComment, lit: "bar"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "p"},
				{typ: tPlainText, lit: "baz"},
				{typ: tEOF, lit: ""},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := newLexer([]byte(tt.input))
			for _, want := range tt.want {
				got := l.nextToken()
				if got.typ != want.typ || got.lit != want.lit {
					t.Errorf("want %v, got %v", want, got)
				}
			}
		})
	}
}

func Test_SlimOutputCode(t *testing.T) {
	tests := map[string]struct {
		input string
		want  []token
	}{
		"simple": {
			input: "@slim test() {\n\t=foo",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tScript, lit: "foo"},
				{typ: tEOF, lit: ""},
			},
		},
		"with space": {
			input: "@slim test() {\n\t= foo",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tScript, lit: "foo"},
				{typ: tEOF, lit: ""},
			},
		},
		"unescaped": {
			input: "@slim test() {\n\t== foo bar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tUnescaped, lit: ""},
				{typ: tScript, lit: "foo bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"multiline": {
			input: "@slim test() {\n\t= foo,\n\t\tbar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tScript, lit: "foo,\nbar"},
				{typ: tEOF, lit: ""},
			},
		},
		"multiline with backslash": {
			input: "@slim test() {\n\t= foo\\\n\t\tbar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tScript, lit: "foo\nbar"},
				{typ: tEOF, lit: ""},
			},
		},
		"multiline and tag": {
			input: "@slim test() {\n\t= foo,\n\t\tbar\n\tp",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tScript, lit: "foo,\nbar\n"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "p"},
				{typ: tEOF, lit: ""},
			},
		},
		"multiline error": {
			input: "@slim test() {\n\t- foo,\n\t\tbar,\n\tp",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tError, lit: "expected continuation of code"},
				{typ: tEOF, lit: ""},
			},
		},
		"after tag": {
			input: "@slim test() {\n\tfoo= bar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tScript, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"without space": {
			input: "@slim test() {\n\t=foo",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tScript, lit: "foo"},
				{typ: tEOF, lit: ""},
			},
		},
		"with parens": {
			input: "@slim test() {\n\t= foo()",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tScript, lit: "foo()"},
				{typ: tEOF, lit: ""},
			},
		},
		"with render command": {
			input: "@slim test() {\n\t= @render foo(\"bar\")",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tRenderCommand, lit: "foo(\"bar\")"},
				{typ: tEOF, lit: ""},
			},
		},
		"with render command and parens": {
			input: "@slim test() {\n\t= @render() foo(\"bar\")",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tRenderCommand, lit: "foo(\"bar\")"},
				{typ: tEOF, lit: ""},
			},
		},
		"with missing render argument": {
			input: "@slim test() {\n\t= @render",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tError, lit: "render argument expected"},
				{typ: tEOF, lit: ""},
			},
		},
		"with multiline render command": {
			input: "@slim test() {\n\t= @render foo(\"bar\",\n\t\tbaz)",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tRenderCommand, lit: "foo(\"bar\",\nbaz)"},
				{typ: tEOF, lit: ""},
			},
		},
		"with multiline render command slash": {
			input: "@slim test() {\n\t= @render foo(\"bar\",\\\n\t\tbaz)",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tRenderCommand, lit: "foo(\"bar\",\nbaz)"},
				{typ: tEOF, lit: ""},
			},
		},
		"with children command": {
			input: "@slim test() {\n\t= @children",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tChildrenCommand, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"with children command and parens": {
			input: "@slim test() {\n\t= @children()",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tChildrenCommand, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"without any children arguments": {
			input: "@slim test() {\n\t= @children() asdfasdf",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tError, lit: "children command does not accept arguments"},
				{typ: tEOF, lit: ""},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := newLexer([]byte(tt.input))
			for _, want := range tt.want {
				got := l.nextToken()
				if got.typ != want.typ || got.lit != want.lit {
					t.Errorf("want %v, got %v", want, got)
				}
			}
		})
	}
}

func Test_SlimSlotCommand(t *testing.T) {
	tests := map[string]struct {
		input string
		want  []token
	}{
		"simple": {
			input: "@slim test() {\n\t= @slot testing",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tSlotCommand, lit: "testing"},
				{typ: tEOF, lit: ""},
			},
		},
		"with parens": {
			input: "@slim test() {\n\t= @slot() testing",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tSlotCommand, lit: "testing"},
				{typ: tEOF, lit: ""},
			},
		},
		"with default content": {
			input: "@slim test() {\n\t= @slot testing\n\t\tp bar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tSlotCommand, lit: "testing"},
				{typ: tIndent, lit: "\t\t"},
				{typ: tTag, lit: "p"},
				{typ: tPlainText, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"missing name": {
			input: "@slim test() {\n\t= @slot",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tError, lit: "slot name expected"},
				{typ: tEOF, lit: ""},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := newLexer([]byte(tt.input))
			for _, want := range tt.want {
				got := l.nextToken()
				if got.typ != want.typ || got.lit != want.lit {
					t.Errorf("want %v, got %v", want, got)
				}
			}
		})
	}
}

func Test_SlimSilentCode(t *testing.T) {
	tests := map[string]struct {
		input string
		want  []token
	}{
		"simple": {
			input: "@slim test() {\n\t-foo",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tSilentScript, lit: "foo"},
				{typ: tEOF, lit: ""},
			},
		},
		"with space": {
			input: "@slim test() {\n\t- foo",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tSilentScript, lit: "foo"},
				{typ: tEOF, lit: ""},
			},
		},
		"not code": {
			input: "@slim test() {\n\tp - bar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "p"},
				{typ: tPlainText, lit: "- bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"multiline code": {
			input: "@slim test() {\n\t- foo(\\\n\t\tbar,\n\t\tbaz,\n\t\t)\n\tp foo\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tSilentScript, lit: "foo(\nbar,\nbaz,\n)\n"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "p"},
				{typ: tPlainText, lit: "foo"},
				{typ: tNewLine, lit: "\n"},
				{typ: tTemplateEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"multiline with backslash": {
			input: "@slim test() {\n\t= foo\\\n\t\tbar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tScript, lit: "foo\nbar"},
				{typ: tEOF, lit: ""},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := newLexer([]byte(tt.input))
			for _, want := range tt.want {
				got := l.nextToken()
				if got.typ != want.typ || got.lit != want.lit {
					t.Errorf("want %v, got %v", want, got)
				}
			}
		})
	}
}

func Test_SlimFilters(t *testing.T) {
	tests := map[string]struct {
		input string
		want  []token
	}{
		"simple javascript": {
			input: "@slim test() {\n\t:javascript\n\t\tfoo\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tFilterStart, lit: "javascript"},
				{typ: tPlainText, lit: "foo\n"},
				{typ: tFilterEnd, lit: ""},
				{typ: tTemplateEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"simple css": {
			input: "@slim test() {\n\t:css\n\t\tfoo\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tFilterStart, lit: "css"},
				{typ: tPlainText, lit: "foo\n"},
				{typ: tFilterEnd, lit: ""},
				{typ: tTemplateEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"indented": {
			input: "@slim test() {\n\t:javascript\n\t\tfoo\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tFilterStart, lit: "javascript"},
				{typ: tPlainText, lit: "foo\n"},
				{typ: tFilterEnd, lit: ""},
				{typ: tTemplateEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"interpolation": {
			input: "@slim test() {\n\t:javascript\n\t\tlet foo = #{bar}\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tFilterStart, lit: "javascript"},
				{typ: tPlainText, lit: "let foo = "},
				{typ: tDynamicText, lit: "bar"},
				{typ: tPlainText, lit: "\n"},
				{typ: tFilterEnd, lit: ""},
				{typ: tTemplateEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"inside tag": {
			input: `@slim test() {
	p
		:javascript
			console.log("hello world")
}`,
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "p"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: "\t\t"},
				{typ: tFilterStart, lit: "javascript"},
				{typ: tPlainText, lit: "console.log(\"hello world\")\n"},
				{typ: tFilterEnd, lit: ""},
				{typ: tTemplateEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"surrounded by tags": {
			input: `@slim test() {
	p
		:javascript
			console.log("hello world")
		p foo
}`,
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "p"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: "\t\t"},
				{typ: tFilterStart, lit: "javascript"},
				{typ: tPlainText, lit: "console.log(\"hello world\")\n"},
				{typ: tFilterEnd, lit: ""},
				{typ: tIndent, lit: "\t\t"},
				{typ: tTag, lit: "p"},
				{typ: tPlainText, lit: "foo"},
				{typ: tNewLine, lit: "\n"},
				{typ: tTemplateEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"filter following filter": {
			input: `@slim test() {
	:javascript
		console.log("Hello");
	:css
		.color { color: red; }
}`,
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tFilterStart, lit: "javascript"},
				{typ: tPlainText, lit: "console.log(\"Hello\");\n"},
				{typ: tFilterEnd, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tFilterStart, lit: "css"},
				{typ: tPlainText, lit: ".color { color: red; }\n"},
				{typ: tFilterEnd, lit: ""},
				{typ: tTemplateEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := newLexer([]byte(tt.input))
			for _, want := range tt.want {
				got := l.nextToken()
				if got.typ != want.typ || got.lit != want.lit {
					t.Errorf("want %v, got %v", want, got)
				}
			}
		})
	}
}
