// Code generated by GoHT - DO NOT EDIT.
// https://github.com/stackus/goht

package formatting

import "bytes"
import "context"
import "io"
import "github.com/stackus/goht"

// Normally, only strings are allowed as the value printed using the
// interpolated value in the template. However, if you provide a format
// before the value that you want outputted then it will be used to
// convert and format the value into a string.

// See: https://pkg.go.dev/fmt

// Include the format that you want, immediately followed by a comma,
// and then the value that you want to format.

var intVar = 123
var floatVar = 123.456
var boolVar = true
var stringVar = "Hello"

func IntExample() goht.Template {
	return goht.TemplateFunc(func(ctx context.Context, __w io.Writer) (__err error) {
		__buf, __isBuf := __w.(*bytes.Buffer)
		if !__isBuf {
			__buf = goht.GetBuffer()
			defer goht.ReleaseBuffer(__buf)
		}
		var __children goht.Template
		ctx, __children = goht.PopChildren(ctx)
		_ = __children
		if _, __err = __buf.WriteString("<p>The integer is ("); __err != nil {
			return
		}
		var __var1 string
		if __var1, __err = goht.CaptureErrors(goht.EscapeString(goht.FormatString("%d", intVar))); __err != nil {
			return
		}
		if _, __err = __buf.WriteString(__var1); __err != nil {
			return
		}
		if _, __err = __buf.WriteString(").</p>\n<p>The integer is ("); __err != nil {
			return
		}
		var __var2 string
		if __var2, __err = goht.CaptureErrors(goht.EscapeString(goht.FormatString("%b", intVar))); __err != nil {
			return
		}
		if _, __err = __buf.WriteString(__var2); __err != nil {
			return
		}
		if _, __err = __buf.WriteString(") in binary.</p>\n<p>The integer is ("); __err != nil {
			return
		}
		var __var3 string
		if __var3, __err = goht.CaptureErrors(goht.EscapeString(goht.FormatString("%o", intVar))); __err != nil {
			return
		}
		if _, __err = __buf.WriteString(__var3); __err != nil {
			return
		}
		if _, __err = __buf.WriteString(") in octal.</p>\n<p>The integer is ("); __err != nil {
			return
		}
		var __var4 string
		if __var4, __err = goht.CaptureErrors(goht.EscapeString(goht.FormatString("%x", intVar))); __err != nil {
			return
		}
		if _, __err = __buf.WriteString(__var4); __err != nil {
			return
		}
		if _, __err = __buf.WriteString(") in hex.</p>\n<p>The integer is ("); __err != nil {
			return
		}
		var __var5 string
		if __var5, __err = goht.CaptureErrors(goht.EscapeString(goht.FormatString("%X", intVar))); __err != nil {
			return
		}
		if _, __err = __buf.WriteString(__var5); __err != nil {
			return
		}
		if _, __err = __buf.WriteString(") in hex with uppercase.</p>\n<p>The integer is ("); __err != nil {
			return
		}
		var __var6 string
		if __var6, __err = goht.CaptureErrors(goht.EscapeString(goht.FormatString("%c", intVar))); __err != nil {
			return
		}
		if _, __err = __buf.WriteString(__var6); __err != nil {
			return
		}
		if _, __err = __buf.WriteString(") as a character.</p>\n"); __err != nil {
			return
		}
		if !__isBuf {
			_, __err = __w.Write(goht.NukeWhitespace(__buf.Bytes()))
		}
		return
	})
}

func FloatExample() goht.Template {
	return goht.TemplateFunc(func(ctx context.Context, __w io.Writer) (__err error) {
		__buf, __isBuf := __w.(*bytes.Buffer)
		if !__isBuf {
			__buf = goht.GetBuffer()
			defer goht.ReleaseBuffer(__buf)
		}
		var __children goht.Template
		ctx, __children = goht.PopChildren(ctx)
		_ = __children
		if _, __err = __buf.WriteString("<p>The float is ("); __err != nil {
			return
		}
		var __var1 string
		if __var1, __err = goht.CaptureErrors(goht.EscapeString(goht.FormatString("%f", floatVar))); __err != nil {
			return
		}
		if _, __err = __buf.WriteString(__var1); __err != nil {
			return
		}
		if _, __err = __buf.WriteString(").</p>\n<p>The float is ("); __err != nil {
			return
		}
		var __var2 string
		if __var2, __err = goht.CaptureErrors(goht.EscapeString(goht.FormatString("%e", floatVar))); __err != nil {
			return
		}
		if _, __err = __buf.WriteString(__var2); __err != nil {
			return
		}
		if _, __err = __buf.WriteString(") in scientific notation.</p>\n<p>The float is ("); __err != nil {
			return
		}
		var __var3 string
		if __var3, __err = goht.CaptureErrors(goht.EscapeString(goht.FormatString("%.2f", floatVar))); __err != nil {
			return
		}
		if _, __err = __buf.WriteString(__var3); __err != nil {
			return
		}
		if _, __err = __buf.WriteString(") with 2 decimal places.</p>\n<p>The float is ("); __err != nil {
			return
		}
		var __var4 string
		if __var4, __err = goht.CaptureErrors(goht.EscapeString(goht.FormatString("%9.2f", floatVar))); __err != nil {
			return
		}
		if _, __err = __buf.WriteString(__var4); __err != nil {
			return
		}
		if _, __err = __buf.WriteString(") with 2 decimal places and padded to 9 characters.</p>\n<p>The float is ("); __err != nil {
			return
		}
		var __var5 string
		if __var5, __err = goht.CaptureErrors(goht.EscapeString(goht.FormatString("%-9.2f", floatVar))); __err != nil {
			return
		}
		if _, __err = __buf.WriteString(__var5); __err != nil {
			return
		}
		if _, __err = __buf.WriteString(") with 2 decimal places and padded to 9 characters and left aligned.</p>\n<p>The float is ("); __err != nil {
			return
		}
		var __var6 string
		if __var6, __err = goht.CaptureErrors(goht.EscapeString(goht.FormatString("%09.2f", floatVar))); __err != nil {
			return
		}
		if _, __err = __buf.WriteString(__var6); __err != nil {
			return
		}
		if _, __err = __buf.WriteString(") with 2 decimal places and padded to 9 characters with 0s.</p>\n"); __err != nil {
			return
		}
		if !__isBuf {
			_, __err = __w.Write(goht.NukeWhitespace(__buf.Bytes()))
		}
		return
	})
}

func BoolExample() goht.Template {
	return goht.TemplateFunc(func(ctx context.Context, __w io.Writer) (__err error) {
		__buf, __isBuf := __w.(*bytes.Buffer)
		if !__isBuf {
			__buf = goht.GetBuffer()
			defer goht.ReleaseBuffer(__buf)
		}
		var __children goht.Template
		ctx, __children = goht.PopChildren(ctx)
		_ = __children
		if _, __err = __buf.WriteString("<p>The bool is ("); __err != nil {
			return
		}
		var __var1 string
		if __var1, __err = goht.CaptureErrors(goht.EscapeString(goht.FormatString("%t", boolVar))); __err != nil {
			return
		}
		if _, __err = __buf.WriteString(__var1); __err != nil {
			return
		}
		if _, __err = __buf.WriteString(").</p>\n"); __err != nil {
			return
		}
		if !__isBuf {
			_, __err = __w.Write(goht.NukeWhitespace(__buf.Bytes()))
		}
		return
	})
}

func StringExample() goht.Template {
	return goht.TemplateFunc(func(ctx context.Context, __w io.Writer) (__err error) {
		__buf, __isBuf := __w.(*bytes.Buffer)
		if !__isBuf {
			__buf = goht.GetBuffer()
			defer goht.ReleaseBuffer(__buf)
		}
		var __children goht.Template
		ctx, __children = goht.PopChildren(ctx)
		_ = __children
		if _, __err = __buf.WriteString("<p>The string is ("); __err != nil {
			return
		}
		var __var1 string
		if __var1, __err = goht.CaptureErrors(goht.EscapeString(stringVar)); __err != nil {
			return
		}
		if _, __err = __buf.WriteString(__var1); __err != nil {
			return
		}
		if _, __err = __buf.WriteString("). Strings do not require any additional formatting.</p>\n<p>The string is ("); __err != nil {
			return
		}
		var __var2 string
		if __var2, __err = goht.CaptureErrors(goht.EscapeString(goht.FormatString("%q", stringVar))); __err != nil {
			return
		}
		if _, __err = __buf.WriteString(__var2); __err != nil {
			return
		}
		if _, __err = __buf.WriteString(") quoted.</p>\n<p>The string is ("); __err != nil {
			return
		}
		var __var3 string
		if __var3, __err = goht.CaptureErrors(goht.EscapeString(goht.FormatString("%x", stringVar))); __err != nil {
			return
		}
		if _, __err = __buf.WriteString(__var3); __err != nil {
			return
		}
		if _, __err = __buf.WriteString(") as hex.</p>\n<p>The string is ("); __err != nil {
			return
		}
		var __var4 string
		if __var4, __err = goht.CaptureErrors(goht.EscapeString(goht.FormatString("%X", stringVar))); __err != nil {
			return
		}
		if _, __err = __buf.WriteString(__var4); __err != nil {
			return
		}
		if _, __err = __buf.WriteString(") as hex with uppercase.</p>\n<p>The string is ("); __err != nil {
			return
		}
		var __var5 string
		if __var5, __err = goht.CaptureErrors(goht.EscapeString(goht.FormatString("%s", stringVar))); __err != nil {
			return
		}
		if _, __err = __buf.WriteString(__var5); __err != nil {
			return
		}
		if _, __err = __buf.WriteString(") as is.</p>\n<p>The string is ("); __err != nil {
			return
		}
		var __var6 string
		if __var6, __err = goht.CaptureErrors(goht.EscapeString(goht.FormatString("%.4s", stringVar))); __err != nil {
			return
		}
		if _, __err = __buf.WriteString(__var6); __err != nil {
			return
		}
		if _, __err = __buf.WriteString("), truncated to 4 characters.</p>\n<p>The string is ("); __err != nil {
			return
		}
		var __var7 string
		if __var7, __err = goht.CaptureErrors(goht.EscapeString(goht.FormatString("%6s", stringVar))); __err != nil {
			return
		}
		if _, __err = __buf.WriteString(__var7); __err != nil {
			return
		}
		if _, __err = __buf.WriteString("), padded to 6 characters.</p>\n<p>The string is ("); __err != nil {
			return
		}
		var __var8 string
		if __var8, __err = goht.CaptureErrors(goht.EscapeString(goht.FormatString("%6.4s", stringVar))); __err != nil {
			return
		}
		if _, __err = __buf.WriteString(__var8); __err != nil {
			return
		}
		if _, __err = __buf.WriteString("), truncated to 4 characters and padded to 6 characters.</p>\n<p>The string is ("); __err != nil {
			return
		}
		var __var9 string
		if __var9, __err = goht.CaptureErrors(goht.EscapeString(goht.FormatString("%6.4q", stringVar))); __err != nil {
			return
		}
		if _, __err = __buf.WriteString(__var9); __err != nil {
			return
		}
		if _, __err = __buf.WriteString("), truncated to 4 characters and padded to 6 characters and quoted.</p>\n"); __err != nil {
			return
		}
		if !__isBuf {
			_, __err = __w.Write(goht.NukeWhitespace(__buf.Bytes()))
		}
		return
	})
}