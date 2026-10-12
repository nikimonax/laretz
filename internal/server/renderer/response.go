package renderer

import (
	"context"
	"net/http"
)

type ResponseRenderer[PageData any] struct {
	renderer Renderer[PageData]
	provider Provider[PageData]
	status   int
}

// WithStatus returns a copy of the renderer that responds with the given
// HTTP status code instead of 200.
func (rr *ResponseRenderer[PageData]) WithStatus(status int) *ResponseRenderer[PageData] {
	updated := *rr
	updated.status = status
	return &updated
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

	status := http.StatusOK

	if rr.status != 0 {
		status = rr.status
	}

	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	writer.WriteHeader(status)
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
