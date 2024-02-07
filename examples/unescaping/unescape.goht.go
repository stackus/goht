// Code generated by GoHT - DO NOT EDIT.
// https://github.com/stackus/goht

package main

import "context"
import "io"
import "github.com/stackus/goht"

// The text from variables are assumed to need HTML escaping. If you
// want to include raw HTML, or you have already processed the variable
// then you will need to unescape the text. This is accomplished with
// the `!` operator.
// Use care when using the unescape operator. You should only use it
// when you are sure that the text is safe. If you are unsure, then
// you should use the default escaping.

func UnescapeCode() goht.Template {
	return goht.TemplateFunc(func(ctx context.Context, __w io.Writer) (__err error) {
		__buf, __isBuf := __w.(goht.Buffer)
		if !__isBuf {
			__buf = goht.GetBuffer()
			defer goht.ReleaseBuffer(__buf)
		}
		var __children goht.Template
		ctx, __children = goht.PopChildren(ctx)
		_ = __children
		if _, __err = __buf.WriteString("<p>"); __err != nil {
			return
		}
		var __var1 string
		if __var1, __err = goht.CaptureErrors(goht.EscapeString("This is <em>not</em> unescaped HTML.")); __err != nil {
			return
		}
		if _, __err = __buf.WriteString(__var1); __err != nil {
			return
		}
		if _, __err = __buf.WriteString("</p>\n<p>"); __err != nil {
			return
		}
		var __var2 string
		if __var2, __err = goht.CaptureErrors("This <em>is</em> unescaped HTML."); __err != nil {
			return
		}
		if _, __err = __buf.WriteString(__var2); __err != nil {
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

// It can also affect the interpolated values.

func UnescapeInterpolation() goht.Template {
	return goht.TemplateFunc(func(ctx context.Context, __w io.Writer) (__err error) {
		__buf, __isBuf := __w.(goht.Buffer)
		if !__isBuf {
			__buf = goht.GetBuffer()
			defer goht.ReleaseBuffer(__buf)
		}
		var __children goht.Template
		ctx, __children = goht.PopChildren(ctx)
		_ = __children
		var html = "<em>is</em>"
		if _, __err = __buf.WriteString("<p>This "); __err != nil {
			return
		}
		var __var1 string
		if __var1, __err = goht.CaptureErrors(goht.EscapeString(html)); __err != nil {
			return
		}
		if _, __err = __buf.WriteString(__var1); __err != nil {
			return
		}
		if _, __err = __buf.WriteString(" HTML.</p>\n<p>This "); __err != nil {
			return
		}
		var __var2 string
		if __var2, __err = goht.CaptureErrors(html); __err != nil {
			return
		}
		if _, __err = __buf.WriteString(__var2); __err != nil {
			return
		}
		if _, __err = __buf.WriteString(" HTML.</p>\n"); __err != nil {
			return
		}
		if !__isBuf {
			_, __err = __w.Write(__buf.Bytes())
		}
		return
	})
}

// The plain text that you write into your Goht templates will not be
// altered by the addition of the `!` operator. It is expected that this
// text has already been HTML escaped properly.

func UnescapeText() goht.Template {
	return goht.TemplateFunc(func(ctx context.Context, __w io.Writer) (__err error) {
		__buf, __isBuf := __w.(goht.Buffer)
		if !__isBuf {
			__buf = goht.GetBuffer()
			defer goht.ReleaseBuffer(__buf)
		}
		var __children goht.Template
		ctx, __children = goht.PopChildren(ctx)
		_ = __children
		if _, __err = __buf.WriteString("<p>This <em>is</em> HTML.</p>\n<p>This <em>is</em> HTML.</p>\n"); __err != nil {
			return
		}
		if !__isBuf {
			_, __err = __w.Write(__buf.Bytes())
		}
		return
	})
}
