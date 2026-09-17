package compiler_test

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/stackus/goht"
	"github.com/stackus/goht/compiler/testdata"
)

func TestGeneratedTypedTemplateAPI(t *testing.T) {
	template := testdata.TypedMultipleSlots().
		WithMainContent(goht.Fragment{}).
		WithSidebar2(goht.Fragment{})
	if template == nil {
		t.Fatal("fluent slot composition returned nil")
	}

	var output strings.Builder
	if err := testdata.TypedOneSlot().Slot("unknown", goht.Fragment{}).Render(context.Background(), &output); err == nil {
		t.Fatal("direct unknown slot assignment did not return a render-time error")
	}
}

func TestGeneratedChildrenAndNestedSlots(t *testing.T) {
	child := text("<span>child</span>\n")
	var output strings.Builder
	if err := testdata.ChildrenTest("passed").WithChildren(child).Render(context.Background(), &output); err != nil {
		t.Fatalf("WithChildren().Render() error = %v", err)
	}
	const wantChildren = "<div class=\"passed-in\">passed</div>\n<div class=\"children\">\n<span>child</span>\n</div>\n<div class=\"after\">After children</div>\n"
	if got := output.String(); got != wantChildren {
		t.Errorf("WithChildren().Render() = %q, want %q", got, wantChildren)
	}

	inner := testdata.SlotTest().WithFirst(text("inner"))
	output.Reset()
	if err := testdata.SlotTest().WithFirst(inner).Render(context.Background(), &output); err != nil {
		t.Fatalf("nested slot Render() error = %v", err)
	}
	const wantNested = "<div class=\"wrapper\">\n<div class=\"wrapper\">\ninner</div>\n</div>\n"
	if got := output.String(); got != wantNested {
		t.Errorf("nested slot Render() = %q, want %q", got, wantNested)
	}

	output.Reset()
	if err := testdata.SlotWithDefaultTest().WithFirst(text("<p>provided</p>\n")).Render(context.Background(), &output); err != nil {
		t.Fatalf("filled fallback Render() error = %v", err)
	}
	const wantFallback = "<div class=\"wrapper\">\n<p>provided</p>\n<p>Default second</p>\n<p>Default third</p>\n</div>\n"
	if got := output.String(); got != wantFallback {
		t.Errorf("filled fallback Render() = %q, want %q", got, wantFallback)
	}
}

func TestGeneratedSlotDirectives(t *testing.T) {
	items := []goht.Template{testdata.SlotDirectiveItem("one"), testdata.SlotDirectiveItem("two")}
	templates := []struct {
		name     string
		template goht.Template
		want     string
	}{
		{name: "haml", template: testdata.HamlSlotDirectives().WithContent(items...), want: "<div class=\"items\">\n<div class=\"item\">one</div>\n<div class=\"item\">two</div>\n</div>\n"},
		{name: "slim", template: testdata.SlimSlotDirectives().WithContent(items...), want: "<div class=\"items\"><div class=\"item\">one</div>\n<div class=\"item\">two</div>\n</div>\n"},
		{name: "ego", template: testdata.EgoSlotDirectives().WithContent(items...), want: "<div class=\"items\">\n\n\n<div class=\"item\">one</div>\n\n<div class=\"item\">two</div>\n\n\n</div>\n"},
		{name: "absent", template: testdata.HamlSlotDirectives(), want: "<div class=\"items\">\n</div>\n"},
		{name: "explicitly empty", template: testdata.HamlSlotDirectives().WithContent(), want: "<div class=\"items\">\n</div>\n"},
	}
	for _, tt := range templates {
		t.Run(tt.name, func(t *testing.T) {
			var output strings.Builder
			if err := tt.template.Render(context.Background(), &output); err != nil {
				t.Fatalf("Render() error = %v", err)
			}
			if got := output.String(); got != tt.want {
				t.Errorf("Render() = %q, want %q", got, tt.want)
			}
		})
	}

	for _, tt := range []struct {
		name     string
		template goht.Template
		want     string
	}{
		{name: "children absent", template: testdata.ChildrenPresence(), want: "<div class=\"container\"></div>\n"},
		{name: "children present", template: testdata.ChildrenPresence().WithChildren(), want: "<div class=\"container\"><div class=\"present\">children are present</div></div>\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var output strings.Builder
			if err := tt.template.Render(context.Background(), &output); err != nil {
				t.Fatalf("Render() error = %v", err)
			}
			if got := output.String(); got != tt.want {
				t.Errorf("Render() = %q, want %q", got, tt.want)
			}
		})
	}
}

func text(value string) goht.TemplateFunc {
	return func(_ context.Context, w io.Writer, _ goht.Slots) error {
		_, err := io.WriteString(w, value)
		return err
	}
}
