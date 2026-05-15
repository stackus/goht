package compiler

import (
	"strings"
	"text/scanner"
)

func lexGoLineStart(l *lexer) lexFn {
	switch l.peek() {
	case 'p':
		if l.current() != "" {
			l.emit(tGoCode)
		}
		return lexPackage
	case 'i':
		if l.current() != "" {
			l.emit(tGoCode)
		}
		return lexImportStart
	case '@':
		return lexTemplate
	case '\n', '\r', scanner.EOF:
		return lexGoLineEnd
	default:
		return lexGoCode
	}
}

func lexGoLineEnd(l *lexer) lexFn {
	switch l.peek() {
	case '\n', '\r':
		l.next()
		if l.current() == "\r" && l.peek() == '\n' {
			l.next()
		}
		l.emit(tNewLine)
		return lexGoLineStart
	case scanner.EOF:
		l.emit(tEOF)
		return nil
	default:
		return l.errorf("unexpected character: %q", l.peek())
	}
}

func lexPackage(l *lexer) lexFn {
	l.acceptUntil(" (\r\n")
	if l.current() != "package" {
		return lexGoCode
	}
	l.ignore()
	l.skipRun(" ")

	l.acceptUntil("\r\n")
	if l.current() == "" {
		return l.errorf("package name expected")
	}
	l.emit(tPackage)
	return lexGoLineEnd
}

func lexImportStart(l *lexer) lexFn {
	l.acceptUntil(" \"(\r\n")
	if l.current() != "import" {
		return lexGoCode
	}
	l.skipRun(" ")
	l.ignore()

	switch l.peek() {
	case '(':
		l.skip()
		l.skipRun(" \t\r\n")
		return lexImports
	default:
		l.acceptUntil("\r\n")
		l.emit(tImport)
		l.skipRun("\r\n")
		return lexGoLineStart
	}
}

func lexImports(l *lexer) lexFn {
	for {
		l.skipRun(" \t\r\n")
		switch l.peek() {
		case ')':
			l.skipRun(")\r\n")
			return lexGoLineStart
		case scanner.EOF:
			return l.errorf("import expected")
		default:
			l.acceptUntil("\r\n")
			if l.current() == "" {
				return l.errorf("import expected")
			}
			l.emit(tImport)
		}
	}
}

func lexGoCode(l *lexer) lexFn {
	l.acceptUntil("\r\n")
	if l.current() != "" {
		l.emit(tGoCode)
	}
	return lexGoLineEnd
}

func lexTemplate(l *lexer) lexFn {
	l.acceptUntil(" ")
	switch l.current() {
	case "@goht", "@haml":
		return lexTemplateStart(l, true, lexHamlLineStart)
	case "@slim":
		return lexTemplateStart(l, false, lexSlimLineStart)
	case "@ego":
		return lexTemplateStart(l, false, lexEgoStart)
	}
	return l.errorf("unknown template type: %q", l.current())
}

func lexTemplateStart(l *lexer, keepNewlines bool, next lexFn) lexFn {
	l.ignore()
	l.skipRun(" ")
	l.acceptUntil(")")

	// reset the indent
	l.indent = 0

	if strings.HasPrefix(l.current(), "(") {
		// we've only captured the receiver, so we need to capture the rest of the function signature
		l.next()
		for {
			l.acceptUntil(")")
			// handle the situation where the function signature contains an `interface{}` type with one or more methods
			openParens := strings.Count(l.current(), "(")
			closeParens := strings.Count(l.current(), ")")
			if openParens == closeParens+1 {
				break
			}
			l.next()
		}
	}
	l.next()
	l.emit(tTemplateStart)
	if keepNewlines {
		l.emit(tKeepNewlines)
	}
	l.skipRun(" {")
	l.skipRun("\n\r")

	return next
}

func continueToMatchingQuote(l *lexer, typ tokenType, captureQuotes bool) rune {
	quote := l.peek()
	if quote != '`' && quote != '"' {
		return quote
	}
	if captureQuotes {
		l.next()
	} else {
		l.skip()
	}
	escaping := false
	for {
		r := l.next()
		if r == scanner.EOF {
			return scanner.EOF
		}
		if escaping {
			escaping = false
			continue
		}
		if quote != '`' && r == '\\' {
			escaping = true
			continue
		}
		if r == quote && !escaping {
			break
		}
	}
	if captureQuotes {
		l.emit(typ)
	} else {
		l.backup()
		l.emit(typ)
		l.skip() // skip closing quote
	}
	return quote
}

func continueToMatchingBrace(l *lexer, endBrace rune, allowNewlines bool) rune {
	startBrace := matchingStartBrace(endBrace)
	depth := 1
	escaping := false
	quote := rune(0)

	for {
		r := l.next()
		if r == scanner.EOF {
			return scanner.EOF
		}

		if !allowNewlines && r == '\n' || r == '\r' {
			return scanner.EOF
		}

		if quote != 0 {
			if quote != '`' && escaping {
				escaping = false
				continue
			}
			if quote != '`' && r == '\\' {
				escaping = true
				continue
			}
			if r == quote {
				quote = 0
			}
			continue
		}

		switch r {
		case '"', '\'', '`':
			quote = r
		case startBrace:
			depth++
		case endBrace:
			depth--
			if depth == 0 {
				return r
			}
		}
	}
}

func matchingStartBrace(endBrace rune) rune {
	switch endBrace {
	case '}':
		return '{'
	case ']':
		return '['
	case ')':
		return '('
	default:
		return 0
	}
}

func ignoreIndentedLines(indent int, next lexFn) lexFn {
	return func(l *lexer) lexFn {
		switch l.peek() {
		case '\n', '\r':
			l.skip()
			return ignoreIndentedLines(indent, next)
		case ' ', '\t':
			priorIndents := l.peekAhead(indent)
			if len(strings.TrimSpace(priorIndents)) != 0 {
				return next
			}
			// validate we have the correct indents
			if lexErr := l.validateIndent(priorIndents); lexErr != nil {
				return lexErr
			}
			l.skipUntil("\n\r")
			return ignoreIndentedLines(indent, next)
		case scanner.EOF:
			l.emit(tEOF)
			return nil
		default:
			return next
		}
	}
}
