package handler

import (
	"net/http"
	"strings"

	"github.com/go-logr/logr"

	"github.com/nikimonax/laretz/internal/app/usecase"
)

type CreatePortfolioHandler struct {
	logger logr.Logger

	portfolioCreate *usecase.CreatePortfolioUseCase
}

// ServeHTTP implements [http.Handler].
func (handler *CreatePortfolioHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	portfolioName := strings.TrimSpace(r.FormValue("name"))

	if portfolioName == "" {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	cmd := usecase.CreatePorfolio{Name: portfolioName}

	if _, err := handler.portfolioCreate.Execute(r.Context(), cmd); err != nil {
		handler.logger.Error(err, "create portfolio")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func NewCreatePortfolioHandler(
	logger logr.Logger,
	portfolioCreate *usecase.CreatePortfolioUseCase,
) http.Handler {
	return &CreatePortfolioHandler{
		logger:          logger,
		portfolioCreate: portfolioCreate,
	}
}
