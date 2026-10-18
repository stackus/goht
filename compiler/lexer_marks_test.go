package compiler

import (
	"reflect"
	"testing"
)

type markText struct {
	kind markKind
	text string
}

func TestLexerMarks(t *testing.T) {
	tests := map[string]struct {
		input string
		want  []markText
	}{
		"haml template keywords": {
			input: "package main\n\n@goht A() {\n}\n\n@haml B() {\n}\n",
			want:  []markText{{markTemplateKeyword, "@goht"}, {markTemplateKeyword, "@haml"}},
		},
		"haml commands": {
			input: "package main\n\n@haml A() {\n\t= @render child()\n\t= @children\n\t= @slot header\n\t= @ifslot header\n\t= @noslot header\n\t= @eachslot items\n\t%p{@attributes: #{attrs}}\n}\n",
			want: []markText{
				{markTemplateKeyword, "@haml"},
				{markCommand, "@render"},
				{markCommand, "@children"},
				{markCommand, "@slot"},
				{markCommand, "@ifslot"},
				{markCommand, "@noslot"},
				{markCommand, "@eachslot"},
				{markCommand, "@attributes"},
			},
		},
		"slim commands": {
			input: "package main\n\n@slim A() {\n\tdoctype html\n\t= @render child()\n\t= @slot header\n\t= @eachslot items\n\tp(@attributes=#{attrs})\n}\n",
			want: []markText{
				{markTemplateKeyword, "@slim"},
				{markDoctype, "doctype"},
				{markCommand, "@render"},
				{markCommand, "@slot"},
				{markCommand, "@eachslot"},
				{markCommand, "@attributes"},
			},
		},
		"ego commands and comments": {
			input: "package main\n\n@ego A() {\n\t<%# a note %>\n\t<%@render child() %>\n\t<%@children %>\n\t<%@slot header %>\n\t<%@ifslot header %>x<%@noslot header %>y<%@eachslot items %>\n}\n",
			want: []markText{
				{markTemplateKeyword, "@ego"},
				{markComment, "<%# a note %>"},
				{markCommand, "@render"},
				{markCommand, "@children"},
				{markCommand, "@slot"},
				{markCommand, "@ifslot"},
				{markCommand, "@noslot"},
				{markCommand, "@eachslot"},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			l := newLexer([]byte(tt.input))
			for tok := l.nextToken(); tok.typ != tEOF; tok = l.nextToken() {
				if tok.typ == tError {
					t.Fatalf("unexpected error token: %s", tok.lit)
				}
			}
			got := make([]markText, 0, len(l.marks))
			for _, m := range l.marks {
				got = append(got, markText{m.kind, tt.input[m.start:m.end]})
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("marks mismatch\n got: %v\nwant: %v", got, tt.want)
			}
		})
	}
}
