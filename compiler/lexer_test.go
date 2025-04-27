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
				{typ: tKeepNewlines, lit: ""},
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
				{typ: tKeepNewlines, lit: ""},
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
				{typ: tKeepNewlines, lit: ""},
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
				{typ: tKeepNewlines, lit: ""},
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
				{typ: tKeepNewlines, lit: ""},
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
