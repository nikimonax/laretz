package handler

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-logr/logr"

	"github.com/nikimonax/laretz/internal/app/usecase"
	"github.com/nikimonax/laretz/internal/lib/spoof"
	"github.com/nikimonax/laretz/internal/server/middleware"
	"github.com/nikimonax/laretz/internal/server/static"
)

func New(
	logger logr.Logger,
	portfolioList *usecase.ListPortfoliosUseCase,
	portfolioCreate *usecase.CreatePortfolioUseCase,
	portfolioUpdate *usecase.UpdatePortfolioUseCase,
	portfolioDelele *usecase.DeletePortfolioUseCase,
) chi.Router {
	router := chi.NewRouter()

	router.Use(
		middleware.OriginCheck,
		spoof.Middleware,
		middleware.NewLoggingMiddleware(logger),
	)

	static.Mount(router)
	router.Get("/", NewHomeHandler(logger, portfolioList).ServeHTTP)
	router.Post("/portfolio", NewCreatePortfolioHandler(logger, portfolioCreate).ServeHTTP)
	router.Post("/portfolio/{id}", NewUpdatePortfolioHandler(logger, portfolioUpdate).ServeHTTP)
	router.Delete("/portfolio/{id}", NewDeletePortfolioHandler(logger, portfolioDelele).ServeHTTP)

	return router
}
