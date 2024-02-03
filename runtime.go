package goht

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html"
	"io"
	"regexp"
	"strings"
	"sync"
)

// Template is a template that can be rendered into a writer.
type Template interface {
	Render(ctx context.Context, w io.Writer) error
}

type TemplateFunc func(ctx context.Context, w io.Writer) error

func (f TemplateFunc) Render(ctx context.Context, w io.Writer) error {
	return f(ctx, w)
}

// little nuke alligators that eat whitespace; silly but important
const (
	NukeAfter  = "~☢<"
	NukeBefore = ">☢~"
)

var nukeWhitespaceRe = regexp.MustCompile(NukeAfter + `\s*|\s*` + NukeBefore)

type contextKey int

const (
	ctxKey contextKey = iota
)

type ctxValue struct {
	children *Template
}

var bufferPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

func GetBuffer() *bytes.Buffer {
	return bufferPool.Get().(*bytes.Buffer)
}

func ReleaseBuffer(buf *bytes.Buffer) {
	buf.Reset()
	bufferPool.Put(buf)
}

func PopChildren(ctx context.Context) (context.Context, Template) {
	var value *ctxValue
	ctx, value = getContext(ctx)
	if value.children == nil {
		return ctx, TemplateFunc(func(ctx context.Context, w io.Writer) error { return nil })
	}
	children := *value.children
	value.children = nil
	return ctx, children
}

func PushChildren(ctx context.Context, children Template) context.Context {
	value := ctx.Value(ctxKey).(*ctxValue)
	value.children = &children
	return ctx
}

func initContext(ctx context.Context) context.Context {
	if _, ok := ctx.Value(ctxKey).(*ctxValue); ok {
		return ctx
	}
	value := &ctxValue{
		children: nil,
	}
	return context.WithValue(ctx, ctxKey, value)
}

func getContext(ctx context.Context) (context.Context, *ctxValue) {
	value, ok := ctx.Value(ctxKey).(*ctxValue)
	if !ok {
		ctx = initContext(ctx)
		value = ctx.Value(ctxKey).(*ctxValue)
	}
	return ctx, value
}

func CaptureErrors(s string, errs ...error) (string, error) {
	return s, errors.Join(errs...)
}

func BuildClassList(classes ...any) (string, error) {
	var classList []string
	for _, class := range classes {
		switch class := class.(type) {
		case string:
			if class == "" {
				continue
			}
			classList = append(classList, class)
		case []string:
			classList = append(classList, class...)
		case map[string]bool:
			for cls, ok := range class {
				if ok {
					if cls == "" {
						continue
					}
					classList = append(classList, cls)
				}
			}
		default:
			return "", errors.New("goht: invalid class type")
		}
	}
	return strings.Join(classList, ` `), nil
}

func BuildAttributeList(attributes ...any) (string, error) {
	attributeList := strings.Builder{}
	for _, attribute := range attributes {
		switch attribute := attribute.(type) {
		case map[string]bool:
			for key, value := range attribute {
				if value {
					attributeList.WriteString(` ` + html.EscapeString(key))
				}
			}
		case map[string]string:
			for key, value := range attribute {
				attributeList.WriteString(` ` + html.EscapeString(key) + `=\"` + html.EscapeString(value) + `\"`)
			}
		default:
			return "", errors.New("goht: invalid attribute type")
		}
	}
	return attributeList.String(), nil
}

func EscapeString(s string) string {
	return html.EscapeString(s)
}

func FormatString(format string, value any) string {
	return fmt.Sprintf(format, value)
}

type ObjectIDer interface {
	ObjectID() string
}

type ObjectClasser interface {
	ObjectClass() string
}

func ObjectID(obj any, prefix ...string) string {
	s := ""
	if len(prefix) > 0 {
		s = prefix[0]
	}
	if v, ok := obj.(ObjectClasser); ok {
		s = s + "_" + v.ObjectClass()
	}
	if v, ok := obj.(ObjectIDer); ok {
		return s + "_" + v.ObjectID()
	}
	return ""
}

func ObjectClass(obj any, prefix ...string) string {
	s := ""
	if len(prefix) > 0 {
		s = prefix[0]
	}
	if v, ok := obj.(ObjectClasser); ok {
		return s + "_" + v.ObjectClass()
	}
	return ""
}

// NukeWhitespace removes whitespace between tags.
//
// Puts those little nuke alligators to work.
func NukeWhitespace(b []byte) []byte {
	return nukeWhitespaceRe.ReplaceAll(b, nil)
}
