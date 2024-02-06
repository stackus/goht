// Code generated by GoHT - DO NOT EDIT.
// https://github.com/stackus/goht

package main

import "bytes"
import "context"
import "io"
import "github.com/stackus/goht"

// The `:plain` filter can be used to display a large amount of text
// without any parsing. Lines may begin with Haml syntax and it will
// be ignored.
// Variable interpolation is still performed.

func PlainExample() goht.Template {
	return goht.TemplateFunc(func(ctx context.Context, __w io.Writer) (__err error) {
		__buf, __isBuf := __w.(*bytes.Buffer)
		if !__isBuf {
			__buf = goht.GetBuffer()
			defer goht.ReleaseBuffer(__buf)
		}
		var __children goht.Template
		ctx, __children = goht.PopChildren(ctx)
		_ = __children
		if _, __err = __buf.WriteString("<p>\nThis is plain text. It <pre>will</pre> be displayed as HTML.\n"); __err != nil {
			return
		}
		var __var1 string
		if __var1, __err = goht.CaptureErrors("This <pre>\"will\"</pre> be interpolated with HTML intact."); __err != nil {
			return
		}
		if _, __err = __buf.WriteString(__var1); __err != nil {
			return
		}
		if _, __err = __buf.WriteString("\n</p>\n"); __err != nil {
			return
		}
		if !__isBuf {
			_, __err = __w.Write(goht.NukeWhitespace(__buf.Bytes()))
		}
		return
	})
}

func EscapedExample() goht.Template {
	return goht.TemplateFunc(func(ctx context.Context, __w io.Writer) (__err error) {
		__buf, __isBuf := __w.(*bytes.Buffer)
		if !__isBuf {
			__buf = goht.GetBuffer()
			defer goht.ReleaseBuffer(__buf)
		}
		var __children goht.Template
		ctx, __children = goht.PopChildren(ctx)
		_ = __children
		if _, __err = __buf.WriteString("<p>\nThis is escaped text. It &lt;pre&gt;will not&lt;/pre&gt; be displayed as HTML.\n"); __err != nil {
			return
		}
		var __var1 string
		if __var1, __err = goht.CaptureErrors(goht.EscapeString("This <pre>\"will not\"</pre> be interpolated with HTML intact.")); __err != nil {
			return
		}
		if _, __err = __buf.WriteString(__var1); __err != nil {
			return
		}
		if _, __err = __buf.WriteString("\n</p>\n"); __err != nil {
			return
		}
		if !__isBuf {
			_, __err = __w.Write(goht.NukeWhitespace(__buf.Bytes()))
		}
		return
	})
}

func PreserveExample() goht.Template {
	return goht.TemplateFunc(func(ctx context.Context, __w io.Writer) (__err error) {
		__buf, __isBuf := __w.(*bytes.Buffer)
		if !__isBuf {
			__buf = goht.GetBuffer()
			defer goht.ReleaseBuffer(__buf)
		}
		var __children goht.Template
		ctx, __children = goht.PopChildren(ctx)
		_ = __children
		if _, __err = __buf.WriteString("<p>\nThis is preserved text. It <pre>will</pre> be displayed as HTML.&#x000A;"); __err != nil {
			return
		}
		var __var1 string
		if __var1, __err = goht.CaptureErrors("This <pre>\"will\"</pre> be interpolated with HTML intact."); __err != nil {
			return
		}
		if _, __err = __buf.WriteString(__var1); __err != nil {
			return
		}
		if _, __err = __buf.WriteString("&#x000A;</p>\n"); __err != nil {
			return
		}
		if !__isBuf {
			_, __err = __w.Write(goht.NukeWhitespace(__buf.Bytes()))
		}
		return
	})
}