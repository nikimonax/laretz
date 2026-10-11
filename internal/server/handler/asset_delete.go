package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-logr/logr"

	"github.com/nikimonax/laretz/internal/app"
	"github.com/nikimonax/laretz/internal/app/usecase"
	"github.com/nikimonax/laretz/internal/domain"
	"github.com/nikimonax/laretz/internal/server/renderer"
	"github.com/nikimonax/laretz/internal/server/views/pages"
)

type DeleteAssetHandler struct {
	logger logr.Logger

	assetsRenderer *renderer.ResponseRenderer[pages.AssetsData]
	assetDelete    *usecase.DeleteAssetUseCase
}

// ServeHTTP implements [http.Handler].
func (handler *DeleteAssetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	assetID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)

	if err != nil {
		handler.renderAssets(w, r, "Invalid path parameter.", "")
		return
	}

	cmd := usecase.DeleteAsset{ID: domain.AssetID(assetID)}

	if err := handler.assetDelete.Execute(r.Context(), cmd); err != nil {
		message := messageSomethingWentWrong

		if errors.Is(err, app.ErrEntityNotFound) {
			message = "Asset not found."
		}

		if errors.Is(err, app.ErrAssetInUse) {
			message = "Asset has operations and cannot be deleted."
		}

		handler.renderAssets(w, r, message, deleteDialog(assetID))
		return
	}

	http.Redirect(w, r, "/assets", http.StatusSeeOther)
}

func (handler *DeleteAssetHandler) renderAssets(w http.ResponseWriter, r *http.Request, message string, dialog string) {
	if err := handler.assetsRenderer.Render(r.Context(), w, renderer.AssetsWithError(message, dialog)); err != nil {
		handler.logger.Error(err, "render")
	}
}

func deleteDialog(assetID int64) string {
	return fmt.Sprintf(pages.DialogDeleteAssetFormat, assetID)
}

func NewDeleteAssetHandler(
	logger logr.Logger,
	assetsRenderer *renderer.ResponseRenderer[pages.AssetsData],
	assetDelete *usecase.DeleteAssetUseCase,
) http.Handler {
	return &DeleteAssetHandler{
		logger:         logger,
		assetsRenderer: assetsRenderer,
		assetDelete:    assetDelete,
	}
}
