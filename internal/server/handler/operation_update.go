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

type UpdateOperationHandler struct {
	logger logr.Logger

	portfolioGet    *usecase.GetPortfolioUseCase
	assetsList      *usecase.ListAssetsUseCase
	operationUpdate *usecase.UpdateOperationUseCase
}

// ServeHTTP implements [http.Handler].
func (handler *UpdateOperationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	dialog := fmt.Sprintf(pages.DialogEditOperationFormat, operationID)

	cmd, message := parseOperationForm(r, domain.PortfolioID(portfolioID))

	if message != "" {
		renderPortfolioPage(w, r, handler.logger, handler.portfolioGet, handler.assetsList, domain.PortfolioID(portfolioID), message, dialog)
		return
	}

	update := usecase.UpdateOperation{CreateOperation: cmd, ID: domain.OperationID(operationID)}

	if err := handler.operationUpdate.Execute(r.Context(), update); err != nil {
		handler.logger.Error(err, "update operation")

		message := messageSomethingWentWrong

		if errors.Is(err, app.ErrEntityNotFound) {
			message = "Operation not found."
		}

		renderPortfolioPage(w, r, handler.logger, handler.portfolioGet, handler.assetsList, domain.PortfolioID(portfolioID), message, dialog)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/portfolio/%d?ops_page=%d", portfolioID, parsePage(r, "ops_page")), http.StatusSeeOther)
}

func NewUpdateOperationHandler(
	logger logr.Logger,
	portfolioGet *usecase.GetPortfolioUseCase,
	assetsList *usecase.ListAssetsUseCase,
	operationUpdate *usecase.UpdateOperationUseCase,
) http.Handler {
	return &UpdateOperationHandler{
		logger:          logger,
		portfolioGet:    portfolioGet,
		assetsList:      assetsList,
		operationUpdate: operationUpdate,
	}
}
