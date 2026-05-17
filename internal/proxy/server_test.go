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
	initializeResult     *protocol.InitializeResult
	completionResult     *protocol.CompletionList
	completionCalls      []protocol.CompletionParams
	definitionResult     []protocol.Location
	definitionCalls      []protocol.DefinitionParams
	implementationResult []protocol.Location
	implementationCalls  []protocol.ImplementationParams
	referencesResult     []protocol.Location
	referencesCalls      []protocol.ReferenceParams
	typeDefinitionResult []protocol.Location
	typeDefinitionCalls  []protocol.TypeDefinitionParams
	didOpenCalls         []protocol.DidOpenTextDocumentParams
	didChangeCalls       []protocol.DidChangeTextDocumentParams
	didCloseCalls        []protocol.DidCloseTextDocumentParams
}

func (s *recordingServer) Initialize(context.Context, *protocol.ParamInitialize) (*protocol.InitializeResult, error) {
	if s.initializeResult != nil {
		return s.initializeResult, nil
	}
	return &protocol.InitializeResult{
		ServerInfo: &protocol.ServerInfo{},
	}, nil
}

func (s *recordingServer) Completion(_ context.Context, params *protocol.CompletionParams) (*protocol.CompletionList, error) {
	s.completionCalls = append(s.completionCalls, *params)
	return s.completionResult, nil
}

func (s *recordingServer) Definition(_ context.Context, params *protocol.DefinitionParams) ([]protocol.Location, error) {
	s.definitionCalls = append(s.definitionCalls, *params)
	return s.definitionResult, nil
}

func (s *recordingServer) Implementation(_ context.Context, params *protocol.ImplementationParams) ([]protocol.Location, error) {
	s.implementationCalls = append(s.implementationCalls, *params)
	return s.implementationResult, nil
}

func (s *recordingServer) References(_ context.Context, params *protocol.ReferenceParams) ([]protocol.Location, error) {
	s.referencesCalls = append(s.referencesCalls, *params)
	return s.referencesResult, nil
}

func (s *recordingServer) TypeDefinition(_ context.Context, params *protocol.TypeDefinitionParams) ([]protocol.Location, error) {
	s.typeDefinitionCalls = append(s.typeDefinitionCalls, *params)
	return s.typeDefinitionResult, nil
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

func TestServerInitializeSanitizesUnsupportedCapabilities(t *testing.T) {
	server := &recordingServer{
		initializeResult: &protocol.InitializeResult{
			Capabilities: protocol.ServerCapabilities{
				CompletionProvider:               &protocol.CompletionOptions{},
				DocumentFormattingProvider:       &protocol.Or_ServerCapabilities_documentFormattingProvider{Value: true},
				DocumentRangeFormattingProvider:  &protocol.Or_ServerCapabilities_documentRangeFormattingProvider{Value: true},
				DocumentOnTypeFormattingProvider: &protocol.DocumentOnTypeFormattingOptions{FirstTriggerCharacter: "."},
				DocumentHighlightProvider:        &protocol.Or_ServerCapabilities_documentHighlightProvider{Value: true},
				DocumentLinkProvider:             &protocol.DocumentLinkOptions{},
				DocumentSymbolProvider:           &protocol.Or_ServerCapabilities_documentSymbolProvider{Value: true},
				ExecuteCommandProvider:           &protocol.ExecuteCommandOptions{Commands: []string{"gopls.test"}},
				FoldingRangeProvider:             &protocol.Or_ServerCapabilities_foldingRangeProvider{Value: true},
				InlayHintProvider:                &protocol.Or_ServerCapabilities_inlayHintProvider{Value: true},
				SemanticTokensProvider:           &protocol.SemanticTokensOptions{},
			},
			ServerInfo: &protocol.ServerInfo{},
		},
	}
	proxy := newTestServer(server, &recordingClient{})

	got, err := proxy.Initialize(context.Background(), &protocol.ParamInitialize{})
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	if got.Capabilities.CompletionProvider == nil {
		t.Fatalf("CompletionProvider = nil, want completion support preserved")
	}
	if got.Capabilities.ExecuteCommandProvider != nil {
		t.Fatalf("ExecuteCommandProvider = %#v, want nil", got.Capabilities.ExecuteCommandProvider)
	}
	if got.Capabilities.DocumentFormattingProvider == nil || got.Capabilities.DocumentFormattingProvider.Value != false {
		t.Fatalf("DocumentFormattingProvider = %#v, want false", got.Capabilities.DocumentFormattingProvider)
	}
	if got.Capabilities.DocumentRangeFormattingProvider == nil || got.Capabilities.DocumentRangeFormattingProvider.Value != false {
		t.Fatalf("DocumentRangeFormattingProvider = %#v, want false", got.Capabilities.DocumentRangeFormattingProvider)
	}
	if got.Capabilities.DocumentOnTypeFormattingProvider != nil {
		t.Fatalf("DocumentOnTypeFormattingProvider = %#v, want nil", got.Capabilities.DocumentOnTypeFormattingProvider)
	}
	if got.Capabilities.DocumentHighlightProvider != nil {
		t.Fatalf("DocumentHighlightProvider = %#v, want nil", got.Capabilities.DocumentHighlightProvider)
	}
	if got.Capabilities.DocumentLinkProvider != nil {
		t.Fatalf("DocumentLinkProvider = %#v, want nil", got.Capabilities.DocumentLinkProvider)
	}
	if got.Capabilities.DocumentSymbolProvider != nil {
		t.Fatalf("DocumentSymbolProvider = %#v, want nil", got.Capabilities.DocumentSymbolProvider)
	}
	if got.Capabilities.FoldingRangeProvider != nil {
		t.Fatalf("FoldingRangeProvider = %#v, want nil", got.Capabilities.FoldingRangeProvider)
	}
	if got.Capabilities.InlayHintProvider != nil {
		t.Fatalf("InlayHintProvider = %#v, want nil", got.Capabilities.InlayHintProvider)
	}
	if got.Capabilities.SemanticTokensProvider != nil {
		t.Fatalf("SemanticTokensProvider = %#v, want nil", got.Capabilities.SemanticTokensProvider)
	}
}

func TestServerCompletionMapsAdditionalTextEdits(t *testing.T) {
	server := &recordingServer{
		completionResult: &protocol.CompletionList{
			Items: []protocol.CompletionItem{
				{
					Label: "Thing",
					TextEdit: &protocol.TextEdit{
						Range:   rangeOf(10, 20, 10, 22),
						NewText: "Thing",
					},
					AdditionalTextEdits: []protocol.TextEdit{
						{
							Range:   rangeOf(10, 20, 10, 22),
							NewText: "mapped edit",
						},
					},
				},
			},
		},
	}
	proxy := newTestServer(server, &recordingClient{})
	proxy.smc.Set(string(testGohtURI), testSourceMap())
	proxy.srcs.Set(string(testGohtURI), NewDocument("package main\n\n@goht Test() {\n}\n"))

	got, err := proxy.Completion(context.Background(), &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: testGohtURI},
			Position:     protocol.Position{Line: 1, Character: 2},
		},
	})
	if err != nil {
		t.Fatalf("Completion() error = %v", err)
	}
	if len(server.completionCalls) != 1 {
		t.Fatalf("completion calls = %d, want 1", len(server.completionCalls))
	}
	if server.completionCalls[0].TextDocument.URI != testGohtGoURI {
		t.Fatalf("forwarded URI = %q, want %q", server.completionCalls[0].TextDocument.URI, testGohtGoURI)
	}
	if server.completionCalls[0].Position != (protocol.Position{Line: 10, Character: 20}) {
		t.Fatalf("forwarded position = %#v, want line 10 character 20", server.completionCalls[0].Position)
	}
	item := got.Items[0]
	if item.TextEdit.Range != rangeOf(1, 2, 1, 4) {
		t.Fatalf("TextEdit.Range = %#v, want mapped GoHT range", item.TextEdit.Range)
	}
	if len(item.AdditionalTextEdits) != 1 {
		t.Fatalf("AdditionalTextEdits = %d, want 1", len(item.AdditionalTextEdits))
	}
	if item.AdditionalTextEdits[0] != (protocol.TextEdit{Range: rangeOf(1, 2, 1, 4), NewText: "mapped edit"}) {
		t.Fatalf("AdditionalTextEdits[0] = %#v, want mapped edit", item.AdditionalTextEdits[0])
	}
}

func TestServerCompletionFallsBackForUnmappedGeneratedImportEdit(t *testing.T) {
	server := &recordingServer{
		completionResult: &protocol.CompletionList{
			Items: []protocol.CompletionItem{
				{
					Label:  "Contains",
					Detail: "func Contains(s, substr string) bool (from \"strings\")",
					AdditionalTextEdits: []protocol.TextEdit{
						{
							Range:   rangeOf(0, 0, 0, 0),
							NewText: "import \"strings\"\n",
						},
					},
				},
			},
		},
	}
	proxy := newTestServer(server, &recordingClient{})
	proxy.smc.Set(string(testGohtURI), testSourceMap())
	proxy.srcs.Set(string(testGohtURI), NewDocument("package main\n\n@goht Test() {\n}\n"))

	got, err := proxy.Completion(context.Background(), &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: testGohtURI},
			Position:     protocol.Position{Line: 1, Character: 2},
		},
	})
	if err != nil {
		t.Fatalf("Completion() error = %v", err)
	}

	edits := got.Items[0].AdditionalTextEdits
	if len(edits) != 1 {
		t.Fatalf("AdditionalTextEdits = %d, want 1", len(edits))
	}
	want := protocol.TextEdit{Range: rangeOf(1, 0, 1, 0), NewText: "import \"strings\"\n\n"}
	if edits[0] != want {
		t.Fatalf("AdditionalTextEdits[0] = %#v, want %#v", edits[0], want)
	}
}

func TestServerCompletionDropsUnmappedNonImportEdit(t *testing.T) {
	server := &recordingServer{
		completionResult: &protocol.CompletionList{
			Items: []protocol.CompletionItem{
				{
					Label:  "Thing",
					Detail: "func()",
					AdditionalTextEdits: []protocol.TextEdit{
						{
							Range:   rangeOf(0, 0, 0, 0),
							NewText: "not an import\n",
						},
					},
				},
			},
		},
	}
	proxy := newTestServer(server, &recordingClient{})
	proxy.smc.Set(string(testGohtURI), testSourceMap())
	proxy.srcs.Set(string(testGohtURI), NewDocument("package main\n\n@goht Test() {\n}\n"))

	got, err := proxy.Completion(context.Background(), &protocol.CompletionParams{
		TextDocumentPositionParams: positionParams(),
	})
	if err != nil {
		t.Fatalf("Completion() error = %v", err)
	}
	if edits := got.Items[0].AdditionalTextEdits; len(edits) != 0 {
		t.Fatalf("AdditionalTextEdits = %#v, want none", edits)
	}
}

func TestServerLocationHandlersPreserveNonGohtResults(t *testing.T) {
	otherURI := protocol.DocumentURI("file:///tmp/other.go")
	virtualLocation := protocol.Location{URI: testGohtGoURI, Range: rangeOf(10, 20, 10, 22)}
	otherLocation := protocol.Location{URI: otherURI, Range: rangeOf(3, 4, 3, 7)}

	tests := map[string]struct {
		setup func(*recordingServer)
		call  func(*Server) ([]protocol.Location, error)
	}{
		"definition": {
			setup: func(s *recordingServer) { s.definitionResult = []protocol.Location{virtualLocation, otherLocation} },
			call: func(proxy *Server) ([]protocol.Location, error) {
				return proxy.Definition(context.Background(), &protocol.DefinitionParams{
					TextDocumentPositionParams: positionParams(),
				})
			},
		},
		"implementation": {
			setup: func(s *recordingServer) { s.implementationResult = []protocol.Location{virtualLocation, otherLocation} },
			call: func(proxy *Server) ([]protocol.Location, error) {
				return proxy.Implementation(context.Background(), &protocol.ImplementationParams{
					TextDocumentPositionParams: positionParams(),
				})
			},
		},
		"references": {
			setup: func(s *recordingServer) { s.referencesResult = []protocol.Location{virtualLocation, otherLocation} },
			call: func(proxy *Server) ([]protocol.Location, error) {
				return proxy.References(context.Background(), &protocol.ReferenceParams{
					TextDocumentPositionParams: positionParams(),
				})
			},
		},
		"type definition": {
			setup: func(s *recordingServer) { s.typeDefinitionResult = []protocol.Location{virtualLocation, otherLocation} },
			call: func(proxy *Server) ([]protocol.Location, error) {
				return proxy.TypeDefinition(context.Background(), &protocol.TypeDefinitionParams{
					TextDocumentPositionParams: positionParams(),
				})
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			server := &recordingServer{}
			tt.setup(server)
			proxy := newTestServer(server, &recordingClient{})
			proxy.smc.Set(string(testGohtURI), testSourceMap())

			got, err := tt.call(proxy)
			if err != nil {
				t.Fatalf("%s() error = %v", name, err)
			}

			want := []protocol.Location{
				{URI: testGohtURI, Range: rangeOf(1, 2, 1, 4)},
				otherLocation,
			}
			assertLocations(t, got, want)
		})
	}
}

func TestGetPackageFromItemDetail(t *testing.T) {
	tests := map[string]struct {
		detail string
		want   string
	}{
		"from detail": {
			detail: "func Contains(s, substr string) bool (from \"strings\")",
			want:   "\"strings\"",
		},
		"already package literal": {
			detail: "\"strings\"",
			want:   "\"strings\"",
		},
		"plain detail": {
			detail: "func()",
			want:   "func()",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := getPackageFromItemDetail(tt.detail); got != tt.want {
				t.Fatalf("getPackageFromItemDetail(%q) = %q, want %q", tt.detail, got, tt.want)
			}
		})
	}
}

func TestAddImport(t *testing.T) {
	tests := map[string]struct {
		text string
		pkg  string
		want importInsert
	}{
		"no imports inserts after package": {
			text: "package main\n\n@goht Test() {\n}\n",
			pkg:  "\"strings\"",
			want: importInsert{line: 1, text: "import \"strings\"\n\n"},
		},
		"before goht declaration": {
			text: "package main\n\n@goht Test() {\n}\n",
			pkg:  "\"fmt\"",
			want: importInsert{line: 1, text: "import \"fmt\"\n\n"},
		},
		"after single-line import": {
			text: "package main\n\nimport \"fmt\"\n\n@goht Test() {\n}\n",
			pkg:  "\"strings\"",
			want: importInsert{line: 3, text: "import \"strings\"\n"},
		},
		"before grouped import close": {
			text: "package main\n\nimport (\n\t\"fmt\"\n)\n\n@goht Test() {\n}\n",
			pkg:  "\"strings\"",
			want: importInsert{line: 4, text: "\t\"strings\"\n"},
		},
		"aliased package": {
			text: "package main\n\nimport \"fmt\"\n\n@goht Test() {\n}\n",
			pkg:  "str \"strings\"",
			want: importInsert{line: 3, text: "import str \"strings\"\n"},
		},
		"dot import": {
			text: "package main\n\nimport \"fmt\"\n\n@goht Test() {\n}\n",
			pkg:  ". \"strings\"",
			want: importInsert{line: 3, text: "import . \"strings\"\n"},
		},
		"blank import": {
			text: "package main\n\nimport \"fmt\"\n\n@goht Test() {\n}\n",
			pkg:  "_ \"strings\"",
			want: importInsert{line: 3, text: "import _ \"strings\"\n"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := addImport(NewDocument(tt.text).lines, tt.pkg)
			if got != tt.want {
				t.Fatalf("addImport() = %#v, want %#v", got, tt.want)
			}
		})
	}
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

func positionParams() protocol.TextDocumentPositionParams {
	return protocol.TextDocumentPositionParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: testGohtURI},
		Position:     protocol.Position{Line: 1, Character: 2},
	}
}

func assertLocations(t *testing.T, got, want []protocol.Location) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("locations = %d, want %d: %#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("locations[%d] = %#v, want %#v", i, got[i], want[i])
		}
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
