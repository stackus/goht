package compiler

import (
	"slices"
	"strings"
	"text/scanner"
)

// func lexHamlStart(l *lexer) lexFn {
// 	l.ignore()
// 	l.skipRun(" ")
// 	l.acceptUntil(")")
// 	if strings.HasPrefix(l.current(), "(") {
// 		// we've only captured the receiver, so we need to capture the rest of the function signature
// 		l.next()
// 		for {
// 			l.acceptUntil(")")
// 			// handle the situation where the function signature contains an `interface{}` type with one or more methods
// 			openParens := strings.Count(l.current(), "(")
// 			closeParens := strings.Count(l.current(), ")")
// 			if openParens == closeParens+1 {
// 				break
// 			}
// 			l.next()
// 		}
// 	}
// 	l.next()
// 	l.emit(tGohtStart)
// 	l.skipRun(" {")
// 	l.skipRun("\n\r")
//
// 	return lexHamlLineStart
// }

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

	// if len(indent) == 0 {
	// 	l.indent = 0
	// 	l.emit(tIndent)
	// 	return lexHamlContentStart
	// }

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
		if s, err := l.peekAhead(3); err != nil {
			return l.errorf("unexpected error: %s", err)
		} else if s == "!!!" {
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
	case '-':
		return lexHamlSilentScript
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
	r := continueToMatchingBrace(l, ']')
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
	r := continueToMatchingBrace(l, '}')
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
		r := continueToMatchingBrace(l, '}')
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
	if s, err := l.peekAhead(2); err != nil {
		return l.errorf("unexpected error: %s", err)
	} else if s != "#{" {
		l.next()
		return lexHamlTextContent
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
	l.acceptUntil("\n\r")
	// TODO: Support multiline silent scripts when they end with a backslash or comma
	// example:
	// - foo = bar \
	//   + baz
	// - foo = bigCall( \
	//   	bar,
	//   	baz,
	//   )
	// Extended lines must be indented once.
	// Additional indentation is captured and emitted with the script
	l.emit(tSilentScript)
	return lexHamlLineEnd
}

func lexHamlOutputCode(l *lexer) lexFn {
	l.skipRun("= \t")
	switch l.peek() {
	case '@':
		return lexHamlCommandCode
	default:
		l.acceptUntil("\n\r")
		// TODO: Support multiline output code when they end with a backslash or comma
		// see the comments in lexHamlSilentScript
		l.emit(tScript)
		return lexHamlLineEnd
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
	}
	return lexHamlLineEnd
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

		// only accept the whitespace that belongs to the indent
		var err error

		// peeking first, in case we've reached the end of the filter
		indents, err = l.peekAhead(indent)
		if err != nil {
			return l.errorf("unexpected error while evaluating filter indents: %s", err)
		}

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
