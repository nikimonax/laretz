package handler

import (
	"net/http"

	"github.com/go-logr/logr"

	"github.com/nikimonax/laretz/internal/app/usecase"
	"github.com/nikimonax/laretz/internal/server/views/pages"
)

type HomeHandler struct {
	logger logr.Logger

	portfolioList *usecase.ListPortfoliosUseCase
}

// ServeHTTP implements [http.Handler].
func (handler *HomeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	portfolios, err := handler.portfolioList.Execute(r.Context())

	if err != nil {
		handler.logger.Error(err, "list portfolios")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	data := pages.HomeData{Portfolios: portfolios}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := pages.Home(data).Render(r.Context(), w); err != nil {
		handler.logger.Error(err, "render home page")
	}
}

func NewHomeHandler(
	logger logr.Logger,
	portfolioList *usecase.ListPortfoliosUseCase,
) http.Handler {
	return &HomeHandler{
		logger:        logger,
		portfolioList: portfolioList,
	}
}
