package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-logr/logr"

	"github.com/nikimonax/laretz/internal/lib/spoof"
	mw "github.com/nikimonax/laretz/internal/server/middleware"
	"github.com/nikimonax/laretz/internal/server/static"
	"github.com/nikimonax/laretz/internal/server/views/pages"
)

func New(logger logr.Logger) http.Handler {
	router := chi.NewRouter()
	router.Use(mw.RequestLog(logger), spoof.Middleware, mw.OriginCheck)

	static.Mount(router)
	router.Get("/", home(logger))

	return router
}

func home(logger logr.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		err := pages.Home().Render(r.Context(), w)
		if err != nil {
			http.Error(w, "render failed", http.StatusInternalServerError)
			logger.Error(err, "render home page")
		}
	}
}
