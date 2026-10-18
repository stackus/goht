package proxy

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stackus/protocol"

	"github.com/stackus/goht/compiler"
)

func TestBuildLegendAppendsGohtTypes(t *testing.T) {
	gopls := &protocol.SemanticTokensLegend{
		TokenTypes:     []string{"namespace", "type", "string", "comment", "keyword"},
		TokenModifiers: []string{"definition", "readonly"},
	}

	got := buildLegend(gopls)

	wantTypes := []string{"namespace", "type", "string", "comment", "keyword", "gohtDirective", "gohtTag", "gohtAttribute"}
	if !reflect.DeepEqual(got.legend.TokenTypes, wantTypes) {
		t.Fatalf("TokenTypes = %v, want %v", got.legend.TokenTypes, wantTypes)
	}
	if !reflect.DeepEqual(got.legend.TokenModifiers, gopls.TokenModifiers) {
		t.Fatalf("TokenModifiers = %v, want %v", got.legend.TokenModifiers, gopls.TokenModifiers)
	}
	wantKinds := map[compiler.HighlightKind]uint32{
		compiler.HighlightDirective: 5,
		compiler.HighlightTag:       6,
		compiler.HighlightAttribute: 7,
		compiler.HighlightString:    2,
		compiler.HighlightComment:   3,
		compiler.HighlightKeyword:   4,
	}
	if !reflect.DeepEqual(got.kinds, wantKinds) {
		t.Fatalf("kinds = %v, want %v", got.kinds, wantKinds)
	}
}

func TestBuildLegendNilGopls(t *testing.T) {
	got := buildLegend(nil)

	wantTypes := []string{"gohtDirective", "gohtTag", "gohtAttribute", "string", "comment", "keyword"}
	if !reflect.DeepEqual(got.legend.TokenTypes, wantTypes) {
		t.Fatalf("TokenTypes = %v, want %v", got.legend.TokenTypes, wantTypes)
	}
	if got.legend.TokenModifiers == nil {
		t.Fatal("TokenModifiers must be an empty list, not nil")
	}
	if got.kinds[compiler.HighlightKeyword] != 5 {
		t.Fatalf("keyword index = %d, want 5", got.kinds[compiler.HighlightKeyword])
	}
}

func TestInitializeAdvertisesSemanticTokens(t *testing.T) {
	t.Run("merges the gopls legend", func(t *testing.T) {
		server := &recordingServer{initializeResult: &protocol.InitializeResult{
			ServerInfo: &protocol.ServerInfo{},
			Capabilities: protocol.ServerCapabilities{
				// capabilities decoded from gopls arrive as generic JSON values
				SemanticTokensProvider: map[string]any{
					"legend": map[string]any{
						"tokenTypes":     []any{"namespace", "type"},
						"tokenModifiers": []any{"definition"},
					},
					"full":  true,
					"range": true,
				},
			},
		}}
		proxy := newTestServer(server, &recordingClient{})

		got, err := proxy.Initialize(context.Background(), &protocol.ParamInitialize{})
		if err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}
		opts, ok := got.Capabilities.SemanticTokensProvider.(*protocol.SemanticTokensOptions)
		if !ok {
			t.Fatalf("SemanticTokensProvider = %T, want *protocol.SemanticTokensOptions", got.Capabilities.SemanticTokensProvider)
		}
		wantTypes := []string{"namespace", "type", "gohtDirective", "gohtTag", "gohtAttribute", "string", "comment", "keyword"}
		if !reflect.DeepEqual(opts.Legend.TokenTypes, wantTypes) {
			t.Fatalf("TokenTypes = %v, want %v", opts.Legend.TokenTypes, wantTypes)
		}
		if !reflect.DeepEqual(opts.Legend.TokenModifiers, []string{"definition"}) {
			t.Fatalf("TokenModifiers = %v", opts.Legend.TokenModifiers)
		}
		if opts.Full == nil || opts.Full.Value != true || opts.Range == nil || opts.Range.Value != true {
			t.Fatalf("Full = %v, Range = %v, want true, true", opts.Full, opts.Range)
		}
		if !proxy.goplsSemantic {
			t.Fatal("goplsSemantic = false, want true")
		}
	})

	t.Run("advertises GoHT tokens without gopls support", func(t *testing.T) {
		proxy := newTestServer(&recordingServer{}, &recordingClient{})

		got, err := proxy.Initialize(context.Background(), &protocol.ParamInitialize{})
		if err != nil {
			t.Fatalf("Initialize() error = %v", err)
		}
		opts, ok := got.Capabilities.SemanticTokensProvider.(*protocol.SemanticTokensOptions)
		if !ok {
			t.Fatalf("SemanticTokensProvider = %T, want *protocol.SemanticTokensOptions", got.Capabilities.SemanticTokensProvider)
		}
		if !reflect.DeepEqual(opts.Legend.TokenTypes, buildLegend(nil).legend.TokenTypes) {
			t.Fatalf("TokenTypes = %v", opts.Legend.TokenTypes)
		}
		if proxy.goplsSemantic {
			t.Fatal("goplsSemantic = true, want false")
		}
	})
}

func TestSourceMapCacheStale(t *testing.T) {
	smc := NewSourceMapCache()
	smc.Set("a", &compiler.SourceMap{})
	if smc.IsStale("a") {
		t.Fatal("new source map is stale")
	}
	smc.MarkStale("a")
	if !smc.IsStale("a") {
		t.Fatal("MarkStale did not mark the source map stale")
	}
	smc.Set("a", &compiler.SourceMap{})
	if smc.IsStale("a") {
		t.Fatal("Set did not clear stale")
	}
}

func TestDidChangeParseErrorMarksStale(t *testing.T) {
	proxy := newTestServer(&recordingServer{}, &recordingClient{})
	ctx := context.Background()
	if err := proxy.DidOpen(ctx, didOpenParams(validGoht)); err != nil {
		t.Fatal(err)
	}
	if proxy.smc.IsStale(string(testGohtURI)) {
		t.Fatal("stale after valid open")
	}
	if err := proxy.DidChange(ctx, didChangeParams(invalidGohtMissingComma)); err != nil {
		t.Fatal(err)
	}
	if !proxy.smc.IsStale(string(testGohtURI)) {
		t.Fatal("not stale after parse error")
	}
	if err := proxy.DidChange(ctx, didChangeParams(validGoht)); err != nil {
		t.Fatal(err)
	}
	if proxy.smc.IsStale(string(testGohtURI)) {
		t.Fatal("still stale after valid change")
	}
}

const nonASCIIGoht = "package main\n\n@slim Test(user User) {\n\tp héllo #{user.Name}\n}\n"

// openForMapping opens src and returns the zero-based Go line and byte column
// of needle in the generated Go source.
func openForMapping(t *testing.T, proxy *Server, src, needle string) (uint32, uint32) {
	t.Helper()
	if err := proxy.DidOpen(context.Background(), didOpenParams(src)); err != nil {
		t.Fatal(err)
	}
	for i, line := range strings.Split(proxy.goSrcs[string(testGohtURI)], "\n") {
		if col := strings.Index(line, needle); col >= 0 {
			return uint32(i), uint32(col)
		}
	}
	t.Fatalf("%q not found in generated Go", needle)
	return 0, 0
}

func TestMapGoplsTokensUTF16(t *testing.T) {
	proxy := newTestServer(&recordingServer{}, &recordingClient{})
	line, col := openForMapping(t, proxy, nonASCIIGoht, "user.Name")

	got := proxy.mapGoplsTokens(testGohtURI, []semanticToken{
		{line: line, char: col, length: 4, tokenType: 1},
		{line: line, char: col + 5, length: 4, tokenType: 2},
	})

	// "\tp héllo #{user.Name}": user starts at UTF-16 column 11
	want := []semanticToken{
		{line: 3, char: 11, length: 4, tokenType: 1},
		{line: 3, char: 16, length: 4, tokenType: 2},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("mapGoplsTokens() = %+v, want %+v", got, want)
	}
}

func TestMapGoplsTokensUTF8(t *testing.T) {
	proxy := newTestServer(&recordingServer{}, &recordingClient{})
	proxy.positionEncoding = protocol.UTF8
	line, col := openForMapping(t, proxy, nonASCIIGoht, "user.Name")

	got := proxy.mapGoplsTokens(testGohtURI, []semanticToken{{line: line, char: col, length: 4}})

	want := []semanticToken{{line: 3, char: 12, length: 4}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("mapGoplsTokens() = %+v, want %+v", got, want)
	}
}

func TestMapGoplsTokensDropsGeneratedOnly(t *testing.T) {
	proxy := newTestServer(&recordingServer{}, &recordingClient{})
	line, col := openForMapping(t, proxy, nonASCIIGoht, "user.Name")

	got := proxy.mapGoplsTokens(testGohtURI, []semanticToken{
		// runs past the end of the mapped expression into generated code
		{line: line, char: col, length: 40},
		// generated-only line
		{line: 0, char: 0, length: 4},
	})

	if len(got) != 0 {
		t.Fatalf("mapGoplsTokens() = %+v, want none", got)
	}
}

func TestMapGoplsTokensSorted(t *testing.T) {
	proxy := newTestServer(&recordingServer{}, &recordingClient{})
	line, col := openForMapping(t, proxy, nonASCIIGoht, "user.Name")

	got := proxy.mapGoplsTokens(testGohtURI, []semanticToken{
		{line: line, char: col + 5, length: 4},
		{line: line, char: col, length: 4},
	})

	if len(got) != 2 || got[0].char > got[1].char {
		t.Fatalf("mapGoplsTokens() = %+v, want sorted", got)
	}
}

// decodedTokens returns the tokens of a response in absolute positions.
func decodedTokens(t *testing.T, got *protocol.SemanticTokens) []semanticToken {
	t.Helper()
	if got == nil {
		t.Fatal("semantic tokens = nil, want non-nil")
	}
	return decodeSemanticTokens(got.Data)
}

func fullParams() *protocol.SemanticTokensParams {
	return &protocol.SemanticTokensParams{TextDocument: protocol.TextDocumentIdentifier{URI: testGohtURI}}
}

const mergeGoht = "package main\n\n@slim Test(user User) {\n\th1= user.Name\n}\n"

func TestSemanticTokensFullMergesGohtAndGopls(t *testing.T) {
	server := &recordingServer{}
	proxy := newTestServer(server, &recordingClient{})
	line, col := openForMapping(t, proxy, mergeGoht, "user.Name")
	server.semanticTokensFullResult = &protocol.SemanticTokens{
		Data: encodeSemanticTokens([]semanticToken{{line: line, char: col, length: 4, tokenType: 1}}),
	}

	got, err := proxy.SemanticTokensFull(context.Background(), fullParams())
	if err != nil {
		t.Fatal(err)
	}

	kinds := proxy.legend.kinds
	want := []semanticToken{
		{line: 2, char: 0, length: 5, tokenType: kinds[compiler.HighlightDirective]},
		{line: 3, char: 1, length: 2, tokenType: kinds[compiler.HighlightTag]},
		{line: 3, char: 5, length: 4, tokenType: 1},
	}
	if gotTokens := decodedTokens(t, got); !reflect.DeepEqual(gotTokens, want) {
		t.Fatalf("tokens = %+v, want %+v", gotTokens, want)
	}
}

func TestSemanticTokensFullStaleSkipsGopls(t *testing.T) {
	server := &recordingServer{semanticTokensFullResult: &protocol.SemanticTokens{Data: []uint32{0, 0, 4, 1, 0}}}
	proxy := newTestServer(server, &recordingClient{})
	ctx := context.Background()
	if err := proxy.DidOpen(ctx, didOpenParams(mergeGoht)); err != nil {
		t.Fatal(err)
	}
	if err := proxy.DidChange(ctx, didChangeParams("package main\n\n@slim Test(user User) {\n\th1= user.Name\n\tp #{user.\n")); err != nil {
		t.Fatal(err)
	}

	got, err := proxy.SemanticTokensFull(ctx, fullParams())
	if err != nil {
		t.Fatal(err)
	}

	if len(server.semanticTokensFullCalls) != 0 {
		t.Fatalf("gopls calls = %d, want 0", len(server.semanticTokensFullCalls))
	}
	kinds := proxy.legend.kinds
	want := []semanticToken{
		{line: 2, char: 0, length: 5, tokenType: kinds[compiler.HighlightDirective]},
		{line: 3, char: 1, length: 2, tokenType: kinds[compiler.HighlightTag]},
		{line: 4, char: 1, length: 1, tokenType: kinds[compiler.HighlightTag]},
	}
	if gotTokens := decodedTokens(t, got); !reflect.DeepEqual(gotTokens, want) {
		t.Fatalf("tokens = %+v, want %+v", gotTokens, want)
	}
}

type failingSemanticServer struct {
	recordingServer
}

func (s *failingSemanticServer) SemanticTokensFull(context.Context, *protocol.SemanticTokensParams) (*protocol.SemanticTokens, error) {
	return nil, errors.New("gopls failed")
}

func TestSemanticTokensFullGoplsErrorFallsBack(t *testing.T) {
	server := &failingSemanticServer{}
	proxy := NewServer(server, &recordingClient{}, NewSourceMapCache(), NewDiagnosticsCache(), NewDocumentContents(), zerolog.Nop())
	if err := proxy.DidOpen(context.Background(), didOpenParams(mergeGoht)); err != nil {
		t.Fatal(err)
	}

	got, err := proxy.SemanticTokensFull(context.Background(), fullParams())
	if err != nil {
		t.Fatalf("SemanticTokensFull() error = %v, want nil", err)
	}
	if n := len(decodedTokens(t, got)); n != 2 {
		t.Fatalf("tokens = %d, want 2 GoHT tokens", n)
	}
}

func TestSemanticTokensFullNoGoplsCapability(t *testing.T) {
	server := &recordingServer{semanticTokensFullResult: &protocol.SemanticTokens{Data: []uint32{0, 0, 4, 1, 0}}}
	proxy := newTestServer(server, &recordingClient{})
	if _, err := proxy.Initialize(context.Background(), &protocol.ParamInitialize{}); err != nil {
		t.Fatal(err)
	}
	if err := proxy.DidOpen(context.Background(), didOpenParams(mergeGoht)); err != nil {
		t.Fatal(err)
	}

	got, err := proxy.SemanticTokensFull(context.Background(), fullParams())
	if err != nil {
		t.Fatal(err)
	}
	if len(server.semanticTokensFullCalls) != 0 {
		t.Fatalf("gopls calls = %d, want 0", len(server.semanticTokensFullCalls))
	}
	if n := len(decodedTokens(t, got)); n != 2 {
		t.Fatalf("tokens = %d, want 2 GoHT tokens", n)
	}
}

func TestSemanticTokensRangeFilters(t *testing.T) {
	proxy := newTestServer(&recordingServer{}, &recordingClient{})
	if err := proxy.DidOpen(context.Background(), didOpenParams(mergeGoht)); err != nil {
		t.Fatal(err)
	}

	got, err := proxy.SemanticTokensRange(context.Background(), &protocol.SemanticTokensRangeParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: testGohtURI},
		Range:        rangeOf(3, 0, 4, 0),
	})
	if err != nil {
		t.Fatal(err)
	}

	want := []semanticToken{{line: 3, char: 1, length: 2, tokenType: proxy.legend.kinds[compiler.HighlightTag]}}
	if gotTokens := decodedTokens(t, got); !reflect.DeepEqual(gotTokens, want) {
		t.Fatalf("tokens = %+v, want %+v", gotTokens, want)
	}
}

func TestGohtTokensSplitMultiline(t *testing.T) {
	proxy := newTestServer(&recordingServer{}, &recordingClient{})
	src := "package main\n\n@slim Test() {\n\t/! first\n\t\tsecond\n\tp\n}\n"
	if err := proxy.DidOpen(context.Background(), didOpenParams(src)); err != nil {
		t.Fatal(err)
	}

	comment := proxy.legend.kinds[compiler.HighlightComment]
	var got []semanticToken
	for _, tok := range proxy.gohtTokens(testGohtURI) {
		if tok.tokenType == comment {
			got = append(got, tok)
		}
	}

	want := []semanticToken{
		{line: 3, char: 1, length: 8, tokenType: comment},
		{line: 4, char: 2, length: 6, tokenType: comment},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("comment tokens = %+v, want %+v", got, want)
	}
}

func TestGohtTokensUTF16(t *testing.T) {
	proxy := newTestServer(&recordingServer{}, &recordingClient{})
	src := "package main\n\n@slim Test() {\n\tp héllo\n\t/ é\n}\n"
	if err := proxy.DidOpen(context.Background(), didOpenParams(src)); err != nil {
		t.Fatal(err)
	}

	comment := proxy.legend.kinds[compiler.HighlightComment]
	for _, tok := range proxy.gohtTokens(testGohtURI) {
		if tok.tokenType == comment {
			if tok.line != 4 || tok.char != 1 || tok.length != 3 {
				t.Fatalf("comment token = %+v, want line 4 char 1 length 3", tok)
			}
			return
		}
	}
	t.Fatal("comment token not found")
}

func TestMergeOverlapGoplsWins(t *testing.T) {
	goht := []semanticToken{
		{line: 1, char: 0, length: 5, tokenType: 9},
		{line: 2, char: 0, length: 3, tokenType: 9},
	}
	gopls := []semanticToken{
		{line: 1, char: 3, length: 4, tokenType: 1},
		{line: 0, char: 0, length: 2, tokenType: 2},
	}

	got := mergeSemanticTokens(goht, gopls)

	want := []semanticToken{
		{line: 0, char: 0, length: 2, tokenType: 2},
		{line: 1, char: 3, length: 4, tokenType: 1},
		{line: 2, char: 0, length: 3, tokenType: 9},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("mergeSemanticTokens() = %+v, want %+v", got, want)
	}
}

func TestMapGoplsTokensNonASCIIInGoCode(t *testing.T) {
	proxy := newTestServer(&recordingServer{}, &recordingClient{})
	src := "package main\n\n@slim Test(x string) {\n\tp= fmt.Sprint(\"é\", x)\n}\n"
	line, col := openForMapping(t, proxy, src, "fmt.Sprint(")

	got := proxy.mapGoplsTokens(testGohtURI, []semanticToken{{line: line, char: col, length: 3, tokenType: 1}})

	// fmt starts at UTF-16 column 4 of "\tp= fmt.Sprint(...)"; a token that
	// cannot be placed exactly must be dropped, never shifted
	for _, tok := range got {
		if tok.line != 3 || tok.char != 4 || tok.length != 3 {
			t.Fatalf("mapGoplsTokens() = %+v, want fmt at 3:4 or no token", got)
		}
	}
}

func TestInitializeEnablesGoplsSemanticTokens(t *testing.T) {
	tests := map[string]struct {
		options any
		want    any
	}{
		"no options": {
			options: nil,
			want:    map[string]any{"semanticTokens": true},
		},
		"keeps other options": {
			options: map[string]any{"staticcheck": true},
			want:    map[string]any{"staticcheck": true, "semanticTokens": true},
		},
		"respects an explicit setting": {
			options: map[string]any{"semanticTokens": false},
			want:    map[string]any{"semanticTokens": false},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			server := &recordingServer{}
			proxy := newTestServer(server, &recordingClient{})
			params := &protocol.ParamInitialize{}
			params.InitializationOptions = tc.options

			if _, err := proxy.Initialize(context.Background(), params); err != nil {
				t.Fatal(err)
			}

			got := server.initializeCalls[0].InitializationOptions
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("forwarded InitializationOptions = %#v, want %#v", got, tc.want)
			}
		})
	}
}
