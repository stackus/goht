// Code generated by hamlet - DO NOT EDIT.

package main

import "bytes"
import "context"
import "io"
import "github.com/stackus/hamlet"

// Adds a doctype to the top of the page
// HTML5 doctype is the default
func Doctype1() hamlet.Template {
	return hamlet.TemplateFunc(func(ctx context.Context, __w io.Writer) (__err error) {
		__buf, __isBuf := __w.(*bytes.Buffer)
		if !__isBuf {
			__buf = hamlet.GetBuffer()
			defer hamlet.ReleaseBuffer(__buf)
		}
		var __children hamlet.Template
		ctx, __children = hamlet.PopChildren(ctx)
		_ = __children
		if _, __err = __buf.WriteString("<!DOCTYPE html>\n"); __err != nil {
			return
		}
		if !__isBuf {
			_, __err = __w.Write(hamlet.NukeWhitespace(__buf.Bytes()))
		}
		return
	})
}

// Indenting the Hamlet code is optional; however, if you do or
// don't keep it consistent throughout the project
func DoctypeIndented() hamlet.Template {
	return hamlet.TemplateFunc(func(ctx context.Context, __w io.Writer) (__err error) {
		__buf, __isBuf := __w.(*bytes.Buffer)
		if !__isBuf {
			__buf = hamlet.GetBuffer()
			defer hamlet.ReleaseBuffer(__buf)
		}
		var __children hamlet.Template
		ctx, __children = hamlet.PopChildren(ctx)
		_ = __children
		if _, __err = __buf.WriteString("<!DOCTYPE html>\n"); __err != nil {
			return
		}
		if !__isBuf {
			_, __err = __w.Write(hamlet.NukeWhitespace(__buf.Bytes()))
		}
		return
	})
}
