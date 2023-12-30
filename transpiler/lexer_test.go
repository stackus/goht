package transpiler

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
			got, err := l.peekAhead(tt.length)
			if err != nil {
				t.Errorf("peekAhead: %v", err)
				return
			}
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
				{typ: tNewLine, lit: "\n"},
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
				{typ: tNewLine, lit: "\n"},
				{typ: tNewLine, lit: "\n"},
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
		"into hmlt": {
			input: "package foo\n\n@hmlt test() {\n}\n",
			want: []token{
				{typ: tPackage, lit: "foo"},
				{typ: tNewLine, lit: "\n"},
				{typ: tNewLine, lit: "\n"},
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tHmltEnd, lit: ""},
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

func Test_HmltTransition(t *testing.T) {
	tests := map[string]struct {
		input string
		want  []token
	}{
		"into go": {
			input: "@hmlt test() {\n}\n\nfunc foo() {\n\tprintln(`bar`)\n}",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tHmltEnd, lit: ""},
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
		"incomplete hmlt": {
			input: "@hmlt test() {\n",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
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

func Test_HamlText(t *testing.T) {
	tests := map[string]struct {
		input string
		want  []token
	}{
		"simple": {
			input: "@hmlt test() {\nfoobar",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tStaticText, lit: "foobar"},
				{typ: tEOF, lit: ""},
			},
		},
		"indented": {
			input: "@hmlt test() {\n\tfoobar",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: "\t"},
				{typ: tStaticText, lit: "foobar"},
				{typ: tEOF, lit: ""},
			},
		},
		"multiple lines": {
			input: "@hmlt test() {\n\tfoobar\n\tbaz",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: "\t"},
				{typ: tStaticText, lit: "foobar"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: "\t"},
				{typ: tStaticText, lit: "baz"},
				{typ: tEOF, lit: ""},
			},
		},
		"with escaped quotes": {
			input: "@hmlt test() {\n\t\"foo\\\"bar\"",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: "\t"},
				{typ: tStaticText, lit: "\"foo\\\"bar\""},
				{typ: tEOF, lit: ""},
			},
		},
		"escape control characters": {
			input: "@hmlt test() {\n\t\\#foo\n\\%bar\n\\.baz",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: "\t"},
				{typ: tStaticText, lit: "#foo"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tStaticText, lit: "%bar"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tStaticText, lit: ".baz"},
			},
		},
		"text with dynamic value": {
			input: "@hmlt test() {\n\tfoo #{bar} baz",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: "\t"},
				{typ: tStaticText, lit: "foo "},
				{typ: tDynamicText, lit: "bar"},
				{typ: tStaticText, lit: " baz"},
				{typ: tEOF, lit: ""},
			},
		},
		"escape dynamic text": {
			input: "@hmlt test() {\n\\#{foo}",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tDynamicText, lit: "foo"},
				{typ: tEOF, lit: ""},
			},
		},
		"error in dynamic syntax": {
			input: "@hmlt test() {\n\tfoo #{bar baz",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: "\t"},
				{typ: tStaticText, lit: "foo "},
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
			input: "@hmlt test() {\n%foo\n",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tTag, lit: "foo"},
				{typ: tNewLine, lit: "\n"},
				{typ: tEOF, lit: ""},
			},
		},
		"multiple tags": {
			input: "@hmlt test() {\n%foo\n%bar",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tTag, lit: "foo"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tTag, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"indented": {
			input: "@hmlt test() {\n\t%foo",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tEOF, lit: ""},
			},
		},
		"two indented": {
			input: "@hmlt test() {\n\t%foo\n\t%bar",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "foo"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"indented 2": {
			input: "@hmlt test() {\n\t\t%foo",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: "\t\t"},
				{typ: tTag, lit: "foo"},
				{typ: tEOF, lit: ""},
			},
		},
		"tag and id": {
			input: "@hmlt test() {\n%foo#bar",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tTag, lit: "foo"},
				{typ: tId, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"tag and class": {
			input: "@hmlt test() {\n%foo.bar",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tTag, lit: "foo"},
				{typ: tClass, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"tag and attribute": {
			input: "@hmlt test() {\n%foo{id:\"bar\"}",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tTag, lit: "foo"},
				{typ: tAttrName, lit: "id"},
				{typ: tAttrOperator, lit: ":"},
				{typ: tAttrStaticValue, lit: "\"bar\""},
				{typ: tEOF, lit: ""},
			},
		},
		"tag and text": {
			input: "@hmlt test() {\n%foo bar",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tTag, lit: "foo"},
				{typ: tStaticText, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"tag and unescaped text": {
			input: "@hmlt test() {\n%foo! bar",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tTag, lit: "foo"},
				{typ: tUnescaped, lit: ""},
				{typ: tStaticText, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"tag and output code": {
			input: "@hmlt test() {\n%foo= bar",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tTag, lit: "foo"},
				{typ: tOutputCode, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"tag and tag again": {
			input: "@hmlt test() {\n%foo%bar",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tTag, lit: "foo"},
				{typ: tStaticText, lit: "%bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"space before tag identifier": {
			input: "@hmlt test() {\n% foo",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
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
			input: "@hmlt test() {\n#foo",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tId, lit: "foo"},
				{typ: tEOF, lit: ""},
			},
		},
		"multiple ids": {
			input: "@hmlt test() {\n#foo\n#bar",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tId, lit: "foo"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tId, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"indented": {
			input: "@hmlt test() {\n\t#foo",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: "\t"},
				{typ: tId, lit: "foo"},
				{typ: tEOF, lit: ""},
			},
		},
		"indented 2": {
			input: "@hmlt test() {\n\t\t#foo",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: "\t\t"},
				{typ: tId, lit: "foo"},
				{typ: tEOF, lit: ""},
			},
		},
		"with underscore": {
			input: "@hmlt test() {\n#foo_bar\n}",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tId, lit: "foo_bar"},
				{typ: tNewLine, lit: "\n"},
				{typ: tHmltEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"with hyphen": {
			input: "@hmlt test() {\n#foo-bar\n}",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tId, lit: "foo-bar"},
				{typ: tNewLine, lit: "\n"},
				{typ: tHmltEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"id and class": {
			input: "@hmlt test() {\n#foo.bar",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tId, lit: "foo"},
				{typ: tClass, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"id and tag": {
			input: "@hmlt test() {\n#foo%bar",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tId, lit: "foo"},
				{typ: tStaticText, lit: "%bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"id and attribute": {
			input: "@hmlt test() {\n#foo{id:\"bar\"}",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tId, lit: "foo"},
				{typ: tAttrName, lit: "id"},
				{typ: tAttrOperator, lit: ":"},
				{typ: tAttrStaticValue, lit: "\"bar\""},
				{typ: tEOF, lit: ""},
			},
		},
		"id and text": {
			input: "@hmlt test() {\n#foo bar",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tId, lit: "foo"},
				{typ: tStaticText, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"id and unescaped text": {
			input: "@hmlt test() {\n#foo! bar",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tId, lit: "foo"},
				{typ: tUnescaped, lit: ""},
				{typ: tStaticText, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"id and output code": {
			input: "@hmlt test() {\n#foo= bar",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tId, lit: "foo"},
				{typ: tOutputCode, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"id and id again": {
			input: "@hmlt test() {\n#foo#bar",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tId, lit: "foo"},
				{typ: tId, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"space before id identifier": {
			input: "@hmlt test() {\n# foo",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
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
			input: "@hmlt test() {\n.foo",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tClass, lit: "foo"},
				{typ: tEOF, lit: ""},
			},
		},
		"multiple ids": {
			input: "@hmlt test() {\n.foo\n.bar",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tClass, lit: "foo"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tClass, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"indented": {
			input: "@hmlt test() {\n\t.foo",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: "\t"},
				{typ: tClass, lit: "foo"},
				{typ: tEOF, lit: ""},
			},
		},
		"indented 2": {
			input: "@hmlt test() {\n\t\t.foo",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: "\t\t"},
				{typ: tClass, lit: "foo"},
				{typ: tEOF, lit: ""},
			},
		},
		"class and id": {
			input: "@hmlt test() {\n.foo#bar",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tClass, lit: "foo"},
				{typ: tId, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"class and tag": {
			input: "@hmlt test() {\n.foo%bar",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tClass, lit: "foo"},
				{typ: tStaticText, lit: "%bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"class and attribute": {
			input: "@hmlt test() {\n.foo{id:\"bar\"}",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tClass, lit: "foo"},
				{typ: tAttrName, lit: "id"},
				{typ: tAttrOperator, lit: ":"},
				{typ: tAttrStaticValue, lit: "\"bar\""},
				{typ: tEOF, lit: ""},
			},
		},
		"class and text": {
			input: "@hmlt test() {\n.foo bar",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tClass, lit: "foo"},
				{typ: tStaticText, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"class and unescaped text": {
			input: "@hmlt test() {\n.foo! bar",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tClass, lit: "foo"},
				{typ: tUnescaped, lit: ""},
				{typ: tStaticText, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"class and output code": {
			input: "@hmlt test() {\n.foo= bar",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tClass, lit: "foo"},
				{typ: tOutputCode, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"class and class again": {
			input: "@hmlt test() {\n.foo.bar",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tClass, lit: "foo"},
				{typ: tClass, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"space before class identifier": {
			input: "@hmlt test() {\n. foo",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
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

func Test_HamlAttributes(t *testing.T) {
	tests := map[string]struct {
		input string
		want  []token
	}{
		"simple": {
			input: "@hmlt test() {\n%foo{id:\"bar\"}",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tTag, lit: "foo"},
				{typ: tAttrName, lit: "id"},
				{typ: tAttrOperator, lit: ":"},
				{typ: tAttrStaticValue, lit: "\"bar\""},
				{typ: tEOF, lit: ""},
			},
		},
		"names with dashes": {
			input: "@hmlt test() {\n%foo{data-foo:\"bar\"}",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tTag, lit: "foo"},
				{typ: tAttrName, lit: "data-foo"},
				{typ: tAttrOperator, lit: ":"},
				{typ: tAttrStaticValue, lit: "\"bar\""},
				{typ: tEOF, lit: ""},
			},
		},
		"names with underscores": {
			input: "@hmlt test() {\n%foo{data_foo:\"bar\"}",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tTag, lit: "foo"},
				{typ: tAttrName, lit: "data_foo"},
				{typ: tAttrOperator, lit: ":"},
				{typ: tAttrStaticValue, lit: "\"bar\""},
				{typ: tEOF, lit: ""},
			},
		},
		"names with numbers": {
			input: "@hmlt test() {\n%foo{data1:\"bar\"}",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tTag, lit: "foo"},
				{typ: tAttrName, lit: "data1"},
				{typ: tAttrOperator, lit: ":"},
				{typ: tAttrStaticValue, lit: "\"bar\""},
				{typ: tEOF, lit: ""},
			},
		},
		"names with colons": {
			input: "@hmlt test() {\n%foo{\":x-data\":\"bar\",`x-on:click`:{onClick}}",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tTag, lit: "foo"},
				{typ: tAttrName, lit: ":x-data"},
				{typ: tAttrOperator, lit: ":"},
				{typ: tAttrStaticValue, lit: "\"bar\""},
				{typ: tAttrName, lit: "x-on:click"},
				{typ: tAttrOperator, lit: ":"},
				{typ: tAttrDynamicValue, lit: "onClick"},
				{typ: tEOF, lit: ""},
			},
		},
		"names with dots": {
			input: "@hmlt test() {\n%foo{data.foo:\"bar\",x.on.click:{onClick}}",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tTag, lit: "foo"},
				{typ: tAttrName, lit: "data.foo"},
				{typ: tAttrOperator, lit: ":"},
				{typ: tAttrStaticValue, lit: "\"bar\""},
				{typ: tAttrName, lit: "x.on.click"},
				{typ: tAttrOperator, lit: ":"},
				{typ: tAttrDynamicValue, lit: "onClick"},
			},
		},
		"names with at signs": {
			input: "@hmlt test() {\n%foo{\"@data\":\"bar\",`x@on@click`:{onClick}}",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tTag, lit: "foo"},
				{typ: tAttrName, lit: "@data"},
				{typ: tAttrOperator, lit: ":"},
				{typ: tAttrStaticValue, lit: "\"bar\""},
				{typ: tAttrName, lit: "x@on@click"},
				{typ: tAttrOperator, lit: ":"},
				{typ: tAttrDynamicValue, lit: "onClick"},
			},
		},
		"several": {
			input: "@hmlt test() {\n%foo{id:\"bar\", class: `baz` , title : \"qux\"}",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tTag, lit: "foo"},
				{typ: tAttrName, lit: "id"},
				{typ: tAttrOperator, lit: ":"},
				{typ: tAttrStaticValue, lit: "\"bar\""},
				{typ: tAttrName, lit: "class"},
				{typ: tAttrOperator, lit: ":"},
				{typ: tAttrStaticValue, lit: "`baz`"},
				{typ: tAttrName, lit: "title"},
				{typ: tAttrOperator, lit: ":"},
				{typ: tAttrStaticValue, lit: "\"qux\""},
				{typ: tEOF, lit: ""},
			},
		},
		"several on multiple lines": {
			input: "@hmlt test() {\n%foo{\n\tid:\"bar\",\n\tclass: `baz` ,\n\ttitle : \"qux\"\n}",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tTag, lit: "foo"},
				{typ: tAttrName, lit: "id"},
				{typ: tAttrOperator, lit: ":"},
				{typ: tAttrStaticValue, lit: "\"bar\""},
				{typ: tAttrName, lit: "class"},
				{typ: tAttrOperator, lit: ":"},
				{typ: tAttrStaticValue, lit: "`baz`"},
				{typ: tAttrName, lit: "title"},
				{typ: tAttrOperator, lit: ":"},
				{typ: tAttrStaticValue, lit: "\"qux\""},
				{typ: tEOF, lit: ""},
			},
		},
		"static value with escaped quotes": {
			input: "@hmlt test() {\n%foo{id:\"bar\\\"baz\"}",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tTag, lit: "foo"},
				{typ: tAttrName, lit: "id"},
				{typ: tAttrOperator, lit: ":"},
				{typ: tAttrStaticValue, lit: "\"bar\\\"baz\""},
				{typ: tEOF, lit: ""},
			},
		},
		"dynamic value": {
			input: "@hmlt test() {\n%foo{id:{bar}}",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tTag, lit: "foo"},
				{typ: tAttrName, lit: "id"},
				{typ: tAttrOperator, lit: ":"},
				{typ: tAttrDynamicValue, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"dynamic value with escaped curly": {
			input: "@hmlt test() {\n%foo{id:{\"big}\"}, class: {\"ba\"+'}'}}",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
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
			input: "@hmlt test() {\n%foo{id:{bar}, class: {baz}}",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
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
			input: "@hmlt test() {\n%foo{bar}",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tTag, lit: "foo"},
				{typ: tAttrName, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"boolean operator": {
			input: "@hmlt test() {\n%foo{bar?:{isBar}, baz ?: {isBaz}}",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tTag, lit: "foo"},
				{typ: tAttrName, lit: "bar"},
				{typ: tAttrOperator, lit: "?:"},
				{typ: tAttrDynamicValue, lit: "isBar"},
				{typ: tAttrName, lit: "baz"},
				{typ: tAttrOperator, lit: "?:"},
				{typ: tAttrDynamicValue, lit: "isBaz"},
				{typ: tEOF, lit: ""},
			},
		},
		"attributes command": {
			input: "@hmlt test() {\n%foo{@attributes:{listA, \"}}\", listB}}",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tTag, lit: "foo"},
				{typ: tAttributesCommand, lit: "listA, \"}}\", listB"},
				{typ: tEOF, lit: ""},
			},
		},
		"missing delimiter": {
			input: "@hmlt test() {\n%foo{id\"bar\", class: \"baz\"}",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tTag, lit: "foo"},
				{typ: tAttrName, lit: "id"},
				{typ: tError, lit: "unexpected character: '\"'"},
				{typ: tEOF, lit: ""},
			},
		},
		"missing separator": {
			input: "@hmlt test() {\n%foo{id:\"bar\" class: \"baz\"}",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tTag, lit: "foo"},
				{typ: tAttrName, lit: "id"},
				{typ: tAttrOperator, lit: ":"},
				{typ: tAttrStaticValue, lit: "\"bar\""},
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
			input: "@hmlt test() {\n!!!",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tDoctype, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"with type": {
			input: "@hmlt test() {\n!!! Strict",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tDoctype, lit: "Strict"},
				{typ: tEOF, lit: ""},
			},
		},
		"not a doctype": {
			input: "@hmlt test() {\n!foo",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tUnescaped, lit: ""},
				{typ: tStaticText, lit: "foo"},
				{typ: tEOF, lit: ""},
			},
		},
		"doctype with content": {
			input: "@hmlt test() {\n!!! 5\n%html\n}",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tDoctype, lit: "5"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tTag, lit: "html"},
				{typ: tNewLine, lit: "\n"},
				{typ: tHmltEnd, lit: ""},
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
			input: "@hmlt test() {\n!foo",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tUnescaped, lit: ""},
				{typ: tStaticText, lit: "foo"},
				{typ: tEOF, lit: ""},
			},
		},
		"indented": {
			input: "@hmlt test() {\n\t! foo",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: "\t"},
				{typ: tUnescaped, lit: ""},
				{typ: tStaticText, lit: "foo"},
				{typ: tEOF, lit: ""},
			},
		},
		"dynamic text": {
			input: "@hmlt test() {\n! #{foo}",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tUnescaped, lit: ""},
				{typ: tDynamicText, lit: "foo"},
				{typ: tEOF, lit: ""},
			},
		},
		"unescaped code": {
			input: "@hmlt test() {\n\t!= foo",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: "\t"},
				{typ: tUnescaped, lit: ""},
				{typ: tOutputCode, lit: "foo"},
			},
		},
		"not unescaped": {
			input: "@hmlt test() {\n%p ! foo",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tTag, lit: "p"},
				{typ: tStaticText, lit: "! foo"},
				{typ: tEOF, lit: ""},
			},
		},
		"tag with unescaped text": {
			input: "@hmlt test() {\n%foo! bar",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tTag, lit: "foo"},
				{typ: tUnescaped, lit: ""},
				{typ: tStaticText, lit: "bar"},
			},
		},
		"tag with unescaped code": {
			input: "@hmlt test() {\n%foo!= bar",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tTag, lit: "foo"},
				{typ: tUnescaped, lit: ""},
				{typ: tOutputCode, lit: "bar"},
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
			input: "@hmlt test() {\n= foo",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tOutputCode, lit: "foo"},
				{typ: tEOF, lit: ""},
			},
		},
		"indented": {
			input: "@hmlt test() {\n\t= foo",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: "\t"},
				{typ: tOutputCode, lit: "foo"},
				{typ: tEOF, lit: ""},
			},
		},
		"after tag": {
			input: "@hmlt test() {\n%foo= bar",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tTag, lit: "foo"},
				{typ: tOutputCode, lit: "bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"without space": {
			input: "@hmlt test() {\n=foo",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tOutputCode, lit: "foo"},
				{typ: tEOF, lit: ""},
			},
		},
		"with parens": {
			input: "@hmlt test() {\n= foo()",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tOutputCode, lit: "foo()"},
				{typ: tEOF, lit: ""},
			},
		},
		"with render command": {
			input: "@hmlt test() {\n= @render foo(\"bar\")",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tRenderCommand, lit: "foo(\"bar\")"},
				{typ: tEOF, lit: ""},
			},
		},
		"with render command and parens": {
			input: "@hmlt test() {\n= @render() foo(\"bar\")",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tRenderCommand, lit: "foo(\"bar\")"},
				{typ: tEOF, lit: ""},
			},
		},
		"with missing render argument": {
			input: "@hmlt test() {\n= @render",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tError, lit: "render argument expected"},
				{typ: tEOF, lit: ""},
			},
		},
		"with children command": {
			input: "@hmlt test() {\n= @children",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tChildrenCommand, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"with children command and parens": {
			input: "@hmlt test() {\n= @children()",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tChildrenCommand, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"without any children arguments": {
			input: "@hmlt test() {\n= @children() asdfasdf",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
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
			input: "@hmlt test() {\n- foo",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tExecuteCode, lit: "foo"},
				{typ: tEOF, lit: ""},
			},
		},
		"indented": {
			input: "@hmlt test() {\n\t- foo",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: "\t"},
				{typ: tExecuteCode, lit: "foo"},
				{typ: tEOF, lit: ""},
			},
		},
		"not code": {
			input: "@hmlt test() {\n%foo- bar\n#foo-bar bar\n%p - bar",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tTag, lit: "foo-"},
				{typ: tStaticText, lit: "bar"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tId, lit: "foo-bar"},
				{typ: tStaticText, lit: "bar"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tTag, lit: "p"},
				{typ: tStaticText, lit: "- bar"},
				{typ: tEOF, lit: ""},
			},
		},
		"without space": {
			input: "@hmlt test() {\n-foo",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tExecuteCode, lit: "foo"},
				{typ: tEOF, lit: ""},
			},
		},
		"nested": {
			input: `@hmlt test()
- if foo != "" {
	%p Foo exists and is #{foo}.
`,
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tExecuteCode, lit: "if foo != \"\" {"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "p"},
				{typ: tStaticText, lit: "Foo exists and is "},
				{typ: tDynamicText, lit: "foo"},
				{typ: tStaticText, lit: "."},
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

func Test_HamlIndent(t *testing.T) {
	tests := map[string]struct {
		input string
		want  []token
	}{
		"simple": {
			input: "@hmlt test() {\n\tfoo\n}",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: "\t"},
				{typ: tStaticText, lit: "foo"},
				{typ: tNewLine, lit: "\n"},
				{typ: tHmltEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"multiple indents": {
			input: "@hmlt test() {\n\t\tfoo\n}",
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: "\t\t"},
				{typ: tStaticText, lit: "foo"},
				{typ: tNewLine, lit: "\n"},
				{typ: tHmltEnd, lit: ""},
				{typ: tEOF, lit: ""},
			},
		},
		"different indents": {
			input: `@hmlt test() {
	%p1
		%span one
	%p2 two
%p3 three
}`,
			want: []token{
				{typ: tHmltStart, lit: "test()"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "p1"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: "\t\t"},
				{typ: tTag, lit: "span"},
				{typ: tStaticText, lit: "one"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: "\t"},
				{typ: tTag, lit: "p2"},
				{typ: tStaticText, lit: "two"},
				{typ: tNewLine, lit: "\n"},
				{typ: tIndent, lit: ""},
				{typ: tTag, lit: "p3"},
				{typ: tStaticText, lit: "three"},
				{typ: tNewLine, lit: "\n"},
				{typ: tHmltEnd, lit: ""},
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
