package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-logr/logr"

	"github.com/nikimonax/laretz/internal/app"
	"github.com/nikimonax/laretz/internal/app/usecase"
	"github.com/nikimonax/laretz/internal/server/renderer"
	"github.com/nikimonax/laretz/internal/server/views/pages"
)

type UpdatePortfolioHandler struct {
	logger logr.Logger

	responseRenderer *renderer.ResponseRenderer[pages.HomeData]
	portfolioUpdate  *usecase.UpdatePortfolioUseCase
}

// ServeHTTP implements [http.Handler].
func (handler *UpdatePortfolioHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	portfolioID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)

	if err != nil {
		mutate := renderer.HomeWithError("Invalid path parameter.")
		if err := handler.responseRenderer.Render(r.Context(), w, mutate); err != nil {
			handler.logger.Error(err, "render")
		}
		return
	}

	dialog := fmt.Sprintf(pages.DialogRenamePortfolioFormat, portfolioID)

	if err := r.ParseForm(); err != nil {
		mutate := renderer.HomeWithError("Invalid form", dialog)
		if err := handler.responseRenderer.Render(r.Context(), w, mutate); err != nil {
			handler.logger.Error(err, "render")
		}
		return
	}

	portfolioName := strings.TrimSpace(r.FormValue("name"))

	if portfolioName == "" {
		mutate := renderer.HomeWithError("Name must not be empty.", dialog)
		if err := handler.responseRenderer.Render(r.Context(), w, mutate); err != nil {
			handler.logger.Error(err, "render")
		}
		return
	}

	cmd := usecase.UpdatePortfolio{
		ID:   portfolioID,
		Name: portfolioName,
	}

	if err := handler.portfolioUpdate.Execute(r.Context(), cmd); err != nil {
		message := messageSomethingWentWrong

		if errors.Is(err, app.ErrEntityNotFound) {
			message = messagePortfolioNotFound
		}

		mutate := renderer.HomeWithError(message, dialog)
		if err := handler.responseRenderer.Render(r.Context(), w, mutate); err != nil {
			handler.logger.Error(err, "render")
		}
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func NewUpdatePortfolioHandler(
	logger logr.Logger,
	responseRenderer *renderer.ResponseRenderer[pages.HomeData],
	portfolioUpdate *usecase.UpdatePortfolioUseCase,
) http.Handler {
	return &UpdatePortfolioHandler{
		logger:           logger,
		responseRenderer: responseRenderer,
		portfolioUpdate:  portfolioUpdate,
	}
}
