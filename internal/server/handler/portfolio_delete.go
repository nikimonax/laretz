package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-logr/logr"

	"github.com/nikimonax/laretz/internal/app"
	"github.com/nikimonax/laretz/internal/app/usecase"
)

type DeletePortfolioHandler struct {
	logger logr.Logger

	portfolioDelete *usecase.DeletePortfolioUseCase
}

// ServeHTTP implements [http.Handler].
func (handler *DeletePortfolioHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	portfolioID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)

	if err != nil {
		http.NotFound(w, r)
		return
	}

	cmd := usecase.DeletePortfolio{ID: portfolioID}

	if err := handler.portfolioDelete.Execute(r.Context(), cmd); err != nil {
		if errors.Is(err, app.ErrEntityNotFound) {
			http.NotFound(w, r)
			return
		}

		handler.logger.Error(err, "delete portfolio")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func NewDeletePortfolioHandler(
	logger logr.Logger,
	portfolioDelete *usecase.DeletePortfolioUseCase,
) http.Handler {
	return &DeletePortfolioHandler{
		logger:          logger,
		portfolioDelete: portfolioDelete,
	}
}
