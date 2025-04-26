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
		if l.peek() == '\r' {
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
	l.acceptUntil(" (\n")
	if l.current() != "package" {
		return lexGoCode
	}
	l.ignore()
	l.skipRun(" ")

	l.acceptUntil("\n")
	if l.current() == "" {
		return l.errorf("package name expected")
	}
	l.emit(tPackage)
	return lexGoLineEnd
}

func lexImportStart(l *lexer) lexFn {
	l.acceptUntil(" \"(\n")
	if l.current() != "import" {
		return lexGoCode
	}
	l.skipRun(" ")
	l.ignore()

	switch l.peek() {
	case '(':
		l.skip()
		l.skipRun(" \t\n\r")
		return lexImports
	default:
		l.acceptUntil("\n\r")
		l.emit(tImport)
		l.skipRun("\n\r")
		return lexGoLineStart
	}
}

func lexImports(l *lexer) lexFn {
	for {
		l.skipRun(" \t\n\r")
		switch l.peek() {
		case ')':
			l.skipRun(")\n\r")
			return lexGoLineStart
		case scanner.EOF:
			return l.errorf("import expected")
		default:
			l.acceptUntil("\n\r")
			if l.current() == "" {
				return l.errorf("import expected")
			}
			l.emit(tImport)
		}
	}
}

func lexGoCode(l *lexer) lexFn {
	l.acceptUntil("\n\r")
	if l.current() != "" {
		l.emit(tGoCode)
	}
	return lexGoLineEnd
}

func lexTemplate(l *lexer) lexFn {
	l.acceptUntil(" ")
	switch l.current() {
	case "@goht", "@haml":
		return lexTemplateStart(l, lexHamlLineStart)
	case "@slim":
		return lexTemplateStart(l, lexSlimLineStart)
	}
	return nil
}

func lexTemplateStart(l *lexer, next lexFn) lexFn {
	l.ignore()
	l.skipRun(" ")
	l.acceptUntil(")")
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
		if r == quote && !escaping {
			break
		}
		escaping = false
		if r == '\\' {
			escaping = true
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

func continueToMatchingBrace(l *lexer, endBrace rune) rune {
	isEscaping := false
	inQuotes := false
	quotes := rune(0)

	for {
		r := l.next()
		if r == scanner.EOF {
			return scanner.EOF
		}

		if r == '"' || r == '\'' {
			if r == quotes && !isEscaping {
				inQuotes = !inQuotes
				quotes = rune(0)
			} else {
				inQuotes = true
				quotes = r
			}
			isEscaping = false
			continue
		}

		if r == endBrace && !inQuotes {
			return r
		}

		if r == '\\' && inQuotes {
			isEscaping = true
		}
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
			// if err != nil {
			// 	return l.errorf("unexpected error while evaluating indents: %s", err)
			// }
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

func acceptIndentedLines(indent int, next lexFn) lexFn {
	return func(l *lexer) lexFn {
		switch l.peek() {
		case '\n', '\r':
			l.next()
			return ignoreIndentedLines(indent, next)
		case ' ', '\t':
			priorIndents := l.peekAhead(indent)
			// if err != nil {
			// 	return l.errorf("unexpected error while evaluating indents: %s", err)
			// }
			if len(strings.TrimSpace(priorIndents)) != 0 {
				return next
			}
			// validate we have the correct indents
			if lexErr := l.validateIndent(priorIndents); lexErr != nil {
				return lexErr
			}
			l.acceptUntil("\n\r")
			return ignoreIndentedLines(indent, next)
		case scanner.EOF:
			l.emit(tEOF)
			return nil
		default:
			return next
		}
	}
}
