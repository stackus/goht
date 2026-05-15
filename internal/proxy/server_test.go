package proxy

import (
	"context"
	"testing"

	"github.com/rs/zerolog"

	"github.com/stackus/goht/internal/protocol"
)

const (
	testGohtURI = protocol.DocumentURI("file:///tmp/test.goht")

	validGoht      = "package main\n\n@goht Test() {\n\t%p hello\n}\n"
	otherValidGoht = "package main\n\n@goht Test() {\n\t%p goodbye\n}\n"
	invalidGoht    = "@goht Test() {\n"
)

type recordingServer struct {
	protocol.Server
	initializeResult *protocol.InitializeResult
	didOpenCalls     []protocol.DidOpenTextDocumentParams
	didChangeCalls   []protocol.DidChangeTextDocumentParams
	didCloseCalls    []protocol.DidCloseTextDocumentParams
}

func (s *recordingServer) Initialize(context.Context, *protocol.ParamInitialize) (*protocol.InitializeResult, error) {
	if s.initializeResult != nil {
		return s.initializeResult, nil
	}
	return &protocol.InitializeResult{
		ServerInfo: &protocol.ServerInfo{},
	}, nil
}

func (s *recordingServer) DidOpen(_ context.Context, params *protocol.DidOpenTextDocumentParams) error {
	s.didOpenCalls = append(s.didOpenCalls, *params)
	return nil
}

func (s *recordingServer) DidChange(_ context.Context, params *protocol.DidChangeTextDocumentParams) error {
	s.didChangeCalls = append(s.didChangeCalls, *params)
	return nil
}

func (s *recordingServer) DidClose(_ context.Context, params *protocol.DidCloseTextDocumentParams) error {
	s.didCloseCalls = append(s.didCloseCalls, *params)
	return nil
}

type recordingClient struct {
	protocol.Client
	diagnostics []protocol.PublishDiagnosticsParams
}

func (c *recordingClient) PublishDiagnostics(_ context.Context, params *protocol.PublishDiagnosticsParams) error {
	c.diagnostics = append(c.diagnostics, *params)
	return nil
}

func TestServerInitializePositionEncoding(t *testing.T) {
	t.Run("selects UTF-8 and incremental sync when client supports UTF-8", func(t *testing.T) {
		server := &recordingServer{}
		client := &recordingClient{}
		proxy := newTestServer(server, client)

		got, err := proxy.Initialize(context.Background(), &protocol.ParamInitialize{
			XInitializeParams: protocol.XInitializeParams{
				Capabilities: protocol.ClientCapabilities{
					General: &protocol.GeneralClientCapabilities{
						PositionEncodings: []protocol.PositionEncodingKind{protocol.UTF8, protocol.UTF16},
					},
				},
			},
		})
		if err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}
		if got.Capabilities.PositionEncoding == nil || *got.Capabilities.PositionEncoding != protocol.UTF8 {
			t.Fatalf("PositionEncoding = %v, want utf-8", got.Capabilities.PositionEncoding)
		}
		sync, ok := got.Capabilities.TextDocumentSync.(protocol.TextDocumentSyncOptions)
		if !ok {
			t.Fatalf("TextDocumentSync = %T, want protocol.TextDocumentSyncOptions", got.Capabilities.TextDocumentSync)
		}
		if sync.Change != protocol.Incremental {
			t.Fatalf("TextDocumentSync.Change = %v, want Incremental", sync.Change)
		}
	})

	t.Run("keeps full sync when client does not support UTF-8", func(t *testing.T) {
		server := &recordingServer{}
		client := &recordingClient{}
		proxy := newTestServer(server, client)

		got, err := proxy.Initialize(context.Background(), &protocol.ParamInitialize{})
		if err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}
		if got.Capabilities.PositionEncoding != nil {
			t.Fatalf("PositionEncoding = %v, want nil UTF-16 default", got.Capabilities.PositionEncoding)
		}
		sync, ok := got.Capabilities.TextDocumentSync.(protocol.TextDocumentSyncOptions)
		if !ok {
			t.Fatalf("TextDocumentSync = %T, want protocol.TextDocumentSyncOptions", got.Capabilities.TextDocumentSync)
		}
		if sync.Change != protocol.Full {
			t.Fatalf("TextDocumentSync.Change = %v, want Full", sync.Change)
		}
	})
}

func TestServerDidOpenValidGohtOpensGeneratedGoDocument(t *testing.T) {
	server := &recordingServer{}
	client := &recordingClient{}
	proxy := newTestServer(server, client)

	err := proxy.DidOpen(context.Background(), didOpenParams(validGoht))
	if err != nil {
		t.Fatalf("DidOpen() error = %v", err)
	}

	if len(server.didOpenCalls) != 1 {
		t.Fatalf("DidOpen calls = %d, want 1", len(server.didOpenCalls))
	}
	if server.didOpenCalls[0].TextDocument.LanguageID != "go" {
		t.Fatalf("opened language = %q, want go", server.didOpenCalls[0].TextDocument.LanguageID)
	}
	if server.didOpenCalls[0].TextDocument.Text == validGoht {
		t.Fatalf("opened text was original GoHT, want generated Go")
	}
	if _, ok := proxy.smc.Get(string(testGohtURI)); !ok {
		t.Fatalf("source map missing for %s", testGohtURI)
	}
	if proxy.goSrcs[string(testGohtURI)] == "" {
		t.Fatalf("generated Go source missing for %s", testGohtURI)
	}
}

func TestServerDidOpenInvalidGohtSkipsGeneratedGoDocument(t *testing.T) {
	server := &recordingServer{}
	client := &recordingClient{}
	proxy := newTestServer(server, client)

	err := proxy.DidOpen(context.Background(), didOpenParams(invalidGoht))
	if err != nil {
		t.Fatalf("DidOpen() error = %v", err)
	}

	if len(server.didOpenCalls) != 0 {
		t.Fatalf("DidOpen calls = %d, want 0", len(server.didOpenCalls))
	}
	if len(client.diagnostics) == 0 {
		t.Fatalf("diagnostics were not published")
	}
	if _, ok := proxy.srcs.Get(string(testGohtURI)); !ok {
		t.Fatalf("GoHT document was not stored")
	}
	if _, ok := proxy.smc.Get(string(testGohtURI)); ok {
		t.Fatalf("source map stored for invalid document")
	}
	if proxy.goSrcs[string(testGohtURI)] != "" {
		t.Fatalf("generated Go source stored for invalid document")
	}
}

func TestServerDidChangeInvalidGohtKeepsLastValidGeneratedState(t *testing.T) {
	server := &recordingServer{}
	client := &recordingClient{}
	proxy := newTestServer(server, client)

	if err := proxy.DidOpen(context.Background(), didOpenParams(validGoht)); err != nil {
		t.Fatalf("DidOpen() error = %v", err)
	}
	initialGoSrc := proxy.goSrcs[string(testGohtURI)]
	initialSourceMap, ok := proxy.smc.Get(string(testGohtURI))
	if !ok {
		t.Fatalf("source map missing after valid open")
	}

	err := proxy.DidChange(context.Background(), didChangeParams(invalidGoht))
	if err != nil {
		t.Fatalf("DidChange() error = %v", err)
	}

	if len(server.didChangeCalls) != 0 {
		t.Fatalf("DidChange calls = %d, want 0", len(server.didChangeCalls))
	}
	if len(client.diagnostics) == 0 {
		t.Fatalf("diagnostics were not published")
	}
	if got := proxy.goSrcs[string(testGohtURI)]; got != initialGoSrc {
		t.Fatalf("generated Go source changed after parse error")
	}
	gotSourceMap, ok := proxy.smc.Get(string(testGohtURI))
	if !ok {
		t.Fatalf("source map missing after parse error")
	}
	if gotSourceMap != initialSourceMap {
		t.Fatalf("source map changed after parse error")
	}
}

func TestServerDidChangeValidAfterInvalidOpenOpensGeneratedGoDocument(t *testing.T) {
	server := &recordingServer{}
	client := &recordingClient{}
	proxy := newTestServer(server, client)

	if err := proxy.DidOpen(context.Background(), didOpenParams(invalidGoht)); err != nil {
		t.Fatalf("DidOpen() error = %v", err)
	}
	if err := proxy.DidChange(context.Background(), didChangeParams(otherValidGoht)); err != nil {
		t.Fatalf("DidChange() error = %v", err)
	}

	if len(server.didOpenCalls) != 1 {
		t.Fatalf("DidOpen calls = %d, want 1", len(server.didOpenCalls))
	}
	if len(server.didChangeCalls) != 0 {
		t.Fatalf("DidChange calls = %d, want 0", len(server.didChangeCalls))
	}
	if proxy.goSrcs[string(testGohtURI)] == "" {
		t.Fatalf("generated Go source missing after valid change")
	}
	if len(client.diagnostics) < 2 {
		t.Fatalf("diagnostic publish calls = %d, want at least 2", len(client.diagnostics))
	}
	if got := client.diagnostics[len(client.diagnostics)-1].Diagnostics; len(got) != 0 {
		t.Fatalf("last diagnostics = %v, want cleared diagnostics", got)
	}
}

func TestServerDidCloseClearsGohtState(t *testing.T) {
	server := &recordingServer{}
	client := &recordingClient{}
	proxy := newTestServer(server, client)

	if err := proxy.DidOpen(context.Background(), didOpenParams(validGoht)); err != nil {
		t.Fatalf("DidOpen() error = %v", err)
	}
	proxy.dc.WithParserDiagnostics(string(testGohtURI), []protocol.Diagnostic{testDiagnostic("parser")})
	proxy.dc.WithGeneratedGoDiagnostics(string(testGohtURI), []protocol.Diagnostic{testDiagnostic("go")})

	err := proxy.DidClose(context.Background(), &protocol.DidCloseTextDocumentParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: testGohtURI},
	})
	if err != nil {
		t.Fatalf("DidClose() error = %v", err)
	}

	if len(server.didCloseCalls) != 1 {
		t.Fatalf("DidClose calls = %d, want 1", len(server.didCloseCalls))
	}
	if server.didCloseCalls[0].TextDocument.URI != testGohtGoURI {
		t.Fatalf("closed URI = %q, want %q", server.didCloseCalls[0].TextDocument.URI, testGohtGoURI)
	}
	if _, ok := proxy.srcs.Get(string(testGohtURI)); ok {
		t.Fatalf("document contents still cached for %s", testGohtURI)
	}
	if _, ok := proxy.goSrcs[string(testGohtURI)]; ok {
		t.Fatalf("generated source still cached for %s", testGohtURI)
	}
	if _, ok := proxy.smc.Get(string(testGohtURI)); ok {
		t.Fatalf("source map still cached for %s", testGohtURI)
	}
	got := proxy.dc.WithParserDiagnostics(string(testGohtURI), nil)
	assertDiagnostics(t, got, []protocol.Diagnostic{})
}

func TestServerRangeConversionRequiresBothEndpoints(t *testing.T) {
	server := &recordingServer{}
	client := &recordingClient{}
	proxy := newTestServer(server, client)
	proxy.smc.Set(string(testGohtURI), testSourceMap())

	tests := map[string]struct {
		convert func(protocol.Range) protocol.Range
		input   protocol.Range
		want    protocol.Range
	}{
		"go range to goht range maps both endpoints": {
			convert: func(r protocol.Range) protocol.Range {
				return proxy.goRangeToGohtRange(testGohtURI, r)
			},
			input: rangeOf(10, 20, 10, 22),
			want:  rangeOf(1, 2, 1, 4),
		},
		"go range to goht range keeps original when end unmapped": {
			convert: func(r protocol.Range) protocol.Range {
				return proxy.goRangeToGohtRange(testGohtURI, r)
			},
			input: rangeOf(10, 20, 99, 1),
			want:  rangeOf(10, 20, 99, 1),
		},
		"goht range to go range maps both endpoints": {
			convert: func(r protocol.Range) protocol.Range {
				return proxy.gohtRangeToGoRange(testGohtURI, r)
			},
			input: rangeOf(1, 2, 1, 4),
			want:  rangeOf(10, 20, 10, 22),
		},
		"goht range to go range keeps original when start unmapped": {
			convert: func(r protocol.Range) protocol.Range {
				return proxy.gohtRangeToGoRange(testGohtURI, r)
			},
			input: rangeOf(99, 1, 1, 4),
			want:  rangeOf(99, 1, 1, 4),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.convert(tt.input); got != tt.want {
				t.Fatalf("range = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestServerUpdatePosition(t *testing.T) {
	server := &recordingServer{}
	client := &recordingClient{}
	proxy := newTestServer(server, client)
	proxy.smc.Set(string(testGohtURI), testSourceMap())

	uri, pos, err := proxy.updatePosition(testGohtURI, protocol.Position{Line: 1, Character: 2})
	if err != nil {
		t.Fatalf("updatePosition() error = %v", err)
	}
	if uri != testGohtGoURI {
		t.Fatalf("URI = %q, want %q", uri, testGohtGoURI)
	}
	if pos != (protocol.Position{Line: 10, Character: 20}) {
		t.Fatalf("position = %#v, want line 10 character 20", pos)
	}

	if _, _, err := proxy.updatePosition(testGohtURI, protocol.Position{Line: 99, Character: 1}); err == nil {
		t.Fatalf("updatePosition() error = nil, want unmapped position error")
	}

	missingMapServer := newTestServer(&recordingServer{}, &recordingClient{})
	if _, _, err := missingMapServer.updatePosition(testGohtURI, protocol.Position{Line: 1, Character: 2}); err == nil {
		t.Fatalf("updatePosition() error = nil, want missing sourcemap error")
	}
}

func newTestServer(server *recordingServer, client *recordingClient) *Server {
	return NewServer(
		server,
		client,
		NewSourceMapCache(),
		NewDiagnosticsCache(),
		NewDocumentContents(),
		zerolog.Nop(),
	)
}

func didOpenParams(text string) *protocol.DidOpenTextDocumentParams {
	return &protocol.DidOpenTextDocumentParams{
		TextDocument: protocol.TextDocumentItem{
			URI:        testGohtURI,
			LanguageID: "goht",
			Version:    1,
			Text:       text,
		},
	}
}

func didChangeParams(text string) *protocol.DidChangeTextDocumentParams {
	return &protocol.DidChangeTextDocumentParams{
		TextDocument: protocol.VersionedTextDocumentIdentifier{
			Version: 2,
			TextDocumentIdentifier: protocol.TextDocumentIdentifier{
				URI: testGohtURI,
			},
		},
		ContentChanges: []protocol.TextDocumentContentChangeEvent{
			{Text: text},
		},
	}
}
