package handler

import (
	"net/http"
	"strings"

	"github.com/go-logr/logr"

	"github.com/nikimonax/laretz/internal/app/usecase"
	"github.com/nikimonax/laretz/internal/domain"
	"github.com/nikimonax/laretz/internal/server/renderer"
	"github.com/nikimonax/laretz/internal/server/views/pages"
)

type SyncExternalAssetsHandler struct {
	logger logr.Logger

	responseRenderer   *renderer.ResponseRenderer[pages.AssetsData]
	externalAssetsSync *usecase.SyncExternalAssetsUseCase
}

// ServeHTTP implements [http.Handler].
func (handler *SyncExternalAssetsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		mutate := renderer.AssetsWithError("Invalid form", pages.DialogSyncAssets)
		if err := handler.responseRenderer.Render(r.Context(), w, mutate); err != nil {
			handler.logger.Error(err, "render")
		}
		return
	}

	sourceID := domain.SourceID(strings.TrimSpace(r.FormValue("source")))

	if sourceID == "" {
		mutate := renderer.AssetsWithError("Source must be selected.", pages.DialogSyncAssets)
		if err := handler.responseRenderer.Render(r.Context(), w, mutate); err != nil {
			handler.logger.Error(err, "render")
		}
		return
	}

	cmd := usecase.SyncExternalAssets{SourceID: sourceID}

	if err := handler.externalAssetsSync.Execute(r.Context(), cmd); err != nil {
		handler.logger.Error(err, "sync external assets")

		mutate := renderer.AssetsWithError(messageSyncFailed, pages.DialogSyncAssets)
		if err := handler.responseRenderer.Render(r.Context(), w, mutate); err != nil {
			handler.logger.Error(err, "render")
		}
		return
	}

	http.Redirect(w, r, "/assets", http.StatusSeeOther)
}

func NewSyncExternalAssetsHandler(
	logger logr.Logger,
	responseRenderer *renderer.ResponseRenderer[pages.AssetsData],
	externalAssetsSync *usecase.SyncExternalAssetsUseCase,
) http.Handler {
	return &SyncExternalAssetsHandler{
		logger:             logger,
		responseRenderer:   responseRenderer,
		externalAssetsSync: externalAssetsSync,
	}
}
