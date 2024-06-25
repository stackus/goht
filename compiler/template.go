package compiler

import (
	"io"
	"strconv"
	"strings"
)

type Template struct {
	Filename string
	Root     nodeBase
}

type templateWriter struct {
	w            io.Writer
	num          *int
	pos          *Position
	sm           *SourceMap
	indent       int
	inStatic     bool
	inErrHandler bool
	isUnescaped  bool
}

func (t *Template) Generate(w io.Writer) error {
	tw := newTemplateWriter(w, nil)

	err := t.Root.Source(tw)
	return err
}

func (t *Template) Compose(w io.Writer) (*SourceMap, error) {
	tw := newTemplateWriter(w, &SourceMap{
		SourceLinesToTarget: make(map[int]map[int]Position),
		TargetLinesToSource: make(map[int]map[int]Position),
	})

	err := t.Root.Source(tw)
	return tw.sm, err
}

func newTemplateWriter(w io.Writer, sm *SourceMap) *templateWriter {
	num := 0
	tw := &templateWriter{
		w:   w,
		num: &num,
		pos: &Position{
			Line: 1,
			Col:  1,
		},
		sm: sm,
	}
	return tw
}

// Add adds a token and its range to the source map.
//
// When the source map is nil, this is a no-op.
func (tw *templateWriter) Add(t token, r Range) {
	if tw.sm == nil {
		return
	}
	tw.sm.Add(t, r)
}

func (tw *templateWriter) Indent(indents uint) *templateWriter {
	itw := *tw
	itw.indent += int(indents)

	return &itw
}

func (tw *templateWriter) GetVarName() string {
	*tw.num++
	return "__var" + strconv.Itoa(*tw.num)
}

func (tw *templateWriter) ResetVarName() {
	*tw.num = 0
}

// WriteVar writes a variable declaration to the template.
//
// (e.g. "var __var1 string")
func (tw *templateWriter) WriteVar(s string) (r Range, err error) {
	if tw.inStatic {
		if _, err = tw.closeStringLiteral(); err != nil {
			return
		}
	}
	return tw.WriteIndent(`var ` + s + " string\n")
}

// Write writes a string to the template.
//
// (e.g. ... whatever you give it ...)
func (tw *templateWriter) Write(s string) (r Range, err error) {
	if tw.inStatic {
		if _, err = tw.closeStringLiteral(); err != nil {
			return
		}
	}
	return tw.write(s)
}

// WriteIndent writes a string to the template with the current indent.
// Close any open string literal before writing.
//
// (e.g. \t\t\t ... whatever you give it ...)
func (tw *templateWriter) WriteIndent(s string) (r Range, err error) {
	if tw.inStatic {
		if _, err = tw.closeStringLiteral(); err != nil {
			return
		}
	}
	if _, err = tw.write(strings.Repeat("\t", tw.indent)); err != nil {
		return
	}
	return tw.write(s)
}

// WriteStringLiteral writes a string literal to the template.
// Continue writing to the string literal if one is already open.
//
// (e.g. " ... whatever you give it ... ")
func (tw *templateWriter) WriteStringLiteral(s string) (r Range, err error) {
	if !tw.inStatic {
		if _, err = tw.write(strings.Repeat("\t", tw.indent)); err != nil {
			return
		}
		if _, err = tw.write(`if _, __err = __buf.WriteString("`); err != nil {
			return
		}
		tw.inStatic = true
		tw.inErrHandler = true
	}

	return tw.write(s)
}

// WriteStringIndent writes a string literal to the template with the current indent.
// Close any open string literal before writing.
//
// (e.g., if, __err := __buf.WriteString( ... whatever you give it ... ); __err != nil { return })
func (tw *templateWriter) WriteStringIndent(s string) (r Range, err error) {
	if tw.inStatic {
		if _, err = tw.closeStringLiteral(); err != nil {
			return
		}
	}

	if _, err = tw.write(strings.Repeat("\t", tw.indent)); err != nil {
		return
	}
	if _, err = tw.write(`if _, __err = __buf.WriteString(`); err != nil {
		return
	}
	if r, err = tw.write(s); err != nil {
		return
	}
	if _, err = tw.write("); __err != nil { return }\n"); err != nil {
		return
	}
	return
}

func (tw *templateWriter) WriteErrorHandler() (Range, error) {
	if tw.inStatic {
		return tw.closeStringLiteral()
	}
	return tw.write(tw.addErrHandler())
}

func (tw *templateWriter) Close() (r Range, err error) {
	if tw.inStatic {
		return tw.closeStringLiteral()
	}
	return
}

func (tw *templateWriter) write(s string) (r Range, err error) {
	r.From = *tw.pos
	nl := strings.Count(s, "\n")
	tw.pos.Line += nl
	if nl > 0 {
		tw.pos.Col = len(s) - strings.LastIndex(s, "\n")
	} else {
		tw.pos.Col += len(s)
	}

	_, err = io.WriteString(tw.w, s)
	r.To = *tw.pos
	return
}

func (tw *templateWriter) closeStringLiteral() (Range, error) {
	var nl = "\n"
	if tw.inErrHandler {
		nl = ""
	}
	s := `")` + nl + tw.addErrHandler()
	tw.inStatic = false
	return tw.write(s)
}

func (tw *templateWriter) addErrHandler() string {
	if tw.inErrHandler {
		tw.inErrHandler = false
		return "; __err != nil { return }\n"
	}
	return strings.Repeat("\t", tw.indent) + "if __err != nil { return }\n"
}
