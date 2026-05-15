package proxy

import (
	"sync"

	"github.com/stackus/goht/internal/protocol"
)

type DiagnosticsCache struct {
	parserDiagnostics      map[string][]protocol.Diagnostic
	generatedGoDiagnostics map[string][]protocol.Diagnostic
	mu                     sync.Mutex
}

func NewDiagnosticsCache() *DiagnosticsCache {
	return &DiagnosticsCache{
		parserDiagnostics:      make(map[string][]protocol.Diagnostic),
		generatedGoDiagnostics: make(map[string][]protocol.Diagnostic),
	}
}

func (dc *DiagnosticsCache) WithParserDiagnostics(uri string, diagnostics []protocol.Diagnostic) []protocol.Diagnostic {
	dc.mu.Lock()
	defer dc.mu.Unlock()
	dc.parserDiagnostics[uri] = normalizeDiagnostics(diagnostics)
	return dc.merged(uri)
}

func (dc *DiagnosticsCache) WithGeneratedGoDiagnostics(uri string, diagnostics []protocol.Diagnostic) []protocol.Diagnostic {
	dc.mu.Lock()
	defer dc.mu.Unlock()
	dc.generatedGoDiagnostics[uri] = normalizeDiagnostics(diagnostics)
	return dc.merged(uri)
}

func (dc *DiagnosticsCache) ClearParserDiagnostics(uri string) []protocol.Diagnostic {
	dc.mu.Lock()
	defer dc.mu.Unlock()
	dc.parserDiagnostics[uri] = make([]protocol.Diagnostic, 0)
	return dc.merged(uri)
}

func (dc *DiagnosticsCache) Delete(uri string) {
	dc.mu.Lock()
	defer dc.mu.Unlock()
	delete(dc.parserDiagnostics, uri)
	delete(dc.generatedGoDiagnostics, uri)
}

func (dc *DiagnosticsCache) merged(uri string) []protocol.Diagnostic {
	parserDiagnostics := normalizeDiagnostics(dc.parserDiagnostics[uri])
	generatedGoDiagnostics := normalizeDiagnostics(dc.generatedGoDiagnostics[uri])
	diagnostics := make([]protocol.Diagnostic, 0, len(parserDiagnostics)+len(generatedGoDiagnostics))
	diagnostics = append(diagnostics, parserDiagnostics...)
	diagnostics = append(diagnostics, generatedGoDiagnostics...)
	return diagnostics
}

func normalizeDiagnostics(diagnostics []protocol.Diagnostic) []protocol.Diagnostic {
	if diagnostics == nil {
		return make([]protocol.Diagnostic, 0)
	}
	return diagnostics
}
