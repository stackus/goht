package compiler

import (
	"strings"
)

type Position struct {
	Line int
	Col  int
}

type Range struct {
	From Position
	To   Position
}

type SourceMap struct {
	SourceLinesToTarget map[int]map[int]Position
	TargetLinesToSource map[int]map[int]Position
}

// Add will create a new source map entry from the Goht template to the generated Go code.
//
// The IDEs will be using zero-based line and column numbers, so we need to convert them
// from the one-based line and column numbers that we've generated during the parsing of the
// Goht template.
//
// When we parse the length of a line we will create mappings for the entire len() because
// the IDEs and LSPs will use the character AFTER the last character in the range as the
// end position.
// For example, "foo" will have a start column of 0 and an end column value of 3, not 2.
func (sm *SourceMap) Add(t token, destRange Range) {
	lines := strings.Split(t.lit, "\n")
	for lineIndex, line := range lines {
		srcLine := t.line + lineIndex - 1
		tgtLine := destRange.From.Line + lineIndex - 1

		var srcCol, tgtCol int
		if lineIndex == 0 {
			srcCol += t.col - 1
			tgtCol += destRange.From.Col - 1
		}

		if _, ok := sm.SourceLinesToTarget[srcLine]; !ok {
			sm.SourceLinesToTarget[srcLine] = make(map[int]Position)
		}
		if _, ok := sm.TargetLinesToSource[tgtLine]; !ok {
			sm.TargetLinesToSource[tgtLine] = make(map[int]Position)
		}

		for colIndex := 0; colIndex <= len(line); colIndex++ {
			sm.SourceLinesToTarget[srcLine][srcCol+colIndex] = Position{Line: tgtLine, Col: tgtCol + colIndex}
			sm.TargetLinesToSource[tgtLine][tgtCol+colIndex] = Position{Line: srcLine, Col: srcCol + colIndex}
		}
	}
}

func (sm *SourceMap) SourcePositionFromTarget(line, col int) (Position, bool) {
	if _, ok := sm.TargetLinesToSource[line]; !ok {
		return Position{}, false
	}
	if _, ok := sm.TargetLinesToSource[line][col]; !ok {
		return Position{}, false
	}
	return sm.TargetLinesToSource[line][col], true
}

func (sm *SourceMap) TargetPositionFromSource(line, col int) (Position, bool) {
	if _, ok := sm.SourceLinesToTarget[line]; !ok {
		return Position{}, false
	}
	if _, ok := sm.SourceLinesToTarget[line][col]; !ok {
		return Position{}, false
	}
	return sm.SourceLinesToTarget[line][col], true
}
