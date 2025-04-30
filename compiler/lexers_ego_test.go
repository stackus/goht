package compiler

import (
	"testing"
)

func Test_EgoText(t *testing.T) {
	tests := map[string]struct {
		input string
		want  []token
	}{
		"simple": {
			input: "@ego test() {\n\tfoobar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tRawText, lit: "foobar"},
				{typ: tEOF, lit: ""},
			},
		},
		"multiple lines": {
			input: "@ego test() {\n\tfoobar\n\tbaz",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tRawText, lit: "foobar\nbaz"},
				{typ: tEOF, lit: ""},
			},
		},
		"multiple indented lines": {
			input: "@ego test() {\n\tfoobar\n\t\tbaz",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tRawText, lit: "foobar\n\tbaz"},
				{typ: tEOF, lit: ""},
			},
		},
		"html": {
			input: "@ego test() {\n\tfoo <bar>\n\tbaz",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tRawText, lit: "foo <bar>\nbaz"},
				{typ: tEOF, lit: ""},
			},
		},
		"text and close": {
			input: "@ego test() {\n\tfoo bar\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tRawText, lit: "foo bar"},
				{typ: tTemplateEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"literal opening tag": {
			input: "@ego test() {\n\tfoo <%% bar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tRawText, lit: "foo <% bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"allow leading spaces": {
			input: "@ego test() {\n\t  foobar\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tRawText, lit: "  foobar"},
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

func Test_EgoComment(t *testing.T) {
	tests := map[string]struct {
		input string
		want  []token
	}{
		"simple": {
			input: "@ego test() {\n\t<%# foobar %>",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tEOF, lit: ""},
			},
		},
		"with text": {
			input: "@ego test() {\n\t<%# foo bar %>\n\tbaz",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tRawText, lit: "\nbaz"},
				{typ: tEOF, lit: ""},
			},
		},
		"multiple lines": {
			input: "@ego test() {\n\t<%# foo\n\n\tbar %>\n\tbaz",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tRawText, lit: "\nbaz"},
				{typ: tEOF, lit: ""},
			},
		},
		"with text and close": {
			input: "@ego test() {\n\t<%# foo bar %>\n\tbaz\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tRawText, lit: "\nbaz"},
				{typ: tTemplateEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"with html": {
			input: "@ego test() {\n\t<%# foobar> %>\n\t<div>baz</div>\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tRawText, lit: "\n<div>baz</div>"},
				{typ: tTemplateEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"with text and html": {
			input: "@ego test() {\n\tfizz<%# foobar> %>\n\t<div>baz</div>\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tRawText, lit: "fizz"},
				{typ: tRawText, lit: "\n<div>baz</div>"},
				{typ: tTemplateEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"eol": {
			input: "@ego test() {\n\t<%#",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tError, lit: "unexpected EOF in tag"},
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

func Test_EgoOutput(t *testing.T) {
	tests := map[string]struct {
		input string
		want  []token
	}{
		"simple": {
			input: "@ego test() {\n\t<%= foobar %>",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tScript, lit: "foobar"},
				{typ: tEOF, lit: ""},
			},
		},
		"with text": {
			input: "@ego test() {\n\t<%= foobar %>\n\tbaz",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tScript, lit: "foobar"},
				{typ: tRawText, lit: "\nbaz"},
				{typ: tEOF, lit: ""},
			},
		},
		"multiple lines": {
			input: "@ego test() {\n\t<%= foo(bar,\n\t\tfizz) %>\n\tbaz",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tScript, lit: "foo(bar,\n\tfizz)"},
				{typ: tRawText, lit: "\nbaz"},
				{typ: tEOF, lit: ""},
			},
		},
		"with text and close": {
			input: "@ego test() {\n\t<%= foobar %>\n\tbaz\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tScript, lit: "foobar"},
				{typ: tRawText, lit: "\nbaz"},
				{typ: tTemplateEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"with html": {
			input: "@ego test() {\n\t<%= foobar %>\n\t<div>baz</div>\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tScript, lit: "foobar"},
				{typ: tRawText, lit: "\n<div>baz</div>"},
				{typ: tTemplateEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"with text and html": {
			input: "@ego test() {\n\tfizz<%= foobar %>\n\t<div>baz</div>\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tRawText, lit: "fizz"},
				{typ: tScript, lit: "foobar"},
				{typ: tRawText, lit: "\n<div>baz</div>"},
				{typ: tTemplateEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"in html": {
			input: "@ego test() {\n\t<title><%= title %></title>",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tRawText, lit: "<title>"},
				{typ: tScript, lit: "title"},
				{typ: tRawText, lit: "</title>"},
				{typ: tEOF, lit: ""},
			},
		},
		"eol": {
			input: "@ego test() {\n\t<%=",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tError, lit: "unexpected EOF in tag"},
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

func Test_EgoUnescapedOutput(t *testing.T) {
	tests := map[string]struct {
		input string
		want  []token
	}{
		"simple": {
			input: "@ego test() {\n\t<%! foobar %>",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tUnescaped, lit: ""},
				{typ: tScript, lit: "foobar"},
				{typ: tEOF, lit: ""},
			},
		},
		"with text": {
			input: "@ego test() {\n\t<%! foobar %>\n\tbaz",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tUnescaped, lit: ""},
				{typ: tScript, lit: "foobar"},
				{typ: tRawText, lit: "\nbaz"},
				{typ: tEOF, lit: ""},
			},
		},
		"with text and close": {
			input: "@ego test() {\n\t<%! foobar %>\n\tbaz\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tUnescaped, lit: ""},
				{typ: tScript, lit: "foobar"},
				{typ: tRawText, lit: "\nbaz"},
				{typ: tTemplateEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"with html": {
			input: "@ego test() {\n\t<%! foobar %>\n\t<div>baz</div>\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tUnescaped, lit: ""},
				{typ: tScript, lit: "foobar"},
				{typ: tRawText, lit: "\n<div>baz</div>"},
				{typ: tTemplateEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"with text and html": {
			input: "@ego test() {\n\tfizz<%! foobar %>\n\t<div>baz</div>\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tRawText, lit: "fizz"},
				{typ: tUnescaped, lit: ""},
				{typ: tScript, lit: "foobar"},
				{typ: tRawText, lit: "\n<div>baz</div>"},
				{typ: tTemplateEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"eol": {
			input: "@ego test() {\n\t<%!",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tUnescaped, lit: ""},
				{typ: tError, lit: "unexpected EOF in tag"},
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

func Test_EgoScript(t *testing.T) {
	tests := map[string]struct {
		input string
		want  []token
	}{
		"simple": {
			input: "@ego test() {\n\t<% foobar %>",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tSilentScript, lit: "foobar"},
				{typ: tEOF, lit: ""},
			},
		},
		"with text": {
			input: "@ego test() {\n\t<% foobar %>\n\tbaz",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tSilentScript, lit: "foobar"},
				{typ: tRawText, lit: "\nbaz"},
				{typ: tEOF, lit: ""},
			},
		},
		"with text and close": {
			input: "@ego test() {\n\t<% foobar %>\n\tbaz\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tSilentScript, lit: "foobar"},
				{typ: tRawText, lit: "\nbaz"},
				{typ: tTemplateEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"with html": {
			input: "@ego test() {\n\t<% foobar %>\n\t<div>baz</div>\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tSilentScript, lit: "foobar"},
				{typ: tRawText, lit: "\n<div>baz</div>"},
				{typ: tTemplateEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"with text and html": {
			input: "@ego test() {\n\tfizz<% foobar %>\n\t<div>baz</div>\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tRawText, lit: "fizz"},
				{typ: tSilentScript, lit: "foobar"},
				{typ: tRawText, lit: "\n<div>baz</div>"},
				{typ: tTemplateEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"increment indent": {
			input: "@ego test() {\n\t<% { %>",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tSilentScript, lit: "{"},
				{typ: tIndent, lit: "\t"},
				{typ: tEOF, lit: ""},
			},
		},
		"increment and decrement indent": {
			input: "@ego test() {\n\t<% { %>\n\t<% } %>",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tSilentScript, lit: "{"},
				{typ: tIndent, lit: "\t"},
				{typ: tRawText, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tSilentScript, lit: "}"},
				{typ: tEOF, lit: ""},
			},
		},
		"eol": {
			input: "@ego test() {\n\t<%",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tError, lit: "unexpected EOF in tag"},
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

func Test_EgoRenderCommand(t *testing.T) {
	tests := map[string]struct {
		input string
		want  []token
	}{
		"simple": {
			input: "@ego test() {\n\t<%@ render foo() %>",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tRenderCommand, lit: "foo()"},
				{typ: tEOF, lit: ""},
			},
		},
		"with text": {
			input: "@ego test() {\n\t<%@ render foo() %>\n\tbaz",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tRenderCommand, lit: "foo()"},
				{typ: tRawText, lit: "\nbaz"},
				{typ: tEOF, lit: ""},
			},
		},
		"with text and close": {
			input: "@ego test() {\n\t<%@ render foo() %>\n\tbaz\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tRenderCommand, lit: "foo()"},
				{typ: tRawText, lit: "\nbaz"},
				{typ: tTemplateEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"with html": {
			input: "@ego test() {\n\t<%@ render foo() %>\n\t<div>baz</div>\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tRenderCommand, lit: "foo()"},
				{typ: tRawText, lit: "\n<div>baz</div>"},
				{typ: tTemplateEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"with text and html": {
			input: "@ego test() {\n\tfizz<%@ render foo() %>\n\t<div>baz</div>\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tRawText, lit: "fizz"},
				{typ: tRenderCommand, lit: "foo()"},
				{typ: tRawText, lit: "\n<div>baz</div>"},
				{typ: tTemplateEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"increment indent": {
			input: "@ego test() {\n\t<%@ render foo() { %>",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tRenderCommand, lit: "foo() {"},
				{typ: tIndent, lit: "\t"},
				{typ: tEOF, lit: ""},
			},
		},
		"with content": {
			input: "@ego test() {\n\t<%@ render foo() { %>\n\t\tbar\n\t<% } %>\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tRenderCommand, lit: "foo() {"},
				{typ: tIndent, lit: "\t"},
				{typ: tRawText, lit: "\n\tbar\n"},
				{typ: tIndent, lit: ""},
				{typ: tSilentScript, lit: "}"},
				{typ: tRawText, lit: ""},
				{typ: tTemplateEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"with scripts": {
			input: "@ego test() {\n\t<%@ render foo() { %>\n\t\t<% bar %>\n\t<% } %>\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tRenderCommand, lit: "foo() {"},
				{typ: tIndent, lit: "\t"},
				{typ: tRawText, lit: "\n\t"},
				{typ: tSilentScript, lit: "bar"},
				{typ: tRawText, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tSilentScript, lit: "}"},
				{typ: tRawText, lit: ""},
				{typ: tTemplateEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"nested in script": {
			input: "@ego test() {\n\t<% if true { %>\n\t\t<%@ render foo() { %>\n\t\tbar\n\t\t<% } %>\n\t<% } %>\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tSilentScript, lit: "if true {"},
				{typ: tIndent, lit: "\t"},
				{typ: tRawText, lit: "\n\t"},
				{typ: tRenderCommand, lit: "foo() {"},
				{typ: tIndent, lit: "\t\t"},
				{typ: tRawText, lit: "\n\tbar\n\t"},
				{typ: tIndent, lit: "\t"},
				{typ: tSilentScript, lit: "}"},
				{typ: tRawText, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tSilentScript, lit: "}"},
				{typ: tRawText, lit: ""},
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

func Test_EgoChildrenCommand(t *testing.T) {
	tests := map[string]struct {
		input string
		want  []token
	}{
		"simple": {
			input: "@ego test() {\n\t<%@ children %>",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tChildrenCommand, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"with text": {
			input: "@ego test() {\n\t<%@ children %>\n\tbaz",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tChildrenCommand, lit: ""},
				{typ: tRawText, lit: "\nbaz"},
				{typ: tEOF, lit: ""},
			},
		},
		"with text and close": {
			input: "@ego test() {\n\t<%@ children %>\n\tbaz\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tChildrenCommand, lit: ""},
				{typ: tRawText, lit: "\nbaz"},
				{typ: tTemplateEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"with html": {
			input: "@ego test() {\n\t<%@ children %>\n\t<div>baz</div>\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tChildrenCommand, lit: ""},
				{typ: tRawText, lit: "\n<div>baz</div>"},
				{typ: tTemplateEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"extra text": {
			input: "@ego test() {\n\t<%@ children foobar %>",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tError, lit: "unexpected content in children command: \"foobar \""},
			},
		},
		"unknown command": {
			input: "@ego test() {\n\t<%@ unknown %>",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tError, lit: "unknown command: \"unknown\""},
			},
		},
		"missing command": {
			input: "@ego test() {\n\t<%@ %>",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tError, lit: "command name expected"},
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

func Test_EgoTrimWhitespace(t *testing.T) {
	tests := map[string]struct {
		input string
		want  []token
	}{
		"leading whitespace": {
			input: "@ego test() {\n\tfoo \t\t\n\t   <%- bar %>\n\tbaz",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tRawText, lit: "foo"},
				{typ: tSilentScript, lit: "bar"},
				{typ: tRawText, lit: "\nbaz"},
				{typ: tEOF, lit: ""},
			},
		},
		"not a closing tag": {
			input: "@ego test() {\n\tfoo<% bar % %>\n\tbaz",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tRawText, lit: "foo"},
				{typ: tSilentScript, lit: "bar %"},
				{typ: tRawText, lit: "\nbaz"},
				{typ: tEOF, lit: ""},
			},
		},
		"trailing newline": {
			input: "@ego test() {\n\tfoo\n\t<% bar $%>\n\tbaz\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tRawText, lit: "foo\n"},
				{typ: tSilentScript, lit: "bar"},
				{typ: tRawText, lit: "baz"},
				{typ: tTemplateEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"trailing newline and carriage return": {
			input: "@ego test() {\n\tfoo\n\t<% bar $%>\n\r\tbaz\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tRawText, lit: "foo\n"},
				{typ: tSilentScript, lit: "bar"},
				{typ: tRawText, lit: "baz"},
				{typ: tTemplateEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"trailing carriage return": {
			input: "@ego test() {\n\tfoo\n\t<% bar $%>\r\tbaz\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tRawText, lit: "foo\n"},
				{typ: tSilentScript, lit: "bar"},
				{typ: tRawText, lit: "baz"},
				{typ: tTemplateEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"not $ closing tag": {
			input: "@ego test() {\n\tfoo\n\t<% bar $ %>\n\tbaz\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tRawText, lit: "foo\n"},
				{typ: tSilentScript, lit: "bar $"},
				{typ: tRawText, lit: "\nbaz"},
				{typ: tTemplateEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"trailing whitespace": {
			input: "@ego test() {\n\tfoo\n\t<% bar -%>\n\t\n\t\tbaz \t \t\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tRawText, lit: "foo\n"},
				{typ: tSilentScript, lit: "bar"},
				{typ: tRawText, lit: "baz"},
				{typ: tTemplateEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"not - closing tag": {
			input: "@ego test() {\n\tfoo\n\t<% bar - %>\n\tbaz \t \t\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tRawText, lit: "foo\n"},
				{typ: tSilentScript, lit: "bar -"},
				{typ: tRawText, lit: "\nbaz"},
				{typ: tTemplateEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"not indented": {
			input: "@ego test() {\n\tfoo\n\t<% bar -%>\nbaz",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tRawText, lit: "foo\n"},
				{typ: tSilentScript, lit: "bar"},
				{typ: tError, lit: "unexpected character at start of line: 'b'"},
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
