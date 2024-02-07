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
	reader     *bytes.Reader
	lex        lexFn
	tokens     chan token
	s          string
	width      int
	pos        []int
	indentChar rune
	indentLen  int
	indent     int
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

func (l *lexer) peek() rune {
	ch := l.next()
	l.backup()
	return ch
}

func (l *lexer) peekAhead(length int) (string, error) {
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
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("failed to peek ahead: %v", err)
	}
	_, err = l.reader.Seek(int64(-width), io.SeekCurrent)
	if err != nil {
		return "", fmt.Errorf("failed to seek back: %v", err)
	}
	return s, nil
}

func (l *lexer) ignore() {
	l.s = ""
}

func (l *lexer) accept(valid string) bool {
	if strings.ContainsRune(valid, l.next()) {
		return true
	}
	l.backup()
	return false
}

func (l *lexer) acceptRun(valid string) {
	for strings.ContainsRune(valid, l.next()) {
	}
	l.backup()
}

func (l *lexer) acceptUntil(invalid string) {
	for r := l.next(); !strings.ContainsRune(invalid, r) && r != scanner.EOF; r = l.next() {
	}
	l.backup()
}

func (l *lexer) acceptAhead(length int) {
	for i := 0; i < length; i++ {
		l.next()
	}
}

func (l *lexer) skip() rune {
	r := l.next()
	l.s = l.s[:len(l.s)-1]
	return r
}

func (l *lexer) skipRun(skipList string) {
	for strings.ContainsRune(skipList, l.next()) {
		l.s = l.s[:len(l.s)-1]
	}
	l.backup()
}

func (l *lexer) skipUntil(stopList string) {
	for r := l.next(); !strings.ContainsRune(stopList, r) && r != scanner.EOF; r = l.next() {
		l.s = l.s[:len(l.s)-1]
	}
	l.backup()
}

func (l *lexer) current() string {
	return l.s
}

func (l *lexer) emit(t tokenType) {
	line, col := l.position()
	l.tokens <- token{typ: t, lit: l.s, line: line, col: col}
	l.s = ""
}

func (l *lexer) errorf(format string, args ...any) lexFn {
	line, col := l.position()
	l.tokens <- token{typ: tError, lit: fmt.Sprintf(format, args...), line: line, col: col}
	return func(l *lexer) lexFn { return nil }
}

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
	isTabs := strings.Contains(indent, "\t")
	isSpaces := strings.Contains(indent, " ")
	if (isTabs && l.indentChar == ' ') || (isSpaces && l.indentChar == '\t') || (currentLen%l.indentLen != 0) || (currentLen/l.indentLen > l.indent+1) {
		if depth := currentLen / l.indentLen; !(isTabs && isSpaces) && depth > l.indent+1 {
			return l.errorf("the line was indented %d levels deeper than the previous line", depth-l.indent)
		}
		var used string
		want := "space(s)"
		if l.indentChar == '\t' {
			want = "tab(s)"
		}
		wanted := fmt.Sprintf("%d %s", l.indentLen, want)
		if isTabs && isSpaces {
			used = fmt.Sprintf("%q", indent)
		} else {
			got := "space(s)"
			if isTabs {
				got = "tab(s)"
			}
			used = fmt.Sprintf("%d %s", currentLen, got)
		}
		return l.errorf("inconsistent indentation: %s used for indentation, but the rest of the template was indented using %s", used, wanted)
	}
	return nil
}
