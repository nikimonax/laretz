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
	"github.com/nikimonax/laretz/internal/server/views/pages"
)

type CreateOperationHandler struct {
	logger logr.Logger

	portfolioGet    *usecase.GetPortfolioUseCase
	assetsList      *usecase.ListAssetsUseCase
	operationCreate *usecase.CreateOperationUseCase
}

// ServeHTTP implements [http.Handler].
func (handler *CreateOperationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	portfolioID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)

	if err != nil || portfolioID <= 0 {
		renderNotFoundPage(handler.logger, w, r)
		return
	}

	cmd, message := parseOperationForm(r, domain.PortfolioID(portfolioID))

	if message != "" {
		renderPortfolioPage(w, r, handler.logger, handler.portfolioGet, handler.assetsList, domain.PortfolioID(portfolioID), message, pages.DialogCreateOperation)
		return
	}

	if err := handler.operationCreate.Execute(r.Context(), cmd); err != nil {
		handler.logger.Error(err, "create operation")

		message := messageSomethingWentWrong

		if errors.Is(err, app.ErrEntityNotFound) {
			message = messageAssetNotFound
		}

		renderPortfolioPage(w, r, handler.logger, handler.portfolioGet, handler.assetsList, domain.PortfolioID(portfolioID), message, pages.DialogCreateOperation)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/portfolio/%d?ops_page=%d", portfolioID, parsePage(r, "ops_page")), http.StatusSeeOther)
}

func NewCreateOperationHandler(
	logger logr.Logger,
	portfolioGet *usecase.GetPortfolioUseCase,
	assetsList *usecase.ListAssetsUseCase,
	operationCreate *usecase.CreateOperationUseCase,
) http.Handler {
	return &CreateOperationHandler{
		logger:          logger,
		portfolioGet:    portfolioGet,
		assetsList:      assetsList,
		operationCreate: operationCreate,
	}
}
