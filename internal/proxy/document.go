package proxy

import (
	"fmt"
	"strings"

	"github.com/stackus/hamlet/internal/protocol"
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
	l, c := len(d.lines), len(d.lines[len(d.lines)-1])
	return r.End.Line >= uint32(l) || r.End.Character >= uint32(c)
}

func (d *Document) isInsert(r *protocol.Range) bool {
	return r.Start.Line == r.End.Line && r.Start.Character == r.End.Character
}

func (d *Document) insert(line, col int, lines []string) {
	before := d.lines[line][:col]
	after := d.lines[line][col:]
	d.lines[line] = before + lines[0]
	if len(lines) > 1 {
		d.lines = append(d.lines[:line+1], append(lines[1:], d.lines[:line+1]...)...)
	}
	d.lines[line+len(lines)-1] = lines[len(lines)-1] + after
}

func (d *Document) isReplace(r *protocol.Range) bool {
	return r.Start.Line != r.End.Line || r.Start.Character != r.End.Character
}

func (d *Document) delete(startLine, startCol, endLine, endCol int) {
	if startLine == endLine {
		d.lines[startLine] = d.lines[startLine][:startCol] + d.lines[startLine][endCol:]
		return
	}
	d.lines[startLine] = d.lines[startLine][:startCol]
	d.lines[endLine] = d.lines[endLine][endCol:]
	d.lines = append(d.lines[:startLine+1], d.lines[endLine:]...)
}

func (d *Document) overwrite(startLine, startCol, endLine, endCol int, lines []string) {
	if startLine == endLine {
		d.lines[startLine] = d.lines[startLine][:startCol] + lines[0] + d.lines[startLine][endCol:]
		return
	}
	d.lines[startLine] = d.lines[startLine][:startCol] + lines[0]
	d.lines[endLine] = lines[len(lines)-1] + d.lines[endLine][endCol:]
	d.lines = append(d.lines[:startLine+1], append(lines[1:], d.lines[endLine:]...)...)
}
