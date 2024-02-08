package compiler

import (
	"slices"
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
	case "@goht":
		return lexGohtStart
	}
	return nil
}

func lexGohtStart(l *lexer) lexFn {
	l.ignore()
	l.skipRun(" ")
	l.acceptUntil(")")
	l.next()
	l.emit(tGohtStart)
	l.skipRun(" {")
	l.skipRun("\n\r")

	l.accept(" \t")
	if l.current() != "" {
		return l.errorf("indenting at the beginning of the template is illegal")
	}

	return lexGohtLineStart
}

func lexGohtLineStart(l *lexer) lexFn {
	switch l.peek() {
	case '}':
		l.emit(tGohtEnd)
		l.skip()
		return lexGoLineStart
	case scanner.EOF:
		l.emit(tEOF)
		return nil
	case '\n', '\r':
		return lexGohtLineEnd
	default:
		return lexGohtIndent
	}
}

func lexGohtIndent(l *lexer) lexFn {
	l.acceptRun(" \t")
	indent := l.current()

	if len(indent) == 0 {
		l.indent = 0
		l.emit(tIndent)
		return lexGohtContentStart
	}

	// set indent char and length
	if l.indentChar == 0 {
		if strings.Contains(indent, " ") && strings.Contains(indent, "\t") {
			return l.errorf("indentation cannot contain both spaces and tabs")
		}
		l.indentChar = ' '
		if strings.Contains(indent, "\t") {
			l.indentChar = '\t'
		}
		l.indentLen = len(indent)
	}

	// validate the indent against the sequence and char
	if lexErr := l.validateIndent(indent); lexErr != nil {
		return lexErr
	}

	l.indent = len(l.current()) / l.indentLen // useful for parsing filters
	l.emit(tIndent)
	return lexGohtContentStart
}

func lexGohtContentStart(l *lexer) lexFn {
	switch l.peek() {
	case '%':
		return lexGohtTag
	case '#':
		return lexGohtId
	case '.':
		return lexGohtClass
	case '\\':
		l.skip()
		return lexGohtTextStart
	case '!':
		if s, err := l.peekAhead(3); err != nil {
			return l.errorf("unexpected error: %s", err)
		} else if s == "!!!" {
			// TODO return an error if we're nesting doctypes
			return lexGohtDoctype
		}
		return lexGohtUnescaped
	case '-':
		return lexGohtSilentScript
	case '=':
		return lexGohtOutputCode
	case '/':
		return lexComment
	case ':':
		return lexFilterStart
	case scanner.EOF, '\n', '\r':
		return lexGohtLineEnd
	default:
		return lexGohtTextStart
	}
}

func lexGohtContent(l *lexer) lexFn {
	switch l.peek() {
	case '#':
		return lexGohtId
	case '.':
		return lexGohtClass
	case '[':
		return lexObjectReference
	case '{':
		return lexGohtAttributesStart
	case '!':
		return lexGohtUnescaped
	case '-':
		return lexGohtSilentScript
	case '=':
		return lexGohtOutputCode
	case '/':
		return lexVoidTag
	case '>', '<':
		return lexWhitespaceRemoval
	case scanner.EOF, '\n', '\r':
		return lexGohtLineEnd
	default:
		return lexGohtTextStart
	}
}

func lexGohtContentEnd(l *lexer) lexFn {
	switch l.peek() {
	case '=':
		return lexGohtOutputCode
	case '/':
		return lexVoidTag
	case '>', '<':
		return lexWhitespaceRemoval
	case scanner.EOF, '\n', '\r':
		return lexGohtLineEnd
	default:
		return lexGohtTextStart
	}
}

func lexGohtLineEnd(l *lexer) lexFn {
	l.skipRun(" \t")

	switch l.peek() {
	case '\n', '\r':
		return lexGohtNewLine
	case scanner.EOF:
		l.emit(tEOF)
		return nil
	default:
		return l.errorf("unexpected character: %q", l.peek())
	}
}

func lexGohtNewLine(l *lexer) lexFn {
	l.acceptRun("\n\r")
	l.emit(tNewLine)
	return lexGohtLineStart
}

func hamlIdentifier(typ tokenType, l *lexer) lexFn {
	l.skip() // eat symbol

	// these characters may follow an identifier
	const mayFollowIdentifier = "%#.[{=!/<> \t\n\r"

	l.acceptUntil(mayFollowIdentifier)
	if l.current() == "" {
		return l.errorf("%s identifier expected", typ)
	}
	l.emit(typ)
	return lexGohtContent
}

func lexGohtTag(l *lexer) lexFn {
	return hamlIdentifier(tTag, l)
}

func lexGohtId(l *lexer) lexFn {
	return hamlIdentifier(tId, l)
}

func lexGohtClass(l *lexer) lexFn {
	return hamlIdentifier(tClass, l)
}

func lexObjectReference(l *lexer) lexFn {
	l.skip() // eat opening bracket
	r := continueToMatchingBrace(l, ']')
	if r == scanner.EOF {
		return l.errorf("object reference not closed: eof")
	}
	l.backup()
	l.emit(tObjectRef)
	l.skip() // skip closing bracket
	return lexGohtContent
}

func lexGohtAttributesStart(l *lexer) lexFn {
	l.skip()
	return lexGohtAttribute
}

func lexGohtAttributesEnd(l *lexer) lexFn {
	l.skip()
	return lexGohtContent
}

func lexGohtAttribute(l *lexer) lexFn {
	// supported attributes
	// key
	// key:value
	// key?value
	// @attributes: []any (string, map[string]string, map[string]bool)

	l.skipRun(", \t\n\r")

	switch l.peek() {
	case '}':
		return lexGohtAttributesEnd
	case '@':
		return lexAttributeCommandStart
	default:
		return lexGohtAttributeName
	}
}

func lexGohtAttributeName(l *lexer) lexFn {
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
		return lexGohtAttributeOperator
	case ',', '}':
		return lexGohtAttributeEnd
	default:
		return l.errorf("unexpected character: %q", l.peek())
	}
}

func lexGohtAttributeOperator(l *lexer) lexFn {
	l.skipRun(" \t\n\r")
	switch l.peek() {
	case '?', ':':
		l.next()
		l.emit(tAttrOperator)
		return lexGohtAttributeValue
	}
	return l.errorf("unexpected character: %q", l.peek())
}

func lexGohtAttributeValue(l *lexer) lexFn {
	l.skipRun(" \t\n\r")

	switch l.peek() {
	case '"', '`':
		return lexGohtAttributeStaticValue
	case '#':
		return lexGohtAttributeDynamicValue
	}
	return l.errorf("unexpected character: %q", l.peek())
}

func lexGohtAttributeStaticValue(l *lexer) lexFn {
	r := continueToMatchingQuote(l, tAttrEscapedValue, true)
	if r == scanner.EOF {
		return l.errorf("attribute value not closed: eof")
	} else if r != '"' && r != '`' {
		return l.errorf("unexpected character: %q", r)
	}
	return lexGohtAttributeEnd
}

func lexGohtAttributeDynamicValue(l *lexer) lexFn {
	l.skip() // skip hash
	if l.peek() != '{' {
		return l.errorf("unexpected character: %q", l.peek())
	}
	l.skip() // skip opening brace
	r := continueToMatchingBrace(l, '}')
	if r == scanner.EOF {
		return l.errorf("attribute value not closed: eof")
	}
	l.backup()
	l.emit(tAttrDynamicValue)
	l.skip() // skip closing brace
	return lexGohtAttributeEnd
}

func lexAttributeCommandStart(l *lexer) lexFn {
	l.skipRun("@")
	l.acceptUntil(": \t\n\r")
	if l.current() == "" {
		return l.errorf("command code expected")
	}
	switch l.current() {
	case "attributes":
		return lexGohtAttributeCommand(tAttributesCommand)
	default:
		return l.errorf("unknown attribute command: %s", l.current())
	}
}

func lexGohtAttributeCommand(command tokenType) lexFn {
	return func(l *lexer) lexFn {
		l.ignore()
		l.skipUntil(":")
		l.skipUntil("{")
		l.skip() // skip opening brace
		r := continueToMatchingBrace(l, '}')
		if r == scanner.EOF {
			return l.errorf("attribute value not closed: eof")
		}
		l.backup()
		l.emit(command)
		l.skip() // skip closing brace

		return lexGohtAttributeEnd
	}
}

func lexGohtAttributeEnd(l *lexer) lexFn {
	l.skipRun(" \t\n\r")
	switch l.peek() {
	case ',':
		l.skip()
		return lexGohtAttribute
	case '}':
		return lexGohtAttributesEnd
	default:
		return l.errorf("unexpected character: %c", l.peek())
	}
}

func lexWhitespaceRemoval(l *lexer) lexFn {
	direction := l.skip()
	switch direction {
	case '>':
		l.emit(tNukeOuterWhitespace)
	case '<':
		l.emit(tNukeInnerWhitespace)
	default:
		return l.errorf("unexpected character: %q", direction)
	}
	return lexGohtContentEnd
}

func lexGohtTextStart(l *lexer) lexFn {
	l.skipRun(" \t")
	return lexGohtTextContent
}

func lexGohtTextContent(l *lexer) lexFn {
	l.acceptUntil("\\#\n\r")
	switch l.peek() {
	case '\\':
		isHashComing, err := l.peekAhead(2)
		if err != nil {
			return l.errorf("unexpected error: %s", err)
		}
		if isHashComing == "\\#" {
			l.skip()
			// was the backslash being escaped?
			if !strings.HasSuffix(l.current(), "\\") {
				l.next()
			}
		} else {
			l.next()
		}
		return lexGohtTextContent
	case '#':
		return lexGohtDynamicText
	default:
		if l.current() != "" {
			l.emit(tPlainText)
		}
		return lexGohtLineEnd
	}
}

func lexGohtDynamicText(l *lexer) lexFn {
	if s, err := l.peekAhead(2); err != nil {
		return l.errorf("unexpected error: %s", err)
	} else if s != "#{" {
		l.next()
		return lexGohtTextContent
	}
	if l.current() != "" {
		l.emit(tPlainText)
	}
	l.skipRun("#{")
	r := continueToMatchingBrace(l, '}')
	if r == scanner.EOF {
		return l.errorf("dynamic text value was not closed: eof")
	}
	l.backup()
	l.emit(tDynamicText)
	l.skip() // skip closing brace
	return lexGohtTextContent
}

func lexGohtDoctype(l *lexer) lexFn {
	l.skipRun("! ")
	l.acceptUntil("\n\r")
	l.emit(tDoctype)
	return lexGohtLineEnd
}

func lexGohtUnescaped(l *lexer) lexFn {
	l.skip()
	l.ignore()
	l.emit(tUnescaped)
	switch l.peek() {
	case '=':
		return lexGohtOutputCode
	default:
		return lexGohtTextStart
	}
}

func lexGohtSilentScript(l *lexer) lexFn {
	l.skip() // eat dash

	// ruby style comment
	if l.peek() == '#' {
		// ignore the rest of the line
		l.skipUntil("\n\r")
		l.emit(tRubyComment)
		return ignoreIndentedLines(l.indent + 1)
	}

	l.skipRun(" \t")
	l.acceptUntil("\n\r")
	l.emit(tSilentScript)
	return lexGohtLineEnd
}

func ignoreIndentedLines(indent int) lexFn {
	return func(l *lexer) lexFn {
		switch l.peek() {
		case '\n', '\r':
			l.skip()
			return ignoreIndentedLines(indent)
		case ' ', '\t':
			priorIndents, err := l.peekAhead(indent * l.indentLen)
			if err != nil {
				return l.errorf("unexpected error while evaluating indents: %s", err)
			}
			if len(strings.TrimSpace(priorIndents)) != 0 {
				return lexGohtLineStart
			}
			// validate we have the correct indents
			if lexErr := l.validateIndent(priorIndents); lexErr != nil {
				return lexErr
			}
			l.skipUntil("\n\r")
			return ignoreIndentedLines(indent)
		case scanner.EOF:
			l.emit(tEOF)
			return nil
		default:
			return lexGohtLineStart
		}
	}
}

func lexGohtOutputCode(l *lexer) lexFn {
	l.skipRun("= \t")
	switch l.peek() {
	case '@':
		return lexGohtCommandCode
	default:
		l.acceptUntil("\n\r")
		l.emit(tScript)
		return lexGohtLineEnd
	}
}

func lexComment(l *lexer) lexFn {
	l.skipRun("/ \t")
	l.acceptUntil("\n\r")
	l.emit(tComment)
	return lexGohtLineEnd
}

func lexVoidTag(l *lexer) lexFn {
	l.skipRun("/ \t")
	l.acceptUntil("\n\r")
	if l.current() != "" {
		l.ignore()
		return l.errorf("self-closing tags can't have content")
	}
	l.emit(tVoidTag)
	return lexGohtLineEnd
}

func lexGohtCommandCode(l *lexer) lexFn {
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
	l.skipRun("\n\r")
	return lexGohtLineStart
}

var filters = []string{"javascript", "css", "plain", "escaped", "preserve"}

func lexFilterStart(l *lexer) lexFn {
	l.skipRun(": \t")
	l.acceptUntil(" \t\n\r")
	if l.current() == "" {
		return l.errorf("filter name expected")
	}
	if !slices.Contains(filters, l.current()) {
		return l.errorf("unknown filter: %s", l.current())
	}
	filter := l.current()
	l.emit(tFilterStart)
	l.skipUntil("\n\r") // ignore the rest of the current line
	l.skipRun("\n\r")   // split so we don't consume the indent on the next line

	switch filter {
	case "javascript", "css", "plain":
		return lexFilterLineStart(l.indent+1, tPlainText)
	case "escaped":
		return lexFilterLineStart(l.indent+1, tEscapedText)
	case "preserve":
		return lexFilterLineStart(l.indent+1, tPreserveText)
	}
	return lexGohtLineEnd
}

func lexFilterLineStart(indent int, textType tokenType) lexFn {
	return func(l *lexer) lexFn {
		switch l.peek() {
		case ' ', '\t':
			return lexFilterIndent(indent, textType)
		case scanner.EOF:
			l.emit(tEOF)
			return nil
		default:
			l.emit(tFilterEnd)
			return lexGohtLineStart
		}
	}
}

func lexFilterIndent(indent int, textType tokenType) lexFn {
	return func(l *lexer) lexFn {
		var indents string

		if l.indentLen == 0 {
			// if the template has not yet been indented, then the first indent
			// used in the filter becomes the templates indent
			l.acceptRun(" \t")
		} else {
			// only accept the whitespace that belongs to the indent
			var err error

			// peeking first, in case we've reached the end of the filter
			indents, err = l.peekAhead(indent * l.indentLen)
			if err != nil {
				return l.errorf("unexpected error while evaluating filter indents: %s", err)
			}

			if len(strings.TrimSpace(indents)) != 0 {
				l.emit(tFilterEnd)
				return lexGohtLineStart
			}
			l.acceptAhead(indent * l.indentLen)
		}
		indents = l.current()
		// throw away the whitespace
		l.ignore()

		if len(indents) == 0 {
			l.emit(tFilterEnd)
			return lexGohtLineStart
		}

		// set indent char and length
		if l.indentChar == 0 {
			if strings.Contains(indents, " ") && strings.Contains(indents, "\t") {
				return l.errorf("indentation cannot contain both spaces and tabs")
			}
			l.indentChar = ' '
			if strings.Contains(indents, "\t") {
				l.indentChar = '\t'
			}
			l.indentLen = len(indents)
		}

		// validate the indent against the sequence and char
		if lexErr := l.validateIndent(indents); lexErr != nil {
			return lexErr
		}

		// l.indent = len(l.current()) / l.indentLen // useful for parsing filters
		// l.emit(tIndent)
		return lexFilterContent(indent, textType)
	}

}

func lexFilterContent(indent int, textType tokenType) lexFn {
	return func(l *lexer) lexFn {
		l.acceptUntil("#\n\r")
		if l.peek() == '#' {
			return lexFilterDynamicText(textType, lexFilterContent(indent, textType))
		}
		l.acceptRun("\n\r")
		if l.current() != "" {
			l.emit(textType)
		}
		return lexFilterLineStart(indent, textType)
	}
}

func lexFilterDynamicText(textType tokenType, next lexFn) lexFn {
	return func(l *lexer) lexFn {
		if s, err := l.peekAhead(2); err != nil {
			return l.errorf("unexpected error: %s", err)
		} else if s != "#{" {
			l.next()
			return next
		}
		if l.current() != "" {
			l.emit(textType)
		}
		l.skipRun("#{")
		r := continueToMatchingBrace(l, '}')
		if r == scanner.EOF {
			return l.errorf("dynamic text value was not closed: eof")
		}
		l.backup()
		l.emit(tDynamicText)
		l.skip() // skip closing brace
		return next
	}
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
