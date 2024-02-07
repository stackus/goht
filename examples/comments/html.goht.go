// Code generated by GoHT - DO NOT EDIT.
// https://github.com/stackus/goht

package main

import "context"
import "io"
import "github.com/stackus/goht"

// HTML comments can be included and will be added to the rendered
// output.
// HTML comments are added using the forward slash at the beginning
// of the line

func HtmlComments() goht.Template {
	return goht.TemplateFunc(func(ctx context.Context, __w io.Writer) (__err error) {
		__buf, __isBuf := __w.(goht.Buffer)
		if !__isBuf {
			__buf = goht.GetBuffer()
			defer goht.ReleaseBuffer(__buf)
		}
		var __children goht.Template
		ctx, __children = goht.PopChildren(ctx)
		_ = __children
		if _, __err = __buf.WriteString("<p>This is a paragraph</p>\n<!--This is a HTML comment-->\n"); __err != nil {
			return
		}
		if !__isBuf {
			_, __err = __w.Write(__buf.Bytes())
		}
		return
	})
}

// You may also use them to comment out nested elements. This does
// not stop the nested elements from being parsed, just from being
// displayed.

func HtmlCommentsNested() goht.Template {
	return goht.TemplateFunc(func(ctx context.Context, __w io.Writer) (__err error) {
		__buf, __isBuf := __w.(goht.Buffer)
		if !__isBuf {
			__buf = goht.GetBuffer()
			defer goht.ReleaseBuffer(__buf)
		}
		var __children goht.Template
		ctx, __children = goht.PopChildren(ctx)
		_ = __children
		if _, __err = __buf.WriteString("<p>This is a paragraph</p>\n<!--\n<p>This is a paragraph that is commented out</p>\n-->\n"); __err != nil {
			return
		}
		if !__isBuf {
			_, __err = __w.Write(__buf.Bytes())
		}
		return
	})
}
