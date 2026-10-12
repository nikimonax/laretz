package handler

import (
	"net/http"
	"strconv"

	"github.com/go-logr/logr"

	"github.com/nikimonax/laretz/internal/server/renderer"
)

func parsePage(r *http.Request, name string) int {
	page, err := strconv.Atoi(r.URL.Query().Get(name))

	if err != nil || page < 1 {
		return 1
	}

	return page
}

func renderNotFoundPage(logger logr.Logger, w http.ResponseWriter, r *http.Request) {
	if err := renderer.NotFound().WithStatus(http.StatusNotFound).Render(r.Context(), w); err != nil {
		logger.Error(err, "render")
	}
}
