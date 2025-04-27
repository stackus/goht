package compiler

import (
	"slices"
	"strings"
	"text/scanner"
)

func lexSlimLineStart(l *lexer) lexFn {
	switch l.peek() {
	case '}':
		l.emit(tTemplateEnd)
		l.skip()
		return lexGoLineStart
	case scanner.EOF:
		l.emit(tEOF)
		return nil
	case '\n', '\r':
		return lexSlimLineEnd
	default:
		return lexSlimIndent
	}
}

func lexSlimIndent(l *lexer) lexFn {
	// accept spaces and tabs so that we can report about improper indentation
	l.acceptRun(" \t")
	indent := l.current()

	// there has not been any indentation yet
	if l.indent == 0 && len(indent) == 0 {
		// return an error that indents are required
		return l.errorf("slim templates must be indented")
	}

	// validate the indent against the sequence and char
	if lexSlimErr := l.validateIndent(indent); lexSlimErr != nil {
		return lexSlimErr
	}

	l.indent = len(l.current()) // useful for parsing filters
	l.emit(tIndent)
	return lexSlimContentStart
}

func lexSlimContentStart(l *lexer) lexFn {
	switch p := l.peek(); p {
	case '#':
		return lexSlimId
	case '.':
		return lexSlimClass
	case '-':
		return lexSlimControlCode
	case '=':
		return lexSlimOutputCode
	case '/':
		return lexSlimComment
	case ':':
		return lexSlimFilterStart
	case '|':
		return lexSlimTextBlock
	case '{':
		return lexSlimAttributesStart
	case scanner.EOF, '\n', '\r':
		return lexSlimLineEnd
	default:
		// if the next character is a letter, we're starting a tag
		if isLetter(p) {
			return lexSlimTag
		}
		return l.errorf("unexpected character: %q", p)
	}
}

func isLetter(r rune) bool {
	return r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z'
}

func lexSlimContent(l *lexer) lexFn {
	switch l.peek() {
	case '#':
		return lexSlimId
	case '.':
		return lexSlimClass
	case '{':
		return lexSlimAttributesStart
	case '-':
		return lexSlimControlCode
	case '=':
		return lexSlimOutputCode
	case '/':
		return lexSlimCloseTag
	case '>', '<':
		return lexSlimWhitespaceAddition
	case ':':
		return lexSlimInlineTag
	case ' ', '\t':
		l.skip()
		return lexSlimTextBlockContent(l.indent+1, 0, tPlainText)
	case scanner.EOF, '\n', '\r':
		return lexSlimLineEnd
	default:
		return lexSlimTextBlockContent(l.indent+1, 0, tPlainText)
	}
}

func lexSlimLineEnd(l *lexer) lexFn {
	l.skipRun(" \t")

	switch l.peek() {
	case '\n', '\r':
		return lexSlimNewLine
	case scanner.EOF:
		l.emit(tEOF)
		return nil
	default:
		return l.errorf("unexpected character: %q", l.peek())
	}
}

func lexSlimNewLine(l *lexer) lexFn {
	l.acceptRun("\n\r")
	l.emit(tNewLine)
	return lexSlimLineStart
}

func slimIdentifier(typ tokenType, l *lexer) lexFn {
	if typ != tTag {
		l.skip() // eat symbol
	}

	// these characters may follow an identifier
	const mayFollowIdentifier = "#.{=!/<>: \t\n\r"

	l.acceptUntil(mayFollowIdentifier)
	if l.current() == "" {
		return l.errorf("%s identifier expected", typ)
	}
	if l.current() == "doctype" {
		return lexSlimDoctype
	}
	l.emit(typ)
	return lexSlimContent
}

func lexSlimDoctype(l *lexer) lexFn {
	l.ignore()
	l.skipRun(" \t")
	l.acceptUntil("\n\r")
	l.emit(tDoctype)
	return lexSlimLineEnd
}

func lexSlimTag(l *lexer) lexFn {
	return slimIdentifier(tTag, l)
}

func lexSlimId(l *lexer) lexFn {
	return slimIdentifier(tId, l)
}

func lexSlimClass(l *lexer) lexFn {
	return slimIdentifier(tClass, l)
}

func lexSlimInlineTag(l *lexer) lexFn {
	l.skip() // eat colon
	l.skipRun(" \t")
	return lexSlimContentStart
}

func lexSlimControlCode(l *lexer) lexFn {
	l.skip() // eat dash

	l.skipRun(" \t")
	l.acceptUntil("\\\n\r")
	// support long code lines split across multiple lines
	if n := l.peek(); n == '\\' || strings.HasSuffix(l.current(), ",") {
		if n == '\\' {
			l.skip()
		}
		return lexSlimCodeBlockContent(l.indent+1, tSilentScript)
	}
	l.emit(tSilentScript)
	return lexSlimLineEnd
}

func lexSlimOutputCode(l *lexer) lexFn {
	l.skip() // eat equals
	// if next character is an equals sign, then this content is not escaped
	if l.peek() == '=' {
		l.skip()
		l.emit(tUnescaped)
	}
	l.skipRun(" \t")
	switch l.peek() {
	case '@':
		return lexSlimCommandCode
	default:
		l.acceptUntil("\\\n\r")
		if n := l.peek(); n == '\\' || strings.HasSuffix(l.current(), ",") {
			if n == '\\' {
				l.skip()
			}
			return lexSlimCodeBlockContent(l.indent+1, tScript)
		}
		l.emit(tScript)
		return lexSlimLineEnd
	}
}

func lexSlimComment(l *lexer) lexFn {
	l.skip() // eat slash
	if l.peek() != '!' {
		// ignore the rest of the line
		l.skipUntil("\n\r")
		l.emit(tRubyComment)
		return ignoreIndentedLines(l.indent+1, lexSlimLineStart)
	}

	l.skip() // eat bang
	l.skipRun(" \t")
	return lexSlimTextBlockContent(l.indent+1, 0, tComment)
}

func lexSlimTextBlock(l *lexer) lexFn {
	l.skip() // eat pipe
	// test for a space after the pipe
	if n := l.peek(); n == ' ' || n == '\t' {
		return lexSlimTextBlockContent(l.indent+1, 1, tPlainText)
	}
	return lexSlimTextBlockContent(l.indent+1, 0, tPlainText)
}

func lexSlimTextBlockLineStart(indent int, spaces int, textType tokenType) lexFn {
	return func(l *lexer) lexFn {
		switch l.peek() {
		case ' ', '\t':
			return lexSlimTextBlockIndent(indent, spaces, textType)
		case scanner.EOF:
			l.emit(tEOF)
			return nil
		default:
			if l.current() != "" {
				l.emit(textType)
			}
			return lexSlimLineStart
		}
	}
}

func lexSlimTextBlockIndent(indent int, spaces int, textType tokenType) lexFn {
	return func(l *lexer) lexFn {
		// only accept the whitespace that belongs to the indent

		// peeking first, in case we've reached the end of the filter
		indents := l.peekAhead(indent)

		if len(strings.Trim(indents, "\t")) != 0 {
			if l.current() != "" {
				l.emit(textType)
			}
			return lexSlimLineStart
		}

		l.skipAhead(indent)

		return lexSlimTextBlockContent(indent, spaces, textType)
	}
}

func lexSlimTextBlockContent(indent int, spaces int, textType tokenType) lexFn {
	return func(l *lexer) lexFn {
		if spaces != 0 {
			if n := l.peek(); n == ' ' || n == '\t' {
				l.skip() // eat space
			}
		}
		l.acceptUntil("#\n\r")
		if len(l.current()) > 0 {
			l.emit(textType)
		}

		if l.peek() == '#' && !strings.HasSuffix(l.current(), "\\") {
			// l.emit(textType)
			return lexSlimFilterDynamicText(textType, lexSlimTextBlockContent(indent, spaces, textType))
		}
		l.acceptRun("\n\r")
		if l.current() != "" {
			l.emit(tNewLine)
		}
		return lexSlimTextBlockLineStart(indent, spaces, textType)
	}
}

func lexSlimCodeBlockLineStart(indent int, textType tokenType) lexFn {
	return func(l *lexer) lexFn {
		switch l.peek() {
		case ' ', '\t':
			return lexSlimCodeBlockIndent(indent, textType)
		case scanner.EOF:
			l.emit(tEOF)
			return nil
		default:
			if l.current() != "" {
				l.emit(textType)
			}
			return lexSlimLineStart
		}
	}
}

func lexSlimCodeBlockIndent(indent int, textType tokenType) lexFn {
	return func(l *lexer) lexFn {
		// only accept the whitespace that belongs to the indent

		// peeking first, in case we've reached the end of the block
		indents := l.peekAhead(indent)

		if len(strings.Trim(indents, "\t")) != 0 {
			if l.current() != "" {
				l.emit(textType)
			}
			return lexSlimLineStart
		}

		l.skipAhead(indent)

		return lexSlimCodeBlockContent(indent, textType)
	}
}

func lexSlimCodeBlockContent(indent int, textType tokenType) lexFn {
	return func(l *lexer) lexFn {
		l.acceptUntil("\n\r")
		l.skipRun("\n\r")
		return lexSlimCodeBlockLineStart(indent, textType)
	}
}

func lexSlimCloseTag(l *lexer) lexFn {
	l.skip() // eat slash
	l.skipRun(" \t")
	l.acceptUntil("\n\r")
	if l.current() != "" {
		return l.errorf("self-closing tags can't have content")
	}
	l.emit(tVoidTag)
	return lexSlimLineEnd
}

func lexSlimCommandCode(l *lexer) lexFn {
	l.skipRun("@")
	l.acceptUntil("() \t\n\r")
	if l.current() == "" {
		return l.errorf("command code expected")
	}
	switch l.current() {
	case "render":
		l.acceptRun("() \t")
		l.ignore()
		l.acceptUntil("\\\n\r")
		if l.current() == "" {
			return l.errorf("render argument expected")
		}
		if n := l.peek(); n == '\\' || strings.HasSuffix(l.current(), ",") {
			if n == '\\' {
				l.skip()
			}
			return lexSlimCodeBlockContent(l.indent+1, tRenderCommand)
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
	return lexSlimLineStart
}

var slimFilters = []string{"javascript", "css"}

func lexSlimFilterStart(l *lexer) lexFn {
	l.skipRun(": \t")
	l.acceptUntil(" \t\n\r")
	if l.current() == "" {
		return l.errorf("filter name expected")
	}
	if !slices.Contains(slimFilters, l.current()) {
		return l.errorf("unknown filter: %s", l.current())
	}
	filter := l.current()
	l.emit(tFilterStart)
	l.skipUntil("\n\r") // ignore the rest of the current line
	l.skipRun("\n\r")   // split so we don't consume the indent on the next line

	switch filter {
	case "javascript", "css":
		return lexSlimFilterLineStart(l.indent+1, tPlainText)
	}
	return lexSlimLineEnd
}

func lexSlimFilterLineStart(indent int, textType tokenType) lexFn {
	return func(l *lexer) lexFn {
		switch l.peek() {
		case ' ', '\t':
			return lexSlimFilterIndent(indent, textType)
		case scanner.EOF:
			l.emit(tEOF)
			return nil
		default:
			l.emit(tFilterEnd)
			return lexSlimLineStart
		}
	}
}

func lexSlimFilterIndent(indent int, textType tokenType) lexFn {
	return func(l *lexer) lexFn {
		var indents string

		// peeking first, in case we've reached the end of the filter
		indents = l.peekAhead(indent)

		// trim the tabs from what we've peeked into; no longer using TrimSpace as that would trim spaces and newlines
		if len(strings.Trim(indents, "\t")) != 0 {
			l.emit(tFilterEnd)
			return lexSlimLineStart
		}

		l.skipAhead(indent)

		return lexSlimFilterContent(indent, textType)
	}
}

func lexSlimFilterContent(indent int, textType tokenType) lexFn {
	return func(l *lexer) lexFn {
		l.acceptUntil("#\n\r")
		// we have reached some interpolation as long as it wasn't escaped
		if l.peek() == '#' && !strings.HasSuffix(l.current(), "\\") {
			return lexSlimFilterDynamicText(textType, lexSlimFilterContent(indent, textType))
		}
		l.acceptRun("\n\r")
		if l.current() != "" {
			l.emit(textType)
		}
		return lexSlimFilterLineStart(indent, textType)
	}
}

// lexSlimFilterDynamicText parses out dynamic text values within a filter block.
func lexSlimFilterDynamicText(textType tokenType, next lexFn) lexFn {
	return func(l *lexer) lexFn {
		if s := l.peekAhead(2); s != "#{" {
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

// Parsing the Slim attributes the same as the Haml attributes

func lexSlimAttributesStart(l *lexer) lexFn {
	l.skip()
	return lexSlimAttribute
}

func lexSlimAttributesEnd(l *lexer) lexFn {
	l.skip()
	return lexSlimContent
}

func lexSlimAttribute(l *lexer) lexFn {
	// supported attributes
	// key
	// key:value
	// key?value
	// @attributes: []any (string, map[string]string, map[string]bool)

	l.skipRun(", \t\n\r")

	switch l.peek() {
	case '}':
		return lexSlimAttributesEnd
	case '@':
		return lexSlimAttributeCommandStart
	default:
		return lexSlimAttributeName
	}
}

func lexSlimAttributeName(l *lexer) lexFn {
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
		return lexSlimAttributeOperator
	case ',', '}':
		return lexSlimAttributeEnd
	default:
		return l.errorf("unexpected character: %q", l.peek())
	}
}

func lexSlimAttributeOperator(l *lexer) lexFn {
	l.skipRun(" \t\n\r")
	switch l.peek() {
	case '?', ':':
		l.next()
		l.emit(tAttrOperator)
		return lexSlimAttributeValue
	}
	return l.errorf("unexpected character: %q", l.peek())
}

func lexSlimAttributeValue(l *lexer) lexFn {
	l.skipRun(" \t\n\r")

	switch l.peek() {
	case '"', '`':
		return lexSlimAttributeStaticValue
	case '#':
		return lexSlimAttributeDynamicValue
	}
	return l.errorf("unexpected character: %q", l.peek())
}

func lexSlimAttributeStaticValue(l *lexer) lexFn {
	r := continueToMatchingQuote(l, tAttrEscapedValue, true)
	if r == scanner.EOF {
		return l.errorf("attribute value not closed: eof")
	} else if r != '"' && r != '`' {
		return l.errorf("unexpected character: %q", r)
	}
	return lexSlimAttributeEnd
}

func lexSlimAttributeDynamicValue(l *lexer) lexFn {
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
	return lexSlimAttributeEnd
}

func lexSlimAttributeCommandStart(l *lexer) lexFn {
	l.skipRun("@")
	l.acceptUntil(": \t\n\r")
	if l.current() == "" {
		return l.errorf("command code expected")
	}
	switch l.current() {
	case "attributes":
		return lexSlimAttributeCommand(tAttributesCommand)
	default:
		return l.errorf("unknown attribute command: %s", l.current())
	}
}

func lexSlimAttributeCommand(command tokenType) lexFn {
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

		return lexSlimAttributeEnd
	}
}

func lexSlimAttributeEnd(l *lexer) lexFn {
	l.skipRun(" \t\n\r")
	switch l.peek() {
	case ',':
		l.skip()
		return lexSlimAttribute
	case '}':
		return lexSlimAttributesEnd
	default:
		return l.errorf("unexpected character: %c", l.peek())
	}
}

func lexSlimWhitespaceAddition(l *lexer) lexFn {
	switch l.peek() {
	case '>':
		l.skip()
		l.emit(tAddWhitespaceAfter)
	case '<':
		l.skip()
		l.emit(tAddWhitespaceBefore)
	default:
		return l.errorf("unexpected character: %q", l.peek())
	}
	return lexSlimContent
}
