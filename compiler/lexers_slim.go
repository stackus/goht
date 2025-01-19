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

	// if len(indent) == 0 {
	// 	l.indent = 0
	// 	l.emit(tIndent)
	// 	return lexSlimContentStart
	// }

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
		return lexSlimVoidTag
	case '>', '<':
		return lexSlimWhitespaceAddition
	case scanner.EOF, '\n', '\r':
		return lexSlimLineEnd
	default:
		return lexSlimTextStart
	}
}

func lexSlimContentEnd(l *lexer) lexFn {
	switch l.peek() {
	case '=':
		return lexSlimOutputCode
	case '/':
		return lexSlimVoidTag
	case '>', '<':
		return lexSlimWhitespaceAddition
	case scanner.EOF, '\n', '\r':
		return lexSlimLineEnd
	default:
		return lexSlimTextStart
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
	const mayFollowIdentifier = "#.{=!/<> \t\n\r"

	l.acceptUntil(mayFollowIdentifier)
	if l.current() == "" {
		return l.errorf("%s identifier expected", typ)
	}
	if l.current() == "doctype" {
		l.emit(tDoctype)
		l.skipUntil("\n\r")
		return lexSlimLineEnd
	}
	l.emit(typ)
	return lexSlimContent
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

func lexSlimControlCode(l *lexer) lexFn {
	l.skip() // eat dash

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
		l.acceptUntil("\n\r")
		// TODO: Support multiline output code when they end with a backslash or comma
		// see the comments in lexHamlSilentScript
		l.emit(tScript)
		return lexSlimLineEnd
	}
}

func lexSlimTextStart(l *lexer) lexFn {
	l.acceptUntil("#.{=!/<> \t\n\r")
	if l.current() == "" {
		return l.errorf("text expected")
	}
	l.emit(tPlainText)
	return lexSlimContentEnd
}

func lexSlimComment(l *lexer) lexFn {
	l.skip() // eat slash
	if l.peek() != '!' {
		// ignore the rest of the line
		l.skipUntil("\n\r")
		l.emit(tRubyComment)
		return ignoreIndentedLines(l.indent+1, lexSlimLineStart)
	}
	// HTML comments in slim can be multiline
	// The comment continues to the next line if:
	// - the number of indents are the same
	// - there is at least one more whitespace characters
	// The additional whitespace characters are removed from the second+ lines if they are
	// to the left of the first non-whitespace character on the first line.
	// TODO: Testing appears that this is space based and when you use tabs for the additional whitespace, it doesn't work as expected.
	// The '/' and '!' will both count as whitespace characters in this case.
	// One additional whitespace character may follow the '!' character.
	// This one additional whitespace character is removed from the first line
	// and will allow the second+ lines to have one additional whitespace character.
	// Example counts:
	// "/!This is a comment" is 2 whitespace characters and up to 2 whitespace characters are removed from the second+ lines
	// "/! This is a comment" is 3 whitespace characters and up to 3 whitespace characters are removed from the second+ lines
	// "/!  This is a comment" is 4 whitespace characters and only the first is removed and the second+ lines can have up to 3 whitespace characters
	// We remove "up to" the count but do not require the count to be the same. (at least one whitespace character rule remains though)
	// In practice, this means if the comments all start at the same column then they are considered to have no whitespace padding.

	l.skip() // eat bang
	return lexSlimFilterLineStart(l.indent+1, tEscapedText)
}

func lexSlimTextBlockLineStart(indent int, spaces int, textType tokenType) lexFn {
	return func(l *lexer) lexFn {
		if l.peek() == '|' {
			l.skip()
			l.skipRun(" \t")
			return lexSlimTextBlockContent(indent, spaces, textType)
		}
		return lexSlimLineEnd
	}
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

		// // only accept the whitespace that belongs to the indent
		// var err error

		// peeking first, in case we've reached the end of the filter
		indents = l.peekAhead(indent)
		// if err != nil {
		// 	return l.errorf("unexpected error while evaluating filter indents: %s", err)
		// }

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
