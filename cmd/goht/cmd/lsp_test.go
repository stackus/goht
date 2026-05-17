package cmd

import (
	"context"
	"testing"

	"github.com/rs/zerolog"
	"go.lsp.dev/jsonrpc2"

	"github.com/stackus/goht/internal/logging"
)

type noopStream struct{}

func (noopStream) Read(context.Context) (jsonrpc2.Message, int64, error) {
	return nil, 0, nil
}

func (noopStream) Write(context.Context, jsonrpc2.Message) (int64, error) {
	return 0, nil
}

func (noopStream) Close() error {
	return nil
}

func TestNewTraceableStream(t *testing.T) {
	logger := zerolog.Nop()

	tests := map[string]struct {
		enabled bool
		label   string
		wantLog bool
	}{
		"disabled leaves stream unwrapped": {
			enabled: false,
			label:   "GOHT-LSP",
			wantLog: false,
		},
		"enabled wraps stream with label": {
			enabled: true,
			label:   "GO-LSP",
			wantLog: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := newTraceableStream(noopStream{}, tt.enabled, tt.label, logger)
			logged, ok := got.(logging.LoggedStream)
			if ok != tt.wantLog {
				t.Fatalf("logged stream = %v, want %v", ok, tt.wantLog)
			}
			if ok && logged.Label != tt.label {
				t.Fatalf("logged label = %q, want %q", logged.Label, tt.label)
			}
		})
	}
}
