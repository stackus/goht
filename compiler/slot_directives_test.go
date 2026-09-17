package compiler

import "testing"

func TestSlotDirectiveLexing(t *testing.T) {
	tests := map[string]struct {
		input     string
		wantTypes []tokenType
	}{
		"haml": {
			input:     "@haml Test() {\n\t= @ifslot content\n\t\t= @eachslot item in content\n",
			wantTypes: []tokenType{tIfSlotCommand, tEachSlotCommand},
		},
		"slim": {
			input:     "@slim Test() {\n\t= @ifslot content\n\t\t= @eachslot item in content\n",
			wantTypes: []tokenType{tIfSlotCommand, tEachSlotCommand},
		},
		"ego": {
			input:     "@ego Test() {\n\t<%@ifslot content { %>\n\t<%@eachslot item in content { %>\n",
			wantTypes: []tokenType{tIfSlotCommand, tEachSlotCommand},
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

type testWriter struct{}

func (testWriter) Write(data []byte) (int, error) { return len(data), nil }
