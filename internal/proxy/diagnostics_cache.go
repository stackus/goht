package proxy

import (
	"sync"

	"github.com/stackus/hamlet/internal/protocol"
)

type DiagnosticsCache struct {
	hamletDiagnostics map[string][]protocol.Diagnostic
	goDiagnostics     map[string][]protocol.Diagnostic
	mu                sync.Mutex
}

func NewDiagnosticsCache() *DiagnosticsCache {
	return &DiagnosticsCache{
		hamletDiagnostics: make(map[string][]protocol.Diagnostic),
		goDiagnostics:     make(map[string][]protocol.Diagnostic),
	}
}

func (dc *DiagnosticsCache) WithHamletDiagnostics(uri string, diagnostics []protocol.Diagnostic) []protocol.Diagnostic {
	if diagnostics == nil {
		diagnostics = make([]protocol.Diagnostic, 0)
	}
	dc.mu.Lock()
	defer dc.mu.Unlock()
	dc.goDiagnostics[uri] = diagnostics
	if dc.hamletDiagnostics[uri] == nil {
		dc.hamletDiagnostics[uri] = make([]protocol.Diagnostic, 0)
	}
	return append(dc.hamletDiagnostics[uri], diagnostics...)
}

func (dc *DiagnosticsCache) WithGoDiagnostics(uri string, diagnostics []protocol.Diagnostic) []protocol.Diagnostic {
	if diagnostics == nil {
		diagnostics = make([]protocol.Diagnostic, 0)
	}
	dc.mu.Lock()
	defer dc.mu.Unlock()
	dc.hamletDiagnostics[uri] = diagnostics
	if dc.goDiagnostics[uri] == nil {
		dc.goDiagnostics[uri] = make([]protocol.Diagnostic, 0)
	}
	return append(dc.goDiagnostics[uri], diagnostics...)
}

func (dc *DiagnosticsCache) ClearHamletDiagnostics(uri string) {
	dc.mu.Lock()
	defer dc.mu.Unlock()
	dc.hamletDiagnostics[uri] = make([]protocol.Diagnostic, 0)
}
