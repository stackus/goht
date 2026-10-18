package compiler

import (
	"bytes"
	"sort"
	"strings"
)

// HighlightKind classifies a span of GoHT source for editor highlighting.
type HighlightKind int

const (
	HighlightDirective HighlightKind = iota
	HighlightTag
	HighlightAttribute
	HighlightString
	HighlightComment
	HighlightKeyword
)

func (k HighlightKind) String() string {
	switch k {
	case HighlightDirective:
		return "directive"
	case HighlightTag:
		return "tag"
	case HighlightAttribute:
		return "attribute"
	case HighlightString:
		return "string"
	case HighlightComment:
		return "comment"
	case HighlightKeyword:
		return "keyword"
	default:
		return "unknown"
	}
}

// HighlightToken is a classified span of GoHT source. Start and End are byte
// offsets into the source; End is exclusive.
type HighlightToken struct {
	Kind  HighlightKind
	Start int
	End   int
}

// Highlight classifies the GoHT syntax in src. Go code is left unclassified so
// it can be classified by a Go language server. Incomplete source yields the
// tokens lexed before the first error. The result is sorted and has no
// overlapping tokens.
func Highlight(src []byte) []HighlightToken {
	l := newLexer(src)
	var tokens []HighlightToken
	for t := l.nextToken(); t.typ != tEOF && t.typ != tError; t = l.nextToken() {
		if tok, ok := highlightToken(src, t); ok {
			tokens = append(tokens, tok)
		}
	}
	for _, m := range l.marks {
		kind := HighlightDirective
		switch m.kind {
		case markDoctype:
			kind = HighlightKeyword
		case markComment:
			kind = HighlightComment
		}
		tokens = append(tokens, HighlightToken{Kind: kind, Start: m.start, End: m.end})
	}

	sort.SliceStable(tokens, func(i, j int) bool { return tokens[i].Start < tokens[j].Start })

	result := tokens[:0]
	prevEnd := 0
	for _, tok := range tokens {
		if tok.Start < prevEnd || tok.Start >= tok.End || tok.End > len(src) {
			continue
		}
		result = append(result, tok)
		prevEnd = tok.End
	}
	return result
}

// highlightToken narrows a lexer token's source span to the text worth
// classifying. Tokens whose literal cannot be found in their span are dropped
// rather than misplaced.
func highlightToken(src []byte, t token) (HighlightToken, bool) {
	if t.start < 0 || t.end > len(src) || t.start >= t.end {
		return HighlightToken{}, false
	}
	span := src[t.start:t.end]
	switch t.typ {
	case tTag, tVoidTag:
		return findLit(span, t, HighlightTag, "%", "")
	case tId:
		return findLit(span, t, HighlightAttribute, "#", "")
	case tClass:
		return findLit(span, t, HighlightAttribute, ".", "")
	case tAttrName:
		return findLit(span, t, HighlightAttribute, "", "")
	case tAttrEscapedValue:
		return findLit(span, t, HighlightString, "\"`", "\"`")
	case tComment, tRubyComment:
		return trimmedSpan(span, t.start, HighlightComment)
	case tDoctype:
		return trimmedSpan(span, t.start, HighlightKeyword)
	case tFilterStart:
		return findLit(span, t, HighlightDirective, ":", "")
	default:
		return HighlightToken{}, false
	}
}

// findLit locates the token literal within its span, extending it by one
// leading rune from prefixes and one trailing rune from suffixes when present.
func findLit(span []byte, t token, kind HighlightKind, prefixes, suffixes string) (HighlightToken, bool) {
	if t.lit == "" {
		return HighlightToken{}, false
	}
	i := bytes.Index(span, []byte(t.lit))
	if i < 0 {
		return HighlightToken{}, false
	}
	start, end := i, i+len(t.lit)
	if start > 0 && strings.IndexByte(prefixes, span[start-1]) >= 0 {
		start--
	}
	if end < len(span) && strings.IndexByte(suffixes, span[end]) >= 0 {
		end++
	}
	return HighlightToken{Kind: kind, Start: t.start + start, End: t.start + end}, true
}

func trimmedSpan(span []byte, offset int, kind HighlightKind) (HighlightToken, bool) {
	trimmed := bytes.TrimLeft(span, " \t\r\n")
	lead := len(span) - len(trimmed)
	trimmed = bytes.TrimRight(trimmed, " \t\r\n")
	if len(trimmed) == 0 {
		return HighlightToken{}, false
	}
	return HighlightToken{Kind: kind, Start: offset + lead, End: offset + lead + len(trimmed)}, true
}
