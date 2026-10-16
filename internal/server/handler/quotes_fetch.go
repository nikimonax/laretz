package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-logr/logr"

	"github.com/nikimonax/laretz/internal/app"
	"github.com/nikimonax/laretz/internal/app/usecase"
	"github.com/nikimonax/laretz/internal/domain"
)

type FetchQuotesHandler struct {
	logger logr.Logger

	quotesFetch  *usecase.FetchQuotesUseCase
	portfolioGet *usecase.GetPortfolioUseCase
	assetsList   *usecase.ListAssetsUseCase
}

// ServeHTTP implements [http.Handler].
func (handler *FetchQuotesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		renderNotFoundPage(handler.logger, w, r)
		return
	}

	portfolioID, err := strconv.ParseInt(strings.TrimSpace(r.FormValue("portfolio")), 10, 64)

	if err != nil || portfolioID <= 0 {
		renderNotFoundPage(handler.logger, w, r)
		return
	}

	cmd := usecase.FetchQuotes{PortfolioID: domain.PortfolioID(portfolioID)}

	if err := handler.quotesFetch.Execute(r.Context(), cmd); err != nil {
		handler.logger.Error(err, "fetch quotes")

		message := messageQuotesFetchFailed

		if errors.Is(err, app.ErrEntityNotFound) {
			message = messagePortfolioNotFound
		}

		renderPortfolioPage(w, r, handler.logger, handler.portfolioGet, handler.assetsList, domain.PortfolioID(portfolioID), message, "")
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/portfolio/%d?ops_page=%d", portfolioID, parsePage(r, "ops_page")), http.StatusSeeOther)
}

func NewFetchQuotesHandler(
	logger logr.Logger,
	quotesFetch *usecase.FetchQuotesUseCase,
	portfolioGet *usecase.GetPortfolioUseCase,
	assetsList *usecase.ListAssetsUseCase,
) http.Handler {
	return &FetchQuotesHandler{
		logger:       logger,
		quotesFetch:  quotesFetch,
		portfolioGet: portfolioGet,
		assetsList:   assetsList,
	}
}
