package proxy

import (
	"context"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stackus/protocol"
)

type forwardingServer struct {
	protocol.Server
	codeActions        int
	codeLenses         int
	colorPresentations int
	documentColors     int
	didSaves           int
	willSaves          int
	onTypeFormats      int
	rangeFormats       int
	declarations       int
	resolvedLinks      int
}

func (s *forwardingServer) CodeAction(context.Context, *protocol.CodeActionParams) ([]protocol.CodeAction, error) {
	s.codeActions++
	return nil, nil
}

func (s *forwardingServer) CodeLens(context.Context, *protocol.CodeLensParams) ([]protocol.CodeLens, error) {
	s.codeLenses++
	return nil, nil
}

func (s *forwardingServer) ColorPresentation(context.Context, *protocol.ColorPresentationParams) ([]protocol.ColorPresentation, error) {
	s.colorPresentations++
	return nil, nil
}

func (s *forwardingServer) DocumentColor(context.Context, *protocol.DocumentColorParams) ([]protocol.ColorInformation, error) {
	s.documentColors++
	return nil, nil
}

func (s *forwardingServer) DidSave(context.Context, *protocol.DidSaveTextDocumentParams) error {
	s.didSaves++
	return nil
}

func (s *forwardingServer) WillSave(context.Context, *protocol.WillSaveTextDocumentParams) error {
	s.willSaves++
	return nil
}

func (s *forwardingServer) OnTypeFormatting(context.Context, *protocol.DocumentOnTypeFormattingParams) ([]protocol.TextEdit, error) {
	s.onTypeFormats++
	return nil, nil
}

func (s *forwardingServer) RangeFormatting(context.Context, *protocol.DocumentRangeFormattingParams) ([]protocol.TextEdit, error) {
	s.rangeFormats++
	return nil, nil
}

func (s *forwardingServer) Declaration(context.Context, *protocol.DeclarationParams) (*protocol.Or_textDocument_declaration, error) {
	s.declarations++
	return &protocol.Or_textDocument_declaration{}, nil
}

func (s *forwardingServer) ResolveDocumentLink(context.Context, *protocol.DocumentLink) (*protocol.DocumentLink, error) {
	s.resolvedLinks++
	return nil, nil
}

type forwardingClient struct {
	protocol.Client
	showMessages  int
	registrations int
}

func (c *forwardingClient) ShowMessage(context.Context, *protocol.ShowMessageParams) error {
	c.showMessages++
	return nil
}
func (c *forwardingClient) RegisterCapability(context.Context, *protocol.RegistrationParams) error {
	c.registrations++
	return nil
}

func TestServerForwardsNonGohtAndSaveRequests(t *testing.T) {
	backend := &forwardingServer{}
	server := NewServer(backend, &recordingClient{}, NewSourceMapCache(), NewDiagnosticsCache(), NewDocumentContents(), zerolog.Nop())
	ctx := context.Background()
	uri := protocol.DocumentURI("file:///tmp/main.go")
	document := protocol.TextDocumentIdentifier{URI: uri}

	if _, err := server.CodeAction(ctx, &protocol.CodeActionParams{TextDocument: document}); err != nil {
		t.Fatal(err)
	}
	if _, err := server.CodeLens(ctx, &protocol.CodeLensParams{TextDocument: document}); err != nil {
		t.Fatal(err)
	}
	if _, err := server.ColorPresentation(ctx, &protocol.ColorPresentationParams{TextDocument: document}); err != nil {
		t.Fatal(err)
	}
	if _, err := server.DocumentColor(ctx, &protocol.DocumentColorParams{TextDocument: document}); err != nil {
		t.Fatal(err)
	}
	if err := server.DidSave(ctx, &protocol.DidSaveTextDocumentParams{TextDocument: document}); err != nil {
		t.Fatal(err)
	}
	if err := server.WillSave(ctx, &protocol.WillSaveTextDocumentParams{TextDocument: document}); err != nil {
		t.Fatal(err)
	}

	if backend.codeActions != 1 || backend.codeLenses != 1 || backend.colorPresentations != 1 || backend.documentColors != 1 || backend.didSaves != 1 || backend.willSaves != 0 {
		t.Fatalf("unexpected forwarding counts: %#v", backend)
	}
}

func TestServerMapsGohtRequestsBeforeForwarding(t *testing.T) {
	backend := &forwardingServer{}
	server := NewServer(backend, &recordingClient{}, NewSourceMapCache(), NewDiagnosticsCache(), NewDocumentContents(), zerolog.Nop())
	server.smc.Set(string(testGohtURI), testSourceMap())
	ctx := context.Background()
	document := protocol.TextDocumentIdentifier{URI: testGohtURI}
	mappedRange := rangeOf(1, 2, 1, 4)

	if _, err := server.CodeAction(ctx, &protocol.CodeActionParams{TextDocument: document, Range: mappedRange}); err != nil {
		t.Fatal(err)
	}
	if _, err := server.CodeLens(ctx, &protocol.CodeLensParams{TextDocument: document}); err != nil {
		t.Fatal(err)
	}
	if _, err := server.ColorPresentation(ctx, &protocol.ColorPresentationParams{TextDocument: document, Range: mappedRange}); err != nil {
		t.Fatal(err)
	}
	if _, err := server.DocumentColor(ctx, &protocol.DocumentColorParams{TextDocument: document}); err != nil {
		t.Fatal(err)
	}
	if err := server.DidSave(ctx, &protocol.DidSaveTextDocumentParams{TextDocument: document}); err != nil {
		t.Fatal(err)
	}
	if err := server.WillSave(ctx, &protocol.WillSaveTextDocumentParams{TextDocument: document}); err != nil {
		t.Fatal(err)
	}

	if backend.codeActions != 1 || backend.codeLenses != 1 || backend.colorPresentations != 1 || backend.documentColors != 1 || backend.didSaves != 1 || backend.willSaves != 1 {
		t.Fatalf("unexpected forwarding counts: %#v", backend)
	}
}

func TestServerForwardsPositionAndFormattingRequests(t *testing.T) {
	backend := &forwardingServer{}
	server := NewServer(backend, &recordingClient{}, NewSourceMapCache(), NewDiagnosticsCache(), NewDocumentContents(), zerolog.Nop())
	server.smc.Set(string(testGohtURI), testSourceMap())
	ctx := context.Background()
	document := protocol.TextDocumentIdentifier{URI: testGohtURI}
	position := protocol.Position{Line: 1, Character: 2}

	if _, err := server.Declaration(ctx, &protocol.DeclarationParams{TextDocumentPositionParams: protocol.TextDocumentPositionParams{TextDocument: document, Position: position}}); err != nil {
		t.Fatal(err)
	}
	if _, err := server.OnTypeFormatting(ctx, &protocol.DocumentOnTypeFormattingParams{TextDocument: document, Position: position}); err != nil {
		t.Fatal(err)
	}
	if _, err := server.RangeFormatting(ctx, &protocol.DocumentRangeFormattingParams{TextDocument: document}); err != nil {
		t.Fatal(err)
	}

	if backend.declarations != 1 || backend.onTypeFormats != 1 || backend.rangeFormats != 1 {
		t.Fatalf("unexpected forwarding counts: %#v", backend)
	}
}

func TestServerRejectsUnsupportedDocumentRequests(t *testing.T) {
	server := NewServer(&forwardingServer{}, &recordingClient{}, NewSourceMapCache(), NewDiagnosticsCache(), NewDocumentContents(), zerolog.Nop())
	ctx := context.Background()
	position := protocol.TextDocumentPositionParams{TextDocument: protocol.TextDocumentIdentifier{URI: protocol.DocumentURI("file:///tmp/main.go")}}

	if got, err := server.Hover(ctx, &protocol.HoverParams{TextDocumentPositionParams: position}); err != nil || got != nil {
		t.Fatalf("Hover() = (%#v, %v)", got, err)
	}
	if got, err := server.Implementation(ctx, &protocol.ImplementationParams{TextDocumentPositionParams: position}); err != nil || got != nil {
		t.Fatalf("Implementation() = (%#v, %v)", got, err)
	}
	if got, err := server.PrepareRename(ctx, &protocol.PrepareRenameParams{TextDocumentPositionParams: position}); err != nil || got != nil {
		t.Fatalf("PrepareRename() = (%#v, %v)", got, err)
	}
	if got, err := server.References(ctx, &protocol.ReferenceParams{TextDocumentPositionParams: position}); err != nil || len(got) != 0 {
		t.Fatalf("References() = (%#v, %v)", got, err)
	}
	if got, err := server.SignatureHelp(ctx, &protocol.SignatureHelpParams{TextDocumentPositionParams: position}); err != nil || got != nil {
		t.Fatalf("SignatureHelp() = (%#v, %v)", got, err)
	}
	if got, err := server.TypeDefinition(ctx, &protocol.TypeDefinitionParams{TextDocumentPositionParams: position}); err != nil || len(got) != 0 {
		t.Fatalf("TypeDefinition() = (%#v, %v)", got, err)
	}
	if got, err := server.Moniker(ctx, &protocol.MonikerParams{TextDocumentPositionParams: position}); err != nil || len(got) != 0 {
		t.Fatalf("Moniker() = (%#v, %v)", got, err)
	}

	target := "file:///tmp/main.go"

	if got, err := server.ResolveDocumentLink(ctx, &protocol.DocumentLink{Target: &target}); err != nil || got != nil {
		t.Fatalf("ResolveDocumentLink() = (%#v, %v)", got, err)
	}
}

func TestClientFiltersGeneratedFileWarnings(t *testing.T) {
	tests := map[string]struct {
		message  string
		wantSent int
	}{
		"generated file warning is ignored": {message: doNotEditMessage, wantSent: 0},
		"editing warning is ignored":        {message: warningEditingMessage, wantSent: 0},
		"other message is forwarded":        {message: "hello", wantSent: 1},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			backend := &forwardingClient{}
			client := NewClient(backend, NewSourceMapCache(), NewDiagnosticsCache(), zerolog.Nop())

			if err := client.ShowMessage(context.Background(), &protocol.ShowMessageParams{Message: tt.message}); err != nil {
				t.Fatal(err)
			}

			if backend.showMessages != tt.wantSent {
				t.Fatalf("ShowMessage forwards = %d, want %d", backend.showMessages, tt.wantSent)
			}
		})
	}
}

func TestClientRegisterCapabilityForwards(t *testing.T) {
	backend := &forwardingClient{}
	client := NewClient(backend, NewSourceMapCache(), NewDiagnosticsCache(), zerolog.Nop())

	if err := client.RegisterCapability(context.Background(), &protocol.RegistrationParams{}); err != nil {
		t.Fatal(err)
	}

	if backend.registrations != 1 {
		t.Fatalf("RegisterCapability forwards = %d, want 1", backend.registrations)
	}
}
