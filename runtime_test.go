package goht_test

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stackus/goht"
)

func TestSlotTemplateComposition(t *testing.T) {
	noSlots := goht.NewSlotTemplate(text("plain"))
	content := func(ctx context.Context, w io.Writer) error {
		slot := goht.GetSlot(ctx, "content")
		if slot == nil {
			_, err := io.WriteString(w, "fallback")
			return err
		}
		return slot.Render(ctx, w)
	}

	base := goht.NewSlotTemplate(content, "content")
	one := base.Slot("content", text("one"))
	multiple := base.Slot("content", text("one"), text("two"))
	replaced := multiple.Slot("content", text("replacement"))
	assigned := []goht.Template{text("assigned")}
	cloned := base.Slot("content", assigned...)
	assigned[0] = text("mutated")

	for _, tt := range []struct {
		name     string
		template goht.Template
		want     string
	}{
		{name: "no declared slots", template: noSlots, want: "plain"},
		{name: "no slots uses fallback", template: base, want: "fallback"},
		{name: "one slot", template: one, want: "one"},
		{name: "multiple templates retain order", template: multiple, want: "onetwo"},
		{name: "repeated assignment replaces", template: replaced, want: "replacement"},
		{name: "assigned list is copied", template: cloned, want: "assigned"},
		{name: "original remains unchanged", template: base, want: "fallback"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var out strings.Builder
			if err := tt.template.Render(context.Background(), &out); err != nil {
				t.Fatalf("Render() error = %v", err)
			}
			if got := out.String(); got != tt.want {
				t.Errorf("Render() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSlotTemplateNestedSlots(t *testing.T) {
	inner := goht.NewSlotTemplate(func(ctx context.Context, w io.Writer) error {
		return goht.GetSlot(ctx, "label").Render(ctx, w)
	}, "label").Slot("label", text("inner"))
	outer := goht.NewSlotTemplate(func(ctx context.Context, w io.Writer) error {
		return goht.GetSlot(ctx, "content").Render(ctx, w)
	}, "content").Slot("content", inner)

	var out strings.Builder
	if err := outer.Render(context.Background(), &out); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if got := out.String(); got != "inner" {
		t.Errorf("Render() = %q, want %q", got, "inner")
	}
}

func TestSlotTemplateUnknownSlotFailsBeforeRendering(t *testing.T) {
	base := goht.NewSlotTemplate(text("rendered"), "known")
	composed := base.Slot("unknown", text("ignored")).Slot("also-unknown", text("ignored"))

	var out strings.Builder
	err := composed.Render(context.Background(), &out)
	if err == nil || !strings.Contains(err.Error(), `unknown slot "unknown"`) || !strings.Contains(err.Error(), `unknown slot "also-unknown"`) {
		t.Fatalf("Render() error = %v, want accumulated unknown-slot errors", err)
	}
	if got := out.String(); got != "" {
		t.Errorf("Render() wrote %q, want no output", got)
	}
}

func TestFragment(t *testing.T) {
	boom := errors.New("boom")
	fragment := goht.Fragment{text("first"), failingTemplate{err: boom}, text("after")}

	var out strings.Builder
	err := fragment.Render(context.Background(), &out)
	if !errors.Is(err, boom) {
		t.Fatalf("Render() error = %v, want %v", err, boom)
	}
	if got := out.String(); got != "first" {
		t.Errorf("Render() = %q, want %q", got, "first")
	}

	out.Reset()
	if err := (goht.Fragment{}).Render(context.Background(), &out); err != nil {
		t.Fatalf("empty Fragment.Render() error = %v", err)
	}
	if got := out.String(); got != "" {
		t.Errorf("empty Fragment.Render() = %q, want empty", got)
	}
}

// generatedLayoutTemplate is a compile-time fixture for Task 002's generated
// shape: an exported runtime carrier is embedded while the generated type keeps
// fluent methods concrete and immutable.
type generatedLayoutTemplate struct {
	goht.SlotTemplate
}

func generatedLayout() *generatedLayoutTemplate {
	return &generatedLayoutTemplate{SlotTemplate: goht.NewSlotTemplate(func(ctx context.Context, w io.Writer) error {
		slot := goht.GetSlot(ctx, "content")
		if slot == nil {
			return text("fallback").Render(ctx, w)
		}
		return slot.Render(ctx, w)
	}, "content")}
}

func (t *generatedLayoutTemplate) Slot(name string, templates ...goht.Template) *generatedLayoutTemplate {
	copy := *t
	copy.SlotTemplate = t.SlotTemplate.Slot(name, templates...)
	return &copy
}

func (t *generatedLayoutTemplate) WithContent(templates ...goht.Template) *generatedLayoutTemplate {
	return t.Slot("content", templates...)
}

var _ goht.Template = (*generatedLayoutTemplate)(nil)

func TestGeneratedShapeFixture(t *testing.T) {
	base := generatedLayout()
	composed := base.WithContent(text("content"))

	for _, tt := range []struct {
		template goht.Template
		want     string
	}{
		{template: base, want: "fallback"},
		{template: composed, want: "content"},
	} {
		var out strings.Builder
		if err := tt.template.Render(context.Background(), &out); err != nil {
			t.Fatalf("Render() error = %v", err)
		}
		if got := out.String(); got != tt.want {
			t.Errorf("Render() = %q, want %q", got, tt.want)
		}
	}
}

func text(value string) goht.TemplateFunc {
	return func(_ context.Context, w io.Writer) error {
		_, err := io.WriteString(w, value)
		return err
	}
}

type failingTemplate struct {
	err error
}

func (t failingTemplate) Render(context.Context, io.Writer) error {
	return t.err
}
