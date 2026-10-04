package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-logr/logr"

	"github.com/nikimonax/laretz/internal/app"
	"github.com/nikimonax/laretz/internal/app/usecase"
)

type UpdatePortfolioHandler struct {
	logger logr.Logger

	portfolioUpdate *usecase.UpdatePortfolioUseCase
}

// ServeHTTP implements [http.Handler].
func (handler *UpdatePortfolioHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	portfolioID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)

	if err != nil {
		http.NotFound(w, r)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	portfolioName := strings.TrimSpace(r.FormValue("name"))

	if portfolioName == "" {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	cmd := usecase.UpdatePortfolio{
		ID:   portfolioID,
		Name: portfolioName,
	}

	if err := handler.portfolioUpdate.Execute(r.Context(), cmd); err != nil {
		if errors.Is(err, app.ErrEntityNotFound) {
			http.NotFound(w, r)
			return
		}

		handler.logger.Error(err, "update portfolio")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func NewUpdatePortfolioHandler(
	logger logr.Logger,
	portfolioUpdate *usecase.UpdatePortfolioUseCase,
) http.Handler {
	return &UpdatePortfolioHandler{
		logger:          logger,
		portfolioUpdate: portfolioUpdate,
	}
}
