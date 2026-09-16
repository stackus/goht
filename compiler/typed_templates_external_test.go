package compiler_test

import (
	"context"
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
