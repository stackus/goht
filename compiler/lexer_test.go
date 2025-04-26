package compiler

import (
	"testing"
	"text/scanner"
)

func Test_LexerNext(t *testing.T) {
	tests := map[string]struct {
		input string
		want  []rune
	}{
		"empty": {
			input: "",
			want:  []rune{scanner.EOF},
		},
		"single character": {
			input: "a",
			want:  []rune{'a', scanner.EOF},
		},
		"multiple characters": {
			input: "abc",
			want:  []rune{'a', 'b', 'c', scanner.EOF},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := newLexer([]byte(tt.input))
			for _, want := range tt.want {
				got := l.next()
				if got != want {
					t.Errorf("want %v, got %v", want, got)
				}
			}
		})
	}
}

func Test_LexerBackup(t *testing.T) {
	tests := map[string]struct {
		input string
		next  int
		want  []rune
	}{
		"empty": {
			input: "",
			want:  []rune{scanner.EOF},
		},
		"single character": {
			input: "a",
			next:  1,
			want:  []rune{'a', scanner.EOF},
		},
		"multiple characters": {
			input: "abc",
			next:  2,
			want:  []rune{'b', 'c', scanner.EOF},
		},
		"to the end": {
			input: "abc",
			next:  4,
			want:  []rune{scanner.EOF},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := newLexer([]byte(tt.input))
			for i := 0; i < tt.next; i++ {
				l.next()
			}
			l.backup()
			for _, want := range tt.want {
				got := l.next()
				if got != want {
					t.Errorf("want %v, got %v", want, got)
				}
			}
		})
	}
}

func Test_LexerPeek(t *testing.T) {
	tests := map[string]struct {
		input    string
		next     int
		want     rune
		wantNext []rune
	}{
		"empty": {
			input:    "",
			want:     scanner.EOF,
			wantNext: []rune{scanner.EOF},
		},
		"single character": {
			input:    "a",
			next:     0,
			want:     'a',
			wantNext: []rune{'a', scanner.EOF},
		},
		"single character skipped": {
			input:    "a",
			next:     1,
			want:     scanner.EOF,
			wantNext: []rune{scanner.EOF},
		},
		"multiple characters": {
			input:    "abc",
			next:     2,
			want:     'c',
			wantNext: []rune{'c', scanner.EOF},
		},
		"to the end": {
			input:    "abc",
			next:     4,
			want:     scanner.EOF,
			wantNext: []rune{scanner.EOF},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := newLexer([]byte(tt.input))
			for i := 0; i < tt.next; i++ {
				l.next()
			}
			got := l.peek()
			if got != tt.want {
				t.Errorf("peek: want %v, got %v", tt.want, got)
				return
			}
			for _, want := range tt.wantNext {
				got := l.next()
				if got != want {
					t.Errorf("next: want %v, got %v", want, got)
				}
			}
		})
	}
}

func Test_LexerPeekAhead(t *testing.T) {
	tests := map[string]struct {
		input    string
		next     int
		length   int
		want     string
		wantNext []rune
	}{
		"empty": {
			input:    "",
			length:   1,
			want:     "",
			wantNext: []rune{scanner.EOF},
		},
		"single character": {
			input:    "a",
			next:     0,
			length:   1,
			want:     "a",
			wantNext: []rune{'a', scanner.EOF},
		},
		"single character skipped": {
			input:    "a",
			next:     1,
			length:   1,
			want:     "",
			wantNext: []rune{scanner.EOF},
		},
		"multiple characters": {
			input:    "abc",
			next:     1,
			length:   2,
			want:     "bc",
			wantNext: []rune{'b', 'c', scanner.EOF},
		},
		"to the end": {
			input:    "abc",
			next:     4,
			length:   1,
			want:     "",
			wantNext: []rune{scanner.EOF},
		},
		"longer than the input": {
			input:    "abc",
			next:     0,
			length:   4,
			want:     "abc",
			wantNext: []rune{'a', 'b', 'c', scanner.EOF},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := newLexer([]byte(tt.input))
			for i := 0; i < tt.next; i++ {
				l.next()
			}
			got := l.peekAhead(tt.length)
			if got != tt.want {
				t.Errorf("peekAhead: want %v, got %v", tt.want, got)
				return
			}
			for _, want := range tt.wantNext {
				got := l.next()
				if got != want {
					t.Errorf("next: want %v, got %v", want, got)
				}
			}
		})
	}
}

func Test_Package(t *testing.T) {
	tests := map[string]struct {
		input string
		want  []token
	}{
		"package": {
			input: "package foo",
			want: []token{
				{typ: tPackage, lit: "foo"},
				{typ: tEOF, lit: ""},
			},
		},
		"package without a name": {
			input: "package \nvar x = 1",
			want: []token{
				{typ: tError, lit: "package name expected"},
			},
		},
		"package with multibyte characters": {
			input: "package 测试",
			want: []token{
				{typ: tPackage, lit: "测试"},
				{typ: tEOF, lit: ""},
			},
		},
		"package with newline": {
			input: "package foo\n",
			want: []token{
				{typ: tPackage, lit: "foo"},
				{typ: tNewLine, lit: "\n"},
				{typ: tEOF, lit: ""},
			},
		},
		"package with comments": {
			input: "// comment for package\npackage foo",
			want: []token{
				{typ: tGoCode, lit: "// comment for package"},
				{typ: tNewLine, lit: "\n"},
				{typ: tPackage, lit: "foo"},
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

func Test_Import(t *testing.T) {
	tests := map[string]struct {
		input string
		want  []token
	}{
		"import": {
			input: "import \"fmt\"",
			want: []token{
				{typ: tImport, lit: "\"fmt\""},
				{typ: tEOF, lit: ""},
			},
		},
		"named import": {
			input: "import foo \"fmt\"",
			want: []token{
				{typ: tImport, lit: "foo \"fmt\""},
				{typ: tEOF, lit: ""},
			},
		},
		"import with newline": {
			input: "import \"fmt\"\n",
			want: []token{
				{typ: tImport, lit: "\"fmt\""},
				{typ: tEOF, lit: ""},
			},
		},
		"import with multibyte characters": {
			input: "import \"测试\"",
			want: []token{
				{typ: tImport, lit: "\"测试\""},
				{typ: tEOF, lit: ""},
			},
		},
		"import with multiple packages": {
			input: "import (\n\t\"fmt\"\n\t\"strings\"\n)",
			want: []token{
				{typ: tImport, lit: "\"fmt\""},
				{typ: tImport, lit: "\"strings\""},
				{typ: tEOF, lit: ""},
			},
		},
		"import with multiple named packages": {
			input: "import (\n\tfoo \"fmt\"\n\tbar \"strings\"\n)",
			want: []token{
				{typ: tImport, lit: "foo \"fmt\""},
				{typ: tImport, lit: "bar \"strings\""},
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

func Test_Go(t *testing.T) {
	tests := map[string]struct {
		input string
		want  []token
	}{
		"empty": {
			input: "",
			want: []token{
				{typ: tEOF, lit: ""},
			},
		},
		"package": {
			input: "package foo",
			want: []token{
				{typ: tPackage, lit: "foo"},
				{typ: tEOF, lit: ""},
			},
		},
		"import": {
			input: "package foo\n\nimport \"fmt\"\n",
			want: []token{
				{typ: tPackage, lit: "foo"},
				{typ: tNewLine, lit: "\n"},
				{typ: tNewLine, lit: "\n"},
				{typ: tImport, lit: "\"fmt\""},
				{typ: tEOF, lit: ""},
			},
		},
		"multibyte runes": {
			input: "package foo\n\nvar x = \"测试\"\n",
			want: []token{
				{typ: tPackage, lit: "foo"},
				{typ: tNewLine, lit: "\n"},
				{typ: tNewLine, lit: "\n"},
				{typ: tGoCode, lit: "var x = \"测试\""},
				{typ: tNewLine, lit: "\n"},
				{typ: tEOF, lit: ""},
			},
		},
		"imports and functions": {
			input: "package foo\n\nimport (\n\t\"fmt\"\n\t\"strings\"\n)\n\nfunc main() {\n\tfmt.Println(strings.ToUpper(\"测试\"))\n}",
			want: []token{
				{typ: tPackage, lit: "foo"},
				{typ: tNewLine, lit: "\n"},
				{typ: tNewLine, lit: "\n"},
				{typ: tImport, lit: "\"fmt\""},
				{typ: tImport, lit: "\"strings\""},
				{typ: tGoCode, lit: "func main() {"},
				{typ: tNewLine, lit: "\n"},
				{typ: tGoCode, lit: "\tfmt.Println(strings.ToUpper(\"测试\"))"},
				{typ: tNewLine, lit: "\n"},
				{typ: tGoCode, lit: "}"},
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

func Test_GoTransition(t *testing.T) {
	tests := map[string]struct {
		input string
		want  []token
	}{
		"into goht": {
			input: "package foo\n\n@goht test() {\n}\n",
			want: []token{
				{typ: tPackage, lit: "foo"},
				{typ: tNewLine, lit: "\n"},
				{typ: tNewLine, lit: "\n"},
				{typ: tTemplateStart, lit: "test()"},
				{typ: tTemplateEnd, lit: ""},
				{typ: tNewLine, lit: "\n"},
				{typ: tEOF, lit: ""},
			},
		},
		"into haml": {
			input: "package foo\n\n@haml test() {\n}\n",
			want: []token{
				{typ: tPackage, lit: "foo"},
				{typ: tNewLine, lit: "\n"},
				{typ: tNewLine, lit: "\n"},
				{typ: tTemplateStart, lit: "test()"},
				{typ: tTemplateEnd, lit: ""},
				{typ: tNewLine, lit: "\n"},
				{typ: tEOF, lit: ""},
			},
		},
		"into slim": {
			input: "package foo\n\n@slim test() {\n}\n",
			want: []token{
				{typ: tPackage, lit: "foo"},
				{typ: tNewLine, lit: "\n"},
				{typ: tNewLine, lit: "\n"},
				{typ: tTemplateStart, lit: "test()"},
				{typ: tTemplateEnd, lit: ""},
				{typ: tNewLine, lit: "\n"},
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

func Test_GohtTransition(t *testing.T) {
	tests := map[string]struct {
		input string
		want  []token
	}{
		"into go from goht": {
			input: "@goht test() {\n}\n\nfunc foo() {\n\tprintln(`bar`)\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tTemplateEnd, lit: ""},
				{typ: tNewLine, lit: "\n"},
				{typ: tNewLine, lit: "\n"},
				{typ: tGoCode, lit: "func foo() {"},
				{typ: tNewLine, lit: "\n"},
				{typ: tGoCode, lit: "\tprintln(`bar`)"},
				{typ: tNewLine, lit: "\n"},
				{typ: tGoCode, lit: "}"},
				{typ: tEOF, lit: ""},
			},
		},
		"into go from haml": {
			input: "@haml test() {\n}\n\nfunc foo() {\n\tprintln(`bar`)\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tTemplateEnd, lit: ""},
				{typ: tNewLine, lit: "\n"},
				{typ: tNewLine, lit: "\n"},
				{typ: tGoCode, lit: "func foo() {"},
				{typ: tNewLine, lit: "\n"},
				{typ: tGoCode, lit: "\tprintln(`bar`)"},
				{typ: tNewLine, lit: "\n"},
				{typ: tGoCode, lit: "}"},
				{typ: tEOF, lit: ""},
			},
		},
		"into go from slim": {
			input: "@slim test() {\n}\n\nfunc foo() {\n\tprintln(`bar`)\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tTemplateEnd, lit: ""},
				{typ: tNewLine, lit: "\n"},
				{typ: tNewLine, lit: "\n"},
				{typ: tGoCode, lit: "func foo() {"},
				{typ: tNewLine, lit: "\n"},
				{typ: tGoCode, lit: "\tprintln(`bar`)"},
				{typ: tNewLine, lit: "\n"},
				{typ: tGoCode, lit: "}"},
				{typ: tEOF, lit: ""},
			},
		},
		"incomplete goht": {
			input: "@goht test() {\n",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
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

func Test_HamlText(t *testing.T) {
	tests := map[string]struct {
		input string
		want  []token
	}{
		"simple": {
			input: "@goht test() {\n\tfoobar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tPlainText, lit: "foobar"},
				{typ: tEOF, lit: ""},
			},
		},
		"multiple lines": {
			input: "@goht test() {\n\tfoobar\n\tbaz",
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
		"with escaped quotes": {
			input: "@goht test() {\n\t\"foo\\\"bar\"",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tPlainText, lit: "\"foo\\\"bar\""},
				{typ: tEOF, lit: ""},
			},
		},
		"escape control characters": {
			input: "@goht test() {\n\t\\#foo\n\t\\%bar\n\t\\.baz",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
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
				{typ: tIndent, lit: "\t"},
				{typ: tDynamicText, lit: "foo"},
				{typ: tEOF, lit: ""},
			},
		},
		"ignore dynamic text in line": {
			input: "@goht test() {\n\tfoo \\#{bar} \\{f} baz",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tPlainText, lit: "foo #{bar} \\{f} baz"},
				{typ: tEOF, lit: ""},
			},
		},
		"error in dynamic syntax": {
			input: "@goht test() {\n\tfoo #{bar baz",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
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

func Test_HamlTag(t *testing.T) {
	tests := map[string]struct {
		input string
		want  []token
	}{
		"simple": {
			input: "@goht test() {\n\t%foo\n",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
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

func Test_HamlId(t *testing.T) {
	tests := map[string]struct {
		input string
		want  []token
	}{
		"simple": {
			input: "@goht test() {\n\t#foo",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tId, lit: "foo"},
				{typ: tEOF, lit: ""},
			},
		},
		"multiple ids": {
			input: "@goht test() {\n\t#foo\n\t#bar",
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
			input: "@goht test() {\n\t#foo_bar\n}",
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
			input: "@goht test() {\n\t#foo-bar\n}",
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
			input: "@goht test() {\n\t#foo.bar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
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

func Test_HamlClass(t *testing.T) {
	tests := map[string]struct {
		input string
		want  []token
	}{
		"simple": {
			input: "@goht test() {\n\t.foo",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tClass, lit: "foo"},
				{typ: tEOF, lit: ""},
			},
		},
		"multiple classes": {
			input: "@goht test() {\n\t.foo\n\t.bar",
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
			input: "@goht test() {\n\t.foo#bar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
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

func Test_WhitespaceRemoval(t *testing.T) {
	tests := map[string]struct {
		input string
		want  []token
	}{
		"remove outer": {
			input: "@goht test() {\n\t%p>\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
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

func Test_HamlDoctype(t *testing.T) {
	tests := map[string]struct {
		input string
		want  []token
	}{
		"simple": {
			input: "@goht test() {\n\t!!!",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tDoctype, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"with type": {
			input: "@goht test() {\n\t!!! Strict",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tDoctype, lit: "Strict"},
				{typ: tEOF, lit: ""},
			},
		},
		"not a doctype": {
			input: "@goht test() {\n\t!foo",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
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

func Test_HamlUnescaped(t *testing.T) {
	tests := map[string]struct {
		input string
		want  []token
	}{
		"simple": {
			input: "@goht test() {\n\t!foo",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
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
				{typ: tIndent, lit: "\t"},
				{typ: tUnescaped, lit: ""},
				{typ: tScript, lit: "foo"},
			},
		},
		"not unescaped": {
			input: "@goht test() {\n\t%p ! foo",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
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
				{typ: tIndent, lit: "\t"},
				{typ: tComment, lit: "foo"},
				{typ: tEOF, lit: ""},
			},
		},
		"with content": {
			input: "@goht test() {\n\t/\n\t%p bar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
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

func Test_HamlVoidTags(t *testing.T) {
	tests := map[string]struct {
		input string
		want  []token
	}{
		"simple": {
			input: "@goht test() {\n\t%foo/",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
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
				{typ: tIndent, lit: "\t"},
				{typ: tScript, lit: "foo"},
				{typ: tEOF, lit: ""},
			},
		},
		"with space": {
			input: "@goht test() {\n\t= foo",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tScript, lit: "foo"},
				{typ: tEOF, lit: ""},
			},
		},
		"after tag": {
			input: "@goht test() {\n\t%foo= bar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
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
				{typ: tIndent, lit: "\t"},
				{typ: tScript, lit: "foo"},
				{typ: tEOF, lit: ""},
			},
		},
		"with parens": {
			input: "@goht test() {\n\t= foo()",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tScript, lit: "foo()"},
				{typ: tEOF, lit: ""},
			},
		},
		"with render command": {
			input: "@goht test() {\n\t= @render foo(\"bar\")",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tRenderCommand, lit: "foo(\"bar\")"},
				{typ: tEOF, lit: ""},
			},
		},
		"with render command and parens": {
			input: "@goht test() {\n\t= @render() foo(\"bar\")",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tRenderCommand, lit: "foo(\"bar\")"},
				{typ: tEOF, lit: ""},
			},
		},
		"with missing render argument": {
			input: "@goht test() {\n\t= @render",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tError, lit: "render argument expected"},
				{typ: tEOF, lit: ""},
			},
		},
		"with children command": {
			input: "@goht test() {\n\t= @children",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tChildrenCommand, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"with children command and parens": {
			input: "@goht test() {\n\t= @children()",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tChildrenCommand, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"without any children arguments": {
			input: "@goht test() {\n\t= @children() asdfasdf",
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

func Test_HamlExecuteCode(t *testing.T) {
	tests := map[string]struct {
		input string
		want  []token
	}{
		"simple": {
			input: "@goht test() {\n\t-foo",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tSilentScript, lit: "foo"},
				{typ: tEOF, lit: ""},
			},
		},
		"with space": {
			input: "@goht test() {\n\t- foo",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
				{typ: tIndent, lit: "\t"},
				{typ: tSilentScript, lit: "foo"},
				{typ: tEOF, lit: ""},
			},
		},
		"not code": {
			input: "@goht test() {\n\t%foo- bar\n\t#foo-bar bar\n\t%p - bar",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
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
				{typ: tIndent, lit: "\t"},
				{typ: tSilentScript, lit: "t.bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"with receiver and interface": {
			input: "@goht (t Tester) test(v interface{}) {\n\t- t.bar",
			want: []token{
				{typ: tTemplateStart, lit: "(t Tester) test(v interface{})"},
				{typ: tIndent, lit: "\t"},
				{typ: tSilentScript, lit: "t.bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"with receiver and interface with methods": {
			input: "@goht (t Tester) test(v interface{ Foo() string }) {\n\t- t.bar",
			want: []token{
				{typ: tTemplateStart, lit: "(t Tester) test(v interface{ Foo() string })"},
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
				{typ: tSilentScript, lit: "foo(bar,baz,)"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "p"},
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

func Test_HamlIndent(t *testing.T) {
	tests := map[string]struct {
		input string
		want  []token
	}{
		"simple": {
			input: "@goht test() {\n\t%foo\n\t\tbar\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
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
				{typ: tIndent, lit: "\t"},
				{typ: tFilterStart, lit: "css"},
				{typ: tPlainText, lit: "foo\n"},
				{typ: tFilterEnd, lit: ""},
				{typ: tTemplateEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"indented": {
			input: "@goht test() {\n\t:javascript\n\t\tfoo\n}",
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
			input: "@goht test() {\n\t:javascript\n\t\tfoo #{bar}\n}",
			want: []token{
				{typ: tTemplateStart, lit: "test()"},
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
		"inside tag": {
			input: `@goht test() {
	%p
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
			input: `@goht test() {
	%p
		:javascript
			console.log("hello world")
		%p foo
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
			input: `@goht test() {
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
