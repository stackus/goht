package proxy

import (
	"strings"
	"testing"

	"github.com/stackus/protocol"
)

func TestDocumentContentsLifecycle(t *testing.T) {
	contents := NewDocumentContents()
	first := NewDocument("first")
	second := NewDocument("second")
	contents.Set("first", first)
	contents.Set("second", second)

	got, ok := contents.Get("first")
	if !ok || got != first {
		t.Fatalf("Get(first) = (%v, %v), want (%v, true)", got, ok, first)
	}
	if _, ok := contents.Get("missing"); ok {
		t.Fatal("Get(missing) found document")
	}

	uris := strings.Join(contents.URIs(), ",")
	if !strings.Contains(uris, "first") || !strings.Contains(uris, "second") {
		t.Fatalf("URIs() = %q, want both documents", uris)
	}

	contents.Delete("first")
	if _, ok := contents.Get("first"); ok {
		t.Fatal("Delete() did not remove document")
	}
}

func TestDocumentContentsApply(t *testing.T) {
	tests := map[string]struct {
		encoding protocol.PositionEncodingKind
		changes  []protocol.TextDocumentContentChangeEvent
		want     string
		wantErr  string
	}{
		"utf8 accepts ranged edits": {
			encoding: protocol.UTF8,
			changes:  []protocol.TextDocumentContentChangeEvent{{Range: rng(0, 1, 0, 2), Text: "X"}},
			want:     "aXc",
		},
		"non utf8 rejects ranged edits": {
			encoding: protocol.UTF16,
			changes:  []protocol.TextDocumentContentChangeEvent{{Range: rng(0, 1, 0, 2), Text: "X"}},
			wantErr:  "utf-8",
		},
		"non utf8 accepts whole document edits": {
			encoding: protocol.UTF16,
			changes:  []protocol.TextDocumentContentChangeEvent{{Text: "replacement"}},
			want:     "replacement",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			contents := NewDocumentContents()
			contents.Set("file", NewDocument("abc"))

			document, err := contents.Apply("file", tt.changes, tt.encoding)

			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("Apply() error = %v, want %q", err, tt.wantErr)
				}
				return
			}
			if err != nil || document.String() != tt.want {
				t.Fatalf("Apply() = (%v, %v), want %q", document, err, tt.want)
			}
		})
	}

	if _, err := NewDocumentContents().Apply("missing", nil, protocol.UTF8); err == nil {
		t.Fatal("Apply(missing) error = nil")
	}
}

func TestSourceMapCacheLifecycle(t *testing.T) {
	cache := NewSourceMapCache()
	mapValue := testSourceMap()
	cache.Set("file", mapValue)

	if got, ok := cache.Get("file"); !ok || got != mapValue {
		t.Fatalf("Get(file) = (%v, %v)", got, ok)
	}

	cache.Delete("file")

	if _, ok := cache.Get("file"); ok {
		t.Fatal("Delete() did not remove source map")
	}
}
