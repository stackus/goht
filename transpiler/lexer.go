package transpiler

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

func (l *lexer) skip() rune {
	r := l.next()
	l.ignore()
	return r
}

func (l *lexer) skipRun(valid string) {
	for strings.ContainsRune(valid, l.next()) {
		l.ignore()
	}
	l.backup()
}

func (l *lexer) skipUntil(invalid string) {
	for r := l.next(); !strings.ContainsRune(invalid, r) && r != scanner.EOF; r = l.next() {
		l.ignore()
	}
	l.backup()
}

func (l *lexer) current() string {
	return l.s
}

func (l *lexer) emit(t tokenType) {
	line, col := l.position()
	tt := token{typ: t, lit: l.s, line: line, col: col}
	// fmt.Printf("emit: %v\n", tt)
	l.tokens <- tt
	l.s = ""
}

func (l *lexer) errorf(format string, args ...any) lexFn {
	line, col := l.position()
	l.tokens <- token{typ: tError, lit: fmt.Sprintf(format, args...), line: line, col: col}
	return nil
}

func (l *lexer) position() (int, int) {
	newLinesInString := strings.Count(l.s, "\n")
	line := len(l.pos) - newLinesInString
	column := 1 + (l.pos[line-1]) - len(l.s)

	return line, column
}
