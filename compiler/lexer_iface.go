package compiler

type ilexer interface {
	// next consumes the next rune from the input.
	next() rune
	// backup steps back one rune.
	backup()
	// peek returns the next rune without consuming it.
	peek() rune
	// peekAhead returns the next length runes without consuming them.
	peekAhead(length int) (string, error)
	// ignore discards the current captured string.
	ignore()
	// accept consumes the next rune if it's contained in the acceptRunes list.
	accept(acceptRunes string) bool
	// acceptRun consumes a run of runes from the acceptRunes list.
	acceptRun(acceptRunes string)
	// acceptUntil consumes runes until it encounters a rune in the stopRunes list.
	acceptUntil(stopRunes string)
	// acceptAhead consumes the next length runes.
	acceptAhead(length int)
	// skip discards the next rune.
	skip() rune
	// skipRun discards a contiguous run of runes from the skipRunes list.
	skipRun(skipRunes string)
	// skipUntil discards runes until it encounters a rune in the stopRunes list.
	skipUntil(stopRunes string)
	// skipAhead consumes the next length runes and discards them.
	skipAhead(length int)
	// current returns the current captured string being built by the lexer.
	current() string
}
