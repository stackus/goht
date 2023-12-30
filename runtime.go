package hamlet

import (
	"bytes"
	"context"
	"errors"
	"html"
	"io"
	"strings"
	"sync"
)

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

func InitContext(ctx context.Context) context.Context {
	if _, ok := ctx.Value(ctxKey).(*ctxValue); ok {
		return ctx
	}
	value := new(ctxValue)
	return context.WithValue(ctx, ctxKey, value)
}

func GetBuffer() *bytes.Buffer {
	return bufferPool.Get().(*bytes.Buffer)
}

func ReleaseBuffer(buf *bytes.Buffer) {
	buf.Reset()
	bufferPool.Put(buf)
}

func PopChildren(ctx context.Context) (context.Context, Template) {
	value := ctx.Value(ctxKey).(*ctxValue)
	children := *value.children
	value.children = nil
	if children == nil {
		children = TemplateFunc(func(ctx context.Context, w io.Writer) error { return nil })
	}
	return ctx, children
}

func PushChildren(ctx context.Context, children Template) context.Context {
	value := ctx.Value(ctxKey).(*ctxValue)
	value.children = &children
	return ctx
}

func CaptureErrors(s string, errs ...error) (string, error) {
	return s, errors.Join(errs...)
}

func BuildClassList(classes ...any) (string, error) {
	var classList []string
	for _, class := range classes {
		switch class := class.(type) {
		case string:
			classList = append(classList, class)
		case []string:
			classList = append(classList, class...)
		case map[string]bool:
			for cls, ok := range class {
				if ok {
					classList = append(classList, cls)
				}
			}
		default:
			return "", errors.New("hamlet: invalid class type")
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
			return "", errors.New("hamlet: invalid attribute type")
		}
	}
	return attributeList.String(), nil
}

func EscapeString(s string) string {
	return html.EscapeString(s)
}
