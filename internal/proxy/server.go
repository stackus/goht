package proxy

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/rs/zerolog"

	"github.com/stackus/goht"
	"github.com/stackus/goht/compiler"
	"github.com/stackus/goht/internal/protocol"
)

type Server struct {
	protocol.Server
	c      protocol.Client
	smc    *SourceMapCache
	dc     *DiagnosticsCache
	srcs   *DocumentContents
	goSrcs map[string]string
	logger zerolog.Logger
}

var _ protocol.Server = (*Server)(nil)

func NewServer(s protocol.Server, c protocol.Client, smc *SourceMapCache, dc *DiagnosticsCache, srcs *DocumentContents, logger zerolog.Logger) *Server {
	return &Server{
		Server: s,
		c:      c,
		smc:    smc,
		dc:     dc,
		srcs:   srcs,
		goSrcs: make(map[string]string),
		logger: logger,
	}
}

// Initialize is called when the client starts up.
//
// It returns the capabilities of the server.
func (s *Server) Initialize(ctx context.Context, params *protocol.ParamInitialize) (*protocol.InitializeResult, error) {
	logger := s.logger.With().
		Str("method", "Initialize").
		Str("clientName", params.ClientInfo.Name).
		Str("clientVersion", params.ClientInfo.Version).
		Logger()

	resp, err := s.Server.Initialize(ctx, params)
	if err != nil {
		logger.Error().Err(err).Msg("unable to initialize server")
	}
	if resp.Capabilities.CompletionProvider == nil {
		resp.Capabilities.CompletionProvider = &protocol.CompletionOptions{}
	}
	if resp.Capabilities.ExecuteCommandProvider == nil {
		resp.Capabilities.ExecuteCommandProvider = &protocol.ExecuteCommandOptions{}
	}
	resp.Capabilities.ExecuteCommandProvider.Commands = []string{}
	resp.Capabilities.DocumentFormattingProvider = &protocol.Or_ServerCapabilities_documentFormattingProvider{Value: false}
	resp.Capabilities.DocumentRangeFormattingProvider = &protocol.Or_ServerCapabilities_documentRangeFormattingProvider{Value: false}
	resp.Capabilities.SemanticTokensProvider = nil
	resp.Capabilities.TextDocumentSync = protocol.TextDocumentSyncOptions{
		OpenClose:         true,
		Change:            protocol.Full,
		WillSave:          false,
		WillSaveWaitUntil: false,
		Save: &protocol.SaveOptions{
			IncludeText: true,
		},
	}

	resp.ServerInfo.Name = "goht-lsp"
	resp.ServerInfo.Version = goht.Version()

	return resp, err
}

// CodeAction is called when the client requests code actions.
func (s *Server) CodeAction(ctx context.Context, params *protocol.CodeActionParams) ([]protocol.CodeAction, error) {
	logger := s.logger.With().
		Str("method", "CodeAction").
		Str("uri", string(params.TextDocument.URI)).
		Logger()

	isGohtFile, goURI := toGohtGoURI(params.TextDocument.URI)
	if !isGohtFile {
		logger.Warn().Msg("not a goht file")
		return s.Server.CodeAction(ctx, params)
	}
	gohtURI := params.TextDocument.URI
	params.TextDocument.URI = goURI

	resp, err := s.Server.CodeAction(ctx, params)
	if err != nil {
		logger.Error().Err(err).Msg("unable to perform code actions")
		return resp, err
	}

	for i, codeAction := range resp {
		for j, diagnostic := range codeAction.Diagnostics {
			diagnostic.Range = s.goRangeToGohtRange(gohtURI, diagnostic.Range)
			codeAction.Diagnostics[j] = diagnostic
		}

		if codeAction.Edit == nil {
			continue
		}

		for j, changes := range codeAction.Edit.DocumentChanges {
			var te protocol.TextEdit
			var ok bool
			for k, textEdit := range changes.TextDocumentEdit.Edits {
				if te, ok = textEdit.Value.(protocol.TextEdit); !ok {
					continue
				}
				te.Range = s.goRangeToGohtRange(gohtURI, te.Range)
				textEdit.Value = te
				changes.TextDocumentEdit.Edits[k] = textEdit
			}
			changes.TextDocumentEdit.TextDocument.URI = gohtURI
			codeAction.Edit.DocumentChanges[j] = changes
		}
		resp[i] = codeAction
	}

	return resp, nil
}

// CodeLens is called when the client requests code lenses.
func (s *Server) CodeLens(ctx context.Context, params *protocol.CodeLensParams) ([]protocol.CodeLens, error) {
	logger := s.logger.With().
		Str("method", "CodeLens").
		Str("uri", string(params.TextDocument.URI)).
		Logger()

	isGohtFile, goURI := toGohtGoURI(params.TextDocument.URI)
	if !isGohtFile {
		return s.Server.CodeLens(ctx, params)
	}
	gohtURI := params.TextDocument.URI
	params.TextDocument.URI = goURI

	resp, err := s.Server.CodeLens(ctx, params)
	if err != nil {
		logger.Error().Err(err).Msg("unable to perform code lens action")
		return resp, err
	}
	if resp == nil {
		return resp, nil
	}

	for i, codeLens := range resp {
		codeLens.Range = s.goRangeToGohtRange(gohtURI, codeLens.Range)
		resp[i] = codeLens
	}

	return resp, nil
}

// ColorPresentation is called when the client requests color presentations.
func (s *Server) ColorPresentation(ctx context.Context, params *protocol.ColorPresentationParams) ([]protocol.ColorPresentation, error) {
	logger := s.logger.With().
		Str("method", "ColorPresentation").
		Str("uri", string(params.TextDocument.URI)).
		Logger()

	isGohtFile, goURI := toGohtGoURI(params.TextDocument.URI)
	if !isGohtFile {
		return s.Server.ColorPresentation(ctx, params)
	}
	gohtURI := params.TextDocument.URI
	params.TextDocument.URI = goURI

	resp, err := s.Server.ColorPresentation(ctx, params)
	if err != nil {
		logger.Error().Err(err).Msg("unable to perform color presentation")
		return resp, err
	}
	if resp == nil {
		return resp, nil
	}

	for i, colorPresentation := range resp {
		colorPresentation.TextEdit.Range = s.goRangeToGohtRange(gohtURI, colorPresentation.TextEdit.Range)
		resp[i] = colorPresentation
	}

	return resp, nil
}

// Completion is called when the client requests completion information.
func (s *Server) Completion(ctx context.Context, params *protocol.CompletionParams) (*protocol.CompletionList, error) {
	logger := s.logger.With().
		Str("method", "Completion").
		Str("uri", string(params.TextDocument.URI)).
		Logger()

	gohtURI := params.TextDocument.URI
	var err error
	params.TextDocument.URI, params.Position, err = s.updatePosition(gohtURI, params.Position)
	if err != nil {
		logger.Error().Err(err).Msg("unable to update position")
		return nil, nil
	}
	resp, err := s.Server.Completion(ctx, params)
	if err != nil {
		logger.Error().Err(err).Msg("unable to perform completions")
		return resp, err
	}
	if resp == nil {
		return resp, nil
	}
	for i, completionItem := range resp.Items {
		if completionItem.TextEdit != nil {
			completionItem.TextEdit.Range = s.goRangeToGohtRange(gohtURI, completionItem.TextEdit.Range)
		}
		if len(completionItem.AdditionalTextEdits) > 0 {
			doc, ok := s.srcs.Get(string(gohtURI))
			if !ok {
				continue
			}
			pkg := getPackageFromItemDetail(completionItem.Detail)
			insert := addImport(doc.lines, pkg)
			completionItem.AdditionalTextEdits = []protocol.TextEdit{
				{
					Range: protocol.Range{
						Start: protocol.Position{
							Line:      uint32(insert.line),
							Character: 0,
						},
						End: protocol.Position{
							Line:      uint32(insert.line),
							Character: 0,
						},
					},
					NewText: insert.text,
				},
			}
		}
		resp.Items[i] = completionItem
	}
	return resp, nil
}

// Declaration is called when the client requests declaration information.
func (s *Server) Declaration(ctx context.Context, params *protocol.DeclarationParams) (*protocol.Or_textDocument_declaration, error) {
	logger := s.logger.With().
		Str("method", "Declaration").
		Str("uri", string(params.TextDocument.URI)).
		Logger()

	gohtURI := params.TextDocument.URI
	var err error
	params.TextDocument.URI, params.Position, err = s.updatePosition(gohtURI, params.Position)
	if err != nil {
		logger.Error().Err(err).Msg("unable to update position")
		return &protocol.Or_textDocument_declaration{
			Value: []protocol.DeclarationLink{},
		}, nil
	}

	resp, err := s.Server.Declaration(ctx, params)
	if err != nil {
		logger.Error().Err(err).Msg("unable to perform declaration lookup")
		return resp, err
	}
	if decls, ok := resp.Value.([]protocol.DeclarationLink); ok {
		for i, decl := range decls {
			if isGohtGoFile, goURI := toGohtURI(decl.TargetURI); isGohtGoFile {
				decl.TargetURI = goURI
				decl.TargetRange = s.goRangeToGohtRange(decl.TargetURI, decl.TargetRange)
				decls[i] = decl
			}
		}
		resp.Value = decls
	}

	return resp, nil
}

// Definition is called when the client requests definition information.
func (s *Server) Definition(ctx context.Context, params *protocol.DefinitionParams) ([]protocol.Location, error) {
	logger := s.logger.With().
		Str("method", "Definition").
		Str("uri", string(params.TextDocument.URI)).
		Logger()

	gohtURI := params.TextDocument.URI
	var err error
	params.TextDocument.URI, params.Position, err = s.updatePosition(gohtURI, params.Position)
	if err != nil {
		logger.Error().Err(err).Msg("unable to update position")
		return []protocol.Location{}, nil
	}
	resp, err := s.Server.Definition(ctx, params)
	if err != nil {
		logger.Error().Err(err).Msg("unable to perform definition lookup")
		return resp, err
	}
	for i, location := range resp {
		if isGohtGoFile, goURI := toGohtURI(location.URI); isGohtGoFile {
			location.URI = goURI
			location.Range = s.goRangeToGohtRange(location.URI, location.Range)
			resp[i] = location
		}
	}
	return resp, nil
}

func (s *Server) DidChange(ctx context.Context, params *protocol.DidChangeTextDocumentParams) error {
	logger := s.logger.With().
		Str("method", "DidChange").
		Str("uri", string(params.TextDocument.URI)).
		Logger()

	isGohtFile, goURI := toGohtGoURI(params.TextDocument.URI)
	if !isGohtFile {
		logger.Warn().Msg("not a goht file")
		return nil
	}

	doc, err := s.srcs.Apply(string(params.TextDocument.URI), params.ContentChanges)
	if err != nil {
		logger.Error().Err(err).Msg("unable to apply changes")
		return err
	}

	template, err := s.parseTemplate(ctx, params.TextDocument.URI, doc.String())
	if err != nil {
		logger.Error().Err(err).Msg("unable to parse template")
	}
	buf := bytes.Buffer{}
	sm, err := template.Compose(&buf)
	if err != nil {
		logger.Error().Err(err).Msg("unable to compose template")
		return err
	}
	s.smc.Set(string(params.TextDocument.URI), sm)
	s.goSrcs[string(params.TextDocument.URI)] = buf.String()
	params.TextDocument.URI = goURI
	params.TextDocument.TextDocumentIdentifier.URI = goURI
	params.ContentChanges = []protocol.TextDocumentContentChangeEvent{
		{
			Text: buf.String(),
		},
	}
	return s.Server.DidChange(ctx, params)
}

func (s *Server) DidClose(ctx context.Context, params *protocol.DidCloseTextDocumentParams) error {
	logger := s.logger.With().
		Str("method", "DidClose").
		Str("uri", string(params.TextDocument.URI)).
		Logger()

	isGohtFile, goURI := toGohtGoURI(params.TextDocument.URI)
	if !isGohtFile {
		logger.Warn().Msg("not a goht file")
		return s.Server.DidClose(ctx, params)
	}
	s.srcs.Delete(string(params.TextDocument.URI))
	delete(s.goSrcs, string(params.TextDocument.URI))
	params.TextDocument.URI = goURI
	err := s.Server.DidClose(ctx, params)
	if err != nil {
		logger.Error().Err(err).Msg("failed to close document")
	}
	return err
}

func (s *Server) DidOpen(ctx context.Context, params *protocol.DidOpenTextDocumentParams) error {
	logger := s.logger.With().
		Str("method", "DidOpen").
		Str("uri", string(params.TextDocument.URI)).
		Logger()

	isGohtFile, goURI := toGohtGoURI(params.TextDocument.URI)
	if !isGohtFile {
		logger.Warn().Msg("not a goht file")
		return s.Server.DidOpen(ctx, params)
	}
	s.srcs.Set(string(params.TextDocument.URI), NewDocument(params.TextDocument.Text))
	template, err := s.parseTemplate(ctx, params.TextDocument.URI, params.TextDocument.Text)
	if err != nil {
		logger.Error().Err(err).Msg("unable to parse template")
		return nil
	}

	buf := bytes.Buffer{}
	sm, err := template.Compose(&buf)
	if err != nil {
		logger.Error().Err(err).Msg("unable to compose template")
		return err
	}
	s.smc.Set(string(params.TextDocument.URI), sm)
	s.goSrcs[string(params.TextDocument.URI)] = buf.String()

	params.TextDocument.LanguageID = "go"
	params.TextDocument.URI = goURI
	params.TextDocument.Text = buf.String()
	err = s.Server.DidOpen(ctx, params)
	if err != nil {
		logger.Error().Err(err).Msg("unable to open document")
	}
	return err
}

func (s *Server) DidSave(ctx context.Context, params *protocol.DidSaveTextDocumentParams) error {
	logger := s.logger.With().
		Str("method", "DidSave").
		Str("uri", string(params.TextDocument.URI)).
		Logger()

	if isGohtFile, goURI := toGohtGoURI(params.TextDocument.URI); isGohtFile {
		params.TextDocument.URI = goURI
	}
	err := s.Server.DidSave(ctx, params)
	if err != nil {
		logger.Error().Err(err).Msg("unable to save document")
	}
	return err
}

func (s *Server) DocumentColor(ctx context.Context, params *protocol.DocumentColorParams) ([]protocol.ColorInformation, error) {
	logger := s.logger.With().
		Str("method", "DocumentColor").
		Str("uri", string(params.TextDocument.URI)).
		Logger()

	isGohtFile, goURI := toGohtGoURI(params.TextDocument.URI)
	if !isGohtFile {
		logger.Warn().Msg("not a goht file")
		return s.Server.DocumentColor(ctx, params)
	}
	gohtURI := params.TextDocument.URI
	params.TextDocument.URI = goURI
	resp, err := s.Server.DocumentColor(ctx, params)
	if err != nil {
		logger.Error().Err(err).Msg("unable to perform document color lookup")
		return resp, err
	}
	for i, colorInfo := range resp {
		colorInfo.Range = s.goRangeToGohtRange(gohtURI, colorInfo.Range)
		resp[i] = colorInfo
	}
	return resp, nil
}

func (s *Server) DocumentHighlight(_ context.Context, _ *protocol.DocumentHighlightParams) ([]protocol.DocumentHighlight, error) {
	return []protocol.DocumentHighlight{}, nil
}

func (s *Server) DocumentLink(_ context.Context, _ *protocol.DocumentLinkParams) ([]protocol.DocumentLink, error) {
	return []protocol.DocumentLink{}, nil
}

func (s *Server) ResolveDocumentLink(ctx context.Context, params *protocol.DocumentLink) (*protocol.DocumentLink, error) {
	logger := s.logger.With().
		Str("method", "ResolveDocumentLink").
		Str("uri", *params.Target).
		Logger()

	gohtURI := *params.Target
	isGohtFile, goURI := toGohtGoURI(protocol.DocumentURI(gohtURI))
	if !isGohtFile {
		logger.Warn().Msg("not a goht file")
		return s.Server.ResolveDocumentLink(ctx, params)
	}
	params.Target = (*protocol.URI)(&goURI)
	params.Range = s.gohtRangeToGoRange(protocol.DocumentURI(gohtURI), params.Range)
	resp, err := s.Server.ResolveDocumentLink(ctx, params)
	if err != nil {
		logger.Error().Err(err).Msg("unable to resolve document link")
		return resp, err
	}
	if resp == nil {
		return resp, nil
	}
	resp.Target = &gohtURI
	resp.Range = s.goRangeToGohtRange(protocol.DocumentURI(gohtURI), resp.Range)
	return resp, nil
}

// DocumentSymbol is called when the client requests document symbols.
func (s *Server) DocumentSymbol(_ context.Context, _ *protocol.DocumentSymbolParams) ([]any, error) {
	return nil, nil
}

func (s *Server) FoldingRanges(_ context.Context, _ *protocol.FoldingRangeParams) ([]protocol.FoldingRange, error) {
	return []protocol.FoldingRange{}, nil
}

func (s *Server) Formatting(_ context.Context, _ *protocol.DocumentFormattingParams) ([]protocol.TextEdit, error) {
	return []protocol.TextEdit{}, nil
}

func (s *Server) Hover(ctx context.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
	logger := s.logger.With().
		Str("method", "Hover").
		Str("uri", string(params.TextDocument.URI)).
		Logger()

	gohtURI := params.TextDocument.URI
	var err error
	params.TextDocument.URI, params.Position, err = s.updatePosition(gohtURI, params.Position)
	if err != nil {
		logger.Error().Err(err).Msg("unable to update position")
		return nil, nil
	}
	resp, err := s.Server.Hover(ctx, params)
	if err != nil {
		logger.Error().Err(err).Msg("unable to perform hover")
		return resp, err
	}
	if resp == nil {
		logger.Warn().Msg("no hover response")
		return resp, nil
	}
	resp.Range = s.goRangeToGohtRange(gohtURI, resp.Range)
	return resp, nil
}

func (s *Server) Implementation(ctx context.Context, params *protocol.ImplementationParams) ([]protocol.Location, error) {
	logger := s.logger.With().
		Str("method", "Implementation").
		Str("uri", string(params.TextDocument.URI)).
		Logger()

	gohtURI := params.TextDocument.URI
	var err error
	params.TextDocument.URI, params.Position, err = s.updatePosition(gohtURI, params.Position)
	if err != nil {
		logger.Error().Err(err).Msg("unable to update position")
		return []protocol.Location{}, nil
	}
	resp, err := s.Server.Implementation(ctx, params)
	if err != nil || resp == nil {
		if err != nil {
			logger.Error().Err(err).Msg("unable to perform implementation lookup")
		}
		return resp, err
	}
	for i, location := range resp {
		location.URI = gohtURI
		location.Range = s.goRangeToGohtRange(gohtURI, location.Range)
		resp[i] = location
	}
	return resp, nil
}

func (s *Server) OnTypeFormatting(ctx context.Context, params *protocol.DocumentOnTypeFormattingParams) ([]protocol.TextEdit, error) {
	logger := s.logger.With().
		Str("method", "OnTypeFormatting").
		Str("uri", string(params.TextDocument.URI)).
		Logger()

	gohtURI := params.TextDocument.URI
	var err error
	params.TextDocument.URI, params.Position, err = s.updatePosition(gohtURI, params.Position)
	if err != nil {
		logger.Error().Err(err).Msg("unable to update position")
		return nil, nil
	}
	resp, err := s.Server.OnTypeFormatting(ctx, params)
	if err != nil || resp == nil {
		if err != nil {
			logger.Error().Err(err).Msg("unable to perform on type formatting")
		}
		return resp, err
	}
	for i, textEdit := range resp {
		textEdit.Range = s.goRangeToGohtRange(gohtURI, textEdit.Range)
		resp[i] = textEdit
	}
	return resp, nil
}

func (s *Server) PrepareRename(ctx context.Context, params *protocol.PrepareRenameParams) (*protocol.PrepareRenameResult, error) {
	logger := s.logger.With().
		Str("method", "PrepareRename").
		Str("uri", string(params.TextDocument.URI)).
		Logger()

	gohtURI := params.TextDocument.URI
	var err error
	params.TextDocument.URI, params.Position, err = s.updatePosition(gohtURI, params.Position)
	if err != nil {
		logger.Error().Err(err).Msg("unable to update position")
		return nil, nil
	}
	resp, err := s.Server.PrepareRename(ctx, params)
	if err != nil || resp == nil {
		if err != nil {
			logger.Error().Err(err).Msg("unable to perform prepare rename")
		}
		return resp, err
	}
	resp.Range = s.goRangeToGohtRange(gohtURI, resp.Range)
	return resp, nil
}

func (s *Server) RangeFormatting(ctx context.Context, params *protocol.DocumentRangeFormattingParams) ([]protocol.TextEdit, error) {
	logger := s.logger.With().
		Str("method", "RangeFormatting").
		Str("uri", string(params.TextDocument.URI)).
		Logger()

	gohtURI := params.TextDocument.URI
	var isGohtURI bool
	isGohtURI, params.TextDocument.URI = toGohtGoURI(params.TextDocument.URI)
	if !isGohtURI {
		logger.Warn().Msg("not a goht file")
		return []protocol.TextEdit{}, nil
	}
	resp, err := s.Server.RangeFormatting(ctx, params)
	if err != nil || resp == nil {
		if err != nil {
			logger.Error().Err(err).Msg("unable to perform range formatting")
		}
		return resp, err
	}
	for i, textEdit := range resp {
		textEdit.Range = s.goRangeToGohtRange(gohtURI, textEdit.Range)
		resp[i] = textEdit
	}
	return resp, nil
}

func (s *Server) References(ctx context.Context, params *protocol.ReferenceParams) ([]protocol.Location, error) {
	logger := s.logger.With().
		Str("method", "References").
		Str("uri", string(params.TextDocument.URI)).
		Logger()

	gohtURI := params.TextDocument.URI
	var isGohtURI bool
	isGohtURI, params.TextDocument.URI = toGohtGoURI(params.TextDocument.URI)
	if !isGohtURI {
		logger.Warn().Msg("not a goht file")
		return []protocol.Location{}, fmt.Errorf("not a goht file")
	}
	resp, err := s.Server.References(ctx, params)
	if err != nil || resp == nil {
		if err != nil {
			logger.Error().Err(err).Msg("unable to perform references lookup")
		}
		return resp, err
	}
	for i, location := range resp {
		location.URI = gohtURI
		location.Range = s.goRangeToGohtRange(gohtURI, location.Range)
		resp[i] = location
	}
	return resp, nil
}

func (s *Server) SignatureHelp(ctx context.Context, params *protocol.SignatureHelpParams) (*protocol.SignatureHelp, error) {
	logger := s.logger.With().
		Str("method", "SignatureHelp").
		Str("uri", string(params.TextDocument.URI)).
		Logger()

	var err error
	params.TextDocument.URI, params.Position, err = s.updatePosition(params.TextDocument.URI, params.Position)
	if err != nil {
		logger.Error().Err(err).Msg("unable to update position")
		return nil, nil
	}
	resp, err := s.Server.SignatureHelp(ctx, params)
	if err != nil {
		logger.Error().Err(err).Msg("unable to perform signature help")
	}
	return resp, err
}

func (s *Server) TypeDefinition(ctx context.Context, params *protocol.TypeDefinitionParams) ([]protocol.Location, error) {
	logger := s.logger.With().
		Str("method", "TypeDefinition").
		Str("uri", string(params.TextDocument.URI)).
		Logger()

	var err error
	params.TextDocument.URI, params.Position, err = s.updatePosition(params.TextDocument.URI, params.Position)
	if err != nil {
		logger.Error().Err(err).Msg("unable to update position")
		return []protocol.Location{}, nil
	}
	resp, err := s.Server.TypeDefinition(ctx, params)
	if err != nil {
		logger.Error().Err(err).Msg("unable to perform type definition lookup")
	}
	return resp, err
}

func (s *Server) WillSave(ctx context.Context, params *protocol.WillSaveTextDocumentParams) error {
	logger := s.logger.With().
		Str("method", "WillSave").
		Str("uri", string(params.TextDocument.URI)).
		Logger()

	if isGohtFile, goURI := toGohtGoURI(params.TextDocument.URI); isGohtFile {
		params.TextDocument.URI = goURI
		return s.Server.WillSave(ctx, params)
	}
	logger.Warn().Msg("not a goht file")
	return nil
}

func (s *Server) SemanticTokensFull(ctx context.Context, params *protocol.SemanticTokensParams) (*protocol.SemanticTokens, error) {
	logger := s.logger.With().
		Str("method", "SemanticTokensFull").
		Str("uri", string(params.TextDocument.URI)).
		Logger()

	isGohtFile, goURI := toGohtGoURI(params.TextDocument.URI)
	if !isGohtFile {
		logger.Warn().Msg("not a goht file")
		return nil, nil
	}
	params.TextDocument.URI = goURI
	resp, err := s.Server.SemanticTokensFull(ctx, params)
	if err != nil {
		logger.Error().Err(err).Msg("unable to perform semantic tokens full")
	}
	return resp, err
}

func (s *Server) SemanticTokensFullDelta(ctx context.Context, params *protocol.SemanticTokensDeltaParams) (any, error) {
	logger := s.logger.With().
		Str("method", "SemanticTokensFullDelta").
		Str("uri", string(params.TextDocument.URI)).
		Logger()

	isGohtFile, goURI := toGohtGoURI(params.TextDocument.URI)
	if !isGohtFile {
		logger.Warn().Msg("not a goht file")
		return nil, nil
	}
	params.TextDocument.URI = goURI
	resp, err := s.Server.SemanticTokensFullDelta(ctx, params)
	if err != nil {
		logger.Error().Err(err).Msg("unable to perform semantic tokens full delta")
	}
	return resp, err
}

func (s *Server) SemanticTokensRange(ctx context.Context, params *protocol.SemanticTokensRangeParams) (*protocol.SemanticTokens, error) {
	logger := s.logger.With().
		Str("method", "SemanticTokensRange").
		Str("uri", string(params.TextDocument.URI)).
		Logger()

	isGohtFile, goURI := toGohtGoURI(params.TextDocument.URI)
	if !isGohtFile {
		logger.Warn().Msg("not a goht file")
		return nil, nil
	}
	params.TextDocument.URI = goURI
	resp, err := s.Server.SemanticTokensRange(ctx, params)
	if err != nil {
		logger.Error().Err(err).Msg("unable to perform semantic tokens range")
	}
	return resp, err
}

func (s *Server) Moniker(ctx context.Context, params *protocol.MonikerParams) ([]protocol.Moniker, error) {
	logger := s.logger.With().
		Str("method", "Moniker").
		Str("uri", string(params.TextDocument.URI)).
		Logger()

	gohtURI := params.TextDocument.URI
	var err error
	params.TextDocument.URI, params.Position, err = s.updatePosition(gohtURI, params.Position)
	if err != nil {
		logger.Error().Err(err).Msg("unable to update position")
		return []protocol.Moniker{}, nil
	}
	resp, err := s.Server.Moniker(ctx, params)
	if err != nil {
		logger.Error().Err(err).Msg("unable to perform moniker lookup")
	}
	return resp, err
}

// InlayHint is called when the client requests inlay hints.
func (s *Server) InlayHint(_ context.Context, _ *protocol.InlayHintParams) ([]protocol.InlayHint, error) {
	return []protocol.InlayHint{}, nil
}

func (s *Server) goRangeToGohtRange(uri protocol.DocumentURI, goRange protocol.Range) protocol.Range {
	gohtRange := goRange
	sm, ok := s.smc.Get(string(uri))
	if !ok {
		return gohtRange
	}

	if start, ok := sm.SourcePositionFromTarget(int(goRange.Start.Line), int(goRange.Start.Character)); ok {
		s.logger.Info().Msgf("goRangeToGohtRange: %s: START [%d,%d] -> [%d,%d]", uri, goRange.Start.Line, goRange.Start.Character, start.Line, start.Col)
		gohtRange.Start.Line = uint32(start.Line)
		gohtRange.Start.Character = uint32(start.Col)
	}

	if end, ok := sm.SourcePositionFromTarget(int(goRange.End.Line), int(goRange.End.Character)); ok {
		s.logger.Info().Msgf("goRangeToGohtRange: %s: END [%d,%d] -> [%d,%d]", uri, goRange.End.Line, goRange.End.Character, end.Line, end.Col)
		gohtRange.End.Line = uint32(end.Line)
		gohtRange.End.Character = uint32(end.Col)
	}

	return gohtRange
}

func (s *Server) gohtRangeToGoRange(uri protocol.DocumentURI, gohtRange protocol.Range) protocol.Range {
	goRange := gohtRange
	sm, ok := s.smc.Get(string(uri))
	if !ok {
		return goRange
	}

	if start, ok := sm.TargetPositionFromSource(int(gohtRange.Start.Line), int(gohtRange.Start.Character)); ok {
		s.logger.Info().Msgf("gohtRangeToGoRange: %s: START [%d,%d] -> [%d,%d]", uri, gohtRange.Start.Line, gohtRange.Start.Character, start.Line, start.Col)
		goRange.Start.Line = uint32(start.Line)
		goRange.Start.Character = uint32(start.Col)
	}

	if end, ok := sm.TargetPositionFromSource(int(gohtRange.End.Line), int(gohtRange.End.Character)); ok {
		s.logger.Info().Msgf("gohtRangeToGoRange: %s: END [%d,%d] -> [%d,%d]", uri, gohtRange.End.Line, gohtRange.End.Character, end.Line, end.Col)
		goRange.End.Line = uint32(end.Line)
		goRange.End.Character = uint32(end.Col)
	}

	return goRange
}

func (s *Server) updatePosition(uri protocol.DocumentURI, pos protocol.Position) (protocol.DocumentURI, protocol.Position, error) {
	logger := s.logger.With().
		Str("uri", string(uri)).
		Uint32("originalLine", pos.Line).
		Uint32("originalColumn", pos.Character).
		Logger()

	isGohtFile, goURI := toGohtGoURI(uri)
	if !isGohtFile {
		return uri, pos, fmt.Errorf("not a goht file")
	}
	sm, ok := s.smc.Get(string(uri))
	if !ok {
		return uri, pos, fmt.Errorf("sourcemap not found")
	}

	to, ok := sm.TargetPositionFromSource(int(pos.Line), int(pos.Character))
	if !ok {
		return uri, pos, fmt.Errorf("mapped position not found")
	}

	logger.Info().
		Int("updatedLine", to.Line).
		Int("updatedColumn", to.Col).
		Msg("updated position")

	return goURI, protocol.Position{
		Line:      uint32(to.Line),
		Character: uint32(to.Col),
	}, nil
}

func (s *Server) parseTemplate(ctx context.Context, uri protocol.DocumentURI, contents string) (*compiler.Template, error) {
	logger := s.logger.With().Str("uri", string(uri)).Logger()

	template, err := compiler.ParseString(contents)
	if err != nil {
		diagnostic := protocol.Diagnostic{
			Severity: protocol.SeverityError,
			Source:   "goht",
			Message:  err.Error(),
		}
		var posErr compiler.PositionalError
		if errors.As(err, &posErr) {
			diagnostic.Range = protocol.Range{
				Start: protocol.Position{
					Line:      uint32(posErr.Line),
					Character: uint32(posErr.Column),
				},
				End: protocol.Position{
					Line:      uint32(posErr.Line),
					Character: uint32(posErr.Column),
				},
			}
		}
		diagnostics := []protocol.Diagnostic{
			diagnostic,
		}
		diagnostics = s.dc.WithGoDiagnostics(string(uri), diagnostics)
		err = s.c.PublishDiagnostics(ctx, &protocol.PublishDiagnosticsParams{
			URI:         uri,
			Diagnostics: diagnostics,
		})
		if err != nil {
			logger.Error().Err(err).Msg("unable to publish diagnostics")
		}
		return template, err
	}
	s.dc.ClearGohtDiagnostics(string(uri))
	err = s.c.PublishDiagnostics(ctx, &protocol.PublishDiagnosticsParams{
		URI:         uri,
		Diagnostics: []protocol.Diagnostic{},
	})
	if err != nil {
		logger.Error().Err(err).Msg("unable to publish diagnostics")
	}
	return template, nil
}

var completionWithImport = regexp.MustCompile(`^.*\(from\s(".+")\)$`)

func getPackageFromItemDetail(pkg string) string {
	if m := completionWithImport.FindStringSubmatch(pkg); len(m) == 2 {
		return m[1]
	}
	return pkg
}

var nonImportKeywordRegexp = regexp.MustCompile(`^(?:goht|func|var|const|type)\s`)

type importInsert struct {
	line int
	text string
}

func addImport(lines []string, pkg string) importInsert {
	var inMultilineImport bool
	lastSingleLineImport := -1
	for i, line := range lines {
		if strings.HasPrefix(line, "import (") {
			inMultilineImport = true
			continue
		}
		if strings.HasPrefix(line, "import ") {
			lastSingleLineImport = i
			continue
		}
		if strings.HasPrefix(line, ")") && inMultilineImport {
			return importInsert{
				line: i,
				text: fmt.Sprintf("\t%s\n", pkg),
			}
		}
		if nonImportKeywordRegexp.MatchString(line) {
			break
		}
	}
	var suffix string
	if lastSingleLineImport == -1 {
		lastSingleLineImport = 1
		suffix = "\n"
	}
	return importInsert{
		line: lastSingleLineImport + 1,
		text: fmt.Sprintf("import %s\n%s", pkg, suffix),
	}
}
