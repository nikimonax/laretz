package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-logr/logr"

	"github.com/nikimonax/laretz/internal/app/usecase"
	"github.com/nikimonax/laretz/internal/domain"
	"github.com/nikimonax/laretz/internal/server/renderer"
	"github.com/nikimonax/laretz/internal/server/views/pages"
)

type AssetSearchHandler struct {
	logger logr.Logger

	search              *usecase.SearchExternalAssetsUseCase
	assetGet            *usecase.GetAssetUseCase
	externalAssetsCount *usecase.CountExternalAssetsUseCase
}

// ServeHTTP implements [http.Handler].
func (handler *AssetSearchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("query"))

	var relinkID domain.AssetID

	if raw := r.URL.Query().Get("relink"); raw != "" {
		id, err := strconv.ParseInt(raw, 10, 64)

		if err != nil || id <= 0 {
			handler.renderSearch(w, r, query, 0, "Invalid relink parameter.")
			return
		}

		relinkID = domain.AssetID(id)
	}

	handler.renderSearch(w, r, query, relinkID, "")
}

func (handler *AssetSearchHandler) renderSearch(w http.ResponseWriter, r *http.Request, query string, relinkID domain.AssetID, message string) {
	responseRenderer := renderer.AssetSearch(
		handler.search,
		handler.assetGet,
		handler.externalAssetsCount,
		query,
		relinkID,
	)

	var mutates []renderer.Mutate[pages.AssetSearchData]

	if message != "" {
		mutates = append(mutates, renderer.AssetSearchWithError(message))
	}

	if err := responseRenderer.Render(r.Context(), w, mutates...); err != nil {
		handler.logger.Error(err, "render")
	}
}

func NewAssetSearchHandler(
	logger logr.Logger,
	search *usecase.SearchExternalAssetsUseCase,
	assetGet *usecase.GetAssetUseCase,
	externalAssetsCount *usecase.CountExternalAssetsUseCase,
) http.Handler {
	return &AssetSearchHandler{
		logger:              logger,
		search:              search,
		assetGet:            assetGet,
		externalAssetsCount: externalAssetsCount,
	}
}
