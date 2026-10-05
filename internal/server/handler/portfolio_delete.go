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
	"github.com/nikimonax/laretz/internal/server/renderer"
	"github.com/nikimonax/laretz/internal/server/views/pages"
)

type DeletePortfolioHandler struct {
	logger logr.Logger

	responseRenderer *renderer.ResponseRenderer[pages.HomeData]
	portfolioDelete  *usecase.DeletePortfolioUseCase
}

// ServeHTTP implements [http.Handler].
func (handler *DeletePortfolioHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	portfolioID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)

	if err != nil {
		mutate := renderer.HomeWithError("Invalid path parameter.")
		if err := handler.responseRenderer.Render(r.Context(), w, mutate); err != nil {
			handler.logger.Error(err, "render")
		}
		return
	}

	cmd := usecase.DeletePortfolio{ID: portfolioID}

	if err := handler.portfolioDelete.Execute(r.Context(), cmd); err != nil {
		message := "Something went wrong."
		dialog := fmt.Sprintf(pages.DialogDeletePortfolioFormat, portfolioID)

		if errors.Is(err, app.ErrEntityNotFound) {
			message = "Portfolio not found."
		}

		mutate := renderer.HomeWithError(message, dialog)
		if err := handler.responseRenderer.Render(r.Context(), w, mutate); err != nil {
			handler.logger.Error(err, "render")
		}
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func NewDeletePortfolioHandler(
	logger logr.Logger,
	responseRenderer *renderer.ResponseRenderer[pages.HomeData],
	portfolioDelete *usecase.DeletePortfolioUseCase,
) http.Handler {
	return &DeletePortfolioHandler{
		logger:           logger,
		responseRenderer: responseRenderer,
		portfolioDelete:  portfolioDelete,
	}
}
