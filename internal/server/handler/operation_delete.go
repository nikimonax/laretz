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

type DeleteOperationHandler struct {
	logger logr.Logger

	portfolioGet    *usecase.GetPortfolioUseCase
	assetsList      *usecase.ListAssetsUseCase
	operationDelete *usecase.DeleteOperationUseCase
}

// ServeHTTP implements [http.Handler].
func (handler *DeleteOperationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	portfolioID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)

	if err != nil || portfolioID <= 0 {
		renderNotFoundPage(handler.logger, w, r)
		return
	}

	operationID, err := strconv.ParseInt(chi.URLParam(r, "operationID"), 10, 64)

	if err != nil || operationID <= 0 {
		renderPortfolioPage(w, r, handler.logger, handler.portfolioGet, handler.assetsList, domain.PortfolioID(portfolioID), "Invalid path parameter.", "")
		return
	}

	cmd := usecase.DeleteOperation{ID: domain.OperationID(operationID)}

	if err := handler.operationDelete.Execute(r.Context(), cmd); err != nil {
		message := messageSomethingWentWrong

		if errors.Is(err, app.ErrEntityNotFound) {
			message = "Operation not found."
		}

		renderPortfolioPage(w, r, handler.logger, handler.portfolioGet, handler.assetsList, domain.PortfolioID(portfolioID), message, fmt.Sprintf(pages.DialogDeleteOperationFormat, operationID))
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/portfolio/%d?ops_page=%d", portfolioID, parsePage(r, "ops_page")), http.StatusSeeOther)
}

func NewDeleteOperationHandler(
	logger logr.Logger,
	portfolioGet *usecase.GetPortfolioUseCase,
	assetsList *usecase.ListAssetsUseCase,
	operationDelete *usecase.DeleteOperationUseCase,
) http.Handler {
	return &DeleteOperationHandler{
		logger:          logger,
		portfolioGet:    portfolioGet,
		assetsList:      assetsList,
		operationDelete: operationDelete,
	}
}
