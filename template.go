package hamlet

import (
	"context"
	"io"
)

type Template interface {
	Render(ctx context.Context, w io.Writer) error
}

type TemplateFunc func(ctx context.Context, w io.Writer) error

func (f TemplateFunc) Render(ctx context.Context, w io.Writer) error {
	return f(ctx, w)
}
