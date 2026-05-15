package proxy

import (
	"testing"

	"github.com/stackus/goht/internal/protocol"
)

func TestDocumentApply(t *testing.T) {
	tests := map[string]struct {
		initial string
		rng     *protocol.Range
		text    string
		want    string
	}{
		"whole document replacement": {
			initial: "one\ntwo",
			text:    "alpha\nbeta",
			want:    "alpha\nbeta",
		},
		"same line insert": {
			initial: "hello world",
			rng:     rng(0, 5, 0, 5),
			text:    ",",
			want:    "hello, world",
		},
		"multi-line insert": {
			initial: "alpha omega\nlast",
			rng:     rng(0, 6, 0, 6),
			text:    "bravo\ncharlie\n",
			want:    "alpha bravo\ncharlie\nomega\nlast",
		},
		"same line delete": {
			initial: "hello, world",
			rng:     rng(0, 5, 0, 6),
			want:    "hello world",
		},
		"multi-line delete": {
			initial: "alpha bravo\ncharlie delta\necho foxtrot",
			rng:     rng(0, 6, 1, 8),
			want:    "alpha delta\necho foxtrot",
		},
		"same line replacement": {
			initial: "hello old world",
			rng:     rng(0, 6, 0, 9),
			text:    "new",
			want:    "hello new world",
		},
		"replacement from document start is not whole document unless range reaches document end": {
			initial: "old first\nsecond",
			rng:     rng(0, 0, 0, len("old first")),
			text:    "new first",
			want:    "new first\nsecond",
		},
		"same line replacement with newline text": {
			initial: "hello old world",
			rng:     rng(0, 6, 0, 9),
			text:    "new\nwide",
			want:    "hello new\nwide world",
		},
		"multi-line replacement": {
			initial: "alpha bravo\ncharlie delta\necho foxtrot",
			rng:     rng(0, 6, 1, 7),
			text:    "one\ntwo\nthree",
			want:    "alpha one\ntwo\nthree delta\necho foxtrot",
		},
		"BMP character before edit range uses UTF-8 byte offsets": {
			initial: "cafe é tail",
			rng:     rng(0, len("cafe é"), 0, len("cafe é")),
			text:    " strong",
			want:    "cafe é strong tail",
		},
		"surrogate-pair character before edit range uses UTF-8 byte offsets": {
			initial: "face 𐐀 tail",
			rng:     rng(0, len("face 𐐀"), 0, len("face 𐐀")),
			text:    " strong",
			want:    "face 𐐀 strong tail",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			doc := NewDocument(tt.initial)
			doc.Apply(tt.rng, tt.text)

			if got := doc.String(); got != tt.want {
				t.Fatalf("Document.Apply() = %q, want %q", got, tt.want)
			}
		})
	}
}

func rng(startLine, startChar, endLine, endChar int) *protocol.Range {
	return &protocol.Range{
		Start: protocol.Position{
			Line:      uint32(startLine),
			Character: uint32(startChar),
		},
		End: protocol.Position{
			Line:      uint32(endLine),
			Character: uint32(endChar),
		},
	}
}
