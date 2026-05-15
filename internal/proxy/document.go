package proxy

import (
	"fmt"
	"strings"

	"github.com/stackus/goht/internal/protocol"
)

type Document struct {
	lines []string
}

var _ fmt.Stringer = (*Document)(nil)

func NewDocument(text string) *Document {
	return &Document{
		lines: strings.Split(text, "\n"),
	}
}

func (d *Document) String() string {
	return strings.Join(d.lines, "\n")
}

func (d *Document) Apply(r *protocol.Range, text string) {
	lines := strings.Split(text, "\n")
	switch {
	case d.isWholeDocument(r):
		d.lines = lines
	case d.isInsert(r) && text != "":
		d.insert(int(r.Start.Line), int(r.Start.Character), lines)
	case d.isReplace(r) && text == "":
		d.delete(int(r.Start.Line), int(r.Start.Character), int(r.End.Line), int(r.End.Character))
	case d.isReplace(r) && text != "":
		d.overwrite(int(r.Start.Line), int(r.Start.Character), int(r.End.Line), int(r.End.Character), lines)
	}
}

func (d *Document) isWholeDocument(r *protocol.Range) bool {
	if r == nil {
		return true
	}
	if r.Start.Line != 0 || r.Start.Character != 0 {
		return false
	}
	lastLine := len(d.lines) - 1
	lastCol := len(d.lines[lastLine])
	return r.End.Line > uint32(lastLine) || r.End.Line == uint32(lastLine) && r.End.Character >= uint32(lastCol)
}

func (d *Document) isInsert(r *protocol.Range) bool {
	return r.Start.Line == r.End.Line && r.Start.Character == r.End.Character
}

func (d *Document) insert(line, col int, lines []string) {
	d.replace(line, col, line, col, lines)
}

func (d *Document) isReplace(r *protocol.Range) bool {
	return r.Start.Line != r.End.Line || r.Start.Character != r.End.Character
}

func (d *Document) delete(startLine, startCol, endLine, endCol int) {
	d.replace(startLine, startCol, endLine, endCol, []string{""})
}

func (d *Document) overwrite(startLine, startCol, endLine, endCol int, lines []string) {
	d.replace(startLine, startCol, endLine, endCol, lines)
}

func (d *Document) replace(startLine, startCol, endLine, endCol int, lines []string) {
	before := d.lines[startLine][:startCol]
	after := d.lines[endLine][endCol:]

	replacement := make([]string, len(lines))
	copy(replacement, lines)
	replacement[0] = before + replacement[0]
	replacement[len(replacement)-1] = replacement[len(replacement)-1] + after

	next := make([]string, 0, len(d.lines)-endLine+startLine+len(replacement))
	next = append(next, d.lines[:startLine]...)
	next = append(next, replacement...)
	next = append(next, d.lines[endLine+1:]...)
	d.lines = next
}
