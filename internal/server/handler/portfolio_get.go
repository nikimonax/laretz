package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-logr/logr"

	"github.com/nikimonax/laretz/internal/app/usecase"
	"github.com/nikimonax/laretz/internal/domain"
	"github.com/nikimonax/laretz/internal/server/renderer"
)

type PortfolioGetHandler struct {
	logger logr.Logger

	portfolioGet *usecase.GetPortfolioUseCase
	assetsList   *usecase.ListAssetsUseCase
}

// ServeHTTP implements [http.Handler].
func (handler *PortfolioGetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	portfolioID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)

	if err != nil || portfolioID <= 0 {
		renderNotFoundPage(handler.logger, w, r)
		return
	}

	responseRenderer := renderer.Portfolio(handler.portfolioGet, handler.assetsList, domain.PortfolioID(portfolioID), parsePage(r, "ops_page"))

	if err := responseRenderer.Render(r.Context(), w); err != nil {
		handler.logger.Error(err, "render")
	}
}

func NewPortfolioGetHandler(
	logger logr.Logger,
	portfolioGet *usecase.GetPortfolioUseCase,
	assetsList *usecase.ListAssetsUseCase,
) http.Handler {
	return &PortfolioGetHandler{
		logger:       logger,
		portfolioGet: portfolioGet,
		assetsList:   assetsList,
	}
}
