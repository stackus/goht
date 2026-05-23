package proxy

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/rs/zerolog"
	"github.com/stackus/protocol"

	"github.com/stackus/goht"
	"github.com/stackus/goht/compiler"
)

type Server struct {
	protocol.Server
	c                protocol.Client
	smc              *SourceMapCache
	dc               *DiagnosticsCache
	srcs             *DocumentContents
	goSrcs           map[string]string
	positionEncoding protocol.PositionEncodingKind
	logger           zerolog.Logger
}

var _ protocol.Server = (*Server)(nil)

func NewServer(s protocol.Server, c protocol.Client, smc *SourceMapCache, dc *DiagnosticsCache, srcs *DocumentContents, logger zerolog.Logger) *Server {
	return &Server{
		Server:           s,
		c:                c,
		smc:              smc,
		dc:               dc,
		srcs:             srcs,
		goSrcs:           make(map[string]string),
		positionEncoding: protocol.UTF16,
		logger:           logger,
	}
}

// Initialize is called when the client starts up.
//
// It returns the capabilities of the server.
func (s *Server) Initialize(ctx context.Context, params *protocol.ParamInitialize) (*protocol.InitializeResult, error) {
	var logger zerolog.Logger

	if params.ClientInfo != nil {
		logger = s.logger.With().
			Str("method", "Initialize").
			Str("clientName", params.ClientInfo.Name).
			Str("clientVersion", params.ClientInfo.Version).
			Logger()
	} else {
		logger = s.logger.With().
			Str("method", "Initialize").
			Logger()
	}

	// log entry and add deferred exit log
	logger.Info().Msg("Initialize Started")
	defer logger.Trace().Msg("Initialize Completed")

	resp, err := s.Server.Initialize(ctx, params)
	if err != nil {
		logger.Error().Err(err).Msg("unable to initialize server")
	}
	logger.Info().Any("capabilities", resp.Capabilities).Msg("capabilities")
	if resp.Capabilities.CompletionProvider == nil {
		resp.Capabilities.CompletionProvider = &protocol.CompletionOptions{}
	}
	s.sanitizeCapabilities(&resp.Capabilities)
	syncKind := protocol.Full
	if clientSupportsUTF8(params) {
		s.positionEncoding = protocol.UTF8
		resp.Capabilities.PositionEncoding = new(protocol.UTF8)
		syncKind = protocol.Incremental
	} else {
		s.positionEncoding = protocol.UTF16
		resp.Capabilities.PositionEncoding = nil
	}
	resp.Capabilities.DocumentRangeFormattingProvider = &protocol.Or_ServerCapabilities_documentRangeFormattingProvider{Value: false}
	resp.Capabilities.TextDocumentSync = protocol.TextDocumentSyncOptions{
		OpenClose:         true,
		Change:            syncKind,
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

func (s *Server) sanitizeCapabilities(capabilities *protocol.ServerCapabilities) {
	capabilities.ExecuteCommandProvider = nil
	capabilities.DocumentHighlightProvider = nil
	capabilities.DocumentLinkProvider = nil
	capabilities.DocumentSymbolProvider = nil
	capabilities.FoldingRangeProvider = nil
	capabilities.InlayHintProvider = nil

	// Go formatting edits target generated Go and do not preserve GoHT/Haml layout.
	capabilities.DocumentFormattingProvider = &protocol.Or_ServerCapabilities_documentFormattingProvider{Value: false}
	capabilities.DocumentRangeFormattingProvider = &protocol.Or_ServerCapabilities_documentRangeFormattingProvider{Value: false}
	capabilities.DocumentOnTypeFormattingProvider = nil
}

// CodeAction is called when the client requests code actions.
func (s *Server) CodeAction(ctx context.Context, params *protocol.CodeActionParams) ([]protocol.CodeAction, error) {
	logger := s.logger.With().
		Str("method", "CodeAction").
		Str("uri", string(params.TextDocument.URI)).
		Logger()

	// log entry and add deferred exit log
	logger.Info().Msg("CodeAction Started")
	defer logger.Info().Msg("CodeAction Ended")

	gohtURI := params.TextDocument.URI
	isGohtFile, goURI := toGohtGoURI(params.TextDocument.URI)
	if !isGohtFile {
		logger.Warn().Msg("not a goht file")
		return s.Server.CodeAction(ctx, params)
	}
	params.TextDocument.URI = goURI
	var ok bool
	if params.Range, ok = s.tryGohtRangeToGoRange(gohtURI, params.Range); !ok {
		return nil, nil
	}

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

	// log entry and add deferred exit log
	logger.Info().Msg("CodeLens Started")
	defer logger.Info().Msg("CodeLens Ended")

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

	// log entry and add deferred exit log
	logger.Info().Msg("ColorPresentation Started")
	defer logger.Info().Msg("ColorPresentation Ended")

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

	// log entry and add deferred exit log
	logger.Info().Msg("Completion Started")
	defer logger.Info().Msg("Completion Ended")

	gohtURI := params.TextDocument.URI
	var err error
	isGohtFile, goURI := toGohtGoURI(params.TextDocument.URI)
	if !isGohtFile {
		return nil, nil
	}
	params.TextDocument.URI = goURI
	_, params.Position, _ = s.updatePosition(gohtURI, params.Position)
	params.TextDocumentPositionParams.Range = s.gohtRangeToGoRange(gohtURI, params.TextDocumentPositionParams.Range)

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
			if textEdit, ok := completionItem.TextEdit.Value.(protocol.TextEdit); ok {
				textEdit.Range = s.goRangeToGohtRange(gohtURI, textEdit.Range)
				completionItem.TextEdit.Value = textEdit
			}
		}
		if len(completionItem.AdditionalTextEdits) > 0 {
			completionItem.AdditionalTextEdits = s.mapCompletionAdditionalTextEdits(gohtURI, completionItem)
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

	// log entry and add deferred exit log
	logger.Info().Msg("Declaration Started")
	defer logger.Info().Msg("Declaration Ended")

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

	// log entry and add deferred exit log
	logger.Info().Msg("Definition Started")
	defer logger.Info().Msg("Definition Ended")

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
	return s.mapLocations(resp), nil
}

func (s *Server) DidChange(ctx context.Context, params *protocol.DidChangeTextDocumentParams) error {
	logger := s.logger.With().
		Str("method", "DidChange").
		Str("uri", string(params.TextDocument.URI)).
		Logger()

	// log entry and add deferred exit log
	logger.Info().Msg("DidChange Started")
	defer logger.Info().Msg("DidChange Ended")

	isGohtFile, goURI := toGohtGoURI(params.TextDocument.URI)
	if !isGohtFile {
		logger.Warn().Msg("not a goht file")
		return nil
	}

	gohtURI := params.TextDocument.URI
	doc, err := s.srcs.Apply(string(gohtURI), params.ContentChanges, s.positionEncoding)
	if err != nil {
		logger.Error().Err(err).Msg("unable to apply changes")
		return err
	}

	template, err := s.parseTemplate(ctx, gohtURI, doc.String())
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
	_, alreadyOpen := s.goSrcs[string(gohtURI)]
	s.smc.Set(string(gohtURI), sm)
	s.goSrcs[string(gohtURI)] = buf.String()
	if !alreadyOpen {
		openParams := &protocol.DidOpenTextDocumentParams{
			TextDocument: protocol.TextDocumentItem{
				URI:        goURI,
				LanguageID: "go",
				Version:    params.TextDocument.Version,
				Text:       buf.String(),
			},
		}
		err = s.Server.DidOpen(ctx, openParams)
		if err != nil {
			logger.Error().Err(err).Msg("unable to open document")
		}
		return err
	}
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

	// log entry and add deferred exit log
	logger.Info().Msg("DidClose Started")
	defer logger.Info().Msg("DidClose Ended")

	isGohtFile, goURI := toGohtGoURI(params.TextDocument.URI)
	if !isGohtFile {
		logger.Warn().Msg("not a goht file")
		return s.Server.DidClose(ctx, params)
	}
	s.srcs.Delete(string(params.TextDocument.URI))
	delete(s.goSrcs, string(params.TextDocument.URI))
	s.smc.Delete(string(params.TextDocument.URI))
	s.dc.Delete(string(params.TextDocument.URI))
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

	// log entry and add deferred exit log
	logger.Info().Msg("DidOpen Started")
	defer logger.Info().Msg("DidOpen Ended")

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

	// log entry and add deferred exit log
	logger.Info().Msg("DidSave Started")
	defer logger.Info().Msg("DidSave Ended")

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

	// log entry and add deferred exit log
	logger.Info().Msg("DocumentColor Started")
	defer logger.Info().Msg("DocumentColor Ended")

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

	// log entry and add deferred exit log
	logger.Info().Msg("ResolveDocumentLink Started")
	defer logger.Info().Msg("ResolveDocumentLink Ended")

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

	// log entry and add deferred exit log
	logger.Info().Msg("Hover Started")
	defer logger.Info().Msg("Hover Ended")

	gohtURI := params.TextDocument.URI
	var err error
	isGohtFile, goURI := toGohtGoURI(params.TextDocument.URI)
	if !isGohtFile {
		return nil, nil
	}
	var ok bool
	params.TextDocument.URI = goURI
	_, params.Position, _ = s.updatePosition(gohtURI, params.Position)
	params.TextDocumentPositionParams.Range, ok = s.tryGohtRangeToGoRange(gohtURI, params.TextDocumentPositionParams.Range)
	if !ok {
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

	// log entry and add deferred exit log
	logger.Info().Msg("Implementation Started")
	defer logger.Info().Msg("Implementation Ended")

	gohtURI := params.TextDocument.URI
	var err error
	isGohtFile, goURI := toGohtGoURI(params.TextDocument.URI)
	if !isGohtFile {
		return nil, nil
	}
	params.TextDocument.URI = goURI
	_, params.Position, _ = s.updatePosition(gohtURI, params.Position)
	params.TextDocumentPositionParams.Range = s.gohtRangeToGoRange(gohtURI, params.TextDocumentPositionParams.Range)

	resp, err := s.Server.Implementation(ctx, params)
	if err != nil || resp == nil {
		if err != nil {
			logger.Error().Err(err).Msg("unable to perform implementation lookup")
		}
		return resp, err
	}
	return s.mapLocations(resp), nil
}

func (s *Server) OnTypeFormatting(ctx context.Context, params *protocol.DocumentOnTypeFormattingParams) ([]protocol.TextEdit, error) {
	logger := s.logger.With().
		Str("method", "OnTypeFormatting").
		Str("uri", string(params.TextDocument.URI)).
		Logger()

	// log entry and add deferred exit log
	logger.Info().Msg("OnTypeFormatting Started")
	defer logger.Info().Msg("OnTypeFormatting Ended")

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

	// log entry and add deferred exit log
	logger.Info().Msg("PrepareRename Started")
	defer logger.Info().Msg("PrepareRename Ended")

	gohtURI := params.TextDocument.URI
	var err error
	isGohtFile, goURI := toGohtGoURI(params.TextDocument.URI)
	if !isGohtFile {
		return nil, nil
	}
	params.TextDocument.URI = goURI
	_, params.Position, _ = s.updatePosition(gohtURI, params.Position)
	params.TextDocumentPositionParams.Range = s.gohtRangeToGoRange(gohtURI, params.TextDocumentPositionParams.Range)

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

	// log entry and add deferred exit log
	logger.Info().Msg("RangeFormatting Started")
	defer logger.Info().Msg("RangeFormatting Ended")

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

	// log entry and add deferred exit log
	logger.Info().Msg("References Started")
	defer logger.Info().Msg("References Ended")

	gohtURI := params.TextDocument.URI
	var err error
	isGohtFile, goURI := toGohtGoURI(params.TextDocument.URI)
	if !isGohtFile {
		return []protocol.Location{}, nil
	}
	params.TextDocument.URI = goURI
	_, params.Position, _ = s.updatePosition(gohtURI, params.Position)
	params.TextDocumentPositionParams.Range = s.gohtRangeToGoRange(gohtURI, params.TextDocumentPositionParams.Range)

	resp, err := s.Server.References(ctx, params)
	if err != nil || resp == nil {
		if err != nil {
			logger.Error().Err(err).Msg("unable to perform references lookup")
		}
		return resp, err
	}
	return s.mapLocations(resp), nil
}

func (s *Server) SignatureHelp(ctx context.Context, params *protocol.SignatureHelpParams) (*protocol.SignatureHelp, error) {
	logger := s.logger.With().
		Str("method", "SignatureHelp").
		Str("uri", string(params.TextDocument.URI)).
		Logger()

	// log entry and add deferred exit log
	logger.Info().Msg("SignatureHelp Started")
	defer logger.Info().Msg("SignatureHelp Ended")

	gohtURI := params.TextDocument.URI
	var err error
	isGohtFile, goURI := toGohtGoURI(params.TextDocument.URI)
	if !isGohtFile {
		return nil, nil
	}
	params.TextDocument.URI = goURI
	_, params.Position, _ = s.updatePosition(gohtURI, params.Position)
	params.TextDocumentPositionParams.Range = s.gohtRangeToGoRange(gohtURI, params.TextDocumentPositionParams.Range)

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

	// log entry and add deferred exit log
	logger.Info().Msg("TypeDefinition Started")
	defer logger.Info().Msg("TypeDefinition Ended")

	gohtURI := params.TextDocument.URI
	var err error
	isGohtFile, goURI := toGohtGoURI(params.TextDocument.URI)
	if !isGohtFile {
		return []protocol.Location{}, nil
	}
	params.TextDocument.URI = goURI
	_, params.Position, _ = s.updatePosition(gohtURI, params.Position)
	params.TextDocumentPositionParams.Range = s.gohtRangeToGoRange(gohtURI, params.TextDocumentPositionParams.Range)

	resp, err := s.Server.TypeDefinition(ctx, params)
	if err != nil {
		logger.Error().Err(err).Msg("unable to perform type definition lookup")
	}
	return s.mapLocations(resp), err
}

func (s *Server) WillSave(ctx context.Context, params *protocol.WillSaveTextDocumentParams) error {
	logger := s.logger.With().
		Str("method", "WillSave").
		Str("uri", string(params.TextDocument.URI)).
		Logger()

	// log entry and add deferred exit log
	logger.Info().Msg("WillSave Started")
	defer logger.Info().Msg("WillSave Ended")

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

	// log entry and add deferred exit log
	logger.Info().Msg("SemanticTokensFull Started")
	defer logger.Info().Msg("SemanticTokensFull Ended")

	gohtURI := params.TextDocument.URI
	isGohtFile, goURI := toGohtGoURI(params.TextDocument.URI)
	if !isGohtFile {
		logger.Warn().Msg("not a goht file")
		return nil, nil
	}
	params.TextDocument.URI = goURI
	resp, err := s.Server.SemanticTokensFull(ctx, params)
	if err != nil {
		logger.Error().Err(err).Msg("unable to perform semantic tokens full")
		return resp, err
	}

	return s.mapSemanticTokens(gohtURI, resp), nil
}

func (s *Server) SemanticTokensFullDelta(_ context.Context, params *protocol.SemanticTokensDeltaParams) (any, error) {
	logger := s.logger.With().
		Str("method", "SemanticTokensFullDelta").
		Str("uri", string(params.TextDocument.URI)).
		Logger()

	// log entry and add deferred exit log
	logger.Info().Msg("SemanticTokensFullDelta Started")
	defer logger.Info().Msg("SemanticTokensFullDelta Ended")

	logger.Warn().Msg("delta not supported; client should request full tokens")
	return nil, nil
}

func (s *Server) SemanticTokensRange(ctx context.Context, params *protocol.SemanticTokensRangeParams) (*protocol.SemanticTokens, error) {
	logger := s.logger.With().
		Str("method", "SemanticTokensRange").
		Str("uri", string(params.TextDocument.URI)).
		Logger()

	// log entry and add deferred exit log
	logger.Info().Msg("SemanticTokensRange Started")
	defer logger.Info().Msg("SemanticTokensRange Ended")

	gohtURI := params.TextDocument.URI
	isGohtFile, goURI := toGohtGoURI(params.TextDocument.URI)
	if !isGohtFile {
		logger.Warn().Msg("not a goht file")
		return nil, nil
	}

	goRange, ok := s.tryGohtRangeToGoRange(gohtURI, params.Range)
	if !ok {
		return nil, nil
	}

	params.TextDocument.URI = goURI
	params.Range = goRange
	resp, err := s.Server.SemanticTokensRange(ctx, params)
	if err != nil {
		logger.Error().Err(err).Msg("unable to perform semantic tokens range")
		return resp, err
	}

	return s.mapSemanticTokens(gohtURI, resp), nil
}

func (s *Server) Moniker(ctx context.Context, params *protocol.MonikerParams) ([]protocol.Moniker, error) {
	logger := s.logger.With().
		Str("method", "Moniker").
		Str("uri", string(params.TextDocument.URI)).
		Logger()

	// log entry and add deferred exit log
	logger.Info().Msg("Moniker Started")
	defer logger.Info().Msg("Moniker Ended")

	gohtURI := params.TextDocument.URI
	var err error
	isGohtFile, goURI := toGohtGoURI(params.TextDocument.URI)
	if !isGohtFile {
		return []protocol.Moniker{}, nil
	}
	params.TextDocument.URI = goURI
	_, params.Position, _ = s.updatePosition(gohtURI, params.Position)
	params.TextDocumentPositionParams.Range = s.gohtRangeToGoRange(gohtURI, params.TextDocumentPositionParams.Range)

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

func (s *Server) mapLocations(locations []protocol.Location) []protocol.Location {
	for i, location := range locations {
		locations[i] = s.mapLocation(location)
	}
	return locations
}

func (s *Server) mapLocation(location protocol.Location) protocol.Location {
	isGohtGoFile, gohtURI := toGohtURI(location.URI)
	if !isGohtGoFile {
		return location
	}
	location.URI = gohtURI
	location.Range = s.goRangeToGohtRange(gohtURI, location.Range)
	return location
}

func (s *Server) mapTextEdit(uri protocol.DocumentURI, edit protocol.TextEdit) (protocol.TextEdit, bool) {
	mappedRange, ok := s.tryGoRangeToGohtRange(uri, edit.Range)
	if !ok {
		return edit, false
	}
	edit.Range = mappedRange
	return edit, true
}

func (s *Server) mapCompletionAdditionalTextEdits(uri protocol.DocumentURI, item protocol.CompletionItem) []protocol.TextEdit {
	edits := make([]protocol.TextEdit, 0, len(item.AdditionalTextEdits))
	for _, edit := range item.AdditionalTextEdits {
		if mapped, ok := s.mapTextEdit(uri, edit); ok {
			edits = append(edits, mapped)
			continue
		}
		if !isGeneratedImportEdit(edit, item.Detail) {
			s.logger.Warn().Str("newText", edit.NewText).Msg("dropping unmapped completion additional text edit")
			continue
		}
		doc, ok := s.srcs.Get(string(uri))
		if !ok {
			s.logger.Warn().Str("uri", string(uri)).Msg("dropping unmapped completion import edit without source document")
			continue
		}
		insert := addImport(doc.lines, getPackageFromItemDetail(item.Detail))
		edits = append(edits, protocol.TextEdit{
			Range: protocol.Range{
				Start: protocol.Position{Line: uint32(insert.line), Character: 0},
				End:   protocol.Position{Line: uint32(insert.line), Character: 0},
			},
			NewText: insert.text,
		})
	}
	return edits
}

func isGeneratedImportEdit(edit protocol.TextEdit, detail string) bool {
	if strings.HasPrefix(strings.TrimSpace(edit.NewText), "import ") {
		return true
	}
	return completionWithImport.MatchString(detail)
}

func (s *Server) goRangeToGohtRange(uri protocol.DocumentURI, goRange protocol.Range) protocol.Range {
	mappedRange, ok := s.tryGoRangeToGohtRange(uri, goRange)
	if !ok {
		return goRange
	}
	return mappedRange
}

func (s *Server) tryGoRangeToGohtRange(uri protocol.DocumentURI, goRange protocol.Range) (protocol.Range, bool) {
	sm, ok := s.smc.Get(string(uri))
	if !ok {
		return goRange, false
	}

	start, ok := sm.SourcePositionFromTarget(int(goRange.Start.Line), int(goRange.Start.Character))
	if !ok {
		return goRange, false
	}

	end, ok := sm.SourcePositionFromTarget(int(goRange.End.Line), int(goRange.End.Character))
	if !ok {
		return goRange, false
	}

	s.logger.Info().Msgf("goRangeToGohtRange: %s: START [%d,%d] -> [%d,%d]", uri, goRange.Start.Line, goRange.Start.Character, start.Line, start.Col)
	s.logger.Info().Msgf("goRangeToGohtRange: %s: END [%d,%d] -> [%d,%d]", uri, goRange.End.Line, goRange.End.Character, end.Line, end.Col)

	return protocol.Range{
		Start: protocol.Position{
			Line:      uint32(start.Line),
			Character: uint32(start.Col),
		},
		End: protocol.Position{
			Line:      uint32(end.Line),
			Character: uint32(end.Col),
		},
	}, true
}

func (s *Server) gohtRangeToGoRange(uri protocol.DocumentURI, gohtRange protocol.Range) protocol.Range {
	mappedRange, ok := s.tryGohtRangeToGoRange(uri, gohtRange)
	if !ok {
		return gohtRange
	}
	return mappedRange
}

func (s *Server) tryGohtRangeToGoRange(uri protocol.DocumentURI, gohtRange protocol.Range) (protocol.Range, bool) {
	sm, ok := s.smc.Get(string(uri))
	if !ok {
		return gohtRange, false
	}

	start, ok := sm.TargetPositionFromSource(int(gohtRange.Start.Line), int(gohtRange.Start.Character))
	if !ok {
		return gohtRange, false
	}

	end, ok := sm.TargetPositionFromSource(int(gohtRange.End.Line), int(gohtRange.End.Character))
	if !ok {
		return gohtRange, false
	}

	s.logger.Info().Msgf("gohtRangeToGoRange: %s: START [%d,%d] -> [%d,%d]", uri, gohtRange.Start.Line, gohtRange.Start.Character, start.Line, start.Col)
	s.logger.Info().Msgf("gohtRangeToGoRange: %s: END [%d,%d] -> [%d,%d]", uri, gohtRange.End.Line, gohtRange.End.Character, end.Line, end.Col)

	return protocol.Range{
		Start: protocol.Position{
			Line:      uint32(start.Line),
			Character: uint32(start.Col),
		},
		End: protocol.Position{
			Line:      uint32(end.Line),
			Character: uint32(end.Col),
		},
	}, true
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

func (s *Server) updateRangeZ(uri protocol.DocumentURI, pos protocol.Range) (protocol.DocumentURI, protocol.Range, error) {
	logger := s.logger.With().
		Str("uri", string(uri)).
		Uint32("originalLineStart", pos.Start.Line).
		Uint32("originalColumnStart", pos.Start.Character).
		Uint32("originalLineEnd", pos.End.Line).
		Uint32("originalColumnEnd", pos.End.Character).
		Logger()

	isGohtFile, goURI := toGohtGoURI(uri)
	if !isGohtFile {
		return uri, pos, fmt.Errorf("not a goht file")
	}
	sm, ok := s.smc.Get(string(uri))
	if !ok {
		return uri, pos, fmt.Errorf("sourcemap not found")
	}

	start, ok := sm.TargetPositionFromSource(int(pos.Start.Line), int(pos.Start.Character))
	if !ok {
		return uri, pos, fmt.Errorf("mapped start position not found")
	}

	end, ok := sm.TargetPositionFromSource(int(pos.End.Line), int(pos.End.Character))
	if !ok {
		return uri, pos, fmt.Errorf("mapped end position not found")
	}

	logger.Info().
		Int("updatedLineStart", start.Line).
		Int("updatedColumnStart", start.Col).
		Int("updatedLineEnd", end.Line).
		Int("updatedColumnEnd", end.Col).
		Msg("updated range")

	return goURI, protocol.Range{
		Start: protocol.Position{
			Line:      uint32(start.Line),
			Character: uint32(start.Col),
		},
		End: protocol.Position{
			Line:      uint32(end.Line),
			Character: uint32(end.Col),
		},
	}, nil
}

func (s *Server) parseTemplate(ctx context.Context, uri protocol.DocumentURI, contents string) (*compiler.Template, error) {
	logger := s.logger.With().Str("uri", string(uri)).Logger()

	template, err := compiler.ParseString(contents)
	if err != nil {
		parseErr := err
		diagnostic := protocol.Diagnostic{
			Severity: protocol.SeverityError,
			Source:   "goht",
			Message:  err.Error(),
		}
		if posErr, ok := errors.AsType[compiler.PositionalError](err); ok {
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
		diagnostics = s.dc.WithParserDiagnostics(string(uri), diagnostics)
		err = s.c.PublishDiagnostics(ctx, &protocol.PublishDiagnosticsParams{
			URI:         uri,
			Diagnostics: diagnostics,
		})
		if err != nil {
			logger.Error().Err(err).Msg("unable to publish diagnostics")
		}
		return template, parseErr
	}
	diagnostics := s.dc.ClearParserDiagnostics(string(uri))
	err = s.c.PublishDiagnostics(ctx, &protocol.PublishDiagnosticsParams{
		URI:         uri,
		Diagnostics: diagnostics,
	})
	if err != nil {
		logger.Error().Err(err).Msg("unable to publish diagnostics")
	}
	return template, nil
}

func clientSupportsUTF8(params *protocol.ParamInitialize) bool {
	if params == nil || params.Capabilities.General == nil {
		return false
	}
	for _, encoding := range params.Capabilities.General.PositionEncodings {
		if encoding == protocol.UTF8 {
			return true
		}
	}
	return false
}

var completionWithImport = regexp.MustCompile(`^.*\(from\s(".+")\)$`)

func getPackageFromItemDetail(pkg string) string {
	if m := completionWithImport.FindStringSubmatch(pkg); len(m) == 2 {
		return m[1]
	}
	return pkg
}

var nonImportKeywordRegexp = regexp.MustCompile(`^(?:@?goht|func|var|const|type)\s`)

type importInsert struct {
	line int
	text string
}

type semanticToken struct {
	line, char, length, tokenType, tokenModifiers uint32
}

func decodeSemanticTokens(data []uint32) []semanticToken {
	if len(data) == 0 {
		return nil
	}
	tokens := make([]semanticToken, 0, len(data)/5)
	var prevLine, prevChar uint32
	for i := 0; i+4 < len(data); i += 5 {
		deltaLine := data[i]
		deltaChar := data[i+1]
		length := data[i+2]
		tokenType := data[i+3]
		tokenMods := data[i+4]

		var line, char uint32
		if deltaLine == 0 {
			line = prevLine
			char = prevChar + deltaChar
		} else {
			line = prevLine + deltaLine
			char = deltaChar
		}

		tokens = append(tokens, semanticToken{
			line:           line,
			char:           char,
			length:         length,
			tokenType:      tokenType,
			tokenModifiers: tokenMods,
		})
		prevLine = line
		prevChar = char
	}
	return tokens
}

func (s *Server) mapSemanticTokens(uri protocol.DocumentURI, tokens *protocol.SemanticTokens) *protocol.SemanticTokens {
	if tokens == nil {
		return nil
	}
	sm, ok := s.smc.Get(string(uri))
	if !ok {
		return nil
	}

	decoded := decodeSemanticTokens(tokens.Data)
	mapped := make([]semanticToken, 0, len(decoded))
	for _, t := range decoded {
		gohtPos, ok := sm.SourcePositionFromTarget(int(t.line), int(t.char))
		if !ok {
			continue
		}
		mapped = append(mapped, semanticToken{
			line:           uint32(gohtPos.Line),
			char:           uint32(gohtPos.Col),
			length:         t.length,
			tokenType:      t.tokenType,
			tokenModifiers: t.tokenModifiers,
		})
	}

	return &protocol.SemanticTokens{Data: encodeSemanticTokens(mapped)}
}

func encodeSemanticTokens(tokens []semanticToken) []uint32 {
	data := make([]uint32, 0, len(tokens)*5)
	var prevLine, prevChar uint32
	for _, t := range tokens {
		deltaLine := t.line - prevLine
		var deltaChar uint32
		if deltaLine == 0 {
			deltaChar = t.char - prevChar
		} else {
			deltaChar = t.char
		}
		data = append(data, deltaLine, deltaChar, t.length, t.tokenType, t.tokenModifiers)
		prevLine = t.line
		prevChar = t.char
	}
	return data
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
		lastSingleLineImport = 0
		suffix = "\n"
	}
	return importInsert{
		line: lastSingleLineImport + 1,
		text: fmt.Sprintf("import %s\n%s", pkg, suffix),
	}
}
