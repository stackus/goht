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
	Render(ctx context.Context, w io.Writer) error
}

type TemplateFunc func(ctx context.Context, w io.Writer) error

func (f TemplateFunc) Render(ctx context.Context, w io.Writer) error {
	return f(ctx, w)
}

// Fragment renders templates in slice order.
type Fragment []Template

func (f Fragment) Render(ctx context.Context, w io.Writer) error {
	for _, template := range f {
		if err := template.Render(ctx, w); err != nil {
			return err
		}
	}
	return nil
}

// SlotTemplate is an immutable template value with named template lists.
// Slot returns a copied value with a named list replaced.
type SlotTemplate struct {
	template TemplateFunc
	declared map[string]struct{}
	slots    map[string]Fragment
	err      error
}

// NewSlotTemplate creates a template that accepts the provided slot names.
func NewSlotTemplate(template TemplateFunc, slotNames ...string) SlotTemplate {
	declared := make(map[string]struct{}, len(slotNames))
	for _, slotName := range slotNames {
		declared[slotName] = struct{}{}
	}
	return SlotTemplate{template: template, declared: declared}
}

// Render fails before rendering when composition contains an unknown slot.
func (t SlotTemplate) Render(ctx context.Context, w io.Writer) error {
	if t.err != nil {
		return t.err
	}
	return t.template.Render(context.WithValue(ctx, slotContextKey{}, t.slots), w)
}

// Slot returns a copied template with slotName replaced by templates. An
// undeclared slot is recorded and returned from Render before any output.
func (t SlotTemplate) Slot(slotName string, templates ...Template) SlotTemplate {
	if _, ok := t.declared[slotName]; !ok {
		t.err = errors.Join(t.err, fmt.Errorf("goht: unknown slot %q", slotName))
		return t
	}

	slots := make(map[string]Fragment, len(t.slots)+1)
	for name, fragment := range t.slots {
		slots[name] = slices.Clone(fragment)
	}
	slots[slotName] = slices.Clone(templates)
	t.slots = slots
	return t
}

type slotContextKey struct{}

// GetSlot returns a copy of the templates assigned to slotName in the current
// slot template render. It returns nil when the slot is absent.
func GetSlot(ctx context.Context, slotName string) Fragment {
	slots, _ := ctx.Value(slotContextKey{}).(map[string]Fragment)
	return slices.Clone(slots[slotName])
}

// little nuke alligators that eat whitespace; silly but important
const (
	NukeAfter  = "~☢<"
	NukeBefore = ">☢~"
)

var nukeWhitespaceRe = regexp.MustCompile(NukeAfter + `\s*|\s*` + NukeBefore)

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
