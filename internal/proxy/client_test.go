package proxy

import (
	"context"
	"testing"

	"github.com/rs/zerolog"

	"github.com/stackus/goht/compiler"
	"github.com/stackus/goht/internal/protocol"
)

const testGohtGoURI = protocol.DocumentURI("file:///tmp/test.goht.go")

func TestClientPublishDiagnosticsPassesThroughNonGohtDiagnostics(t *testing.T) {
	client := &recordingClient{}
	proxyClient := NewClient(client, NewSourceMapCache(), NewDiagnosticsCache(), zerolog.Nop())
	params := &protocol.PublishDiagnosticsParams{
		URI:         protocol.DocumentURI("file:///tmp/main.go"),
		Diagnostics: []protocol.Diagnostic{diagnosticWithRange("go", rangeOf(3, 4, 3, 7))},
	}

	err := proxyClient.PublishDiagnostics(context.Background(), params)
	if err != nil {
		t.Fatalf("PublishDiagnostics() error = %v", err)
	}

	if len(client.diagnostics) != 1 {
		t.Fatalf("diagnostic publishes = %d, want 1", len(client.diagnostics))
	}
	if client.diagnostics[0].URI != params.URI {
		t.Fatalf("URI = %q, want %q", client.diagnostics[0].URI, params.URI)
	}
	assertDiagnostics(t, client.diagnostics[0].Diagnostics, params.Diagnostics)
}

func TestClientPublishDiagnosticsRemapsGohtGoDiagnostics(t *testing.T) {
	client := &recordingClient{}
	smc := NewSourceMapCache()
	smc.Set(string(testGohtURI), testSourceMap())
	proxyClient := NewClient(client, smc, NewDiagnosticsCache(), zerolog.Nop())

	err := proxyClient.PublishDiagnostics(context.Background(), &protocol.PublishDiagnosticsParams{
		URI: testGohtGoURI,
		Diagnostics: []protocol.Diagnostic{
			diagnosticWithRange("mapped", rangeOf(10, 20, 10, 22)),
		},
	})
	if err != nil {
		t.Fatalf("PublishDiagnostics() error = %v", err)
	}

	if len(client.diagnostics) != 1 {
		t.Fatalf("diagnostic publishes = %d, want 1", len(client.diagnostics))
	}
	if client.diagnostics[0].URI != testGohtURI {
		t.Fatalf("URI = %q, want %q", client.diagnostics[0].URI, testGohtURI)
	}
	assertDiagnostics(t, client.diagnostics[0].Diagnostics, []protocol.Diagnostic{
		diagnosticWithRange("mapped", rangeOf(1, 2, 1, 4)),
	})
}

func TestClientPublishDiagnosticsDropsUnmappableGohtGoDiagnostics(t *testing.T) {
	tests := map[string]protocol.Range{
		"unmapped start":           rangeOf(99, 1, 10, 22),
		"unmapped end":             rangeOf(10, 20, 99, 1),
		"generated only full span": rangeOf(99, 1, 99, 2),
	}

	for name, diagnosticRange := range tests {
		t.Run(name, func(t *testing.T) {
			client := &recordingClient{}
			smc := NewSourceMapCache()
			smc.Set(string(testGohtURI), testSourceMap())
			proxyClient := NewClient(client, smc, NewDiagnosticsCache(), zerolog.Nop())

			err := proxyClient.PublishDiagnostics(context.Background(), &protocol.PublishDiagnosticsParams{
				URI:         testGohtGoURI,
				Diagnostics: []protocol.Diagnostic{diagnosticWithRange(name, diagnosticRange)},
			})
			if err != nil {
				t.Fatalf("PublishDiagnostics() error = %v", err)
			}

			if len(client.diagnostics) != 1 {
				t.Fatalf("diagnostic publishes = %d, want 1", len(client.diagnostics))
			}
			if got := client.diagnostics[0].Diagnostics; len(got) != 0 {
				t.Fatalf("diagnostics = %#v, want none", got)
			}
			if client.diagnostics[0].Diagnostics == nil {
				t.Fatalf("diagnostics = nil, want non-nil empty slice")
			}
		})
	}
}

func TestClientPublishDiagnosticsUsesMappedEndPosition(t *testing.T) {
	client := &recordingClient{}
	smc := NewSourceMapCache()
	smc.Set(string(testGohtURI), testSourceMap())
	proxyClient := NewClient(client, smc, NewDiagnosticsCache(), zerolog.Nop())

	err := proxyClient.PublishDiagnostics(context.Background(), &protocol.PublishDiagnosticsParams{
		URI: testGohtGoURI,
		Diagnostics: []protocol.Diagnostic{
			diagnosticWithRange("same line", rangeOf(10, 20, 10, 25)),
		},
	})
	if err != nil {
		t.Fatalf("PublishDiagnostics() error = %v", err)
	}

	assertDiagnostics(t, client.diagnostics[0].Diagnostics, []protocol.Diagnostic{
		diagnosticWithRange("same line", rangeOf(1, 2, 1, 6)),
	})
}

func diagnosticWithRange(message string, diagnosticRange protocol.Range) protocol.Diagnostic {
	diagnostic := testDiagnostic(message)
	diagnostic.Range = diagnosticRange
	return diagnostic
}

func rangeOf(startLine, startChar, endLine, endChar uint32) protocol.Range {
	return protocol.Range{
		Start: protocol.Position{Line: startLine, Character: startChar},
		End:   protocol.Position{Line: endLine, Character: endChar},
	}
}

func testSourceMap() *compiler.SourceMap {
	return &compiler.SourceMap{
		SourceLinesToTarget: map[int]map[int]compiler.Position{
			1: {
				2: {Line: 10, Col: 20},
				4: {Line: 10, Col: 22},
				6: {Line: 10, Col: 25},
			},
		},
		TargetLinesToSource: map[int]map[int]compiler.Position{
			10: {
				20: {Line: 1, Col: 2},
				22: {Line: 1, Col: 4},
				25: {Line: 1, Col: 6},
			},
		},
	}
}
