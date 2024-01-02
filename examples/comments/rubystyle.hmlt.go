// Code generated by hamlet - DO NOT EDIT.

package main

import (
	"bytes"
	"context"
	"github.com/stackus/hamlet"
	"io"
)

// You may use ruby style comments to completely remove a line or
// even a block of nested elements.
// This is accomplished by using the `-#` syntax.
// This is useful for removing elements that are only used for
// documentation purposes.
// The nested elements commented out with this syntax will be
// not the parsed by the compiler and will not be included in the
// output.

func RubyStyle() hamlet.Template {
	return hamlet.TemplateFunc(func(ctx context.Context, __w io.Writer) (__err error) {
		__buf, __isBuf := __w.(*bytes.Buffer)
		if !__isBuf {
			__buf = hamlet.GetBuffer()
			defer hamlet.ReleaseBuffer(__buf)
		}
		var __children hamlet.Template
		ctx, __children = hamlet.PopChildren(ctx)
		_ = __children
		if _, __err = __buf.WriteString("<p>This is the only paragraph in the output.</p>\n"); __err != nil {
			return
		}
		if !__isBuf {
			_, __err = __w.Write(hamlet.NukeWhitespace(__buf.Bytes()))
		}
		return
	})
}
