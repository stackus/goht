package compiler

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"text/scanner"
)

type lexFn func(*lexer) lexFn

type lexer struct {
	reader *bytes.Reader
	lex    lexFn
	tokens chan token
	s      string
	width  int
	pos    []int
	indent int
}

func newLexer(input []byte) *lexer {
	return &lexer{
		reader: bytes.NewReader(input),
		lex:    lexGoLineStart,
		tokens: make(chan token, 64),
		pos:    []int{0},
	}
}

func (l *lexer) nextToken() token {
	for {
		select {
		case t := <-l.tokens:
			return t
		default:
			if l.lex == nil {
				return token{typ: tEOF}
			}
			l.lex = l.lex(l)
		}
	}
}

// next consumes the next rune from the input.
func (l *lexer) next() rune {
	ch, size, err := l.reader.ReadRune()
	if err != nil {
		l.width = 0
		return scanner.EOF
	}

	l.pos[len(l.pos)-1]++
	if ch == '\n' {
		l.pos = append(l.pos, 0)
	}
	l.width = size
	l.s += string(ch)
	return ch
}

// backup steps back one rune.
func (l *lexer) backup() {
	if l.width == 0 {
		return
	}
	if l.pos[len(l.pos)-1] == 0 {
		l.pos = l.pos[:len(l.pos)-1]
	}
	l.pos[len(l.pos)-1]--

	_ = l.reader.UnreadRune()
	l.s = l.s[:len(l.s)-l.width]
}

// peek returns the next rune without consuming it.
func (l *lexer) peek() rune {
	ch := l.next()
	l.backup()
	return ch
}

// peekAhead returns the next length runes without consuming them.
func (l *lexer) peekAhead(length int) string {
	width := 0
	s := ""
	var err error
	for i := 0; i < length; i++ {
		var ch rune
		var size int
		ch, size, err = l.reader.ReadRune()
		if err != nil {
			break
		}
		width += size
		s += string(ch)
	}
	// if err != nil && err != io.EOF {
	// 	return "", fmt.Errorf("failed to peek ahead: %v", err)
	// }
	_, _ = l.reader.Seek(int64(-width), io.SeekCurrent)
	// if err != nil {
	// 	return "", fmt.Errorf("failed to seek back: %v", err)
	// }
	return s
}

// ignore discards the current captured string.
func (l *lexer) ignore() {
	l.s = ""
}

// accept consumes the next rune if it's contained in the acceptRunes list.
func (l *lexer) accept(acceptRunes string) bool {
	if strings.ContainsRune(acceptRunes, l.next()) {
		return true
	}
	l.backup()
	return false
}

// acceptRun consumes a run of runes from the acceptRunes list.
func (l *lexer) acceptRun(acceptRunes string) {
	for strings.ContainsRune(acceptRunes, l.next()) {
	}
	l.backup()
}

// acceptUntil consumes runes until it encounters a rune in the stopRunes list.
func (l *lexer) acceptUntil(stopRunes string) {
	for r := l.next(); !strings.ContainsRune(stopRunes, r) && r != scanner.EOF; r = l.next() {
	}
	l.backup()
}

// acceptAhead consumes the next length runes.
func (l *lexer) acceptAhead(length int) {
	for i := 0; i < length; i++ {
		l.next()
	}
}

// skip discards the next rune.
func (l *lexer) skip() rune {
	r := l.next()
	l.s = l.s[:len(l.s)-1]
	return r
}

// skipRun discards a contiguous run of runes from the skipRunes list.
func (l *lexer) skipRun(skipRunes string) {
	for strings.ContainsRune(skipRunes, l.next()) {
		l.s = l.s[:len(l.s)-1]
	}
	l.backup()
}

// skipUntil discards runes until it encounters a rune in the stopRunes list.
func (l *lexer) skipUntil(stopRunes string) {
	for r := l.next(); !strings.ContainsRune(stopRunes, r) && r != scanner.EOF; r = l.next() {
		l.s = l.s[:len(l.s)-1]
	}
	l.backup()
}

// skipAhead consumes the next length runes and discards them.
func (l *lexer) skipAhead(length int) {
	for i := 0; i < length; i++ {
		l.next()
	}
	l.ignore()
}

// current returns the current captured string being built by the lexer.
func (l *lexer) current() string {
	return l.s
}

// emit creates a new token with the current string and sends it to the tokens channel.
func (l *lexer) emit(t tokenType) {
	line, col := l.position()
	l.tokens <- token{typ: t, lit: l.s, line: line, col: col}
	l.s = ""
}

// errorf creates a new error token with the formatted message and sends it to the tokens channel.
func (l *lexer) errorf(format string, args ...any) lexFn {
	line, col := l.position()
	l.tokens <- token{typ: tError, lit: fmt.Sprintf(format, args...), line: line, col: col}
	return func(l *lexer) lexFn { return nil }
}

// position returns the current line and column of the content being lexed.
func (l *lexer) position() (int, int) {
	newLinesInString := strings.Count(l.s, "\n")
	line := len(l.pos) - newLinesInString
	column := 1 + (l.pos[line-1]) - len(l.s)
	return line, column
}

func (l *lexer) validateIndent(indent string) lexFn {
	if indent == "" {
		return nil
	}
	// validate the indent against the sequence and char
	currentLen := len(indent)

	// if the indent is less than or equal to the current indent, return
	if currentLen <= l.indent {
		return nil
	}

	// require tabs for indenting; report the use of spaces as an error
	if strings.Contains(indent, " ") {
		return l.errorf("the line was indented using spaces, templates must be indented using tabs")
	}

	if currentLen > l.indent+1 {
		return l.errorf("the line was indented %d levels deeper than the previous line", currentLen-l.indent)
	}

	return nil
}
