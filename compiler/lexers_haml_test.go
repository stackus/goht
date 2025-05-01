package compiler

import (
	"testing"
)

func Test_HamlText(t *testing.T) {
	tests := map[string]struct {
		input string
		want  []token
	}{
		"simple": {
			input: "@goht test() {\n\tfoobar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tPlainText, lit: "foobar"},
				{typ: tEOF, lit: ""},
			},
		},
		"multiple lines": {
			input: "@goht test() {\n\tfoobar\n\tbaz",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tPlainText, lit: "foobar"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: "\t"},
				{typ: tPlainText, lit: "baz"},
				{typ: tEOF, lit: ""},
			},
		},
		"with escaped quotes": {
			input: "@goht test() {\n\t\"foo\\\"bar\"",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tPlainText, lit: "\"foo\\\"bar\""},
				{typ: tEOF, lit: ""},
			},
		},
		"escape control characters": {
			input: "@goht test() {\n\t\\#foo\n\t\\%bar\n\t\\.baz",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tPlainText, lit: "#foo"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: "\t"},
				{typ: tPlainText, lit: "%bar"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: "\t"},
				{typ: tPlainText, lit: ".baz"},
			},
		},
		"text with dynamic value": {
			input: "@goht test() {\n\tfoo #{bar} baz",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tPlainText, lit: "foo "},
				{typ: tDynamicText, lit: "bar"},
				{typ: tPlainText, lit: " baz"},
				{typ: tEOF, lit: ""},
			},
		},
		"escape dynamic text at start of line": {
			input: "@goht test() {\n\t\\#{foo}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tDynamicText, lit: "foo"},
				{typ: tEOF, lit: ""},
			},
		},
		"ignore dynamic text in line": {
			input: "@goht test() {\n\tfoo \\#{bar} \\{f} baz",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tPlainText, lit: "foo #{bar} \\{f} baz"},
				{typ: tEOF, lit: ""},
			},
		},
		"error in dynamic syntax": {
			input: "@goht test() {\n\tfoo #{bar baz",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tPlainText, lit: "foo "},
				{typ: tError, lit: "dynamic text value was not closed: eof"},
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

func Test_HamlTag(t *testing.T) {
	tests := map[string]struct {
		input string
		want  []token
	}{
		"simple": {
			input: "@goht test() {\n\t%foo\n",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tNewLine, lit: "\n"},
				{typ: tEOF, lit: ""},
			},
		},
		"multiple tags": {
			input: "@goht test() {\n\t%foo\n\t%bar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"tag and id": {
			input: "@goht test() {\n\t%foo#bar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tId, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"tag and class": {
			input: "@goht test() {\n\t%foo.bar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tClass, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"tag and attribute": {
			input: "@goht test() {\n\t%foo{id:\"bar\"}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tAttrName, lit: "id"},
				{typ: tAttrOperator, lit: ":"},
				{typ: tAttrEscapedValue, lit: "\"bar\""},
				{typ: tEOF, lit: ""},
			},
		},
		"tag and text": {
			input: "@goht test() {\n\t%foo bar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tPlainText, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"tag and text and close": {
			input: "@goht test() {\n\t%foo bar\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tPlainText, lit: "bar"},
				{typ: tNewLine, lit: "\n"},
				{typ: tTemplateEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		}, "tag and unescaped text": {
			input: "@goht test() {\n\t%foo! bar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tUnescaped, lit: ""},
				{typ: tPlainText, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"tag and output code": {
			input: "@goht test() {\n\t%foo= bar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tScript, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"tag and tag again": {
			input: "@goht test() {\n\t%foo%bar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tPlainText, lit: "%bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"space before tag identifier": {
			input: "@goht test() {\n\t% foo",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tError, lit: "Tag identifier expected"},
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

func Test_HamlId(t *testing.T) {
	tests := map[string]struct {
		input string
		want  []token
	}{
		"simple": {
			input: "@goht test() {\n\t#foo",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tId, lit: "foo"},
				{typ: tEOF, lit: ""},
			},
		},
		"multiple ids": {
			input: "@goht test() {\n\t#foo\n\t#bar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tId, lit: "foo"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: "\t"},
				{typ: tId, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"with underscore": {
			input: "@goht test() {\n\t#foo_bar\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tId, lit: "foo_bar"},
				{typ: tNewLine, lit: "\n"},
				{typ: tTemplateEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"with hyphen": {
			input: "@goht test() {\n\t#foo-bar\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tId, lit: "foo-bar"},
				{typ: tNewLine, lit: "\n"},
				{typ: tTemplateEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"id and class": {
			input: "@goht test() {\n\t#foo.bar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tId, lit: "foo"},
				{typ: tClass, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"id and tag": {
			input: "@goht test() {\n\t#foo%bar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tId, lit: "foo"},
				{typ: tPlainText, lit: "%bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"id and attribute": {
			input: "@goht test() {\n\t#foo{id:\"bar\"}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tId, lit: "foo"},
				{typ: tAttrName, lit: "id"},
				{typ: tAttrOperator, lit: ":"},
				{typ: tAttrEscapedValue, lit: "\"bar\""},
				{typ: tEOF, lit: ""},
			},
		},
		"id and text": {
			input: "@goht test() {\n\t#foo bar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tId, lit: "foo"},
				{typ: tPlainText, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"id and unescaped text": {
			input: "@goht test() {\n\t#foo! bar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tId, lit: "foo"},
				{typ: tUnescaped, lit: ""},
				{typ: tPlainText, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"id and output code": {
			input: "@goht test() {\n\t#foo= bar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tId, lit: "foo"},
				{typ: tScript, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"id and id again": {
			input: "@goht test() {\n\t#foo#bar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tId, lit: "foo"},
				{typ: tId, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"space before id identifier": {
			input: "@goht test() {\n\t# foo",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tError, lit: "Id identifier expected"},
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

func Test_HamlClass(t *testing.T) {
	tests := map[string]struct {
		input string
		want  []token
	}{
		"simple": {
			input: "@goht test() {\n\t.foo",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tClass, lit: "foo"},
				{typ: tEOF, lit: ""},
			},
		},
		"multiple classes": {
			input: "@goht test() {\n\t.foo\n\t.bar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tClass, lit: "foo"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: "\t"},
				{typ: tClass, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"class and id": {
			input: "@goht test() {\n\t.foo#bar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tClass, lit: "foo"},
				{typ: tId, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"class and tag": {
			input: "@goht test() {\n\t.foo%bar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tClass, lit: "foo"},
				{typ: tPlainText, lit: "%bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"class and attribute": {
			input: "@goht test() {\n\t.foo{id:\"bar\"}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tClass, lit: "foo"},
				{typ: tAttrName, lit: "id"},
				{typ: tAttrOperator, lit: ":"},
				{typ: tAttrEscapedValue, lit: "\"bar\""},
				{typ: tEOF, lit: ""},
			},
		},
		"class and text": {
			input: "@goht test() {\n\t.foo bar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tClass, lit: "foo"},
				{typ: tPlainText, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"class and unescaped text": {
			input: "@goht test() {\n\t.foo! bar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tClass, lit: "foo"},
				{typ: tUnescaped, lit: ""},
				{typ: tPlainText, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"class and output code": {
			input: "@goht test() {\n\t.foo= bar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tClass, lit: "foo"},
				{typ: tScript, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"class and class again": {
			input: "@goht test() {\n\t.foo.bar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tClass, lit: "foo"},
				{typ: tClass, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"space before class identifier": {
			input: "@goht test() {\n\t. foo",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tError, lit: "Class identifier expected"},
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

func Test_HamlWhitespaceRemoval(t *testing.T) {
	tests := map[string]struct {
		input string
		want  []token
	}{
		"remove outer": {
			input: "@goht test() {\n\t%p>\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "p"},
				{typ: tNukeOuterWhitespace, lit: ""},
				{typ: tNewLine, lit: "\n"},
				{typ: tTemplateEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"remove inner": {
			input: "@goht test() {\n\t%p<\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "p"},
				{typ: tNukeInnerWhitespace, lit: ""},
				{typ: tNewLine, lit: "\n"},
				{typ: tTemplateEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"remove both": {
			input: "@goht test() {\n\t%p<>\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "p"},
				{typ: tNukeInnerWhitespace, lit: ""},
				{typ: tNukeOuterWhitespace, lit: ""},
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

func Test_HamlObjectRef(t *testing.T) {
	tests := map[string]struct {
		input string
		want  []token
	}{
		"simple": {
			input: "@goht test() {\n\t%p[foo]",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "p"},
				{typ: tObjectRef, lit: "foo"},
				{typ: tEOF, lit: ""},
			},
		},
		"with prefix": {
			input: "@goht test() {\n\t%p[foo, \"bar\"]",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "p"},
				{typ: tObjectRef, lit: "foo, \"bar\""},
				{typ: tEOF, lit: ""},
			},
		},
		"on tag from class": {
			input: "@goht test() {\n\t.foo[foo, \"bar\"]",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tClass, lit: "foo"},
				{typ: tObjectRef, lit: "foo, \"bar\""},
				{typ: tEOF, lit: ""},
			},
		},
		"on tag from id": {
			input: "@goht test() {\n\t#foo[foo, \"bar\"]",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tId, lit: "foo"},
				{typ: tObjectRef, lit: "foo, \"bar\""},
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

func Test_HamlAttributes(t *testing.T) {
	tests := map[string]struct {
		input string
		want  []token
	}{
		"simple": {
			input: "@goht test() {\n\t%foo{id:\"bar\"}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tAttrName, lit: "id"},
				{typ: tAttrOperator, lit: ":"},
				{typ: tAttrEscapedValue, lit: "\"bar\""},
				{typ: tEOF, lit: ""},
			},
		},
		"no tag": {
			input: "@goht test() {\n\t{id:\"bar\"}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tAttrName, lit: "id"},
				{typ: tAttrOperator, lit: ":"},
				{typ: tAttrEscapedValue, lit: "\"bar\""},
				{typ: tEOF, lit: ""},
			},
		},
		"names with dashes": {
			input: "@goht test() {\n\t%foo{data-foo:\"bar\"}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tAttrName, lit: "data-foo"},
				{typ: tAttrOperator, lit: ":"},
				{typ: tAttrEscapedValue, lit: "\"bar\""},
				{typ: tEOF, lit: ""},
			},
		},
		"names with underscores": {
			input: "@goht test() {\n\t%foo{data_foo:\"bar\"}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tAttrName, lit: "data_foo"},
				{typ: tAttrOperator, lit: ":"},
				{typ: tAttrEscapedValue, lit: "\"bar\""},
				{typ: tEOF, lit: ""},
			},
		},
		"names with numbers": {
			input: "@goht test() {\n\t%foo{data1:\"bar\"}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tAttrName, lit: "data1"},
				{typ: tAttrOperator, lit: ":"},
				{typ: tAttrEscapedValue, lit: "\"bar\""},
				{typ: tEOF, lit: ""},
			},
		},
		"names with colons": {
			input: "@goht test() {\n\t%foo{\":x-data\":\"bar\",`x-on:click`:#{onClick}}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
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
			input: "@goht test() {\n\t%foo{data.foo:\"bar\",x.on.click:#{onClick}}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
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
			input: "@goht test() {\n\t%foo{\"@data\":\"bar\",`x@on@click`:#{onClick}}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
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
			input: "@goht test() {\n\t%foo{id:\"bar\", class: `baz` , title : \"qux\"}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
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
			input: "@goht test() {\n\t%foo{\n\tid:\"bar\",\n\tclass: `baz` ,\n\ttitle : \"qux\"\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
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
			input: "@goht test() {\n\t%foo{id:\"bar\\\"baz\"}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tAttrName, lit: "id"},
				{typ: tAttrOperator, lit: ":"},
				{typ: tAttrEscapedValue, lit: "\"bar\\\"baz\""},
				{typ: tEOF, lit: ""},
			},
		},
		"dynamic value": {
			input: "@goht test() {\n\t%foo{id:#{bar}}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tAttrName, lit: "id"},
				{typ: tAttrOperator, lit: ":"},
				{typ: tAttrDynamicValue, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"dynamic value with escaped curly": {
			input: "@goht test() {\n\t%foo{id:#{\"big}\"}, class: #{\"ba\"+'}'}}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
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
			input: "@goht test() {\n\t%foo{id:#{bar}, class: #{baz}}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
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
			input: "@goht test() {\n\t%foo{bar}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tAttrName, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"boolean operator": {
			input: "@goht test() {\n\t%foo{bar?#{isBar}, baz ? #{isBaz}}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
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
			input: "@goht test() {\n\t%foo{@attributes:#{listA, \"}}\", listB}}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tAttributesCommand, lit: "listA, \"}}\", listB"},
				{typ: tEOF, lit: ""},
			},
		},
		"missing delimiter": {
			input: "@goht test() {\n\t%foo{id\"bar\", class: \"baz\"}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tAttrName, lit: "id"},
				{typ: tError, lit: "unexpected character: '\"'"},
				{typ: tEOF, lit: ""},
			},
		},
		"missing separator": {
			input: "@goht test() {\n\t%foo{id:\"bar\" class: \"baz\"}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
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

func Test_HamlDoctype(t *testing.T) {
	tests := map[string]struct {
		input string
		want  []token
	}{
		"simple": {
			input: "@goht test() {\n\t!!!",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tDoctype, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"with type": {
			input: "@goht test() {\n\t!!! Strict",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tDoctype, lit: "Strict"},
				{typ: tEOF, lit: ""},
			},
		},
		"not a doctype": {
			input: "@goht test() {\n\t!foo",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tUnescaped, lit: ""},
				{typ: tPlainText, lit: "foo"},
				{typ: tEOF, lit: ""},
			},
		},
		"doctype with content": {
			input: "@goht test() {\n\t!!! 5\n\t%html\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tDoctype, lit: "5"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "html"},
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

func Test_HamlUnescaped(t *testing.T) {
	tests := map[string]struct {
		input string
		want  []token
	}{
		"simple": {
			input: "@goht test() {\n\t!foo",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tUnescaped, lit: ""},
				{typ: tPlainText, lit: "foo"},
				{typ: tEOF, lit: ""},
			},
		},
		"with space": {
			input: "@goht test() {\n\t! foo",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tUnescaped, lit: ""},
				{typ: tPlainText, lit: "foo"},
				{typ: tEOF, lit: ""},
			},
		},
		"dynamic text": {
			input: "@goht test() {\n\t! #{foo}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tUnescaped, lit: ""},
				{typ: tDynamicText, lit: "foo"},
				{typ: tEOF, lit: ""},
			},
		},
		"unescaped code": {
			input: "@goht test() {\n\t!= foo",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tUnescaped, lit: ""},
				{typ: tScript, lit: "foo"},
			},
		},
		"not unescaped": {
			input: "@goht test() {\n\t%p ! foo",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "p"},
				{typ: tPlainText, lit: "! foo"},
				{typ: tEOF, lit: ""},
			},
		},
		"tag with unescaped text": {
			input: "@goht test() {\n\t%foo! bar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tUnescaped, lit: ""},
				{typ: tPlainText, lit: "bar"},
			},
		},
		"tag with unescaped code": {
			input: "@goht test() {\n\t%foo!= bar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tUnescaped, lit: ""},
				{typ: tScript, lit: "bar"},
			},
		},
		"mixed with tags": {
			input: `@goht test() {
	- var html = "<em>is</em>"
	%p This #{html} HTML.
	%p! This #{html} HTML.
}`,
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tSilentScript, lit: "var html = \"<em>is</em>\""},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "p"},
				{typ: tPlainText, lit: "This "},
				{typ: tDynamicText, lit: "html"},
				{typ: tPlainText, lit: " HTML."},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "p"},
				{typ: tUnescaped, lit: ""},
				{typ: tPlainText, lit: "This "},
				{typ: tDynamicText, lit: "html"},
				{typ: tPlainText, lit: " HTML."},
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

func Test_HamlComment(t *testing.T) {
	tests := map[string]struct {
		input string
		want  []token
	}{
		"simple": {
			input: "@goht test() {\n\t/ foo",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tComment, lit: "foo"},
				{typ: tEOF, lit: ""},
			},
		},
		"with content": {
			input: "@goht test() {\n\t/\n\t%p bar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tComment, lit: ""},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "p"},
				{typ: tPlainText, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"haml comment": {
			input: "@goht test() {\n\t-# foo",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tRubyComment, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"nested haml comment": {
			input: `@goht test() {
	%p foo
	-#
		%p bar
}`,
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "p"},
				{typ: tPlainText, lit: "foo"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: "\t"},
				{typ: tRubyComment, lit: ""},
				{typ: tTemplateEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"both comments": {
			input: "@goht test() {\n\t/ foo\n\t-# bar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tComment, lit: "foo"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: "\t"},
				{typ: tRubyComment, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"both comments with nested content": {
			input: `@goht test() {
	/
		foo
	-#
		%p bar
}`,
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tComment, lit: ""},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: "\t\t"},
				{typ: tPlainText, lit: "foo"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: "\t"},
				{typ: tRubyComment, lit: ""},
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

func Test_HamlVoidTags(t *testing.T) {
	tests := map[string]struct {
		input string
		want  []token
	}{
		"simple": {
			input: "@goht test() {\n\t%foo/",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tVoidTag, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"line is ignored": {
			input: "@goht test() {\n\t%foo/ bar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
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

func Test_HamlOutputCode(t *testing.T) {
	tests := map[string]struct {
		input string
		want  []token
	}{
		"simple": {
			input: "@goht test() {\n\t=foo",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tScript, lit: "foo"},
				{typ: tEOF, lit: ""},
			},
		},
		"with space": {
			input: "@goht test() {\n\t= foo",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tScript, lit: "foo"},
				{typ: tEOF, lit: ""},
			},
		},
		"multiline": {
			input: "@goht test() {\n\t= foo,\n\t\tbar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tScript, lit: "foo,\nbar"},
				{typ: tEOF, lit: ""},
			},
		},
		"multiline with backslash": {
			input: "@goht test() {\n\t= foo\\\n\t\tbar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tScript, lit: "foo\nbar"},
				{typ: tEOF, lit: ""},
			},
		},
		"multiline and tag": {
			input: "@goht test() {\n\t= foo,\n\t\tbar\n\t%p",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tScript, lit: "foo,\nbar\n"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "p"},
				{typ: tEOF, lit: ""},
			},
		},
		"after tag": {
			input: "@goht test() {\n\t%foo= bar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tScript, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"without space": {
			input: "@goht test() {\n\t=foo",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tScript, lit: "foo"},
				{typ: tEOF, lit: ""},
			},
		},
		"with parens": {
			input: "@goht test() {\n\t= foo()",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tScript, lit: "foo()"},
				{typ: tEOF, lit: ""},
			},
		},
		"with missing command": {
			input: "@goht test() {\n\t= @ foo(\"bar\")",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tError, lit: "command code expected"},
				{typ: tEOF, lit: ""},
			},
		},
		"with unknown command": {
			input: "@goht test() {\n\t= @unknown foo(\"bar\")",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tError, lit: "unknown command: unknown"},
				{typ: tEOF, lit: ""},
			},
		},
		"with render command": {
			input: "@goht test() {\n\t= @render foo(\"bar\")",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tRenderCommand, lit: "foo(\"bar\")"},
				{typ: tEOF, lit: ""},
			},
		},
		"with multiline render command": {
			input: "@goht test() {\n\t= @render foo(\"bar\",\n\t\tbaz)",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tRenderCommand, lit: "foo(\"bar\",\nbaz)"},
				{typ: tEOF, lit: ""},
			},
		},
		"with multiline render command slash": {
			input: "@goht test() {\n\t= @render foo(\"bar\",\\\n\t\tbaz)",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tRenderCommand, lit: "foo(\"bar\",\nbaz)"},
				{typ: tEOF, lit: ""},
			},
		},
		"with render command and parens": {
			input: "@goht test() {\n\t= @render() foo(\"bar\")",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tRenderCommand, lit: "foo(\"bar\")"},
				{typ: tEOF, lit: ""},
			},
		},
		"with missing render argument": {
			input: "@goht test() {\n\t= @render",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tError, lit: "render argument expected"},
				{typ: tEOF, lit: ""},
			},
		},
		"with children command": {
			input: "@goht test() {\n\t= @children",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tChildrenCommand, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"with children command and parens": {
			input: "@goht test() {\n\t= @children()",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tChildrenCommand, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"without any children arguments": {
			input: "@goht test() {\n\t= @children() asdfasdf",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
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

func Test_HamlSlotCommand(t *testing.T) {
	tests := map[string]struct {
		input string
		want  []token
	}{
		"simple": {
			input: "@goht test() {\n\t= @slot testing",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tSlotCommand, lit: "testing"},
				{typ: tEOF, lit: ""},
			},
		},
		"with parens": {
			input: "@goht test() {\n\t= @slot() testing",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tSlotCommand, lit: "testing"},
				{typ: tEOF, lit: ""},
			},
		},
		"with default content": {
			input: "@goht test() {\n\t= @slot testing\n\t\t%p bar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tSlotCommand, lit: "testing"},
				{typ: tIndent, lit: "\t\t"},
				{typ: tTag, lit: "p"},
				{typ: tPlainText, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"missing name": {
			input: "@goht test() {\n\t= @slot",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
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

func Test_HamlExecuteCode(t *testing.T) {
	tests := map[string]struct {
		input string
		want  []token
	}{
		"simple": {
			input: "@goht test() {\n\t-foo",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tSilentScript, lit: "foo"},
				{typ: tEOF, lit: ""},
			},
		},
		"with space": {
			input: "@goht test() {\n\t- foo",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tSilentScript, lit: "foo"},
				{typ: tEOF, lit: ""},
			},
		},
		"multiline": {
			input: "@goht test() {\n\t- foo,\n\t\tbar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tSilentScript, lit: "foo,\nbar"},
				{typ: tEOF, lit: ""},
			},
		},
		"multiline with backslash": {
			input: "@goht test() {\n\t- foo\\\n\t\tbar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tSilentScript, lit: "foo\nbar"},
				{typ: tEOF, lit: ""},
			},
		},
		"multiline and tag": {
			input: "@goht test() {\n\t- foo,\n\t\tbar\n\t%p",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tSilentScript, lit: "foo,\nbar\n"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "p"},
				{typ: tEOF, lit: ""},
			},
		},
		"multiline error": {
			input: "@goht test() {\n\t- foo,\n\t\tbar,\n\t%p",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tError, lit: "expected continuation of code"},
				{typ: tEOF, lit: ""},
			},
		},
		"not code": {
			input: "@goht test() {\n\t%foo- bar\n\t#foo-bar bar\n\t%p - bar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo-"},
				{typ: tPlainText, lit: "bar"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: "\t"},
				{typ: tId, lit: "foo-bar"},
				{typ: tPlainText, lit: "bar"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "p"},
				{typ: tPlainText, lit: "- bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"without space": {
			input: "@goht test() {\n\t-foo",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tSilentScript, lit: "foo"},
				{typ: tEOF, lit: ""},
			},
		},
		"nested": {
			input: `@goht test() {
	- if foo != "" {
		%p Foo exists and is #{foo}.
	- }
`,
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tSilentScript, lit: "if foo != \"\" {"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: "\t\t"},
				{typ: tTag, lit: "p"},
				{typ: tPlainText, lit: "Foo exists and is "},
				{typ: tDynamicText, lit: "foo"},
				{typ: tPlainText, lit: "."},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: "\t"},
				{typ: tSilentScript, lit: "}"},
				{typ: tNewLine, lit: "\n"},
				{typ: tEOF, lit: ""},
			},
		},
		"ruby style comment": {
			input: "@goht test() {\n\t-# comment\n\t- foo",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tRubyComment, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tSilentScript, lit: "foo"},
				{typ: tEOF, lit: ""},
			},
		},
		"nested ruby style comment": {
			input: "@goht test() {\n\t-#\n\t\tcomment\n\t- foo",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tRubyComment, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tSilentScript, lit: "foo"},
				{typ: tEOF, lit: ""},
			},
		},
		"with receiver": {
			input: "@goht (t Tester) test() {\n\t- t.bar",
			want: []token{
				{typ: tTemplateStart, lit: "(t Tester) test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tSilentScript, lit: "t.bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"with receiver and interface": {
			input: "@goht (t Tester) test(v interface{}) {\n\t- t.bar",
			want: []token{
				{typ: tTemplateStart, lit: "(t Tester) test(v interface{})"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tSilentScript, lit: "t.bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"with receiver and interface with methods": {
			input: "@goht (t Tester) test(v interface{ Foo() string }) {\n\t- t.bar",
			want: []token{
				{typ: tTemplateStart, lit: "(t Tester) test(v interface{ Foo() string })"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tSilentScript, lit: "t.bar"},
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

func Test_HamlIndent(t *testing.T) {
	tests := map[string]struct {
		input string
		want  []token
	}{
		"simple": {
			input: "@goht test() {\n\t%foo\n\t\tbar\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: "\t\t"},
				{typ: tPlainText, lit: "bar"},
				{typ: tNewLine, lit: "\n"},
				{typ: tTemplateEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"indented too deep": {
			input: "@goht test() {\n\t%foo\n\t\t\tbar\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tNewLine, lit: "\n"},
				{typ: tError, lit: "the line was indented 2 levels deeper than the previous line"},
			},
		},
		"different indents": {
			input: `@goht test() {
	%p1
		%span one
		%p2
			%p3 three
}`,
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "p1"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: "\t\t"},
				{typ: tTag, lit: "span"},
				{typ: tPlainText, lit: "one"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: "\t\t"},
				{typ: tTag, lit: "p2"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: "\t\t\t"},
				{typ: tTag, lit: "p3"},
				{typ: tPlainText, lit: "three"},
				{typ: tNewLine, lit: "\n"},
				{typ: tTemplateEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"indents must be tabs": {
			input: `@goht test() {
	%p1
    %p2
}`,
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "p1"},
				{typ: tNewLine, lit: "\n"},
				{typ: tError, lit: "the line was indented using spaces, templates must be indented using tabs"},
				{typ: tEOF, lit: ""},
			},
		},
		"wrong indent size": {
			input: `@goht test() {
	%p1
		%p2
				%p3
}`,
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "p1"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: "\t\t"},
				{typ: tTag, lit: "p2"},
				{typ: tNewLine, lit: "\n"},
				{typ: tError, lit: "the line was indented 2 levels deeper than the previous line"},
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

func Test_HamlFilters(t *testing.T) {
	tests := map[string]struct {
		input string
		want  []token
	}{
		"simple javascript": {
			input: "@goht test() {\n\t:javascript\n\t\tfoo\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tFilterStart, lit: "javascript"},
				{typ: tPlainText, lit: "foo\n"},
				{typ: tFilterEnd, lit: ""},
				{typ: tTemplateEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"simple css": {
			input: "@goht test() {\n\t:css\n\t\tfoo\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tFilterStart, lit: "css"},
				{typ: tPlainText, lit: "foo\n"},
				{typ: tFilterEnd, lit: ""},
				{typ: tTemplateEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"multiline css": {
			input: "@goht test() {\n\t:css\n\t\tfoo\n\t\tbar\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tFilterStart, lit: "css"},
				{typ: tPlainText, lit: "foo\n"},
				{typ: tPlainText, lit: "bar\n"},
				{typ: tFilterEnd, lit: ""},
				{typ: tTemplateEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"indented": {
			input: "@goht test() {\n\t:javascript\n\t\tfoo\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tFilterStart, lit: "javascript"},
				{typ: tPlainText, lit: "foo\n"},
				{typ: tFilterEnd, lit: ""},
				{typ: tTemplateEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"missing filter": {
			input: "@goht test() {\n\t:\n\t\tfoo\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tError, lit: "filter name expected"},
				{typ: tEOF, lit: ""},
			},
		},
		"unknown filter": {
			input: "@goht test() {\n\t:unknown\n\t\tfoo\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tError, lit: "unknown filter: unknown"},
				{typ: tEOF, lit: ""},
			},
		},
		"eol": {
			input: "@goht test() {\n\t:javascript\n",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tFilterStart, lit: "javascript"},
				{typ: tEOF, lit: ""},
			},
		},
		"interpolation": {
			input: "@goht test() {\n\t:javascript\n\t\tfoo #{bar}\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tFilterStart, lit: "javascript"},
				{typ: tPlainText, lit: "foo "},
				{typ: tDynamicText, lit: "bar"},
				{typ: tPlainText, lit: "\n"},
				{typ: tFilterEnd, lit: ""},
				{typ: tTemplateEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"interpolation not closed": {
			input: "@goht test() {\n\t:javascript\n\t\tfoo #{bar\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tFilterStart, lit: "javascript"},
				{typ: tPlainText, lit: "foo "},
				{typ: tError, lit: "dynamic text value was not closed: eof"},
				{typ: tEOF, lit: ""},
			},
		},
		"is not interpolation": {
			input: "@goht test() {\n\t:javascript\n\t\tfoo # {bar}\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
				{typ: tIndent, lit: "\t"},
				{typ: tFilterStart, lit: "javascript"},
				{typ: tPlainText, lit: "foo # {bar}\n"},
				{typ: tFilterEnd, lit: ""},
				{typ: tTemplateEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"inside tag": {
			input: `@goht test() {
	%p
		:javascript
			console.log("hello world")
}`,
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
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
			input: `@goht test() {
	%p
		:javascript
			console.log("hello world")
		%p foo
}`,
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
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
			input: `@goht test() {
	:javascript
		console.log("Hello");
	:css
		.color { color: red; }
}`,
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tKeepNewlines, lit: ""},
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
