package compiler

import (
	"strings"
	"testing"
)

func TestLexerTokenSpans(t *testing.T) {
	tests := map[string]string{
		"haml": "package main\n\n@haml Foo() {\n\t%p.bg-red/[0.5]#main{class: \"x\"} héllo #{x}\n\t%a(@click=\"go()\" data-[state]=\"on\") Link\n\t-# a comment\n}\n",
		"slim": "package main\n\n@slim Foo() {\n\tp.hover:bg-red/80 héllo #{x}\n\ta(@click=\"go()\" x-on:click=\"y\") Link\n\t| Save \\#{title} now\n}\n",
		"ego":  "package main\n\n@ego Foo() {\n\t<p>héllo <%= x %></p>\n\t<%@slot header -%>\n}\n",
	}
	for name, input := range tests {
		t.Run(name, func(t *testing.T) {
			l := newLexer([]byte(input))
			prevEnd := 0
			for {
				tok := l.nextToken()
				if tok.typ == tEOF {
					break
				}
				if tok.typ == tError {
					t.Fatalf("unexpected error token: %s", tok.lit)
				}
				if tok.start < prevEnd || tok.start > tok.end || tok.end > len(input) {
					t.Fatalf("%s: bad span [%d,%d) after %d (len %d)", tok, tok.start, tok.end, prevEnd, len(input))
				}
				prevEnd = tok.end
				switch tok.typ {
				case tTag, tId, tClass, tAttrName, tAttrEscapedValue, tGoCode:
					if span := input[tok.start:tok.end]; !strings.Contains(span, tok.lit) {
						t.Errorf("%s: span %q does not contain lit %q", tok, span, tok.lit)
					}
				}
			}
		})
	}
}
