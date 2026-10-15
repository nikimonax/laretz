package renderer

import (
	"context"
	"io"
)

type Renderer[PageData any] interface {
	Render(ctx context.Context, w io.Writer, data PageData) error
}

type RendererFunc[PageData any] func(ctx context.Context, w io.Writer, data PageData) error

func (f RendererFunc[PageData]) Render(ctx context.Context, w io.Writer, data PageData) error {
	return f(ctx, w, data)
}

type Provider[PageData any] interface {
	Provide(ctx context.Context) (PageData, error)
}

type ProviderFunc[PageData any] func(ctx context.Context) (PageData, error)

func (f ProviderFunc[PageData]) Provide(ctx context.Context) (PageData, error) {
	return f(ctx)
}

type Mutate[PageData any] = func(data *PageData)
