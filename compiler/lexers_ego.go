package compiler

import (
	"fmt"
	"strings"
	"text/scanner"
)

func lexEgoStart(l *lexer) lexFn {
	return lexEgoLineStart(lexEgoText)
}

func increaseEgoIndent(l *lexer) lexFn {
	if l.current() != "" {
		return l.errorf("uncommitted content during block start: %q", l.current())
	}
	l.indent++
	l.s = strings.Repeat("\t", l.indent)
	l.emit(tIndent)
	return nil
}

func decreaseEgoIndent(l *lexer) lexFn {
	if l.current() != "" {
		return l.errorf("uncommitted content during block end: %q", l.current())
	}
	l.indent--
	l.s = strings.Repeat("\t", l.indent)
	l.emit(tIndent)
	return nil
}

func lexEgoLineStart(next lexFn) lexFn {
	return func(l *lexer) lexFn {
		switch l.peek() {
		case scanner.EOF:
			l.emit(tEOF)
			return nil
		case '\n', '\r':
			l.acceptRun("\n\r")
			return lexEgoLineStart(next)
		case '\t':
			// we require all templates to be indented with one tab; the rest of the line is content
			l.skip()
			return next
		case '}':
			// end of the template
			if l.current() != "" {
				// if the indent is not at 1, then we know that we're ending early and should report an error
				if l.indent != 0 {
					fmt.Println("indent", l.indent)
					fmt.Printf("current %q\n", l.current())
					return l.errorf("unexpected closing brace: %q", l.current())
				}
				// assumption: if there is anything in the buffer, then it is text, AND we can trim it
				l.s = strings.TrimRight(l.s, " \t\n\r")
				l.emit(tRawText)
			}
			l.emit(tTemplateEnd)
			l.skip()
			return lexGoLineStart
		default:
			return l.errorf("unexpected character: %q", l.peek())
		}
	}
}

func lexEgoText(l *lexer) lexFn {
	for {
		l.acceptUntil("<\n\r")
		switch l.peek() {
		case '\n', '\r':
			return lexEgoLineStart(lexEgoText)
		case '<':
			if l.peekAhead(2) == "<%" {
				return lexEgoTagStart
			}
			l.next()
		case scanner.EOF:
			if l.current() != "" {
				l.emit(tRawText)
			}
			l.emit(tEOF)
			return nil
		}
	}
}

func lexEgoTagStart(l *lexer) lexFn {
	l.skipAhead(2) // consume the '<%'
	r := l.peek()
	switch r {
	case '#':
		if l.current() != "" {
			l.emit(tRawText)
		}
		// comment
		l.skip() // consume the '#'
		return lexEgoCommentStart
	case '=':
		if l.current() != "" {
			l.emit(tRawText)
		}
		// output
		l.skip() // consume the '='
		return lexEgoOutputStart
	case '!':
		if l.current() != "" {
			l.emit(tRawText)
		}
		// unescaped output
		l.skip() // consume the '!'
		return lexEgoUnescapedOutputStart
	case '@':
		if l.current() != "" {
			l.emit(tRawText)
		}
		// command
		l.skip() // consume the '@'
		return lexEgoCommandStart
	case '%':
		// literal "<%"
		l.s += "<%"
		l.skip() // consume the '%'
		return lexEgoText
	case scanner.EOF:
		return l.errorf("unexpected EOF in tag")
	default:
		// script
		if r == '-' {
			// strip the whitespace on current()
			l.s = strings.TrimRight(l.s, " \t\n\r")
			l.emit(tRawText)
			l.skip() // consume the '-'
			return lexEgoScriptStart
		}
		if l.current() != "" {
			l.emit(tRawText)
		}
		return lexEgoScriptStart
	}
}

func findClosingTag(l *lexer, next lexFn) lexFn {
	for {
		l.acceptRun(" \t")
		l.acceptUntil("-$%\n\r")
		if l.peek() == scanner.EOF {
			return l.errorf("unexpected EOF in tag")
		}
		switch l.peek() {
		case '\n', '\r':
			// enforce indents even inside of multiline tags
			l.acceptRun("\n\r") // skip newlines
			if l.peek() != '\t' {
				return l.errorf("unexpected character at start of line: %q", l.peek())
			}
			l.skip() // skip the tab
			continue
		case '-':
			if l.peekAhead(3) != "-%>" {
				l.next()
				continue
			}
		case '$':
			if l.peekAhead(3) != "$%>" {
				l.next()
				continue
			}
		case '%':
			if l.peekAhead(2) != "%>" {
				l.next()
				continue
			}
		}
		if err := next(l); err != nil {
			return err
		}
		return lexEgoClosingTag
	}
}

func lexEgoClosingTag(l *lexer) lexFn {
	switch l.peek() {
	case '%':
		// just a closing tag
		l.skipAhead(2) // consume the '%>'
		return lexEgoText
	case '$':
		// closing tag with next newline removal
		l.skipAhead(3) // consume the '$%>'
		switch l.peek() {
		case '\n':
			l.skip() // skip the newline
			if l.peek() == '\r' {
				l.skip() // skip the carriage return
			}
		case '\r':
			l.skip() // skip the carriage return
		}
		return lexEgoLineStart(lexEgoText)
	case '-':
		// closing tag with whitespace removal
		l.skipAhead(3) // consume the '-%>'
		return lexStripWhitespace
	default:
		// unexpected character
		return l.errorf("unexpected character in closing tag: %q", l.peek())
	}
}

func lexEgoCommentStart(l *lexer) lexFn {
	l.skipRun(" \t\n\r") // skip whitespace

	return findClosingTag(l, func(l *lexer) lexFn {
		l.ignore() // ignore the comment
		return nil
	})
}

func lexEgoOutputStart(l *lexer) lexFn {
	l.skipRun(" \t\n\r") // skip whitespace

	return findClosingTag(l, func(l *lexer) lexFn {
		l.s = strings.TrimSpace(l.s)
		l.emit(tScript)
		return nil
	})
}

func lexEgoUnescapedOutputStart(l *lexer) lexFn {
	l.skipRun(" \t\n\r") // skip whitespace

	l.emit(tUnescaped)
	return findClosingTag(l, func(l *lexer) lexFn {
		l.s = strings.TrimSpace(l.s)
		l.emit(tScript)
		return nil
	})
}

func lexEgoCommandStart(l *lexer) lexFn {
	l.skipRun(" \t\n\r") // skip whitespace

	// TODO look for the command text
	l.acceptUntil(" \t%")
	if l.current() == "" {
		return l.errorf("command name expected")
	}
	switch l.current() {
	case "render":
		return lexEgoRenderStart
	case "children":
		return lexEgoChildrenStart
	case "slot":
		return lexEgoSlotStart
	default:
		return l.errorf("unknown command: %q", l.current())
	}
}

func lexEgoRenderStart(l *lexer) lexFn {
	l.skipRun(" \t") // skip whitespace
	l.ignore()       // ignore the command keyword and the whitespace

	return findClosingTag(l, func(l *lexer) lexFn {
		l.s = strings.TrimSpace(l.s)
		s := l.current()
		l.emit(tRenderCommand)
		// if the content ends with a '{' then increase the indent (after emitting)
		if strings.HasSuffix(s, "{") {
			if err := increaseEgoIndent(l); err != nil {
				return err
			}
		}
		return nil
	})
}

func lexEgoChildrenStart(l *lexer) lexFn {
	l.skipRun(" \t") // skip whitespace
	l.ignore()       // ignore the command keyword and the whitespace

	return findClosingTag(l, func(l *lexer) lexFn {
		if strings.TrimSpace(l.s) != "" {
			return l.errorf("unexpected content in children command: %q", l.s)
		}
		l.ignore()
		l.emit(tChildrenCommand)
		return nil
	})
}

func lexEgoSlotStart(l *lexer) lexFn {
	l.skipRun(" \t") // skip whitespace
	l.ignore()       // ignore the command keyword and the whitespace

	return findClosingTag(l, func(l *lexer) lexFn {
		l.s = strings.TrimSpace(l.s)
		s := l.current()
		l.s = strings.TrimRight(l.s, " \t{") // remove trailing whitespace and '{'
		if l.current() == "" {
			return l.errorf("slot name expected")
		}
		l.emit(tSlotCommand)
		// if the content originally ends with a '{' then increase the indent (after emitting)
		if strings.HasSuffix(s, "{") {
			if err := increaseEgoIndent(l); err != nil {
				return err
			}
		}
		return nil
	})
}

func lexEgoScriptStart(l *lexer) lexFn {
	l.skipRun(" \t\n\r") // skip whitespace

	// if the content starts with a '}' then decrease the indent (before emitting)
	if l.peek() == '}' {
		if err := decreaseEgoIndent(l); err != nil {
			return err
		}
	}

	return findClosingTag(l, func(l *lexer) lexFn {
		l.s = strings.TrimSpace(l.s)
		s := l.current()
		l.emit(tSilentScript)
		// if the content ends with a '{' then increase the indent (after emitting)
		if strings.HasSuffix(s, "{") {
			return increaseEgoIndent(l)
		}
		return nil
	})
}

func lexStripWhitespace(l *lexer) lexFn {
	for {
		l.skipRun(" \t") // skip whitespace
		switch l.peek() {
		case scanner.EOF:
			l.emit(tEOF)
			return nil
		case '\n', '\r':
			// enforce indents even when we are stripping whitespace
			l.skipRun("\n\r") // skip newlines
			if l.peek() != '\t' {
				return l.errorf("unexpected character at start of line: %q", l.peek())
			}
			l.skip() // skip the tab
			continue
		default:
			return lexEgoText
		}
	}
}
