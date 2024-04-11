// Code generated by GoHT - DO NOT EDIT.
// https://github.com/stackus/goht

package commands

import "context"
import "io"
import "github.com/stackus/goht"

// Any template can be included into another template; assuming that
// you have not created a circular reference, this will not cause the
// compiler to loop but will instead cause the generated code to
// run into errors.
// When you are creating a template that will be included into another
// template, you can use the `@children` command to indicate where any
// optional content from the calling template should be inserted.
// The `@children` command is used in combination with the rendering
// code syntax `=`.

func ChildrenExample() goht.Template {
	return goht.TemplateFunc(func(ctx context.Context, __w io.Writer) (__err error) {
		__buf, __isBuf := __w.(goht.Buffer)
		if !__isBuf {
			__buf = goht.GetBuffer()
			defer goht.ReleaseBuffer(__buf)
		}
		var __children goht.Template
		ctx, __children = goht.PopChildren(ctx)
		_ = __children
		if _, __err = __buf.WriteString("<p>\nThe following was passed in from the calling template:\n"); __err != nil {
			return
		}
		if __err = __children.Render(ctx, __buf); __err != nil {
			return
		}
		if _, __err = __buf.WriteString("</p>\n"); __err != nil {
			return
		}
		if !__isBuf {
			_, __err = __w.Write(__buf.Bytes())
		}
		return
	})
}
