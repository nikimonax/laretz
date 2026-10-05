package renderer

import (
	"context"
	"net/http"
)

type ResponseRenderer[PageData any] struct {
	renderer Renderer[PageData]
	provider Provider[PageData]
}

func (rr *ResponseRenderer[PageData]) Render(
	ctx context.Context,
	writer http.ResponseWriter,
	mutates ...Mutate[PageData],
) error {
	data := rr.provider.Provide(ctx)

	for _, mutate := range mutates {
		mutate(&data)
	}

	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	return rr.renderer.Render(ctx, writer, data)
}

func NewResponseRenderer[PageData any](
	renderer Renderer[PageData],
	provider Provider[PageData],
) *ResponseRenderer[PageData] {
	return &ResponseRenderer[PageData]{
		renderer: renderer,
		provider: provider,
	}
}
