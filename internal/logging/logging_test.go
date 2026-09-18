package logging

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stackus/protocol/jsonrpc2"
)

type recordingStream struct {
	readMsg    jsonrpc2.Message
	readCount  int64
	readErr    error
	writeCount int64
	writeErr   error
	closed     bool
}

func (s *recordingStream) Read(context.Context) (jsonrpc2.Message, int64, error) {
	return s.readMsg, s.readCount, s.readErr
}

func (s *recordingStream) Write(context.Context, jsonrpc2.Message) (int64, error) {
	return s.writeCount, s.writeErr
}

func (s *recordingStream) Close() error {
	s.closed = true
	return nil
}

func resetMaps() {
	maps.mu.Lock()
	defer maps.mu.Unlock()
	maps.clientCalls = make(map[string]req)
	maps.serverCalls = make(map[string]req)
}

func TestNewLoggerFormatsConsoleOutput(t *testing.T) {
	var output bytes.Buffer
	logger := NewLogger(&output, zerolog.DebugLevel)
	logger.Info().Str("key", "value").Msg("hello")

	got := output.String()
	for _, want := range []string{"[INFO ]", "hello", "key:value"} {
		if !strings.Contains(got, want) {
			t.Fatalf("log output = %q, want %q", got, want)
		}
	}
}

func TestLoggedStreamReadWriteAndClose(t *testing.T) {
	call, err := jsonrpc2.NewCall(jsonrpc2.NewIntID(1), "initialize", map[string]string{"x": "y"})
	if err != nil {
		t.Fatal(err)
	}

	tests := map[string]struct {
		read      bool
		readErr   error
		writeErr  error
		wantCount int64
	}{
		"read forwards message and count":  {read: true, wantCount: 7},
		"write forwards message and count": {wantCount: 9},
		"read preserves error":             {read: true, readErr: io.EOF, wantCount: 7},
		"write preserves error":            {writeErr: errors.New("write failed"), wantCount: 9},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			resetMaps()
			stream := &recordingStream{readMsg: call, readCount: 7, readErr: tt.readErr, writeCount: 9, writeErr: tt.writeErr}
			logged := LoggedStream{Label: "test", Stream: stream, Logger: zerolog.Nop()}

			if tt.read {
				msg, count, err := logged.Read(context.Background())

				if msg != call || count != tt.wantCount || !errors.Is(err, tt.readErr) {
					t.Fatalf("Read() = (%v, %d, %v)", msg, count, err)
				}
			} else {
				count, err := logged.Write(context.Background(), call)

				if count != tt.wantCount || !errors.Is(err, tt.writeErr) {
					t.Fatalf("Write() = (%d, %v)", count, err)
				}
			}

			if err := logged.Close(); err != nil || !stream.closed {
				t.Fatalf("Close() = %v, closed = %v", err, stream.closed)
			}
		})
	}
}

func TestLoggedStreamLogsMessageKinds(t *testing.T) {
	call, _ := jsonrpc2.NewCall(jsonrpc2.NewStringID("call"), "request", map[string]string{})
	notification, _ := jsonrpc2.NewNotification("notice", map[string]string{})
	response, _ := jsonrpc2.NewResponse(jsonrpc2.NewStringID("call"), map[string]string{"ok": "yes"}, nil)
	failedResponse, _ := jsonrpc2.NewResponse(jsonrpc2.NewStringID("failed"), nil, errors.New("failed"))

	tests := map[string]struct {
		message jsonrpc2.Message
		isRead  bool
	}{
		"nil message":         {},
		"outgoing call":       {message: call},
		"incoming call":       {message: call, isRead: true},
		"notification":        {message: notification},
		"successful response": {message: response},
		"failed response":     {message: failedResponse},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			resetMaps()

			logged := LoggedStream{Label: "test", Logger: zerolog.Nop()}

			logged.logMsg(tt.message, tt.isRead)
		})
	}
}
