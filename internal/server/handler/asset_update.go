package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-logr/logr"

	"github.com/nikimonax/laretz/internal/app"
	"github.com/nikimonax/laretz/internal/app/usecase"
	"github.com/nikimonax/laretz/internal/domain"
	"github.com/nikimonax/laretz/internal/server/renderer"
	"github.com/nikimonax/laretz/internal/server/views/pages"
)

type UpdateAssetHandler struct {
	logger logr.Logger

	assetsRenderer      *renderer.ResponseRenderer[pages.AssetsData]
	search              *usecase.SearchExternalAssetsUseCase
	assetGet            *usecase.GetAssetUseCase
	externalAssetsCount *usecase.CountExternalAssetsUseCase
	assetUpdate         *usecase.UpdateAssetUseCase
}

// ServeHTTP implements [http.Handler].
func (handler *UpdateAssetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	assetID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)

	if err != nil {
		handler.renderAssets(w, r, "Invalid path parameter.", "")
		return
	}

	if err := r.ParseForm(); err != nil {
		handler.renderAssets(w, r, "Invalid form", renameDialog(assetID))
		return
	}

	cmd := usecase.UpdateAsset{ID: domain.AssetID(assetID)}

	symbol := strings.TrimSpace(r.FormValue("symbol"))
	name := strings.TrimSpace(r.FormValue("name"))

	if r.Form.Has("symbol") || r.Form.Has("name") {
		if symbol == "" || name == "" {
			handler.renderAssets(w, r, "Symbol and name must not be empty.", renameDialog(assetID))
			return
		}

		cmd.Symbol, cmd.Name = &symbol, &name
	}

	if raw := strings.TrimSpace(r.FormValue("external_asset_id")); raw != "" {
		externalAssetID, err := strconv.ParseInt(raw, 10, 64)

		if err != nil || externalAssetID <= 0 {
			handler.renderSearchRelink(w, r, assetID, "Invalid form")
			return
		}

		cmd.ExternalAssetID = &externalAssetID
	}

	if err := handler.assetUpdate.Execute(r.Context(), cmd); err != nil {
		handler.logger.Error(err, "update asset")

		message := messageSomethingWentWrong

		if errors.Is(err, app.ErrEntityNotFound) {
			message = messageAssetNotFound
		}

		if errors.Is(err, app.ErrSourceMismatch) {
			message = "Cannot re-link to a different source."
		}

		if r.Form.Has("relink") {
			handler.renderSearchRelink(w, r, assetID, message)
			return
		}

		handler.renderAssets(w, r, message, renameDialog(assetID))
		return
	}

	http.Redirect(w, r, "/assets", http.StatusSeeOther)
}

func (handler *UpdateAssetHandler) renderAssets(w http.ResponseWriter, r *http.Request, message string, dialog string) {
	if err := handler.assetsRenderer.Render(r.Context(), w, renderer.AssetsWithError(message, dialog)); err != nil {
		handler.logger.Error(err, "render")
	}
}

func (handler *UpdateAssetHandler) renderSearchRelink(w http.ResponseWriter, r *http.Request, assetID int64, message string) {
	query := strings.TrimSpace(r.FormValue("query"))

	responseRenderer := renderer.AssetSearch(
		handler.search,
		handler.assetGet,
		handler.externalAssetsCount,
		query,
		domain.AssetID(assetID),
	)

	if err := responseRenderer.Render(r.Context(), w, renderer.AssetSearchWithError(message)); err != nil {
		handler.logger.Error(err, "render")
	}
}

func renameDialog(assetID int64) string {
	return fmt.Sprintf(pages.DialogRenameAssetFormat, assetID)
}

func NewUpdateAssetHandler(
	logger logr.Logger,
	assetsRenderer *renderer.ResponseRenderer[pages.AssetsData],
	search *usecase.SearchExternalAssetsUseCase,
	assetGet *usecase.GetAssetUseCase,
	externalAssetsCount *usecase.CountExternalAssetsUseCase,
	assetUpdate *usecase.UpdateAssetUseCase,
) http.Handler {
	return &UpdateAssetHandler{
		logger:              logger,
		assetsRenderer:      assetsRenderer,
		search:              search,
		assetGet:            assetGet,
		externalAssetsCount: externalAssetsCount,
		assetUpdate:         assetUpdate,
	}
}
