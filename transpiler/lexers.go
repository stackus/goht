package transpiler

import (
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
	}
	return l.errorf("unexpected character: %q", l.peek())
}

func lexPackage(l *lexer) lexFn {
	l.acceptUntil(" (\n")
	if l.current() != "package" {
		return lexGoCode
	}
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
	}
	return lexGoLineEnd
}

func lexImports(l *lexer) lexFn {
	for {
		l.skipRun(" \t\n\r")
		switch l.peek() {
		case ')':
			l.skip()
			return lexGoLineEnd
		case scanner.EOF:
			return l.errorf("import expected")
		default:
			l.acceptUntil(" \t\n\r")
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
	case "@hmlt":
		return lexHmltStart
	}
	return nil
}

func lexHmltStart(l *lexer) lexFn {
	l.skipRun(" ")
	l.acceptUntil(")")
	l.next()
	l.emit(tHmltStart)
	l.skipRun(" {")

	return lexHmltLineEnd
}

func lexHmltLineStart(l *lexer) lexFn {
	switch l.peek() {
	case '}':
		l.emit(tHmltEnd)
		l.skip()
		return lexGoLineStart
	case scanner.EOF:
		l.emit(tEOF)
		return nil
	case '\n', '\r':
		return lexHmltLineEnd
	default:
		return lexHmltIndent
	}
}

func lexHmltIndent(l *lexer) lexFn {
	l.acceptRun(" \t")
	l.emit(tIndent)
	return lexHmltContentStart
}

func lexHmltContentStart(l *lexer) lexFn {
	switch l.peek() {
	case '%':
		return lexHmltTag
	case '#':
		return lexHmltId
	case '.':
		return lexHmltClass
	case '\\':
		l.skip()
		return lexHmltTextStart
	case '!':
		if s, err := l.peekAhead(3); err != nil {
			return l.errorf("unexpected error: %s", err)
		} else if s == "!!!" {
			// TODO return an error if we're nesting doctypes
			return lexHmltDoctype
		}
		return lexHmltUnescaped
	case '-':
		return lexHmltExecuteCode
	case '=':
		return lexHmltOutputCode
	case scanner.EOF, '\n', '\r':
		return lexHmltLineEnd
	default:
		return lexHmltTextStart
	}
}

func lexHmltContent(l *lexer) lexFn {
	switch l.peek() {
	case '#':
		return lexHmltId
	case '.':
		return lexHmltClass
	case '{':
		return lexHmltAttributesStart
	case '!':
		return lexHmltUnescaped
	case '-':
		return lexHmltExecuteCode
	case '=':
		return lexHmltOutputCode
	case scanner.EOF, '\n', '\r':
		return lexHmltLineEnd
	default:
		return lexHmltTextStart
	}
}

func lexHmltLineEnd(l *lexer) lexFn {
	l.skipRun(" \t")

	switch l.peek() {
	case '\n', '\r':
		return lexHmltNewLine
	case scanner.EOF:
		l.emit(tEOF)
		return nil
	}

	return l.errorf("unexpected character: %q", l.peek())
}

func lexHmltNewLine(l *lexer) lexFn {
	l.acceptRun("\n\r")
	l.emit(tNewLine)
	return lexHmltLineStart
}

func hamlIdentifier(typ tokenType, l *lexer) lexFn {
	l.skip() // eat symbol

	// these characters may follow an identifier
	const mayFollowIdentifier = "%#.{=! \t\n\r"

	l.acceptUntil(mayFollowIdentifier)
	if l.current() == "" {
		return l.errorf("%s identifier expected", typ)
	}
	l.emit(typ)
	return lexHmltContent
}

func lexHmltTag(l *lexer) lexFn {
	return hamlIdentifier(tTag, l)
}

func lexHmltId(l *lexer) lexFn {
	return hamlIdentifier(tId, l)
}

func lexHmltClass(l *lexer) lexFn {
	return hamlIdentifier(tClass, l)
}

func lexHmltAttributesStart(l *lexer) lexFn {
	l.skip()
	return lexHmltAttribute
}

func lexHmltAttributesEnd(l *lexer) lexFn {
	l.skip()
	return lexHmltContent
}

func lexHmltAttribute(l *lexer) lexFn {
	// supported attributes
	// key
	// key:value
	// key?:value
	// @attributes: []any (string, map[string]string, map[string]bool)
	// @classes: []any (string, map[string]bool)
	// @styles: []any (string, map[string]string)

	l.skipRun(", \t\n\r")

	switch l.peek() {
	case '}':
		return lexHmltAttributesEnd
	case '@':
		return lexAttributeCommandStart
	default:
		return lexHmltAttributeName
	}
}

func lexHmltAttributeName(l *lexer) lexFn {
	if l.peek() == '"' || l.peek() == '`' {
		r := continueToMatchingQuote(l, tAttrName, false)
		if r == scanner.EOF {
			return l.errorf("attribute name not closed: eof")
		} else if r != '"' && r != '`' {
			return l.errorf("unexpected character: %q", r)
		}
	} else {
		l.acceptUntil("?:,}{\" \t\n\r")
		if l.current() == "" {
			return l.errorf("attribute name expected")
		}
		l.emit(tAttrName)
	}

	l.skipRun(" \t\n\r")
	switch l.peek() {
	case '?', ':':
		return lexHmltAttributeOperator
	case ',', '}':
		return lexHmltAttributeEnd
	default:
		return l.errorf("unexpected character: %q", l.peek())
	}
}

func lexHmltAttributeOperator(l *lexer) lexFn {
	l.skipRun(" \t\n\r")
	switch l.peek() {
	case '?':
		l.next()
		if l.peek() == ':' {
			l.next()
			l.emit(tAttrOperator)
			return lexHmltAttributeValue
		}
		return l.errorf("unexpected character: %q", l.peek())
	case ':':
		l.next()
		l.emit(tAttrOperator)
		return lexHmltAttributeValue
	}
	return l.errorf("unexpected character: %q", l.peek())
}

func lexHmltAttributeValue(l *lexer) lexFn {
	l.skipRun(" \t\n\r")

	switch l.peek() {
	case '"', '`':
		return lexHmltAttributeStaticValue
	case '{':
		return lexHmltAttributeDynamicValue
	}
	return l.errorf("unexpected character: %q", l.peek())
}

func lexHmltAttributeStaticValue(l *lexer) lexFn {
	r := continueToMatchingQuote(l, tAttrStaticValue, true)
	if r == scanner.EOF {
		return l.errorf("attribute value not closed: eof")
	} else if r != '"' && r != '`' {
		return l.errorf("unexpected character: %q", r)
	}
	return lexHmltAttributeEnd
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

func lexHmltAttributeDynamicValue(l *lexer) lexFn {
	// match opening curly brace
	// advance until closing curly brace
	// skip over any escaped curly braces
	// emit value
	l.skip() // skip opening brace
	r := continueToMatchingBrace(l)
	if r == scanner.EOF {
		return l.errorf("attribute value not closed: eof")
	}
	l.backup()
	l.emit(tAttrDynamicValue)
	l.skip() // skip closing brace
	return lexHmltAttributeEnd
}

func continueToMatchingBrace(l *lexer) rune {
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

		if r == '}' && !inQuotes {
			return r
		}

		if r == '\\' && inQuotes {
			isEscaping = true
		}
	}
}

func lexAttributeCommandStart(l *lexer) lexFn {
	l.skipRun("@")
	l.acceptUntil(": \t\n\r")
	if l.current() == "" {
		return l.errorf("command code expected")
	}
	switch l.current() {
	case "attributes":
		return lexHmltAttributeCommand(tAttributesCommand)
	default:
		return l.errorf("unknown attribute command: %s", l.current())
	}
}

func lexHmltAttributeCommand(command tokenType) lexFn {
	return func(l *lexer) lexFn {
		l.skipUntil(":")
		l.skipUntil("{")
		l.skip() // skip opening brace
		r := continueToMatchingBrace(l)
		if r == scanner.EOF {
			return l.errorf("attribute value not closed: eof")
		}
		l.backup()
		l.emit(command)
		l.skip() // skip closing brace

		return lexHmltAttributeEnd
	}
}

func lexHmltAttributeEnd(l *lexer) lexFn {
	l.skipRun(" \t\n\r")
	switch l.peek() {
	case ',':
		l.skip()
		return lexHmltAttribute
	case '}':
		return lexHmltAttributesEnd
	default:
		return l.errorf("unexpected character: %c", l.peek())
	}
}

func lexHmltTextStart(l *lexer) lexFn {
	l.skipRun(" \t")
	return lexHmltTextContent
}

func lexHmltTextContent(l *lexer) lexFn {
	l.acceptUntil("#\n\r")
	if l.peek() == '#' {
		return lexHmltDynamicText
	}
	if l.current() != "" {
		l.emit(tStaticText)
	}
	return lexHmltLineEnd
}

func lexHmltDynamicText(l *lexer) lexFn {
	if s, err := l.peekAhead(2); err != nil {
		return l.errorf("unexpected error: %s", err)
	} else if s != "#{" {
		l.next()
		return lexHmltTextContent
	}
	if l.current() != "" {
		l.emit(tStaticText)
	}
	l.skipRun("#{")
	r := continueToMatchingBrace(l)
	if r == scanner.EOF {
		return l.errorf("dynamic text value was not closed: eof")
	}
	l.backup()
	l.emit(tDynamicText)
	l.skip() // skip closing brace
	return lexHmltTextContent
}

func lexHmltDoctype(l *lexer) lexFn {
	l.skipRun("! ")
	l.acceptUntil("\n\r")
	l.emit(tDoctype)
	return lexHmltLineEnd
}

func lexHmltUnescaped(l *lexer) lexFn {
	l.skip()
	l.ignore()
	l.emit(tUnescaped)
	switch l.peek() {
	case '=':
		return lexHmltOutputCode
	default:
		return lexHmltTextStart
	}
}

func lexHmltExecuteCode(l *lexer) lexFn {
	l.skipRun("- \t")
	l.acceptUntil("\n\r")
	l.emit(tExecuteCode)
	return lexHmltLineEnd
}

func lexHmltOutputCode(l *lexer) lexFn {
	l.skipRun("= \t")
	switch l.peek() {
	case '@':
		return lexHmltCommandCode
	default:
		l.acceptUntil("\n\r")
		l.emit(tOutputCode)
		return lexHmltLineEnd
	}
}

func lexHmltCommandCode(l *lexer) lexFn {
	l.skipRun("@")
	l.acceptUntil("() \t\n\r")
	if l.current() == "" {
		return l.errorf("command code expected")
	}
	switch l.current() {
	case "render":
		l.acceptRun("() \t")
		l.ignore()
		l.acceptUntil("\n\r")
		if l.current() == "" {
			return l.errorf("render argument expected")
		}
		l.emit(tRenderCommand)
	case "children":
		l.acceptRun("() \t")
		l.ignore()
		l.acceptUntil("\n\r")
		if l.current() != "" {
			return l.errorf("children command does not accept arguments")
		}
		l.emit(tChildrenCommand)
	}
	return lexHmltLineEnd
}
