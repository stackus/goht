package cmd

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stackus/protocol/jsonrpc2"

	"github.com/stackus/goht/internal/logging"
)

type testReadCloser struct{ *bytes.Reader }

func (testReadCloser) Close() error { return nil }

type testWriteCloser struct{ bytes.Buffer }

func (testWriteCloser) Close() error { return nil }

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

func TestRWCDelegates(t *testing.T) {
	reader := testReadCloser{Reader: bytes.NewReader([]byte("input"))}
	writer := &testWriteCloser{}
	stream := rwc{r: reader, w: writer}

	got, err := io.ReadAll(stream)

	if err != nil || string(got) != "input" {
		t.Fatalf("ReadAll() = (%q, %v)", got, err)
	}
	if _, err := stream.Write([]byte("output")); err != nil {
		t.Fatal(err)
	}
	if err := stream.Close(); err != nil || writer.String() != "output" {
		t.Fatalf("Close() = %v, output = %q", err, writer.String())
	}
}

func TestFindAndStartGoPlsRequiresExecutable(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PATH", dir)
	if _, err := findAndStartGoPls(context.Background()); err == nil {
		t.Fatal("findAndStartGoPls() error = nil")
	}

	if _, err := os.Stat(filepath.Join(dir, "gopls")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("temporary PATH unexpectedly contains gopls: %v", err)
	}
}

func TestFindAndStartGoPlsStartsExecutable(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "gopls")

	if err := os.WriteFile(path, []byte("#!/bin/sh\nexit 0\n"), 0700); err != nil {
		t.Fatal(err)
	}

	t.Setenv("PATH", dir)

	stream, err := findAndStartGoPls(context.Background())

	if err != nil {
		t.Fatalf("findAndStartGoPls() error = %v", err)
	}
	if err := stream.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}

func TestRunLspContextStopsOnCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	input := testReadCloser{Reader: bytes.NewReader(nil)}
	output := &testWriteCloser{}
	started := false

	err := runLspContext(ctx, input, output, func(context.Context) (io.ReadWriteCloser, error) {
		started = true
		return rwc{r: testReadCloser{Reader: bytes.NewReader(nil)}, w: &testWriteCloser{}}, nil
	})

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("runLspContext() error = %v, want context canceled", err)
	}
	if !started {
		t.Fatal("runLspContext() did not start gopls")
	}
}
