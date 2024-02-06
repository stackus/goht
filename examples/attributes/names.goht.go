// Code generated by GoHT - DO NOT EDIT.
// https://github.com/stackus/goht

package main

import "bytes"
import "context"
import "io"
import "github.com/stackus/goht"

// For most attribute names you can include the name in the list
// of attributes just as you expect it to appear in the HTML. Names
// that contain alphanumeric characters, dashes (-), and
// underscores (_) are all acceptable as-is.

func SimpleNames() goht.Template {
	return goht.TemplateFunc(func(ctx context.Context, __w io.Writer) (__err error) {
		__buf, __isBuf := __w.(*bytes.Buffer)
		if !__isBuf {
			__buf = goht.GetBuffer()
			defer goht.ReleaseBuffer(__buf)
		}
		var __children goht.Template
		ctx, __children = goht.PopChildren(ctx)
		_ = __children
		if _, __err = __buf.WriteString("<a href=\"https://github.com/stackus/goht\" data-foo=\"bar\" odd_name=\"baz\" _=\"I&#39;m a _hyperscript attribute!\">Goht</a>\n"); __err != nil {
			return
		}
		if !__isBuf {
			_, __err = __w.Write(goht.NukeWhitespace(__buf.Bytes()))
		}
		return
	})
}

// For more complex names, such as data attributes, you can use
// enclose the name in in double quotes or backticks.
// - Names that start with an at sign (@).
// - Names that contain a colon (:).
// - Names that contain a question mark (?).
// The names will be rendered into the HTML without the quotes.

func ComplexNames() goht.Template {
	return goht.TemplateFunc(func(ctx context.Context, __w io.Writer) (__err error) {
		__buf, __isBuf := __w.(*bytes.Buffer)
		if !__isBuf {
			__buf = goht.GetBuffer()
			defer goht.ReleaseBuffer(__buf)
		}
		var __children goht.Template
		ctx, __children = goht.PopChildren(ctx)
		_ = __children
		if _, __err = __buf.WriteString("<a href=\"https://github.com/stackus/goht\" :class=\"show ? &#39;&#39; : &#39;hidden&#39;\" @click=\"show = !show\">Goht</a>\n"); __err != nil {
			return
		}
		if !__isBuf {
			_, __err = __w.Write(goht.NukeWhitespace(__buf.Bytes()))
		}
		return
	})
}