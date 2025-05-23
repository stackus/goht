package compiler

import (
	"slices"
	"strings"
	"text/scanner"
)

func lexHamlLineStart(l *lexer) lexFn {
	switch l.peek() {
	case '}':
		l.emit(tTemplateEnd)
		l.skip()
		return lexGoLineStart
	case scanner.EOF:
		l.emit(tEOF)
		return nil
	case '\n', '\r':
		return lexHamlLineEnd
	default:
		return lexHamlIndent
	}
}

func lexHamlIndent(l *lexer) lexFn {
	// accept spaces and tabs so that we can report about improper indentation
	l.acceptRun(" \t")
	indent := l.current()

	// there has not been any indentation yet
	if l.indent == 0 && len(indent) == 0 {
		// return an error that indents are required
		return l.errorf("haml templates must be indented")
	}

	// validate the indent against the sequence and char
	if lexHamlErr := l.validateIndent(indent); lexHamlErr != nil {
		return lexHamlErr
	}

	l.indent = len(l.current()) // useful for parsing filters
	l.emit(tIndent)
	return lexHamlContentStart
}

func lexHamlContentStart(l *lexer) lexFn {
	switch l.peek() {
	case '%':
		return lexHamlTag
	case '#':
		return lexHamlId
	case '.':
		return lexHamlClass
	case '\\':
		l.skip()
		return lexHamlTextStart
	case '!':
		if s := l.peekAhead(3); s == "!!!" {
			// TODO return an error if we're nesting doctypes
			return lexHamlDoctype
		}
		return lexHamlUnescaped
	case '-':
		return lexHamlSilentScript
	case '=':
		return lexHamlOutputCode
	case '/':
		return lexHamlComment
	case ':':
		return lexHamlFilterStart
	case '{':
		return lexHamlAttributesStart
	case scanner.EOF, '\n', '\r':
		return lexHamlLineEnd
	default:
		return lexHamlTextStart
	}
}

func lexHamlContent(l *lexer) lexFn {
	switch l.peek() {
	case '#':
		return lexHamlId
	case '.':
		return lexHamlClass
	case '[':
		return lexHamlObjectReference
	case '{':
		return lexHamlAttributesStart
	case '!':
		return lexHamlUnescaped
	case '=':
		return lexHamlOutputCode
	case '/':
		return lexHamlVoidTag
	case '>', '<':
		return lexHamlWhitespaceRemoval
	case scanner.EOF, '\n', '\r':
		return lexHamlLineEnd
	default:
		return lexHamlTextStart
	}
}

func lexHamlContentEnd(l *lexer) lexFn {
	switch l.peek() {
	case '=':
		return lexHamlOutputCode
	case '/':
		return lexHamlVoidTag
	case '>', '<':
		return lexHamlWhitespaceRemoval
	case scanner.EOF, '\n', '\r':
		return lexHamlLineEnd
	default:
		return lexHamlTextStart
	}
}

func lexHamlLineEnd(l *lexer) lexFn {
	l.skipRun(" \t")

	switch l.peek() {
	case '\n', '\r':
		return lexHamlNewLine
	case scanner.EOF:
		l.emit(tEOF)
		return nil
	default:
		return l.errorf("unexpected character: %q", l.peek())
	}
}

func lexHamlNewLine(l *lexer) lexFn {
	l.acceptRun("\n\r")
	l.emit(tNewLine)
	return lexHamlLineStart
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
	return lexHamlContent
}

func lexHamlTag(l *lexer) lexFn {
	return hamlIdentifier(tTag, l)
}

func lexHamlId(l *lexer) lexFn {
	return hamlIdentifier(tId, l)
}

func lexHamlClass(l *lexer) lexFn {
	return hamlIdentifier(tClass, l)
}

func lexHamlObjectReference(l *lexer) lexFn {
	l.skip() // eat opening bracket
	r := continueToMatchingBrace(l, ']', false)
	if r == scanner.EOF {
		return l.errorf("object reference not closed: eof")
	}
	l.backup()
	l.emit(tObjectRef)
	l.skip() // skip closing bracket
	return lexHamlContent
}

func lexHamlAttributesStart(l *lexer) lexFn {
	l.skip()
	return lexHamlAttribute
}

func lexHamlAttributesEnd(l *lexer) lexFn {
	l.skip()
	return lexHamlContent
}

func lexHamlAttribute(l *lexer) lexFn {
	// supported attributes
	// key
	// key:value
	// key?value
	// @attributes: []any (string, map[string]string, map[string]bool)

	l.skipRun(", \t\n\r")

	switch l.peek() {
	case '}':
		return lexHamlAttributesEnd
	case '@':
		return lexHamlAttributeCommandStart
	default:
		return lexHamlAttributeName
	}
}

func lexHamlAttributeName(l *lexer) lexFn {
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
		return lexHamlAttributeOperator
	case ',', '}':
		return lexHamlAttributeEnd
	default:
		return l.errorf("unexpected character: %q", l.peek())
	}
}

func lexHamlAttributeOperator(l *lexer) lexFn {
	l.skipRun(" \t\n\r")
	switch l.peek() {
	case '?', ':':
		l.next()
		l.emit(tAttrOperator)
		return lexHamlAttributeValue
	}
	return l.errorf("unexpected character: %q", l.peek())
}

func lexHamlAttributeValue(l *lexer) lexFn {
	l.skipRun(" \t\n\r")

	switch l.peek() {
	case '"', '`':
		return lexHamlAttributeStaticValue
	case '#':
		return lexHamlAttributeDynamicValue
	}
	return l.errorf("unexpected character: %q", l.peek())
}

func lexHamlAttributeStaticValue(l *lexer) lexFn {
	r := continueToMatchingQuote(l, tAttrEscapedValue, true)
	if r == scanner.EOF {
		return l.errorf("attribute value not closed: eof")
	} else if r != '"' && r != '`' {
		return l.errorf("unexpected character: %q", r)
	}
	return lexHamlAttributeEnd
}

func lexHamlAttributeDynamicValue(l *lexer) lexFn {
	l.skip() // skip hash
	if l.peek() != '{' {
		return l.errorf("unexpected character: %q", l.peek())
	}
	l.skip() // skip opening brace
	r := continueToMatchingBrace(l, '}', false)
	if r == scanner.EOF {
		return l.errorf("attribute value not closed: eof")
	}
	l.backup()
	l.emit(tAttrDynamicValue)
	l.skip() // skip closing brace
	return lexHamlAttributeEnd
}

func lexHamlAttributeCommandStart(l *lexer) lexFn {
	l.skipRun("@")
	l.acceptUntil(": \t\n\r")
	if l.current() == "" {
		return l.errorf("command code expected")
	}
	switch l.current() {
	case "attributes":
		return lexHamlAttributeCommand(tAttributesCommand)
	default:
		return l.errorf("unknown attribute command: %s", l.current())
	}
}

func lexHamlAttributeCommand(command tokenType) lexFn {
	return func(l *lexer) lexFn {
		l.ignore()
		l.skipUntil(":")
		l.skipUntil("{")
		l.skip() // skip opening brace
		r := continueToMatchingBrace(l, '}', true)
		if r == scanner.EOF {
			return l.errorf("attribute value not closed: eof")
		}
		l.backup()
		l.emit(command)
		l.skip() // skip closing brace

		return lexHamlAttributeEnd
	}
}

func lexHamlAttributeEnd(l *lexer) lexFn {
	l.skipRun(" \t\n\r")
	switch l.peek() {
	case ',':
		l.skip()
		return lexHamlAttribute
	case '}':
		return lexHamlAttributesEnd
	default:
		return l.errorf("unexpected character: %c", l.peek())
	}
}

func lexHamlWhitespaceRemoval(l *lexer) lexFn {
	direction := l.skip()
	switch direction {
	case '>':
		l.emit(tNukeOuterWhitespace)
	case '<':
		l.emit(tNukeInnerWhitespace)
	default:
		return l.errorf("unexpected character: %q", direction)
	}
	return lexHamlContentEnd
}

func lexHamlTextStart(l *lexer) lexFn {
	l.skipRun(" \t")
	return lexHamlTextContent
}

func lexHamlTextContent(l *lexer) lexFn {
	l.acceptUntil("\\#\n\r")
	switch l.peek() {
	case '\\':
		isHashComing := l.peekAhead(2)
		if isHashComing == "\\#" {
			l.skip()
			// was the backslash being escaped?
			if !strings.HasSuffix(l.current(), "\\") {
				l.next()
			}
		} else {
			l.next()
		}
		return lexHamlTextContent
	case '#':
		return lexHamlDynamicText
	default:
		if l.current() != "" {
			l.emit(tPlainText)
		}
		return lexHamlLineEnd
	}
}

func lexHamlDynamicText(l *lexer) lexFn {
	if s := l.peekAhead(2); s != "#{" {
		l.next()
		return lexHamlTextContent
	}
	if l.current() != "" {
		l.emit(tPlainText)
	}
	l.skipRun("#{")
	r := continueToMatchingBrace(l, '}', false)
	if r == scanner.EOF {
		return l.errorf("dynamic text value was not closed: eof")
	}
	l.backup()
	l.emit(tDynamicText)
	l.skip() // skip closing brace
	return lexHamlTextContent
}

func lexHamlDoctype(l *lexer) lexFn {
	l.skipRun("! ")
	l.acceptUntil("\n\r")
	l.emit(tDoctype)
	return lexHamlLineEnd
}

func lexHamlUnescaped(l *lexer) lexFn {
	l.skip()
	l.ignore()
	l.emit(tUnescaped)
	switch l.peek() {
	case '=':
		return lexHamlOutputCode
	default:
		return lexHamlTextStart
	}
}

func lexHamlSilentScript(l *lexer) lexFn {
	l.skip() // eat dash

	// ruby style comment
	if l.peek() == '#' {
		// ignore the rest of the line
		l.skipUntil("\n\r")
		l.emit(tRubyComment)
		return ignoreIndentedLines(l.indent+1, lexHamlLineStart)
	}

	l.skipRun(" \t")
	l.acceptUntil("\\\n\r")
	if n := l.peek(); n == '\\' || strings.HasSuffix(l.current(), ",") {
		if n == '\\' {
			l.skip()
		}
		l.acceptRun("\n\r")
		return lexHamlCodeBlockIndent(l.indent+1, tSilentScript)
	}
	l.emit(tSilentScript)
	return lexHamlLineEnd
}

func lexHamlOutputCode(l *lexer) lexFn {
	l.skipRun("= \t")
	switch l.peek() {
	case '@':
		return lexHamlCommandCode
	default:
		l.acceptUntil("\\\n\r")
		if n := l.peek(); n == '\\' || strings.HasSuffix(l.current(), ",") {
			if n == '\\' {
				l.skip()
			}
			l.acceptRun("\n\r")
			return lexHamlCodeBlockIndent(l.indent+1, tScript)
		}
		l.emit(tScript)
		return lexHamlLineEnd
	}
}

func lexHamlCodeBlockIndent(indent int, textType tokenType) lexFn {
	return func(l *lexer) lexFn {
		// only accept the whitespace that belongs to the indent

		// peeking first, in case we've reached the end of the block
		indents := l.peekAhead(indent)

		if len(strings.Trim(indents, "\t")) != 0 {
			return l.errorf("expected continuation of code")
		}

		l.skipAhead(indent)

		return lexHamlCodeBlockContent(indent, textType)
	}
}

func lexHamlCodeBlockContent(indent int, textType tokenType) lexFn {
	return func(l *lexer) lexFn {
		l.acceptUntil("\\\n\r")
		if n := l.peek(); n == '\\' || strings.HasSuffix(l.current(), ",") {
			if n == '\\' {
				l.skip()
			}
			l.acceptRun("\n\r")
			return lexHamlCodeBlockIndent(indent, textType)
		}
		l.acceptRun("\n\r")
		l.emit(textType)

		return lexHamlLineStart
	}
}

func lexHamlComment(l *lexer) lexFn {
	l.skipRun("/ \t")
	l.acceptUntil("\n\r")
	l.emit(tComment)
	return lexHamlLineEnd
}

func lexHamlVoidTag(l *lexer) lexFn {
	l.skipRun("/ \t")
	l.acceptUntil("\n\r")
	if l.current() != "" {
		l.ignore()
		return l.errorf("self-closing tags can't have content")
	}
	l.emit(tVoidTag)
	return lexHamlLineEnd
}

func lexHamlCommandCode(l *lexer) lexFn {
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
			l.acceptRun("\n\r")
			return lexHamlCodeBlockIndent(l.indent+1, tRenderCommand)
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
	case "slot":
		l.acceptRun("() \t")
		l.ignore()
		l.acceptUntil("\n\r")
		if l.current() == "" {
			return l.errorf("slot name expected")
		}
		l.emit(tSlotCommand)
	default:
		return l.errorf("unknown command: %s", l.current())
	}
	l.skipRun("\n\r")
	return lexHamlLineStart
}

var hamlFilters = []string{"javascript", "css", "plain", "escaped", "preserve"}

func lexHamlFilterStart(l *lexer) lexFn {
	l.skipRun(": \t")
	l.acceptUntil(" \t\n\r")
	if l.current() == "" {
		return l.errorf("filter name expected")
	}
	if !slices.Contains(hamlFilters, l.current()) {
		return l.errorf("unknown filter: %s", l.current())
	}
	filter := l.current()
	l.emit(tFilterStart)
	l.skipUntil("\n\r") // ignore the rest of the current line
	l.skipRun("\n\r")   // split so we don't consume the indent on the next line

	switch filter {
	case "javascript", "css", "plain":
		return lexHamlFilterLineStart(l.indent+1, tPlainText)
	case "escaped":
		return lexHamlFilterLineStart(l.indent+1, tEscapedText)
	case "preserve":
		return lexHamlFilterLineStart(l.indent+1, tPreserveText)
	default:
		return l.errorf("unsupported filter: %s", filter)
	}
}

func lexHamlFilterLineStart(indent int, textType tokenType) lexFn {
	return func(l *lexer) lexFn {
		switch l.peek() {
		case ' ', '\t':
			return lexHamlFilterIndent(indent, textType)
		case scanner.EOF:
			l.emit(tEOF)
			return nil
		default:
			l.emit(tFilterEnd)
			return lexHamlLineStart
		}
	}
}

func lexHamlFilterIndent(indent int, textType tokenType) lexFn {
	return func(l *lexer) lexFn {
		var indents string

		// peeking first, in case we've reached the end of the filter
		indents = l.peekAhead(indent)

		// trim the tabs from what we've peeked into; no longer using TrimSpace as that would trim spaces and newlines
		if len(strings.Trim(indents, "\t")) != 0 {
			l.emit(tFilterEnd)
			return lexHamlLineStart
		}

		l.skipAhead(indent)

		return lexHamlFilterContent(indent, textType)
	}
}

func lexHamlFilterContent(indent int, textType tokenType) lexFn {
	return func(l *lexer) lexFn {
		l.acceptUntil("#\n\r")
		if l.peek() == '#' {
			return lexHamlFilterDynamicText(textType, lexHamlFilterContent(indent, textType))
		}
		l.acceptRun("\n\r")
		if l.current() != "" {
			l.emit(textType)
		}
		return lexHamlFilterLineStart(indent, textType)
	}
}

func lexHamlFilterDynamicText(textType tokenType, next lexFn) lexFn {
	return func(l *lexer) lexFn {
		if s := l.peekAhead(2); s != "#{" {
			l.next()
			return next
		}
		if l.current() != "" {
			l.emit(textType)
		}
		l.skipAhead(2) // skip the hash and opening brace
		r := continueToMatchingBrace(l, '}', false)
		if r == scanner.EOF {
			return l.errorf("dynamic text value was not closed: eof")
		}
		l.backup()
		l.emit(tDynamicText)
		l.skip() // skip closing brace
		return next
	}
}
