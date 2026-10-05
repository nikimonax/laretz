package handler

import (
	"net/http"

	"github.com/go-logr/logr"

	"github.com/nikimonax/laretz/internal/server/renderer"
)

type PageHandler[PageData any] struct {
	logger           logr.Logger
	mutates          []renderer.Mutate[PageData]
	responseRenderer *renderer.ResponseRenderer[PageData]
}

// ServeHTTP implements [http.Handler].
func (handler *PageHandler[PageData]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := handler.responseRenderer.Render(r.Context(), w, handler.mutates...); err != nil {
		handler.logger.Error(err, "render")
	}
}

func NewPageHandler[PageData any](
	logger logr.Logger,
	responseRenderer *renderer.ResponseRenderer[PageData],
	mutates ...renderer.Mutate[PageData],
) http.Handler {
	return &PageHandler[PageData]{
		logger:           logger,
		mutates:          mutates,
		responseRenderer: responseRenderer,
	}
}
