package goht

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html"
	"io"
	"regexp"
	"slices"
	"strings"
	"sync"
)

// Template is a template that can be rendered into a writer.
type Template interface {
	Render(ctx context.Context, w io.Writer, slottedTemplates ...SlottedTemplate) error
	Slot(slotName string, slottedTemplates ...SlottedTemplate) SlottedTemplate
}

type SlottedTemplate interface {
	Template
	SlotName() string
	SlottedTemplates() []SlottedTemplate
}

type slottedTemplate struct {
	slotName         string
	slottedTemplates []SlottedTemplate
}

type TemplateFunc func(ctx context.Context, w io.Writer, slottedTemplates ...SlottedTemplate) error

func (f TemplateFunc) Render(ctx context.Context, w io.Writer, slottedTemplates ...SlottedTemplate) error {
	return f(ctx, w, slottedTemplates...)
}

func (f TemplateFunc) Slot(slotName string, slottedTemplates ...SlottedTemplate) SlottedTemplate {
	return &slottedTemplate{
		slotName:         slotName,
		slottedTemplates: slottedTemplates,
	}
}

func (f *slottedTemplate) Render(ctx context.Context, w io.Writer, slottedTemplates ...SlottedTemplate) error {
	if err := f.Render(ctx, w, slottedTemplates...); err != nil {
		return err
	}
	return nil
}

func (f *slottedTemplate) Slot(slotName string, slottedTemplates ...SlottedTemplate) SlottedTemplate {
	return &slottedTemplate{
		slotName:         slotName,
		slottedTemplates: slottedTemplates,
	}
}

func (f *slottedTemplate) SlotName() string {
	return f.slotName
}

func (f *slottedTemplate) SlottedTemplates() []SlottedTemplate {
	return f.slottedTemplates
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

type Buffer struct {
	*bytes.Buffer
}

func (b *Buffer) Bytes() []byte {
	return nukeWhitespaceRe.ReplaceAll(b.Buffer.Bytes(), nil)
}

var bufferPool = sync.Pool{
	New: func() any {
		return Buffer{new(bytes.Buffer)}
	},
}

func GetBuffer() Buffer {
	return bufferPool.Get().(Buffer)
}

func ReleaseBuffer(buf Buffer) {
	buf.Reset()
	bufferPool.Put(buf)
}

func PopChildren(ctx context.Context) (context.Context, Template) {
	var value *ctxValue
	ctx, value = getContext(ctx)
	if value.children == nil {
		return ctx, TemplateFunc(func(ctx context.Context, w io.Writer, slottedTemplates ...SlottedTemplate) error { return nil })
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

func GetSlottedTemplate(slottedTemplates []SlottedTemplate, slotName string) SlottedTemplate {
	for _, st := range slottedTemplates {
		if st.SlotName() == slotName {
			return st
		}
	}
	return nil
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
			return "", fmt.Errorf("goht: invalid class type: %T", class)
		}
	}
	return strings.Join(classList, ` `), nil
}

func BuildAttributeList(attributes ...any) (string, error) {
	var attributeList []string
	for _, attribute := range attributes {
		switch attribute := attribute.(type) {
		case map[string]bool:
			for key, value := range attribute {
				if value {
					attributeList = append(attributeList, html.EscapeString(key))
				}
			}
		case map[string]string:
			for key, value := range attribute {
				attributeList = append(attributeList, html.EscapeString(key)+`="`+html.EscapeString(value)+`"`)
			}
		default:
			return "", fmt.Errorf("goht: invalid attribute type: %T", attribute)
		}
	}
	// for stable ordering of the attributes
	slices.Sort(attributeList)
	return strings.Join(attributeList, " "), nil
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
	ref, ok := obj.(ObjectIDer)
	if !ok {
		return ""
	}

	var s []string
	if len(prefix) > 0 {
		s = append(s, prefix[0])
	}
	if v, ok := obj.(ObjectClasser); ok {
		s = append(s, v.ObjectClass())
	}
	s = append(s, ref.ObjectID())
	return strings.Join(s, "_")
}

func ObjectClass(obj any, prefix ...string) string {
	ref, ok := obj.(ObjectClasser)
	if !ok {
		return ""
	}

	var s []string
	if len(prefix) > 0 {
		s = append(s, prefix[0])
	}
	s = append(s, ref.ObjectClass())
	return strings.Join(s, "_")
}
