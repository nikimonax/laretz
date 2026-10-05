package handler

import (
	"net/http"
	"strings"

	"github.com/go-logr/logr"

	"github.com/nikimonax/laretz/internal/app/usecase"
	"github.com/nikimonax/laretz/internal/server/renderer"
	"github.com/nikimonax/laretz/internal/server/views/pages"
)

type CreatePortfolioHandler struct {
	logger logr.Logger

	responseRenderer *renderer.ResponseRenderer[pages.HomeData]
	portfolioCreate  *usecase.CreatePortfolioUseCase
}

// ServeHTTP implements [http.Handler].
func (handler *CreatePortfolioHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		mutate := renderer.HomeWithError("Invalid form", pages.DialogCreatePortfolio)
		if err := handler.responseRenderer.Render(r.Context(), w, mutate); err != nil {
			handler.logger.Error(err, "render")
		}
		return
	}

	portfolioName := strings.TrimSpace(r.FormValue("name"))

	if portfolioName == "" {
		mutate := renderer.HomeWithError("Name must not be empty.", pages.DialogCreatePortfolio)
		if err := handler.responseRenderer.Render(r.Context(), w, mutate); err != nil {
			handler.logger.Error(err, "render")
		}
		return
	}

	cmd := usecase.CreatePorfolio{Name: portfolioName}

	if _, err := handler.portfolioCreate.Execute(r.Context(), cmd); err != nil {
		mutate := renderer.HomeWithError("Something went wrong.", pages.DialogCreatePortfolio)
		if err := handler.responseRenderer.Render(r.Context(), w, mutate); err != nil {
			handler.logger.Error(err, "render")
		}
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func NewCreatePortfolioHandler(
	logger logr.Logger,
	responseRenderer *renderer.ResponseRenderer[pages.HomeData],
	portfolioCreate *usecase.CreatePortfolioUseCase,
) http.Handler {
	return &CreatePortfolioHandler{
		logger:           logger,
		responseRenderer: responseRenderer,
		portfolioCreate:  portfolioCreate,
	}
}
