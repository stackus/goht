package proxy

import (
	"context"
	"encoding/json"
	"maps"
	"slices"
	"sort"
	"strings"
	"unicode/utf16"

	"github.com/rs/zerolog"
	"github.com/stackus/protocol"

	"github.com/stackus/goht/compiler"
)

// gohtTokenTypes are the semantic token types GoHT adds to the gopls legend.
var gohtTokenTypes = []string{"gohtDirective", "gohtTag", "gohtAttribute"}

// gohtStandardTypes are the standard semantic token types GoHT emits; each is
// appended to the legend only when gopls does not already provide it.
var gohtStandardTypes = map[compiler.HighlightKind]string{
	compiler.HighlightString:  "string",
	compiler.HighlightComment: "comment",
	compiler.HighlightKeyword: "keyword",
}

type tokenLegend struct {
	legend protocol.SemanticTokensLegend
	kinds  map[compiler.HighlightKind]uint32
}

// buildLegend keeps the gopls token types and modifiers at their original
// indices so gopls tokens pass through unchanged, then appends the GoHT types.
func buildLegend(gopls *protocol.SemanticTokensLegend) tokenLegend {
	types := []string{}
	modifiers := []string{}
	if gopls != nil {
		types = append(types, gopls.TokenTypes...)
		modifiers = append(modifiers, gopls.TokenModifiers...)
	}
	index := func(name string) uint32 {
		if i := slices.Index(types, name); i >= 0 {
			return uint32(i)
		}
		types = append(types, name)
		return uint32(len(types) - 1)
	}
	kinds := map[compiler.HighlightKind]uint32{
		compiler.HighlightDirective: index(gohtTokenTypes[0]),
		compiler.HighlightTag:       index(gohtTokenTypes[1]),
		compiler.HighlightAttribute: index(gohtTokenTypes[2]),
	}
	for _, kind := range []compiler.HighlightKind{compiler.HighlightString, compiler.HighlightComment, compiler.HighlightKeyword} {
		kinds[kind] = index(gohtStandardTypes[kind])
	}
	return tokenLegend{
		legend: protocol.SemanticTokensLegend{TokenTypes: types, TokenModifiers: modifiers},
		kinds:  kinds,
	}
}

// withGoplsSemanticTokens turns on the gopls semanticTokens setting, which is
// off by default, unless the client set it explicitly.
func withGoplsSemanticTokens(options any) any {
	opts := map[string]any{}
	switch o := options.(type) {
	case nil:
	case map[string]any:
		if _, ok := o["semanticTokens"]; ok {
			return options
		}
		maps.Copy(opts, o)
	default:
		return options
	}
	opts["semanticTokens"] = true
	return opts
}

// goplsSemanticLegend extracts the legend from the semantic tokens capability
// gopls advertised, which arrives as a generic JSON value.
func goplsSemanticLegend(provider any) (*protocol.SemanticTokensLegend, bool) {
	if provider == nil {
		return nil, false
	}
	data, err := json.Marshal(provider)
	if err != nil {
		return nil, false
	}
	var opts protocol.SemanticTokensOptions
	if err := json.Unmarshal(data, &opts); err != nil || len(opts.Legend.TokenTypes) == 0 {
		return nil, false
	}
	return &opts.Legend, true
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
	if _, ok := s.smc.Get(string(uri)); !ok {
		return nil
	}
	mapped := s.mapGoplsTokens(uri, decodeSemanticTokens(tokens.Data))
	return &protocol.SemanticTokens{Data: encodeSemanticTokens(mapped)}
}

// mapGoplsTokens translates gopls tokens on the generated Go source into GoHT
// document positions. Tokens are dropped unless their whole span maps onto a
// single contiguous GoHT range, so generated-only code never produces tokens.
//
// Columns are converted through the generated Go and GoHT line text: gopls
// columns use the negotiated position encoding, the source map uses byte
// columns for Go and rune columns for GoHT.
func (s *Server) mapGoplsTokens(uri protocol.DocumentURI, goTokens []semanticToken) []semanticToken {
	sm, ok := s.smc.Get(string(uri))
	if !ok {
		return nil
	}
	goLines := strings.Split(s.goSrcs[string(uri)], "\n")
	var gohtLines []string
	if doc, ok := s.srcs.Get(string(uri)); ok {
		gohtLines = doc.lines
	}
	lineText := func(lines []string, line int) string {
		if line < len(lines) {
			return lines[line]
		}
		return ""
	}

	mapped := make([]semanticToken, 0, len(goTokens))
	for _, t := range goTokens {
		goLine := lineText(goLines, int(t.line))
		goStart := s.toByteCol(goLine, int(t.char))
		goEnd := s.toByteCol(goLine, int(t.char+t.length))
		start, ok := sm.SourcePositionFromTarget(int(t.line), goStart)
		if !ok {
			continue
		}
		end, ok := sm.SourcePositionFromTarget(int(t.line), goEnd)
		if !ok || end.Line != start.Line || end.Col-start.Col != goEnd-goStart {
			continue
		}
		gohtLine := lineText(gohtLines, start.Line)
		byteStart := runeColToByte(gohtLine, start.Col)
		byteEnd := byteStart + (goEnd - goStart)
		// the source map can drift on lines with non-ASCII Go code; drop
		// tokens whose text differs rather than misplace them
		if gohtLine != "" && goLine != "" &&
			(byteEnd > len(gohtLine) || goEnd > len(goLine) || gohtLine[byteStart:byteEnd] != goLine[goStart:goEnd]) {
			continue
		}
		char := s.fromByteCol(gohtLine, byteStart)
		mapped = append(mapped, semanticToken{
			line:           uint32(start.Line),
			char:           uint32(char),
			length:         uint32(s.fromByteCol(gohtLine, byteEnd) - char),
			tokenType:      t.tokenType,
			tokenModifiers: t.tokenModifiers,
		})
	}
	sortSemanticTokens(mapped)
	return mapped
}

func sortSemanticTokens(tokens []semanticToken) {
	sort.SliceStable(tokens, func(i, j int) bool {
		if tokens[i].line != tokens[j].line {
			return tokens[i].line < tokens[j].line
		}
		return tokens[i].char < tokens[j].char
	})
}

// toByteCol converts a column in the negotiated position encoding to a byte column.
func (s *Server) toByteCol(line string, col int) int {
	if s.positionEncoding == protocol.UTF8 {
		return col
	}
	return utf16ColToByte(line, col)
}

// fromByteCol converts a byte column to the negotiated position encoding.
func (s *Server) fromByteCol(line string, col int) int {
	if s.positionEncoding == protocol.UTF8 {
		return col
	}
	return byteColToUTF16(line, col)
}

// byteColToUTF16 converts a byte column to a UTF-16 column. Columns past the
// end of the line are treated as ASCII.
func byteColToUTF16(line string, col int) int {
	if col > len(line) {
		return byteColToUTF16(line, len(line)) + col - len(line)
	}
	n := 0
	for _, r := range line[:col] {
		n += len(utf16.Encode([]rune{r}))
	}
	return n
}

// utf16ColToByte converts a UTF-16 column to a byte column. Columns past the
// end of the line are treated as ASCII.
func utf16ColToByte(line string, col int) int {
	n := 0
	for i, r := range line {
		if n >= col {
			return i
		}
		n += len(utf16.Encode([]rune{r}))
	}
	return len(line) + col - n
}

// runeColToByte converts a rune column to a byte column. Columns past the end
// of the line are treated as ASCII.
func runeColToByte(line string, col int) int {
	n := 0
	for i := range line {
		if n == col {
			return i
		}
		n++
	}
	return len(line) + col - n
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

// semanticTokens returns the merged GoHT and gopls tokens for a document.
// gopls tokens are skipped when gopls lacks semantic tokens, the source map is
// stale, or the gopls request fails; GoHT tokens are always returned.
func (s *Server) semanticTokens(ctx context.Context, uri protocol.DocumentURI, logger zerolog.Logger) []semanticToken {
	goht := s.gohtTokens(uri)
	if !s.goplsSemantic || s.smc.IsStale(string(uri)) {
		return goht
	}
	if _, ok := s.smc.Get(string(uri)); !ok {
		return goht
	}
	_, goURI := toGohtGoURI(uri)
	resp, err := s.Server.SemanticTokensFull(ctx, &protocol.SemanticTokensParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: goURI},
	})
	if err != nil {
		logger.Error().Err(err).Msg("unable to perform gopls semantic tokens full")
		return goht
	}
	if resp == nil {
		return goht
	}
	return mergeSemanticTokens(goht, s.mapGoplsTokens(uri, decodeSemanticTokens(resp.Data)))
}

// gohtTokens classifies the GoHT syntax of the current document text. Tokens
// spanning several lines are split per line, because clients need not support
// multiline tokens.
func (s *Server) gohtTokens(uri protocol.DocumentURI) []semanticToken {
	doc, ok := s.srcs.Get(string(uri))
	if !ok {
		return nil
	}
	src := doc.String()
	lineStarts := []int{0}
	for i := 0; i < len(src); i++ {
		if src[i] == '\n' {
			lineStarts = append(lineStarts, i+1)
		}
	}

	var tokens []semanticToken
	for _, h := range compiler.Highlight([]byte(src)) {
		tokenType, ok := s.legend.kinds[h.Kind]
		if !ok {
			continue
		}
		line := sort.SearchInts(lineStarts, h.Start+1) - 1
		for start := h.Start; start < h.End && line < len(lineStarts); line++ {
			lineStart := lineStarts[line]
			lineEnd := len(src)
			if line+1 < len(lineStarts) {
				lineEnd = lineStarts[line+1] - 1
			}
			end := min(h.End, lineEnd)
			text := src[lineStart:lineEnd]
			startCol := start - lineStart
			endCol := end - lineStart
			// skip the indentation of continuation lines
			for startCol < endCol && (text[startCol] == ' ' || text[startCol] == '\t') {
				startCol++
			}
			endCol = startCol + len(strings.TrimRight(text[startCol:endCol], " \t\r"))
			if endCol > startCol {
				char := s.fromByteCol(text, startCol)
				tokens = append(tokens, semanticToken{
					line:      uint32(line),
					char:      uint32(char),
					length:    uint32(s.fromByteCol(text, endCol) - char),
					tokenType: tokenType,
				})
			}
			start = lineEnd + 1
		}
	}
	return tokens
}

// mergeSemanticTokens combines GoHT and gopls tokens into one sorted,
// non-overlapping list. gopls owns embedded Go, so it wins any overlap.
func mergeSemanticTokens(goht, gopls []semanticToken) []semanticToken {
	merged := make([]semanticToken, 0, len(goht)+len(gopls))
	merged = append(merged, gopls...)
	sortSemanticTokens(merged)
	merged = dropOverlaps(merged)
	for _, t := range goht {
		if !overlapsAny(t, merged) {
			merged = append(merged, t)
		}
	}
	sortSemanticTokens(merged)
	return dropOverlaps(merged)
}

func overlaps(a, b semanticToken) bool {
	return a.line == b.line && a.char < b.char+b.length && b.char < a.char+a.length
}

func overlapsAny(t semanticToken, tokens []semanticToken) bool {
	i := sort.Search(len(tokens), func(i int) bool { return tokens[i].line >= t.line })
	for ; i < len(tokens) && tokens[i].line == t.line; i++ {
		if overlaps(t, tokens[i]) {
			return true
		}
	}
	return false
}

// dropOverlaps keeps the first of any overlapping tokens in a sorted list.
func dropOverlaps(tokens []semanticToken) []semanticToken {
	result := tokens[:0]
	for _, t := range tokens {
		if n := len(result); n > 0 && overlaps(result[n-1], t) {
			continue
		}
		result = append(result, t)
	}
	return result
}

// filterTokensToRange keeps the tokens that intersect r.
func filterTokensToRange(tokens []semanticToken, r protocol.Range) []semanticToken {
	result := make([]semanticToken, 0, len(tokens))
	for _, t := range tokens {
		if t.line < r.Start.Line || t.line > r.End.Line {
			continue
		}
		if t.line == r.Start.Line && t.char+t.length <= r.Start.Character {
			continue
		}
		if t.line == r.End.Line && t.char >= r.End.Character {
			continue
		}
		result = append(result, t)
	}
	return result
}
