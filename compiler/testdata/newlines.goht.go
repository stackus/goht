// Code generated by GoHT - DO NOT EDIT.
// https://github.com/stackus/goht

package testdata

import "context"
import "io"
import "github.com/stackus/goht"

func NewlinesTest() goht.Template {
	return goht.TemplateFunc(func(ctx context.Context, __w io.Writer) (__err error) {
		__buf, __isBuf := __w.(goht.Buffer)
		if !__isBuf {
			__buf = goht.GetBuffer()
			defer goht.ReleaseBuffer(__buf)
		}
		var __children goht.Template
		ctx, __children = goht.PopChildren(ctx)
		_ = __children
		if _, __err = __buf.WriteString("<div class=\"first\"></div>\n<div class=\"second\">Content</div>\n<div class=\"third\"></div>\n"); __err != nil {
			return
		}
		if !__isBuf {
			_, __err = __w.Write(__buf.Bytes())
		}
		return
	})
}
