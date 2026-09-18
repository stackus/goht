package compiler

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestPositionalErrorUnwrap(t *testing.T) {
	cause := errors.New("cause")
	err := PositionalError{Line: 2, Column: 3, Err: cause}

	if !errors.Is(err, cause) || err.Error() != "[2:3]: cause" {
		t.Fatalf("error = %q, unwrap = %v", err, errors.Unwrap(err))
	}
}

func TestLexerAcceptHelpers(t *testing.T) {
	tests := map[string]struct {
		input  string
		accept string
		want   bool
		left   string
	}{
		"accepts matching rune":     {input: "abc", accept: "a", want: true, left: "bc"},
		"backs up nonmatching rune": {input: "abc", accept: "z", want: false, left: "abc"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			lexer := newLexer([]byte(tt.input))
			if got := lexer.accept(tt.accept); got != tt.want {
				t.Fatalf("accept() = %v, want %v", got, tt.want)
			}
			if got := lexer.peekAhead(len(tt.left)); got != tt.left {
				t.Fatalf("remaining input = %q, want %q", got, tt.left)
			}
		})
	}

	lexer := newLexer([]byte("abcdef"))
	lexer.acceptAhead(3)
	if got := lexer.current(); got != "abc" {
		t.Fatalf("current() = %q, want abc", got)
	}
}

func TestParseFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "template.goht")
	if err := os.WriteFile(path, []byte("package example\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := ParseFile(path); err != nil {
		t.Fatalf("ParseFile() error = %v", err)
	}
	if _, err := ParseFile(path + ".missing"); err == nil {
		t.Fatal("ParseFile(missing) error = nil")
	}
}

func TestTokenAccessors(t *testing.T) {
	token := token{typ: tTag, lit: "p", line: 4, col: 5}
	if token.Type() != tTag || token.Lit() != "p" || token.Line() != 4 || token.Col() != 5 {
		t.Fatalf("accessors returned unexpected token values: %#v", token)
	}
}

func TestStackAndOrderedMap(t *testing.T) {
	var values stack[int]
	if values.pop() != 0 || values.peek() != 0 {
		t.Fatal("empty stack did not return zero value")
	}

	values.push(1)
	values.push(2)
	if values.pop() != 2 || values.peek() != 1 {
		t.Fatalf("stack = %#v", values)
	}

	ordered := NewOrderedMap[int]()
	ordered.Set("first", 1)
	ordered.Set("second", 2)
	ordered.Set("first", 3)
	ordered.Delete("missing")
	ordered.Delete("second")

	if ordered.Len() != 1 || !reflect.DeepEqual(ordered.keys, []string{"first"}) {
		t.Fatalf("ordered map = %#v", ordered)
	}
	stopped := 0
	if err := ordered.Range(func(string, int) (bool, error) { stopped++; return false, nil }); err != nil || stopped != 1 {
		t.Fatalf("Range stop = (%v, %d)", err, stopped)
	}
	cause := errors.New("stop")
	if err := ordered.Range(func(string, int) (bool, error) { return true, cause }); !errors.Is(err, cause) {
		t.Fatalf("Range error = %v", err)
	}
}

func TestTemplateWriterTransitions(t *testing.T) {
	var output bytes.Buffer
	writer := newTemplateWriter(&output, nil).Indent(1)

	if _, err := writer.WriteStringLiteral("one"); err != nil {
		t.Fatal(err)
	}
	if _, err := writer.WriteVar("value"); err != nil {
		t.Fatal(err)
	}
	if _, err := writer.WriteStringIndent("\"two\""); err != nil {
		t.Fatal(err)
	}
	if _, err := writer.WriteErrorHandler(); err != nil {
		t.Fatal(err)
	}

	if got := output.String(); got == "" || writer.GetVarName() != "__var1" {
		t.Fatalf("output = %q", got)
	}
	writer.ResetVarName()
	if writer.GetVarName() != "__var1" {
		t.Fatal("ResetVarName() did not reset counter")
	}
}
