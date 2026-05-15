package proxy

import (
	"testing"

	"github.com/stackus/goht/internal/protocol"
)

func TestDiagnosticsCacheMergesSources(t *testing.T) {
	uri := string(testGohtURI)
	parserDiagnostic := testDiagnostic("parser")
	goDiagnostic := testDiagnostic("go")

	tests := map[string]struct {
		arrange func(*DiagnosticsCache) []protocol.Diagnostic
		want    []protocol.Diagnostic
	}{
		"parser only": {
			arrange: func(dc *DiagnosticsCache) []protocol.Diagnostic {
				return dc.WithParserDiagnostics(uri, []protocol.Diagnostic{parserDiagnostic})
			},
			want: []protocol.Diagnostic{parserDiagnostic},
		},
		"generated go only": {
			arrange: func(dc *DiagnosticsCache) []protocol.Diagnostic {
				return dc.WithGeneratedGoDiagnostics(uri, []protocol.Diagnostic{goDiagnostic})
			},
			want: []protocol.Diagnostic{goDiagnostic},
		},
		"parser diagnostics before generated go diagnostics": {
			arrange: func(dc *DiagnosticsCache) []protocol.Diagnostic {
				dc.WithParserDiagnostics(uri, []protocol.Diagnostic{parserDiagnostic})
				return dc.WithGeneratedGoDiagnostics(uri, []protocol.Diagnostic{goDiagnostic})
			},
			want: []protocol.Diagnostic{parserDiagnostic, goDiagnostic},
		},
		"clearing parser diagnostics leaves generated go diagnostics": {
			arrange: func(dc *DiagnosticsCache) []protocol.Diagnostic {
				dc.WithParserDiagnostics(uri, []protocol.Diagnostic{parserDiagnostic})
				dc.WithGeneratedGoDiagnostics(uri, []protocol.Diagnostic{goDiagnostic})
				return dc.ClearParserDiagnostics(uri)
			},
			want: []protocol.Diagnostic{goDiagnostic},
		},
		"nil parser diagnostics normalize to empty slice": {
			arrange: func(dc *DiagnosticsCache) []protocol.Diagnostic {
				return dc.WithParserDiagnostics(uri, nil)
			},
			want: []protocol.Diagnostic{},
		},
		"nil generated go diagnostics normalize to empty slice": {
			arrange: func(dc *DiagnosticsCache) []protocol.Diagnostic {
				return dc.WithGeneratedGoDiagnostics(uri, nil)
			},
			want: []protocol.Diagnostic{},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := tt.arrange(NewDiagnosticsCache())
			assertDiagnostics(t, got, tt.want)
			if got == nil {
				t.Fatalf("diagnostics = nil, want non-nil empty slice")
			}
		})
	}
}

func TestDiagnosticsCacheDeleteRemovesBothSources(t *testing.T) {
	uri := string(testGohtURI)
	dc := NewDiagnosticsCache()
	dc.WithParserDiagnostics(uri, []protocol.Diagnostic{testDiagnostic("parser")})
	dc.WithGeneratedGoDiagnostics(uri, []protocol.Diagnostic{testDiagnostic("go")})

	dc.Delete(uri)

	got := dc.WithParserDiagnostics(uri, nil)
	assertDiagnostics(t, got, []protocol.Diagnostic{})
}

func testDiagnostic(message string) protocol.Diagnostic {
	return protocol.Diagnostic{
		Range: protocol.Range{
			Start: protocol.Position{Line: 1, Character: 2},
			End:   protocol.Position{Line: 1, Character: 3},
		},
		Severity: protocol.SeverityError,
		Source:   "test",
		Message:  message,
	}
}

func assertDiagnostics(t *testing.T, got, want []protocol.Diagnostic) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("diagnostics = %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i].Message != want[i].Message {
			t.Fatalf("diagnostics[%d].Message = %q, want %q", i, got[i].Message, want[i].Message)
		}
		if got[i].Range != want[i].Range {
			t.Fatalf("diagnostics[%d].Range = %#v, want %#v", i, got[i].Range, want[i].Range)
		}
	}
}
