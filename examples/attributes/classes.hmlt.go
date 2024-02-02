// Code generated by hamlet - DO NOT EDIT.

package main

import "bytes"
import "context"
import "io"
import "github.com/stackus/hamlet"

// The class attribute is a bit special. You will often find yourself
// working with not one, but several classes. This is why repeating
// use of the class `.` operator is allowed. You may also run into
// situations where you want to add a class conditionally or want to
// assign a number of classes all at once.
// If you provide a dynamic value to the class attribute, it will be
// interpreted as a parameter list.
// The types of parameters that are allowed are:
// - `string` - will be added as a class if it is not blank
// - `[]string` - each non-blank string will be added as a class
// - `map[string]bool` - each key with a true value will be added
// If you have any dynamic sources for a class, from an object
// reference, or from the class attribute, they will be merged and
// deduplicated.
// If you have all static values for your classes, then they are
// rendered as-is avoiding any extra processing.

var myClassList = []string{"foo", "bar"}
var myOptionalClasses = map[string]bool{
	"baz": true,
	"qux": false,
}

func Classes() hamlet.Template {
	return hamlet.TemplateFunc(func(ctx context.Context, __w io.Writer) (__err error) {
		__buf, __isBuf := __w.(*bytes.Buffer)
		if !__isBuf {
			__buf = hamlet.GetBuffer()
			defer hamlet.ReleaseBuffer(__buf)
		}
		var __children hamlet.Template
		ctx, __children = hamlet.PopChildren(ctx)
		_ = __children
		if _, __err = __buf.WriteString("<p"); __err != nil {
			return
		}
		var __var1 string
		__var1, __err = hamlet.BuildClassList("fizz", "buzz", myClassList, myOptionalClasses)
		if __err != nil {
			return
		}
		if _, __err = __buf.WriteString(" class=\"" + __var1 + "\""); __err != nil {
			return
		}
		if _, __err = __buf.WriteString(">\n</p>\n"); __err != nil {
			return
		}
		if !__isBuf {
			_, __err = __w.Write(hamlet.NukeWhitespace(__buf.Bytes()))
		}
		return
	})
}
