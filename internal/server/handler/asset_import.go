package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-logr/logr"

	"github.com/nikimonax/laretz/internal/app"
	"github.com/nikimonax/laretz/internal/app/usecase"
	"github.com/nikimonax/laretz/internal/domain"
	"github.com/nikimonax/laretz/internal/server/renderer"
)

type ImportAssetHandler struct {
	logger logr.Logger

	search              *usecase.SearchExternalAssetsUseCase
	assetGet            *usecase.GetAssetUseCase
	externalAssetsCount *usecase.CountExternalAssetsUseCase
	assetImport         *usecase.ImportAssetUseCase
}

// ServeHTTP implements [http.Handler].
func (handler *ImportAssetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		handler.renderSearch(w, r, "", "Invalid form")
		return
	}

	query := strings.TrimSpace(r.FormValue("query"))

	externalAssetID, err := strconv.ParseInt(strings.TrimSpace(r.FormValue("external_asset_id")), 10, 64)

	if err != nil || externalAssetID <= 0 {
		handler.renderSearch(w, r, query, "Invalid form")
		return
	}

	symbol := strings.TrimSpace(r.FormValue("symbol"))
	name := strings.TrimSpace(r.FormValue("name"))

	if symbol == "" {
		handler.renderSearch(w, r, query, "Symbol must not be empty.")
		return
	}

	if name == "" {
		handler.renderSearch(w, r, query, "Name must not be empty.")
		return
	}

	cmd := usecase.ImportAsset{
		ExternalAssetID: domain.ExternalAssetID(externalAssetID),
		Symbol:          symbol,
		Name:            name,
	}

	if err := handler.assetImport.Execute(r.Context(), cmd); err != nil {
		handler.logger.Error(err, "import asset")

		message := messageSomethingWentWrong

		if errors.Is(err, app.ErrEntityNotFound) {
			message = "External asset not found."
		}

		handler.renderSearch(w, r, query, message)
		return
	}

	http.Redirect(w, r, "/assets", http.StatusSeeOther)
}

func (handler *ImportAssetHandler) renderSearch(w http.ResponseWriter, r *http.Request, query string, message string) {
	responseRenderer := renderer.AssetSearch(
		handler.search,
		handler.assetGet,
		handler.externalAssetsCount,
		query,
		0,
	)

	if err := responseRenderer.Render(r.Context(), w, renderer.AssetSearchWithError(message)); err != nil {
		handler.logger.Error(err, "render")
	}
}

func NewImportAssetHandler(
	logger logr.Logger,
	search *usecase.SearchExternalAssetsUseCase,
	assetGet *usecase.GetAssetUseCase,
	externalAssetsCount *usecase.CountExternalAssetsUseCase,
	assetImport *usecase.ImportAssetUseCase,
) http.Handler {
	return &ImportAssetHandler{
		logger:              logger,
		search:              search,
		assetGet:            assetGet,
		externalAssetsCount: externalAssetsCount,
		assetImport:         assetImport,
	}
}
