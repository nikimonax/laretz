package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-logr/logr"

	"github.com/nikimonax/laretz/internal/app/usecase"
	"github.com/nikimonax/laretz/internal/domain"
	"github.com/nikimonax/laretz/internal/server/renderer"
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
		if id, err := strconv.ParseInt(raw, 10, 64); err == nil && id > 0 {
			relinkID = domain.AssetID(id)
		}
	}

	responseRenderer := renderer.AssetSearch(
		handler.search,
		handler.assetGet,
		handler.externalAssetsCount,
		query,
		relinkID,
	)

	if err := responseRenderer.Render(r.Context(), w); err != nil {
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
