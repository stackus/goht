package proxy

import (
	"sync"

	"github.com/stackus/goht/internal/protocol"
)

type DiagnosticsCache struct {
	gohtDiagnostics map[string][]protocol.Diagnostic
	goDiagnostics   map[string][]protocol.Diagnostic
	mu              sync.Mutex
}

func NewDiagnosticsCache() *DiagnosticsCache {
	return &DiagnosticsCache{
		gohtDiagnostics: make(map[string][]protocol.Diagnostic),
		goDiagnostics:   make(map[string][]protocol.Diagnostic),
	}
}

func (dc *DiagnosticsCache) WithGohtDiagnostics(uri string, diagnostics []protocol.Diagnostic) []protocol.Diagnostic {
	if diagnostics == nil {
		diagnostics = make([]protocol.Diagnostic, 0)
	}
	dc.mu.Lock()
	defer dc.mu.Unlock()
	dc.goDiagnostics[uri] = diagnostics
	if dc.gohtDiagnostics[uri] == nil {
		dc.gohtDiagnostics[uri] = make([]protocol.Diagnostic, 0)
	}
	return append(dc.gohtDiagnostics[uri], diagnostics...)
}

func (dc *DiagnosticsCache) WithGoDiagnostics(uri string, diagnostics []protocol.Diagnostic) []protocol.Diagnostic {
	if diagnostics == nil {
		diagnostics = make([]protocol.Diagnostic, 0)
	}
	dc.mu.Lock()
	defer dc.mu.Unlock()
	dc.gohtDiagnostics[uri] = diagnostics
	if dc.goDiagnostics[uri] == nil {
		dc.goDiagnostics[uri] = make([]protocol.Diagnostic, 0)
	}
	return append(dc.goDiagnostics[uri], diagnostics...)
}

func (dc *DiagnosticsCache) ClearGohtDiagnostics(uri string) {
	dc.mu.Lock()
	defer dc.mu.Unlock()
	dc.gohtDiagnostics[uri] = make([]protocol.Diagnostic, 0)
}
