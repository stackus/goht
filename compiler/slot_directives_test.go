package compiler

import (
	"strings"
	"testing"
)

func TestSlotDirectiveLexing(t *testing.T) {
	tests := map[string]struct {
		input     string
		wantTypes []tokenType
	}{
		"haml": {
			input:     "@haml Test() {\n\t= @ifslot content\n\t= @noslot empty\n\t\t= @eachslot item in content\n",
			wantTypes: []tokenType{tIfSlotCommand, tNoSlotCommand, tEachSlotCommand},
		},
		"slim": {
			input:     "@slim Test() {\n\t= @ifslot content\n\t= @noslot empty\n\t\t= @eachslot item in content\n",
			wantTypes: []tokenType{tIfSlotCommand, tNoSlotCommand, tEachSlotCommand},
		},
		"ego": {
			input:     "@ego Test() {\n\t<%@ifslot content { %>\n\t<%@noslot empty { %>\n\t<%@eachslot item in content { %>\n",
			wantTypes: []tokenType{tIfSlotCommand, tNoSlotCommand, tEachSlotCommand},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			lexer := newLexer([]byte(tt.input))
			for _, want := range tt.wantTypes {
				for {
					got := lexer.nextToken()
					if got.typ == want {
						break
					}
					if got.typ == tError || got.typ == tEOF {
						t.Fatalf("did not find %s; stopped at %v", want, got)
					}
				}
			}
		})
	}
}

func TestSlotDirectiveErrors(t *testing.T) {
	tests := map[string]string{
		"missing ifslot argument":   "@slim Test() {\n\t= @ifslot\n}\n",
		"missing noslot argument":   "@haml Test() {\n\t= @noslot\n}\n",
		"missing eachslot argument": "@haml Test() {\n\t= @eachslot\n}\n",
		"invalid variable":          "@ego Test() {\n\t<%@eachslot 1item in content %>\n}\n",
		"missing in":                "@slim Test() {\n\t= @eachslot item content\n}\n",
	}
	for name, source := range tests {
		t.Run(name, func(t *testing.T) {
			template, err := ParseString(source)
			if err != nil {
				return
			}
			if err := template.Generate(testWriter{}); err == nil {
				t.Fatal("expected directive error")
			}
		})
	}
}

func TestNoSlotDirectiveGeneratesNegatedPresenceCheck(t *testing.T) {
	template, err := ParseString("package testdata\n@slim Test() {\n\t= @noslot content\n\t\tp Empty\n}\n")
	if err != nil {
		t.Fatalf("ParseString() error = %v", err)
	}

	var output strings.Builder
	if err := template.Generate(&output); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	generated := output.String()
	if !strings.Contains(generated, `if _, __hasSlot := __slots.Has("content"); !__hasSlot {`) {
		t.Errorf("generated source did not negate slot presence:\n%s", generated)
	}
	if !strings.Contains(generated, `}, "children", "content")`) {
		t.Errorf("generated source did not declare the referenced slot:\n%s", generated)
	}
}

type testWriter struct{}

func (testWriter) Write(data []byte) (int, error) { return len(data), nil }
