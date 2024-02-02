// Code generated by hamlet - DO NOT EDIT.

package example

import "bytes"
import "context"
import "io"
import "github.com/stackus/hamlet"

// You can include Go code to handle conditional and loop statements
// in your Hamlet templates.

var isAdmin = true

func Conditional() hamlet.Template {
	return hamlet.TemplateFunc(func(ctx context.Context, __w io.Writer) (__err error) {
		__buf, __isBuf := __w.(*bytes.Buffer)
		if !__isBuf {
			__buf = hamlet.GetBuffer()
			defer hamlet.ReleaseBuffer(__buf)
		}
		var __children hamlet.Template
		ctx, __children = hamlet.PopChildren(ctx)
		_ = __children
		if _, __err = __buf.WriteString("<div class=\"actions\">\n"); __err != nil {
			return
		}
		if isAdmin {
			if _, __err = __buf.WriteString("<button>~☢<\nEdit content\n>☢~</button>\n"); __err != nil {
				return
			}
		} else {
			if _, __err = __buf.WriteString("<button>Login</button>\n"); __err != nil {
				return
			}
		}
		if _, __err = __buf.WriteString("<button>View content</button>\n</div>\n"); __err != nil {
			return
		}
		if !__isBuf {
			_, __err = __w.Write(hamlet.NukeWhitespace(__buf.Bytes()))
		}
		return
	})
}

// However, we are using Haml and so we're into shortcuts. We can
// continue to write out the brackets or we can use the shorthand
// syntax.
// The short hand form simply drops the opening and closing brackets
// for statements that wrap around children.
// Shorthand statements include:
// for, if, else, else if, switch

func ShorthandConditional() hamlet.Template {
	return hamlet.TemplateFunc(func(ctx context.Context, __w io.Writer) (__err error) {
		__buf, __isBuf := __w.(*bytes.Buffer)
		if !__isBuf {
			__buf = hamlet.GetBuffer()
			defer hamlet.ReleaseBuffer(__buf)
		}
		var __children hamlet.Template
		ctx, __children = hamlet.PopChildren(ctx)
		_ = __children
		if _, __err = __buf.WriteString("<div class=\"actions\">\n"); __err != nil {
			return
		}
		if isAdmin {
			if _, __err = __buf.WriteString("<button>~☢<\nEdit content\n>☢~</button>\n"); __err != nil {
				return
			}
		} else if _, __err = __buf.WriteString("<button>Login</button>\n"); __err != nil {
			return
		}
		if _, __err = __buf.WriteString("<button>View content</button>\n</div>\n"); __err != nil {
			return
		}
		if !__isBuf {
			_, __err = __w.Write(hamlet.NukeWhitespace(__buf.Bytes()))
		}
		return
	})
}

// With a switch statement, we can use the case and default keywords
// but we will need to nest these statements if we're using the
// shorthand syntax. (win some, lose some)

func ShorthandSwitch() hamlet.Template {
	return hamlet.TemplateFunc(func(ctx context.Context, __w io.Writer) (__err error) {
		__buf, __isBuf := __w.(*bytes.Buffer)
		if !__isBuf {
			__buf = hamlet.GetBuffer()
			defer hamlet.ReleaseBuffer(__buf)
		}
		var __children hamlet.Template
		ctx, __children = hamlet.PopChildren(ctx)
		_ = __children
		if _, __err = __buf.WriteString("<div class=\"actions\">\n"); __err != nil {
			return
		}
		switch isAdmin {
		case true:
			if _, __err = __buf.WriteString("<button>~☢<\nEdit content\n>☢~</button>\n"); __err != nil {
				return
			}
		case false:
			if _, __err = __buf.WriteString("<button>Login</button>\n"); __err != nil {
				return
			}
		}
		if _, __err = __buf.WriteString("<button>View content</button>\n</div>\n"); __err != nil {
			return
		}
		if !__isBuf {
			_, __err = __w.Write(hamlet.NukeWhitespace(__buf.Bytes()))
		}
		return
	})
}
