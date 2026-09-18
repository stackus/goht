package proxy

import (
	"context"
	"testing"

	"github.com/stackus/protocol"
)

func TestServerUnsupportedOperationsReturnEmptyResults(t *testing.T) {
	server := newTestServer(&recordingServer{}, &recordingClient{})
	ctx := context.Background()
	uri := protocol.TextDocumentIdentifier{URI: testGohtURI}

	if got, err := server.DocumentHighlight(ctx, &protocol.DocumentHighlightParams{TextDocumentPositionParams: protocol.TextDocumentPositionParams{TextDocument: uri}}); err != nil || len(got) != 0 {
		t.Fatalf("DocumentHighlight() = (%#v, %v)", got, err)
	}
	if got, err := server.DocumentLink(ctx, &protocol.DocumentLinkParams{TextDocument: uri}); err != nil || len(got) != 0 {
		t.Fatalf("DocumentLink() = (%#v, %v)", got, err)
	}
	if got, err := server.DocumentSymbol(ctx, &protocol.DocumentSymbolParams{TextDocument: uri}); err != nil || got != nil {
		t.Fatalf("DocumentSymbol() = (%#v, %v)", got, err)
	}
	if got, err := server.FoldingRanges(ctx, &protocol.FoldingRangeParams{}); err != nil || len(got) != 0 {
		t.Fatalf("FoldingRanges() = (%#v, %v)", got, err)
	}
	if got, err := server.Formatting(ctx, &protocol.DocumentFormattingParams{TextDocument: uri}); err != nil || len(got) != 0 {
		t.Fatalf("Formatting() = (%#v, %v)", got, err)
	}
	if got, err := server.InlayHint(ctx, &protocol.InlayHintParams{TextDocument: uri}); err != nil || len(got) != 0 {
		t.Fatalf("InlayHint() = (%#v, %v)", got, err)
	}
	if got, err := server.SemanticTokensFullDelta(ctx, &protocol.SemanticTokensDeltaParams{TextDocument: uri}); err != nil || got != nil {
		t.Fatalf("SemanticTokensFullDelta() = (%#v, %v)", got, err)
	}
}

func TestServerUpdateRange(t *testing.T) {
	server := newTestServer(&recordingServer{}, &recordingClient{})
	server.smc.Set(string(testGohtURI), testSourceMap())
	tests := map[string]struct {
		uri     protocol.DocumentURI
		rangeIn protocol.Range
		wantErr bool
	}{
		"maps GoHT range":        {uri: testGohtURI, rangeIn: rangeOf(1, 2, 1, 4)},
		"rejects non GoHT URI":   {uri: protocol.DocumentURI("file:///tmp/main.go"), rangeIn: rangeOf(1, 2, 1, 4), wantErr: true},
		"rejects unmapped range": {uri: testGohtURI, rangeIn: rangeOf(9, 2, 9, 4), wantErr: true},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, got, err := server.updateRangeZ(tt.uri, tt.rangeIn)

			if (err != nil) != tt.wantErr {
				t.Fatalf("updateRangeZ() error = %v, want error = %v", err, tt.wantErr)
			}
			if !tt.wantErr && got != rangeOf(10, 20, 10, 22) {
				t.Fatalf("updateRangeZ() = %#v", got)
			}
		})
	}
}
