// Code generated by GoHT - DO NOT EDIT.
// https://github.com/stackus/goht

package example

import "bytes"
import "context"
import "io"
import "github.com/stackus/goht"

// Package example is an examples package for the Goht language.

func Doc() goht.Template {
	return goht.TemplateFunc(func(ctx context.Context, __w io.Writer) (__err error) {
		__buf, __isBuf := __w.(*bytes.Buffer)
		if !__isBuf {
			__buf = goht.GetBuffer()
			defer goht.ReleaseBuffer(__buf)
		}
		var __children goht.Template
		ctx, __children = goht.PopChildren(ctx)
		_ = __children
		if _, __err = __buf.WriteString("<div class=\"doc\">An example of package documentation.</div>\n"); __err != nil {
			return
		}
		if !__isBuf {
			_, __err = __w.Write(goht.NukeWhitespace(__buf.Bytes()))
		}
		return
	})
}
